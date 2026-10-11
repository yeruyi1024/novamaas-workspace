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
	"net/url"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	coreService "github.com/QuantumNous/new-api/service"
	assetService "github.com/QuantumNous/new-api/service/assetlibrary"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestStudioRequestRejectsDisabledGenerationAndUnsupportedPayloads(t *testing.T) {
	settings := config.GlobalConfig.Get("image_studio_setting").(*operation_setting.ImageStudioSetting)
	previous := *settings
	t.Cleanup(func() { *settings = previous })
	for _, tc := range []struct {
		name, code   string
		disabled     bool
		image, extra json.RawMessage
	}{
		{name: "root disabled", disabled: true, code: "image_studio_disabled"},
		{name: "provider count override", extra: json.RawMessage(`{"sequential_image_generation_options":{"max_images":999}}`), code: "unsupported_parameters"},
		{name: "inline image", image: json.RawMessage(`"data:image/png;base64,secret"`), code: "asset_ids_required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			settings.Enabled = !tc.disabled
			request := dto.ImageRequest{Model: "seedream", Prompt: "mountain", Image: tc.image}
			err := validateStudioImageRequest(&request, tc.extra, 0)
			require.Error(t, err)
			assert.Equal(t, tc.code, ErrorCode(err))
		})
	}
}

func TestSaveResultRequiresOwnedImageAndExplicitOwnedGroup(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	settings := config.GlobalConfig.Get("image_studio_setting").(*operation_setting.ImageStudioSetting)
	previous := *settings
	t.Cleanup(func() { *settings = previous })
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.AssetGroup{}, &model.MediaAsset{}))
	group := &model.AssetGroup{PublicID: "owned-group", OwnerUserID: 7, Status: model.AssetStatusReady}
	foreign := &model.AssetGroup{PublicID: "foreign-group", OwnerUserID: 8, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(group).Error)
	require.NoError(t, db.Create(foreign).Error)
	run := &model.ImageStudioRun{PublicID: "owned-run", UserID: 7, IdempotencyKey: "owned", Status: model.ImageStudioRunStatusCompleted}
	require.NoError(t, db.Create(run).Error)
	asset := &model.MediaAsset{PublicID: "saved-image", OwnerUserID: 7, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(asset).Error)
	result := &model.ImageStudioResult{RunID: run.ID, MediaAssetID: asset.ID, SaveStatus: "saved"}
	require.NoError(t, db.Create(result).Error)
	for _, tc := range []struct {
		name, group, code string
		user              int
		disabled          bool
	}{
		{name: "owner retry", user: 7, group: group.PublicID},
		{name: "foreign image", user: 8, group: foreign.PublicID, code: "result_not_found"},
		{name: "foreign group", user: 7, group: foreign.PublicID, code: "output_group_unavailable"},
		{name: "no group", user: 7, code: "invalid_group"},
		{name: "root disabled", user: 7, group: group.PublicID, disabled: true, code: "asset_library_disabled"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			settings.AssetLibraryEnabled = !tc.disabled
			err := SaveResult(context.Background(), tc.user, result.PublicID, tc.group)
			if tc.code == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tc.code, ErrorCode(err))
			}
		})
	}
	var groups, assets int64
	require.NoError(t, db.Model(&model.AssetGroup{}).Count(&groups).Error)
	require.NoError(t, db.Model(&model.MediaAsset{}).Count(&assets).Error)
	assert.EqualValues(t, 2, groups)
	assert.EqualValues(t, 1, assets)
}

func TestSaveResultRejectsUnfinishedOrUnavailableGeneration(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	settings := config.GlobalConfig.Get("image_studio_setting").(*operation_setting.ImageStudioSetting)
	previous := *settings
	settings.AssetLibraryEnabled = true
	t.Cleanup(func() { *settings = previous })
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.AssetGroup{}, &model.MediaAsset{}))
	group := &model.AssetGroup{PublicID: "owned-group", OwnerUserID: 7, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(group).Error)
	for _, status := range []string{model.ImageStudioRunStatusRunning, model.ImageStudioRunStatusFailed, model.ImageStudioRunStatusCanceled} {
		t.Run(status, func(t *testing.T) {
			run := &model.ImageStudioRun{PublicID: status, UserID: 7, IdempotencyKey: status, Status: status}
			require.NoError(t, db.Create(run).Error)
			result := &model.ImageStudioResult{RunID: run.ID, SaveStatus: "not_requested", PreviewURL: "data:image/png;base64,invalid!"}
			require.NoError(t, db.Create(result).Error)
			if status == model.ImageStudioRunStatusRunning {
				view, _, err := GetRun(context.Background(), Actor{UserID: 7}, run.PublicID, false)
				require.NoError(t, err)
				assert.Empty(t, view.Results, "uncommitted generation results must not be exposed")
			}
			err := SaveResult(context.Background(), 7, result.PublicID, group.PublicID)
			require.Error(t, err)
			assert.Equal(t, http.StatusConflict, StatusCode(err))
			require.NoError(t, db.First(result, result.ID).Error)
			assert.Equal(t, "not_requested", result.SaveStatus)
			assert.Empty(t, result.SaveAttemptID)
			var assets int64
			require.NoError(t, db.Model(&model.MediaAsset{}).Count(&assets).Error)
			assert.Zero(t, assets)
		})
	}
}

func TestImageStudioUnsavedResultRemainsVisibleWithoutAsset(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{}))
	run := &model.ImageStudioRun{PublicID: "run-unsaved", UserID: 7, IdempotencyKey: "unsaved", Status: model.ImageStudioRunStatusCompleted}
	require.NoError(t, db.Create(run).Error)
	image := &model.ImageStudioResult{RunID: run.ID, PreviewURL: "data:image/png;base64,aW1hZ2U="}
	require.NoError(t, db.Create(image).Error)
	view, err := buildImageStudioRunView(context.Background(), run, 7, true)
	require.NoError(t, err)
	require.Len(t, view.Results, 1)
	assert.Empty(t, view.Results[0].AssetID)
	assert.Equal(t, string(image.PreviewURL), view.Results[0].PreviewURL)
	var count int64
	require.NoError(t, db.Model(&model.MediaAsset{}).Count(&count).Error)
	assert.Zero(t, count)
	hidden, err := buildImageStudioRunView(context.Background(), run, 8, false)
	require.NoError(t, err)
	assert.Empty(t, hidden.Results)
}

func TestDeletingLibraryAssetKeepsGeneratedImageReadable(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{}, &model.AssetReplica{}, &model.StorageObject{}))
	object := &model.StorageObject{ObjectID: "stored-image", Purpose: model.StorageObjectPurposeAssetLibrary, Status: model.StorageObjectStatusUploaded}
	require.NoError(t, db.Create(object).Error)
	asset := &model.MediaAsset{PublicID: "saved-image", OwnerUserID: 7, AssetType: model.AssetTypeImage, Status: model.AssetStatusReady, StorageObjectID: object.ID}
	require.NoError(t, db.Create(asset).Error)
	run := &model.ImageStudioRun{PublicID: "run-saved", UserID: 7, IdempotencyKey: "saved", Status: model.ImageStudioRunStatusCompleted, Quota: 100}
	require.NoError(t, db.Create(run).Error)
	preview := model.ImageStudioPreview("data:image/png;base64,aW1hZ2U=")
	result := &model.ImageStudioResult{RunID: run.ID, PreviewURL: preview, MediaAssetID: asset.ID, SaveStatus: "saved"}
	require.NoError(t, db.Create(result).Error)

	require.NoError(t, assetService.DeleteAsset(asset.PublicID, 7, false))
	view, _, err := GetRun(context.Background(), Actor{UserID: 7, Role: common.RoleCommonUser}, run.PublicID, false)
	require.NoError(t, err)
	require.Len(t, view.Results, 1)
	assert.Equal(t, string(preview), view.Results[0].PreviewURL)
	assert.Empty(t, view.Results[0].AssetID)
	assert.Equal(t, "not_requested", view.Results[0].SaveStatus)
	assert.Equal(t, 100, view.Quota)
	require.NoError(t, db.First(asset, asset.ID).Error)
	assert.Equal(t, model.AssetStatusDeleted, asset.Status)
	require.NoError(t, db.First(object, object.ID).Error)
	assert.Equal(t, model.StorageObjectStatusDeletePending, object.Status)

	group := &model.AssetGroup{PublicID: "restore-group", OwnerUserID: 7, Name: "Restored", Status: model.AssetStatusReady}
	require.NoError(t, db.AutoMigrate(group))
	require.NoError(t, db.Create(group).Error)
	// The retained fixture is deliberately invalid image data. Reimport must
	// attempt to save it and report that error instead of claiming success.
	require.Error(t, SaveResult(context.Background(), 7, result.PublicID, group.PublicID))
	require.NoError(t, db.First(result, result.ID).Error)
	assert.Zero(t, result.MediaAssetID)
	assert.Equal(t, "failed", result.SaveStatus)
}

func TestHistoryReturnsThumbnailsWithoutOriginalImagesAndKeepsOwnerIsolation(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{}))
	var original bytes.Buffer
	require.NoError(t, png.Encode(&original, image.NewRGBA(image.Rect(0, 0, 1024, 512))))
	preview := "data:image/png;base64," + base64.StdEncoding.EncodeToString(original.Bytes())
	var small bytes.Buffer
	require.NoError(t, png.Encode(&small, image.NewRGBA(image.Rect(0, 0, 320, 160))))
	storedThumbnail := "data:image/png;base64," + base64.StdEncoding.EncodeToString(small.Bytes())
	for _, owner := range []int{7, 8} {
		run := &model.ImageStudioRun{UserID: owner, PublicID: common.GetUUID(), IdempotencyKey: "thumbnail", Model: "dall-e-3", Status: model.ImageStudioRunStatusCompleted}
		require.NoError(t, db.Create(run).Error)
		require.NoError(t, db.Create(&model.ImageStudioResult{RunID: run.ID, PreviewURL: model.ImageStudioPreview(preview), ThumbnailURL: model.ImageStudioPreview(storedThumbnail)}).Error)
	}
	list, _, err := ListRuns(context.Background(), Actor{UserID: 7}, false, 8, &common.PageInfo{Page: 1, PageSize: 12})
	require.NoError(t, err)
	require.Len(t, list.Items, 1)
	assert.Equal(t, 7, list.Items[0].UserID)
	require.Len(t, list.Items[0].Results, 1)
	encoded, err := common.Marshal(list)
	require.NoError(t, err)
	assert.False(t, strings.Contains(string(encoded), preview), "history must not transfer original image bytes")
	thumbnail := list.Items[0].Results[0].ThumbnailURL
	require.True(t, strings.HasPrefix(thumbnail, "data:image/png;base64,"))
	_, payload, _ := strings.Cut(thumbnail, ",")
	data, err := base64.StdEncoding.DecodeString(payload)
	require.NoError(t, err)
	size, _, err := image.DecodeConfig(bytes.NewReader(data))
	require.NoError(t, err)
	assert.LessOrEqual(t, size.Width, 320)
	assert.Equal(t, size.Width/2, size.Height)
	// The detail endpoint continues to expose the owned original image.
	detail, _, err := GetRun(context.Background(), Actor{UserID: 7}, list.Items[0].ID, false)
	require.NoError(t, err)
	assert.Equal(t, preview, detail.Results[0].PreviewURL)
}

func TestImageStudioSaveFailureRetainsResultForRetry(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioResult{}, &model.AssetGroup{}, &model.MediaAsset{}))
	group := &model.AssetGroup{PublicID: "group-save", OwnerUserID: 7, Name: "Saved", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(group).Error)
	result := &model.ImageStudioResult{PreviewURL: "data:image/png;base64,invalid!", SaveStatus: "not_requested"}
	require.NoError(t, db.Create(result).Error)
	for range 2 {
		err := saveImageStudioResult(context.Background(), result, 7, group.PublicID, "seedream")
		require.Error(t, err)
		var persisted model.ImageStudioResult
		require.NoError(t, db.First(&persisted, result.ID).Error)
		assert.Equal(t, "failed", persisted.SaveStatus)
		assert.Equal(t, result.PreviewURL, persisted.PreviewURL)
		assert.Equal(t, group.PublicID, persisted.SaveGroupID)
		assert.Zero(t, persisted.MediaAssetID)
	}
	var count int64
	require.NoError(t, db.Model(&model.MediaAsset{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestImageStudioSaveClaimIsIdempotentAndRecoversExpiredAttempt(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioResult{}, &model.MediaAsset{}))
	result := &model.ImageStudioResult{SaveStatus: "not_requested"}
	require.NoError(t, db.Create(result).Error)
	claimed, err := model.ClaimImageStudioResultSave(result.ID, "group-a", "attempt-a", 1000)
	require.NoError(t, err)
	assert.True(t, claimed)
	claimed, err = model.ClaimImageStudioResultSave(result.ID, "group-a", "attempt-b", 1001)
	require.NoError(t, err)
	assert.False(t, claimed)
	claimed, err = model.ClaimImageStudioResultSave(result.ID, "group-a", "attempt-c", 1361)
	require.NoError(t, err)
	assert.True(t, claimed)
	require.NoError(t, db.Model(result).Update("media_asset_id", 99).Error)
	claimed, err = model.ClaimImageStudioResultSave(result.ID, "group-b", "attempt-d", 2000)
	require.NoError(t, err)
	assert.False(t, claimed)
	result.MediaAssetID = 99
	require.NoError(t, db.Create(&model.MediaAsset{ID: 99, PublicID: "existing-import", OwnerUserID: 7, Status: model.AssetStatusReady}).Error)
	require.NoError(t, saveImageStudioResult(context.Background(), result, 7, "group-b", "seedream"))
}

func TestImageStudioAssetSwitchBlocksReferencesAndSavingOnly(t *testing.T) {
	settings := config.GlobalConfig.Get("image_studio_setting").(*operation_setting.ImageStudioSetting)
	previous := *settings
	t.Cleanup(func() { *settings = previous })
	settings.AssetLibraryEnabled = false
	err := SaveResult(context.Background(), 7, "result-a", "group-a")
	require.Error(t, err)
	assert.Equal(t, "asset_library_disabled", ErrorCode(err))
	_, err = resolveStudioAssetURLs(context.Background(), 7, []string{"asset-a"})
	require.Error(t, err)
	assert.Equal(t, "asset_library_disabled", ErrorCode(err))
	err = SaveResult(context.Background(), 7, "result-a", "")
	require.Error(t, err)
	assert.Equal(t, "invalid_group", ErrorCode(err))
	require.NoError(t, validateStudioImageRequest(&dto.ImageRequest{Model: "seedream", Prompt: "mountain"}, nil, 0))
}

func TestReferenceImagesRequireReadyOwnedImages(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	settings := config.GlobalConfig.Get("image_studio_setting").(*operation_setting.ImageStudioSetting)
	previous := *settings
	settings.AssetLibraryEnabled = true
	t.Cleanup(func() { *settings = previous })
	require.NoError(t, db.AutoMigrate(&model.MediaAsset{}))
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
	assert.Equal(t, "owned", assets[0].PublicID)
	for _, rejected := range []string{"foreign", "deleted", "video"} {
		_, err := loadInputAssets(7, []string{"owned", rejected})
		require.Error(t, err)
		assert.Equal(t, "input_asset_unavailable", ErrorCode(err))
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

func TestImageStudioRunViewOnlyLinksOwnerAssets(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{}))
	asset := &model.MediaAsset{PublicID: "foreign-asset", OwnerUserID: 8, AssetType: model.AssetTypeImage, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(asset).Error)
	run := &model.ImageStudioRun{PublicID: "run-bad-link", UserID: 7, IdempotencyKey: "bad-link"}
	require.NoError(t, db.Create(run).Error)
	require.NoError(t, db.Create(&model.ImageStudioResult{RunID: run.ID, MediaAssetID: asset.ID}).Error)
	view, err := buildImageStudioRunView(context.Background(), run, 7, true)
	require.NoError(t, err)
	require.Len(t, view.Results, 1)
	assert.Empty(t, view.Results[0].AssetID)
	assert.Empty(t, view.Results[0].PreviewURL)
}

func TestGenerateImagesRecordsResultsAndReplaysWithoutSecondCharge(t *testing.T) {
	db, token := setupImageGeneration(t)
	var body bytes.Buffer
	require.NoError(t, png.Encode(&body, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	response, err := common.Marshal(dto.ImageResponse{Data: []dto.ImageData{{B64Json: base64.StdEncoding.EncodeToString(body.Bytes())}}})
	require.NoError(t, err)
	input := RunInput{TokenID: token.Id, IdempotencyKey: "one-request", Request: dto.ImageRequest{Model: "dall-e-3", Prompt: "A mountain", Quality: "hd", OutputCompression: json.RawMessage(`0`), Extra: map[string]json.RawMessage{"parameters": json.RawMessage(`{"n":999}`)}}}
	execution := studioStorageExecution(t, db)
	calls := 0
	var relay RelayExecutor
	relay = func(ctx RelayContext, selected *model.Token, path string, request []byte, format types.RelayFormat) RelayResult {
		var forwarded dto.ImageRequest
		require.NoError(t, common.Unmarshal(request, &forwarded))
		assert.Equal(t, "hd", forwarded.Quality)
		assert.Empty(t, forwarded.OutputCompression)
		assert.NotContains(t, string(request), "parameters")
		require.NotNil(t, forwarded.N)
		assert.Equal(t, uint(1), *forwarded.N)
		calls++
		assert.Equal(t, token.Id, selected.Id)
		assert.Equal(t, "/v1/images/generations", path)
		assert.Equal(t, types.RelayFormat(types.RelayFormatOpenAIImage), format)
		replay, repeated, replayErr := CreateRun(ctx, 501, input, relay)
		require.NoError(t, replayErr)
		assert.True(t, repeated)
		assert.Equal(t, model.ImageStudioRunStatusRunning, replay.Status)
		return RelayResult{StatusCode: http.StatusOK, Body: response, RequestID: "billed-request"}
	}
	run, repeated, err := CreateRun(execution, 501, input, relay)
	require.NoError(t, err)
	assert.False(t, repeated)
	assert.Equal(t, model.ImageStudioRunStatusCompleted, run.Status)
	require.Len(t, run.Results, 1)
	assert.Empty(t, run.Results[0].AssetID)
	assert.Contains(t, run.Results[0].PreviewURL, ".aliyuncs.com/")
	assert.NotContains(t, run.Results[0].PreviewURL, "base64")
	thumbnail, err := url.Parse(run.Results[0].ThumbnailURL)
	require.NoError(t, err)
	assert.Contains(t, thumbnail.Query().Get("x-oss-process"), "image/resize")
	assert.NotEmpty(t, thumbnail.Query().Get("x-oss-signature"))
	var result model.ImageStudioResult
	require.NoError(t, db.First(&result).Error)
	assert.Empty(t, result.PreviewURL)
	assert.Empty(t, result.ThumbnailURL)
	assert.Positive(t, result.StorageObjectID)
	retry, repeated, err := CreateRun(execution, 501, input, relay)
	require.NoError(t, err)
	assert.True(t, repeated)
	assert.Equal(t, run.ID, retry.ID)
	assert.Equal(t, 1, calls)
	var assets int64
	require.NoError(t, db.Model(&model.MediaAsset{}).Count(&assets).Error)
	assert.Zero(t, assets)
	list, _, err := ListRuns(context.Background(), Actor{UserID: 501}, false, 0, &common.PageInfo{Page: 1, PageSize: 12})
	require.NoError(t, err)
	require.Len(t, list.Items, 1)
	assert.Empty(t, list.Items[0].Results[0].PreviewURL)
	assert.Contains(t, list.Items[0].Results[0].ThumbnailURL, "x-oss-process")
	_, _, err = GetRun(context.Background(), Actor{UserID: 502}, run.ID, false)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Error(t, DeleteResult(502, result.PublicID))
	var object model.StorageObject
	require.NoError(t, db.First(&object, result.StorageObjectID).Error)
	assert.Equal(t, model.StorageObjectStatusUploaded, object.Status)
	assert.Zero(t, object.DeleteAfter)
	// An explicitly saved library image is an independent object.
	libraryObject := &model.StorageObject{ObjectID: "library-copy", OwnerUserID: 501, Purpose: model.StorageObjectPurposeAssetLibrary, Status: model.StorageObjectStatusUploaded}
	require.NoError(t, db.Create(libraryObject).Error)
	asset := &model.MediaAsset{PublicID: "library-copy", OwnerUserID: 501, StorageObjectID: libraryObject.ID, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(asset).Error)
	require.NoError(t, db.Model(&result).Update("media_asset_id", asset.ID).Error)
	require.NoError(t, DeleteResult(501, result.PublicID))
	require.NoError(t, db.First(&object, result.StorageObjectID).Error)
	assert.Equal(t, model.StorageObjectStatusDeletePending, object.Status)
	require.NoError(t, db.First(libraryObject, libraryObject.ID).Error)
	assert.Equal(t, model.StorageObjectStatusUploaded, libraryObject.Status)
	require.NoError(t, db.First(asset, asset.ID).Error)
	assert.Equal(t, model.AssetStatusReady, asset.Status)

}

func TestImageCountBoundsRejectBillingMultipliers(t *testing.T) {
	for _, count := range []uint{0, 129, ^uint(0)} {
		request := dto.ImageRequest{Model: "dall-e-3", Prompt: "A mountain", N: &count}
		err := validateStudioImageRequest(&request, nil, 0)
		require.Error(t, err)
		assert.Equal(t, "invalid_image_count", ErrorCode(err))
	}
}

func TestFailedGenerationReturnsReasonWithoutKeepingRecords(t *testing.T) {
	for _, response := range []struct {
		name, body, reason string
		status             int
	}{
		{"upstream failure", `{"error":{"message":"provider unavailable sk-private"}}`, "provider unavailable sk-[redacted]", http.StatusBadGateway},
		{"empty response", `{"data":[]}`, "image provider returned no results", http.StatusOK},
		{"invalid image", `{"data":[{"b64_json":"not-base64!"}]}`, "generated image contains invalid base64 data", http.StatusOK},
		{"corrupt PNG", `{"data":[{"b64_json":"iVBORw0KGgo="}]}`, "image provider returned an invalid or oversized image", http.StatusOK},
	} {
		t.Run(response.name, func(t *testing.T) {
			db, token := setupImageGeneration(t)
			input := RunInput{TokenID: token.Id, IdempotencyKey: "retry-failed-request", Request: dto.ImageRequest{Model: "dall-e-3", Prompt: "A mountain"}}
			relay := func(RelayContext, *model.Token, string, []byte, types.RelayFormat) RelayResult {
				return RelayResult{StatusCode: response.status, Body: []byte(response.body)}
			}
			run, _, err := CreateRun(studioStorageExecution(t, db), 501, input, relay)
			require.Error(t, err)
			assert.Nil(t, run)
			assert.Equal(t, response.reason, err.Error())
			for _, record := range []any{&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{}} {
				var count int64
				require.NoError(t, db.Model(record).Count(&count).Error)
				assert.Zero(t, count)
			}
		})
	}
}

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

func TestMissingImageStorageRejectsBeforeRelayAndBilling(t *testing.T) {
	db, token := setupImageGeneration(t)
	calls := 0
	_, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, RunInput{
		TokenID: token.Id, IdempotencyKey: "no-storage", Request: dto.ImageRequest{Model: "dall-e-3", Prompt: "A mountain"},
	}, func(RelayContext, *model.Token, string, []byte, types.RelayFormat) RelayResult {
		calls++
		return RelayResult{}
	})
	require.Error(t, err)
	assert.Equal(t, "image_storage_unavailable", ErrorCode(err))
	assert.Zero(t, calls)
	var count int64
	require.NoError(t, db.Model(&model.ImageStudioRun{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestImageStorageFailureDoesNotExposeImagesOrRebillReplay(t *testing.T) {
	for _, failure := range []string{"transport", "record", "finalize"} {
		t.Run(failure, func(t *testing.T) {
			db, token := setupImageGeneration(t)
			execution := studioStorageExecution(t, db)
			if failure == "transport" {
				execution.uploadImage = func(context.Context, int, string, string, string, int64, io.Reader) (*model.StorageObject, error) {
					return nil, errors.New("OSS unavailable")
				}
			} else if failure == "record" {
				require.NoError(t, db.Callback().Create().Before("gorm:create").Register("fail_image_record", func(tx *gorm.DB) {
					if tx.Statement.Table == "image_studio_results" {
						tx.AddError(errors.New("record unavailable"))
					}
				}))
				t.Cleanup(func() { db.Callback().Create().Remove("fail_image_record") })
			} else {
				require.NoError(t, db.Callback().Update().Before("gorm:update").Register("fail_image_finalize", func(tx *gorm.DB) {
					if run, ok := tx.Statement.Dest.(*model.ImageStudioRun); ok && run.Status == model.ImageStudioRunStatusCompleted {
						tx.AddError(errors.New("completion unavailable"))
					}
				}))
				t.Cleanup(func() { db.Callback().Update().Remove("fail_image_finalize") })
			}
			var body bytes.Buffer
			require.NoError(t, png.Encode(&body, image.NewRGBA(image.Rect(0, 0, 2, 2))))
			response, err := common.Marshal(dto.ImageResponse{Data: []dto.ImageData{{B64Json: base64.StdEncoding.EncodeToString(body.Bytes())}}})
			require.NoError(t, err)
			input := RunInput{TokenID: token.Id, IdempotencyKey: "storage-failure", Request: dto.ImageRequest{Model: "dall-e-3", Prompt: "A mountain"}}
			calls := 0
			relay := func(RelayContext, *model.Token, string, []byte, types.RelayFormat) RelayResult {
				calls++
				return RelayResult{StatusCode: 200, Body: response, RequestID: "charged-request"}
			}
			_, _, err = CreateRun(execution, 501, input, relay)
			require.Error(t, err)
			assert.Equal(t, "image_storage_failed", ErrorCode(err))
			_, repeated, err := CreateRun(execution, 501, input, relay)
			require.Error(t, err)
			assert.True(t, repeated)
			assert.Equal(t, 1, calls)
			list, _, err := ListRuns(context.Background(), Actor{UserID: 501}, false, 0, &common.PageInfo{Page: 1, PageSize: 12})
			require.NoError(t, err)
			assert.Empty(t, list.Items)
			var count int64
			require.NoError(t, db.Model(&model.ImageStudioResult{}).Count(&count).Error)
			assert.Zero(t, count)
			if failure != "transport" {
				var object model.StorageObject
				require.NoError(t, db.First(&object).Error)
				assert.Equal(t, model.StorageObjectStatusDeletePending, object.Status)
			}
		})
	}
}
