package imagestudio

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	coreService "github.com/QuantumNous/new-api/service"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupImageGeneration(t *testing.T) (*gorm.DB, *model.Token) {
	t.Helper()
	db := setupStudioServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Channel{}, &model.Ability{}, &model.Model{}, &model.Vendor{}, &model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{}))
	require.NoError(t, db.Create(&model.User{Id: 501, Username: "studio-owner", Group: "default", Status: common.UserStatusEnabled}).Error)
	channel := &model.Channel{Type: constant.ChannelTypeOpenAI, Name: "image-test", Status: common.ChannelStatusEnabled, Models: "dall-e-3", Group: "default"}
	require.NoError(t, db.Create(channel).Error)
	require.NoError(t, db.Create(&model.Ability{Group: "default", Model: "dall-e-3", ChannelId: channel.Id, Enabled: true}).Error)
	token := &model.Token{UserId: 501, Key: "generationfixture", Name: "Images", Status: common.TokenStatusEnabled, ExpiredTime: -1, UnlimitedQuota: true, Group: "default"}
	require.NoError(t, db.Create(token).Error)
	model.InvalidatePricingCache()
	t.Cleanup(model.InvalidatePricingCache)
	return db, token
}

// The fixture replaces the object transport; signing, metadata and ownership use production code.
func studioStorageExecution(t *testing.T, db *gorm.DB) RelayContext {
	t.Helper()
	require.NoError(t, db.AutoMigrate(&model.StorageProfile{}, &model.StorageCredential{}, &model.StoragePolicy{}, &model.StorageObject{}))
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "studio-storage-test-encryption-key")
	profile, err := storageService.SaveProfile(0, storageService.ProfileInput{
		Name: "Studio OSS", ProviderType: model.StorageProviderAliyunOSS, Status: model.StorageProfileStatusEnabled,
		Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", Region: "cn-hangzhou", Bucket: "studio-test-bucket",
		AuthType: model.StorageCredentialAuthStatic, AccessKeyID: "test-id", AccessKeySecret: "test-secret",
	})
	require.NoError(t, err)
	policy := model.DefaultAssetLibraryStoragePolicy()
	policy.StorageProfileID, policy.Enabled = profile.ID, true
	require.NoError(t, model.SaveStoragePolicy(policy))
	return RelayContext{Context: context.Background(), uploadImage: func(ctx context.Context, owner int, run, result, contentType string, size int64, body io.Reader) (*model.StorageObject, error) {
		payload, err := io.ReadAll(body)
		require.NoError(t, err)
		assert.EqualValues(t, len(payload), size)
		assert.Equal(t, "image/png", contentType)
		object := &model.StorageObject{ObjectID: result, OwnerUserID: owner, TaskID: run,
			Purpose: model.StorageObjectPurposeImageStudio, StorageProfileID: profile.ID, StoragePolicyID: policy.ID,
			ObjectKey: "assets/image-studio/" + run + "/" + result + ".png", ContentType: contentType, Size: size, Status: model.StorageObjectStatusUploaded}
		require.NoError(t, db.Create(object).Error)
		return object, nil
	}}
}

func imageResponse(t *testing.T) []byte {
	t.Helper()
	var imageData bytes.Buffer
	require.NoError(t, png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	response, err := common.Marshal(dto.ImageResponse{Data: []dto.ImageData{{B64Json: base64.StdEncoding.EncodeToString(imageData.Bytes())}}})
	require.NoError(t, err)
	return response
}

func TestStudioGenerationStorageAndHistory(t *testing.T) {
	db, token := setupImageGeneration(t)
	execution := studioStorageExecution(t, db)
	input := RunInput{TokenID: token.Id, IdempotencyKey: "generate", Request: dto.ImageRequest{Model: "dall-e-3", Prompt: "mountain", Quality: "hd", Extra: map[string]json.RawMessage{"parameters": json.RawMessage(`{"n":999}`)}}}
	calls := 0
	var relay RelayExecutor
	relay = func(ctx RelayContext, selected *model.Token, path string, request []byte, format types.RelayFormat) RelayResult {
		calls++
		assert.Equal(t, token.Id, selected.Id)
		assert.Equal(t, "/v1/images/generations", path)
		assert.JSONEq(t, `{"model":"dall-e-3","prompt":"mountain"}`, string(request))
		inflight, repeated, err := CreateRun(ctx, 501, input, relay)
		require.NoError(t, err)
		assert.True(t, repeated)
		assert.Equal(t, model.ImageStudioRunStatusRunning, inflight.Status)
		return RelayResult{StatusCode: 200, Body: imageResponse(t), RequestID: "request"}
	}
	view, _, err := CreateRun(execution, 501, input, relay)
	require.NoError(t, err)
	require.Len(t, view.Results, 1)
	assert.Contains(t, view.Results[0].PreviewURL, ".aliyuncs.com/")
	assert.Contains(t, view.Results[0].ThumbnailURL, "x-oss-process")
	var assets int64
	require.NoError(t, db.Model(&model.MediaAsset{}).Count(&assets).Error)
	assert.Zero(t, assets)
	replay, repeated, err := CreateRun(execution, 501, input, relay)
	require.NoError(t, err)
	assert.True(t, repeated)
	assert.Equal(t, view.ID, replay.ID)
	assert.Equal(t, 1, calls)
	list, _, err := ListRuns(context.Background(), Actor{UserID: 501}, false, 502, &common.PageInfo{Page: 1, PageSize: 12})
	require.NoError(t, err)
	require.Len(t, list.Items, 1)
	assert.Empty(t, list.Items[0].Results[0].PreviewURL)
	_, _, err = GetRun(context.Background(), Actor{UserID: 502}, view.ID, false)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	result := view.Results[0]
	require.Error(t, DeleteResult(502, result.ID))
	var stored model.ImageStudioResult
	require.NoError(t, db.First(&stored).Error)
	var object model.StorageObject
	require.NoError(t, db.First(&object, stored.StorageObjectID).Error)
	assert.Zero(t, object.DeleteAfter)
	libraryObject := model.StorageObject{ObjectID: "library-copy", OwnerUserID: 501, Purpose: model.StorageObjectPurposeAssetLibrary, Status: model.StorageObjectStatusUploaded}
	require.NoError(t, db.Create(&libraryObject).Error)
	asset := model.MediaAsset{PublicID: "saved-copy", OwnerUserID: 501, StorageObjectID: libraryObject.ID, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	require.NoError(t, db.Model(&stored).Update("media_asset_id", asset.ID).Error)
	// Library deletion clears the link while the independent generation stays readable.
	require.NoError(t, db.Model(&asset).Update("status", model.AssetStatusDeleted).Error)
	detail, _, err := GetRun(context.Background(), Actor{UserID: 501}, view.ID, false)
	require.NoError(t, err)
	assert.Empty(t, detail.Results[0].AssetID)
	assert.NotEmpty(t, detail.Results[0].PreviewURL)
	require.NoError(t, db.Model(&asset).Update("status", model.AssetStatusReady).Error)
	require.NoError(t, db.Model(&stored).Updates(map[string]any{"save_status": "saving", "save_started_at": common.GetTimestamp()}).Error)
	require.Error(t, DeleteResult(501, result.ID))
	require.NoError(t, db.Model(&stored).Update("save_status", "saved").Error)
	require.NoError(t, DeleteResult(501, result.ID))
	require.NoError(t, db.First(&libraryObject, libraryObject.ID).Error)
	assert.Equal(t, model.StorageObjectStatusUploaded, libraryObject.Status)
	require.NoError(t, db.First(&asset, asset.ID).Error)
	assert.Equal(t, model.AssetStatusReady, asset.Status)
	require.NoError(t, db.First(&object, object.ID).Error)
	assert.Equal(t, model.StorageObjectStatusDeletePending, object.Status)
	_, repeated, err = CreateRun(execution, 501, input, relay)
	require.Error(t, err)
	assert.True(t, repeated)
	assert.Equal(t, 1, calls)
}

func TestStudioFailuresCleanOutputsAndKeepAcceptedRetryIdentity(t *testing.T) {
	for _, failure := range []string{"missing storage", "provider", "empty", "invalid base64", "corrupt image", "transport", "record", "finalize"} {
		t.Run(failure, func(t *testing.T) {
			db, token := setupImageGeneration(t)
			execution := studioStorageExecution(t, db)
			response := imageResponse(t)
			status, requestID := 200, "accepted"
			switch failure {
			case "missing storage":
				require.NoError(t, db.Model(&model.StoragePolicy{}).Where("1 = 1").Update("enabled", false).Error)
			case "provider":
				status, requestID, response = 502, "", []byte(`{"error":{"message":"provider unavailable"}}`)
			case "empty":
				response = []byte(`{"data":[]}`)
			case "invalid base64":
				response = []byte(`{"data":[{"b64_json":"invalid!"}]}`)
			case "corrupt image":
				response = []byte(`{"data":[{"b64_json":"iVBORw0KGgo="}]}`)
			case "transport":
				execution.uploadImage = func(context.Context, int, string, string, string, int64, io.Reader) (*model.StorageObject, error) {
					return nil, errors.New("OSS unavailable")
				}
			case "record":
				require.NoError(t, db.Callback().Create().Before("gorm:create").Register("fail_result", func(tx *gorm.DB) {
					if tx.Statement.Table == "image_studio_results" {
						tx.AddError(errors.New("record failed"))
					}
				}))
			case "finalize":
				require.NoError(t, db.Callback().Update().Before("gorm:update").Register("fail_finalize", func(tx *gorm.DB) {
					if run, ok := tx.Statement.Dest.(*model.ImageStudioRun); ok && run.Status == model.ImageStudioRunStatusCompleted {
						tx.AddError(errors.New("finalize failed"))
					}
				}))
			}
			calls := 0
			relay := func(RelayContext, *model.Token, string, []byte, types.RelayFormat) RelayResult {
				calls++
				return RelayResult{StatusCode: status, Body: response, RequestID: requestID}
			}
			input := RunInput{TokenID: token.Id, IdempotencyKey: "failure", Request: dto.ImageRequest{Model: "dall-e-3", Prompt: "mountain"}}
			_, _, err := CreateRun(execution, 501, input, relay)
			require.Error(t, err)
			if failure == "missing storage" {
				assert.Zero(t, calls)
			} else if failure != "provider" {
				_, repeated, err := CreateRun(execution, 501, input, relay)
				require.Error(t, err)
				assert.True(t, repeated)
				assert.Equal(t, 1, calls)
			}
			var count int64
			require.NoError(t, db.Model(&model.ImageStudioResult{}).Count(&count).Error)
			assert.Zero(t, count)
			list, _, err := ListRuns(context.Background(), Actor{UserID: 501}, false, 0, &common.PageInfo{Page: 1, PageSize: 12})
			require.NoError(t, err)
			assert.Empty(t, list.Items)
			if failure == "record" || failure == "finalize" {
				var object model.StorageObject
				require.NoError(t, db.First(&object).Error)
				assert.Equal(t, model.StorageObjectStatusDeletePending, object.Status)
			}
		})
	}
}

func TestStudioValidationPreservesDefaultsAndRejectsOverrides(t *testing.T) {
	settings := config.GlobalConfig.Get("image_studio_setting").(*operation_setting.ImageStudioSetting)
	previous := *settings
	settings.Enabled = true
	t.Cleanup(func() { *settings = previous })
	for _, body := range []string{
		`{"model":"seedream","prompt":"mountain"}`,
		`{"model":"seedream","prompt":"mountain","size":"2048x2048","n":2,"response_format":"url"}`,
	} {
		var request dto.ImageRequest
		require.NoError(t, common.Unmarshal([]byte(body), &request))
		require.NoError(t, validateStudioImageRequest(&request, nil, 0))
		actual, err := common.Marshal(request)
		require.NoError(t, err)
		assert.JSONEq(t, body, string(actual))
	}
	for _, body := range []string{`{"n":0}`, `{"n":129}`, `{"n":18446744073709551615}`, `{"image":"inline"}`, `{"stream":true}`} {
		request := dto.ImageRequest{Model: "seedream", Prompt: "mountain"}
		require.NoError(t, common.Unmarshal([]byte(body), &request))
		require.Error(t, validateStudioImageRequest(&request, nil, 0))
	}
	request := dto.ImageRequest{Model: "seedream", Prompt: "mountain"}
	require.Error(t, validateStudioImageRequest(&request, json.RawMessage(`{"parameters":{"n":999}}`), 0))
	settings.Enabled = false
	require.Error(t, validateStudioImageRequest(&request, nil, 0))
}

func TestStudioSaveOwnershipAndBusyResults(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.AssetGroup{}, &model.MediaAsset{}, &model.StorageObject{}))
	settings := config.GlobalConfig.Get("image_studio_setting").(*operation_setting.ImageStudioSetting)
	previous := *settings
	settings.AssetLibraryEnabled = true
	t.Cleanup(func() { *settings = previous })
	group := model.AssetGroup{PublicID: "group", OwnerUserID: 7, Status: model.AssetStatusReady}
	run := model.ImageStudioRun{PublicID: "run", UserID: 7, IdempotencyKey: "save", Status: model.ImageStudioRunStatusCompleted}
	asset := model.MediaAsset{PublicID: "asset", OwnerUserID: 7, Status: model.AssetStatusReady}
	for _, row := range []any{&group, &run, &asset} {
		require.NoError(t, db.Create(row).Error)
	}
	result := model.ImageStudioResult{RunID: run.ID, MediaAssetID: asset.ID, SaveStatus: "saved"}
	require.NoError(t, db.Create(&result).Error)
	require.NoError(t, SaveResult(context.Background(), 7, result.PublicID, group.PublicID))
	require.Error(t, SaveResult(context.Background(), 8, result.PublicID, group.PublicID))
	require.Error(t, SaveResult(context.Background(), 7, result.PublicID, "foreign"))
	require.Error(t, SaveResult(context.Background(), 7, result.PublicID, ""))
	settings.AssetLibraryEnabled = false
	assert.Equal(t, "asset_library_disabled", ErrorCode(SaveResult(context.Background(), 7, result.PublicID, group.PublicID)))
	settings.AssetLibraryEnabled = true
	for _, status := range []string{model.ImageStudioRunStatusRunning, model.ImageStudioRunStatusFailed, model.ImageStudioRunStatusCanceled} {
		require.NoError(t, db.Model(&run).Update("status", status).Error)
		require.Error(t, SaveResult(context.Background(), 7, result.PublicID, group.PublicID))
	}
	require.NoError(t, db.Model(&run).Update("status", model.ImageStudioRunStatusCompleted).Error)
	require.NoError(t, db.Model(&asset).Update("status", model.AssetStatusDeleted).Error)
	// An unavailable original fails import, leaves the result and permits another attempt.
	for range 2 {
		require.Error(t, SaveResult(context.Background(), 7, result.PublicID, group.PublicID))
		require.NoError(t, db.First(&result, result.ID).Error)
		assert.Equal(t, "failed", result.SaveStatus)
		assert.Zero(t, result.MediaAssetID)
		assert.NotEmpty(t, result.SaveError)
	}
	for _, attempt := range []struct {
		at      int64
		claimed bool
	}{{1000, true}, {1001, false}, {1361, true}} {
		claimed, err := model.ClaimImageStudioResultSave(result.ID, group.PublicID, common.GetUUID(), attempt.at)
		require.NoError(t, err)
		assert.Equal(t, attempt.claimed, claimed)
	}
}

func TestStudioReferenceImagesRejectForeignDeletedAndVideoAssets(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.MediaAsset{}))
	settings := config.GlobalConfig.Get("image_studio_setting").(*operation_setting.ImageStudioSetting)
	previous := *settings
	t.Cleanup(func() { *settings = previous })
	settings.AssetLibraryEnabled = false
	_, err := loadInputAssets(7, []string{"owned"})
	assert.Equal(t, "asset_library_disabled", ErrorCode(err))
	settings.AssetLibraryEnabled = true
	for _, asset := range []model.MediaAsset{
		{PublicID: "owned", StorageObjectID: 1, OwnerUserID: 7, AssetType: model.AssetTypeImage, Status: model.AssetStatusReady},
		{PublicID: "foreign", StorageObjectID: 2, OwnerUserID: 8, AssetType: model.AssetTypeImage, Status: model.AssetStatusReady},
		{PublicID: "deleted", StorageObjectID: 3, OwnerUserID: 7, AssetType: model.AssetTypeImage, Status: model.AssetStatusDeleted},
		{PublicID: "video", StorageObjectID: 4, OwnerUserID: 7, AssetType: model.AssetTypeVideo, Status: model.AssetStatusReady},
	} {
		require.NoError(t, db.Create(&asset).Error)
	}
	assets, err := loadInputAssets(7, []string{"owned"})
	require.NoError(t, err)
	require.Len(t, assets, 1)
	for _, id := range []string{"foreign", "deleted", "video"} {
		_, err := loadInputAssets(7, []string{id})
		require.Error(t, err)
	}
}
func TestPartialURLGenerationRetainsFailureReasonAndOnlyValidImages(t *testing.T) {
	db, token := setupImageGeneration(t)
	fetch := system_setting.GetFetchSetting()
	previous := *fetch
	fetch.EnableSSRFProtection = false
	coreService.InitHttpClient()
	t.Cleanup(func() { *fetch = previous; coreService.InitHttpClient() })
	var imageData bytes.Buffer
	require.NoError(t, png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/failed" {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write(imageData.Bytes())
	}))
	t.Cleanup(server.Close)
	body, err := common.Marshal(dto.ImageResponse{Data: []dto.ImageData{{Url: server.URL + "/failed"}, {Url: server.URL + "/image"}}})
	require.NoError(t, err)
	view, _, err := CreateRun(studioStorageExecution(t, db), 501, RunInput{
		TokenID: token.Id, IdempotencyKey: "partial", Request: dto.ImageRequest{Model: "dall-e-3", Prompt: "A mountain"},
	}, func(RelayContext, *model.Token, string, []byte, types.RelayFormat) RelayResult {
		return RelayResult{StatusCode: http.StatusOK, Body: body, RequestID: "partial-request"}
	})
	require.NoError(t, err)
	assert.Equal(t, model.ImageStudioRunStatusPartial, view.Status)
	assert.Equal(t, "generated image download is unavailable or oversized", view.ErrorMessage)
	require.Len(t, view.Results, 1)
	assert.Equal(t, 1, view.Results[0].Index)
}
