package imagestudio

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	storageService "github.com/QuantumNous/new-api/service/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupImageStudio(t *testing.T) (*gorm.DB, *model.Token) {
	t.Helper()
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousRedis, previousPath, previousMaster := common.RedisEnabled, common.SQLitePath, common.IsMasterNode
	previousDBType, previousLogDBType := common.MainDatabaseType(), common.LogDatabaseType()
	t.Setenv("SQL_DSN", "")
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "image-studio-test-encryption-key")
	common.SQLitePath = filepath.Join(t.TempDir(), "image-studio.db")
	common.IsMasterNode, common.RedisEnabled = false, false
	require.NoError(t, model.InitDB())
	db := model.DB
	model.LOG_DB = db
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.RedisEnabled, common.SQLitePath, common.IsMasterNode = previousRedis, previousPath, previousMaster
		common.SetMainDatabaseType(previousDBType)
		common.SetLogDatabaseType(previousLogDBType)
		model.InvalidatePricingCache()
	})
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Token{}, &model.Channel{}, &model.Ability{}, &model.Model{}, &model.Vendor{}, &model.Log{},
		&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{},
		&model.StorageProfile{}, &model.StorageCredential{}, &model.StoragePolicy{}, &model.StorageObject{}))
	require.NoError(t, db.Create(&model.User{Id: 501, Username: "owner", Group: "default", Status: common.UserStatusEnabled}).Error)
	channel := &model.Channel{Type: constant.ChannelTypeOpenAI, Name: "image", Status: common.ChannelStatusEnabled, Models: "dall-e-3", Group: "default"}
	require.NoError(t, db.Create(channel).Error)
	require.NoError(t, db.Create(&model.Ability{Group: "default", Model: "dall-e-3", ChannelId: channel.Id, Enabled: true}).Error)
	token := &model.Token{UserId: 501, Key: "studiofixture", Name: "Images", Status: common.TokenStatusEnabled, ExpiredTime: -1, UnlimitedQuota: true, Group: "default"}
	require.NoError(t, db.Create(token).Error)
	model.InvalidatePricingCache()

	profile, err := storageService.SaveProfile(0, storageService.ProfileInput{
		Name: "Studio OSS", ProviderType: model.StorageProviderAliyunOSS, Status: model.StorageProfileStatusEnabled,
		Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", Region: "cn-hangzhou", Bucket: "studio-test-bucket",
		AuthType: model.StorageCredentialAuthStatic, AccessKeyID: "test-id", AccessKeySecret: "test-secret",
	})
	require.NoError(t, err)
	policy := model.DefaultAssetLibraryStoragePolicy()
	policy.StorageProfileID, policy.Enabled = profile.ID, true
	require.NoError(t, model.SaveStoragePolicy(policy))
	// Only the object transport is replaced; metadata, signing and ownership use production code.
	previousUpload := uploadImage
	uploadImage = func(_ context.Context, owner int, run, result, contentType string, data []byte) (*model.StorageObject, error) {
		object := &model.StorageObject{ObjectID: result, OwnerUserID: owner, Purpose: model.StorageObjectPurposeImageStudio,
			StorageProfileID: profile.ID, StoragePolicyID: policy.ID, ObjectKey: "assets/" + run + "/" + result + ".png",
			ContentType: contentType, Size: int64(len(data)), Status: model.StorageObjectStatusUploaded}
		return object, db.Create(object).Error
	}
	t.Cleanup(func() { uploadImage = previousUpload })
	return db, token
}

func pngResponse(t *testing.T, count int) []byte {
	t.Helper()
	var buffer bytes.Buffer
	require.NoError(t, png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	data := make([]dto.ImageData, count)
	for i := range data {
		data[i] = dto.ImageData{B64Json: base64.StdEncoding.EncodeToString(buffer.Bytes())}
	}
	body, err := common.Marshal(dto.ImageResponse{Data: data})
	require.NoError(t, err)
	return body
}

func TestGPTImage2IsAvailableForImagesAndExcludedFromPromptAssist(t *testing.T) {
	db, token := setupImageStudio(t)
	require.NoError(t, db.Model(&model.Channel{}).Where("name = ?", "image").Update("models", "gpt-image-2-1K").Error)
	require.NoError(t, db.Model(&model.Ability{}).Where("model = ?", "dall-e-3").Update("model", "gpt-image-2-1K").Error)
	model.InvalidatePricingCache()

	images, err := ModelsForToken(501, token.Id, constant.EndpointTypeImageGeneration)
	require.NoError(t, err)
	assert.Contains(t, images, "gpt-image-2-1K")
	text, err := ModelsForToken(501, token.Id, constant.EndpointTypeOpenAI)
	require.NoError(t, err)
	assert.NotContains(t, text, "gpt-image-2-1K")
}

func TestCreateRunCallsRelayOncePerIdempotencyKey(t *testing.T) {
	db, token := setupImageStudio(t)
	execution := RelayContext{Context: context.Background()}
	input := RunInput{TokenID: token.Id, IdempotencyKey: "once", Request: RunRequest{Model: "dall-e-3", Prompt: "mountain"}}
	original := []byte(`{"token_id":1,"idempotency_key":"once","request":{"model":"dall-e-3","prompt":"mountain"}}`)
	calls := 0
	var relay RelayExecutor
	relay = func(ctx RelayContext, selected *model.Token, path string, body []byte, format types.RelayFormat, snapshot *RequestSnapshot) RelayResult {
		calls++
		assert.Equal(t, token.Id, selected.Id)
		assert.Equal(t, "/v1/images/generations", path)
		assert.EqualValues(t, types.RelayFormatOpenAIImage, format)
		assert.JSONEq(t, `{"model":"dall-e-3","prompt":"mountain"}`, string(body))
		require.NotNil(t, snapshot)
		assert.Equal(t, original, snapshot.Original)
		// A duplicate submitted while generating sees the running run.
		inflight, replay, err := CreateRun(ctx, 501, input, original, relay)
		require.NoError(t, err)
		assert.True(t, replay)
		assert.Equal(t, model.ImageStudioRunStatusRunning, inflight.Status)
		require.NoError(t, db.Create(&model.Log{UserId: 501, Type: model.LogTypeConsume, RequestId: "req-1", Quota: 42}).Error)
		return RelayResult{StatusCode: http.StatusOK, Body: pngResponse(t, 2), RequestID: "req-1"}
	}
	view, replay, err := CreateRun(execution, 501, input, original, relay)
	require.NoError(t, err)
	assert.False(t, replay)
	assert.Equal(t, model.ImageStudioRunStatusCompleted, view.Status)
	assert.Empty(t, view.RequestID, "owners do not see the archived request reference")
	assert.Equal(t, 42, view.Quota)
	require.Len(t, view.Results, 2)
	assert.Contains(t, view.Results[0].PreviewURL, "studio-test-bucket")
	assert.Contains(t, view.Results[0].ThumbnailURL, "x-oss-process")

	again, replay, err := CreateRun(execution, 501, input, original, relay)
	require.NoError(t, err)
	assert.True(t, replay)
	assert.Equal(t, view.PublicID, again.PublicID)
	assert.Equal(t, 1, calls)
	var assets int64
	require.NoError(t, db.Model(&model.MediaAsset{}).Count(&assets).Error)
	assert.Zero(t, assets, "generated images are not saved to the library by default")
}

func TestFailedRunKeepsRequestIDForAudit(t *testing.T) {
	_, token := setupImageStudio(t)
	input := RunInput{TokenID: token.Id, IdempotencyKey: "fail", Request: RunRequest{Model: "dall-e-3", Prompt: "mountain"}}
	relay := func(RelayContext, *model.Token, string, []byte, types.RelayFormat, *RequestSnapshot) RelayResult {
		return RelayResult{StatusCode: http.StatusBadRequest, Body: []byte(`{"error":{"message":"content rejected by Bearer secret-key"}}`), RequestID: "req-fail"}
	}
	_, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, input, nil, relay)
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	assert.NotContains(t, err.Error(), "secret-key")
	run, err := model.FindImageStudioRunByIdempotencyKey(501, "fail")
	require.NoError(t, err)
	assert.Equal(t, model.ImageStudioRunStatusFailed, run.Status)
	assert.Equal(t, "req-fail", run.RequestID)
	var objects int64
	require.NoError(t, model.DB.Model(&model.StorageObject{}).Count(&objects).Error)
	assert.Zero(t, objects)
}

func TestHistoryIsOwnerScoped(t *testing.T) {
	db, token := setupImageStudio(t)
	relay := func(RelayContext, *model.Token, string, []byte, types.RelayFormat, *RequestSnapshot) RelayResult {
		return RelayResult{StatusCode: http.StatusOK, Body: pngResponse(t, 1), RequestID: "req"}
	}
	view, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, RunInput{TokenID: token.Id, IdempotencyKey: "own", Request: RunRequest{Model: "dall-e-3", Prompt: "lake"}}, nil, relay)
	require.NoError(t, err)
	resultID := view.Results[0].ID
	page := &common.PageInfo{Page: 1, PageSize: 10}

	foreign, err := ListRuns(context.Background(), ListFilter{OwnerUserID: 502, Page: page})
	require.NoError(t, err)
	assert.Zero(t, foreign.Total)
	_, err = GetRun(context.Background(), 502, view.PublicID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.ErrorIs(t, DeleteResult(502, resultID, LinkedAssetsAsk), gorm.ErrRecordNotFound)
	_, err = SaveResult(context.Background(), 502, resultID, "group")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.ErrorIs(t, DeleteRun(502, view.PublicID, LinkedAssetsAsk), gorm.ErrRecordNotFound)

	review, err := ListRuns(context.Background(), ListFilter{Review: true, Page: page})
	require.NoError(t, err)
	require.Len(t, review.Items, 1)
	assert.Equal(t, "owner", review.Items[0].Username)
	require.Len(t, review.Users, 1)
	assert.Equal(t, "req", review.Items[0].RequestID, "reviewers can open the archived request")

	require.NoError(t, DeleteResult(501, resultID, LinkedAssetsAsk))
	var object model.StorageObject
	require.NoError(t, db.Where("owner_user_id = ?", 501).First(&object).Error)
	assert.Equal(t, model.StorageObjectStatusDeletePending, object.Status)
	require.NoError(t, DeleteRun(501, view.PublicID, LinkedAssetsAsk))
	own, err := ListRuns(context.Background(), ListFilter{OwnerUserID: 501, Page: page})
	require.NoError(t, err)
	assert.Zero(t, own.Total)
	// The deleted run keeps its key, so a delayed retry cannot generate again.
	_, replay, err := CreateRun(RelayContext{Context: context.Background()}, 501, RunInput{TokenID: token.Id, IdempotencyKey: "own", Request: RunRequest{Model: "dall-e-3", Prompt: "lake"}}, nil, relay)
	assert.True(t, replay)
	assert.Equal(t, http.StatusGone, StatusCode(err))
}

func TestSaveClaimAllowsOneConcurrentSave(t *testing.T) {
	db, _ := setupImageStudio(t)
	result := &model.ImageStudioResult{RunID: 1}
	require.NoError(t, db.Create(result).Error)
	claimedAt, err := model.ClaimImageStudioResultSave(result.ID)
	require.NoError(t, err)
	_, err = model.ClaimImageStudioResultSave(result.ID)
	assert.ErrorIs(t, err, model.ErrImageStudioSaveInProgress)
	linked, err := model.CompleteImageStudioResultSave(result.ID, claimedAt, 9)
	require.NoError(t, err)
	assert.True(t, linked)
	_, err = model.ClaimImageStudioResultSave(result.ID)
	assert.ErrorIs(t, err, model.ErrImageStudioSaveInProgress, "a saved image is not saved twice")
}

func imageResponse(t *testing.T, encode func(*bytes.Buffer, image.Image) error) []byte {
	t.Helper()
	var buffer bytes.Buffer
	require.NoError(t, encode(&buffer, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	body, err := common.Marshal(dto.ImageResponse{Data: []dto.ImageData{{B64Json: base64.StdEncoding.EncodeToString(buffer.Bytes())}}})
	require.NoError(t, err)
	return body
}

func TestStoredImageKeepsItsActualFormat(t *testing.T) {
	db, token := setupImageStudio(t)
	relay := func(RelayContext, *model.Token, string, []byte, types.RelayFormat, *RequestSnapshot) RelayResult {
		return RelayResult{StatusCode: http.StatusOK, RequestID: "req-jpeg", Body: imageResponse(t, func(buffer *bytes.Buffer, img image.Image) error {
			return jpeg.Encode(buffer, img, nil)
		})}
	}
	view, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, RunInput{TokenID: token.Id, IdempotencyKey: "jpeg", Request: RunRequest{Model: "dall-e-3", Prompt: "lake"}}, nil, relay)
	require.NoError(t, err)
	var object model.StorageObject
	require.NoError(t, db.First(&object).Error)
	assert.Equal(t, "image/jpeg", object.ContentType)
	require.Len(t, view.Results, 1)
	download, err := url.Parse(view.Results[0].DownloadURL)
	require.NoError(t, err)
	assert.Equal(t, `attachment; filename="`+view.Results[0].ID+`.jpg"`, download.Query().Get("response-content-disposition"))
}

// When the results cannot be committed with the run, the uploaded objects are
// queued for the storage cleanup job instead of being left behind.
func TestUncommittedResultsQueueObjectsForCleanup(t *testing.T) {
	db, token := setupImageStudio(t)
	relay := func(ctx RelayContext, _ *model.Token, _ string, _ []byte, _ types.RelayFormat, snapshot *RequestSnapshot) RelayResult {
		// Another request already failed this run as interrupted.
		require.NoError(t, db.Model(&model.ImageStudioRun{}).Where("id = ?", snapshot.RunID).Update("status", model.ImageStudioRunStatusFailed).Error)
		return RelayResult{StatusCode: http.StatusOK, Body: pngResponse(t, 2), RequestID: "req-late"}
	}
	view, replay, err := CreateRun(RelayContext{Context: context.Background()}, 501, RunInput{TokenID: token.Id, IdempotencyKey: "late", Request: RunRequest{Model: "dall-e-3", Prompt: "lake"}}, nil, relay)
	require.NoError(t, err)
	assert.True(t, replay)
	assert.Equal(t, model.ImageStudioRunStatusFailed, view.Status)
	assert.Empty(t, view.Results)
	var results, pending int64
	require.NoError(t, db.Model(&model.ImageStudioResult{}).Count(&results).Error)
	assert.Zero(t, results)
	require.NoError(t, db.Model(&model.StorageObject{}).Where("status = ?", model.StorageObjectStatusDeletePending).Count(&pending).Error)
	assert.EqualValues(t, 2, pending)
}

func TestDeletingSavedImageAsksAboutTheLibraryAsset(t *testing.T) {
	db, token := setupImageStudio(t)
	require.NoError(t, db.AutoMigrate(&model.AssetReplica{}))
	relay := func(RelayContext, *model.Token, string, []byte, types.RelayFormat, *RequestSnapshot) RelayResult {
		return RelayResult{StatusCode: http.StatusOK, Body: pngResponse(t, 2), RequestID: "req-saved"}
	}
	view, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, RunInput{TokenID: token.Id, IdempotencyKey: "saved", Request: RunRequest{Model: "dall-e-3", Prompt: "lake"}}, nil, relay)
	require.NoError(t, err)
	require.Len(t, view.Results, 2)
	assets := make([]*model.MediaAsset, 2)
	for i, result := range view.Results {
		object := &model.StorageObject{ObjectID: "asset-object-" + result.ID, OwnerUserID: 501, Purpose: model.StorageObjectPurposeAssetLibrary, Status: model.StorageObjectStatusUploaded}
		require.NoError(t, db.Create(object).Error)
		assets[i] = &model.MediaAsset{PublicID: "asset-" + result.ID, OwnerUserID: 501, StorageObjectID: object.ID, Status: model.AssetStatusReady}
		require.NoError(t, db.Create(assets[i]).Error)
		require.NoError(t, db.Model(&model.ImageStudioResult{}).Where("public_id = ?", result.ID).Update("media_asset_id", assets[i].ID).Error)
	}

	err = DeleteResult(501, view.Results[0].ID, LinkedAssetsAsk)
	assert.Equal(t, http.StatusConflict, StatusCode(err))
	assert.Equal(t, "linked_assets_exist", ErrorCode(err))

	require.NoError(t, DeleteResult(501, view.Results[0].ID, LinkedAssetsKeep))
	var kept model.MediaAsset
	require.NoError(t, db.First(&kept, assets[0].ID).Error)
	assert.Equal(t, model.AssetStatusReady, kept.Status)

	assert.Equal(t, http.StatusConflict, StatusCode(DeleteRun(501, view.PublicID, LinkedAssetsAsk)))
	require.NoError(t, DeleteRun(501, view.PublicID, LinkedAssetsDelete))
	var deleted model.MediaAsset
	require.NoError(t, db.First(&deleted, assets[1].ID).Error)
	assert.Equal(t, model.AssetStatusDeleted, deleted.Status)
	var object model.StorageObject
	require.NoError(t, db.First(&object, deleted.StorageObjectID).Error)
	assert.Equal(t, model.StorageObjectStatusDeletePending, object.Status)
}
