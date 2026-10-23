package imagestudio

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"net/url"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/QuantumNous/new-api/setting/operation_setting"

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
	require.NoError(t, model.InitializeImageStudioRetention(db))
	// Only the object transport is replaced; metadata, signing and ownership use production code.
	previousUpload := uploadImage
	uploadImage = func(_ context.Context, owner int, run, result, contentType string, data []byte) (*model.StorageObject, error) {
		object := &model.StorageObject{ObjectID: result, OwnerUserID: owner, Purpose: model.StorageObjectPurposeImageStudio,
			StorageProfileID: profile.ID, StoragePolicyID: policy.ID, ObjectKey: "assets/" + run + "/" + result + ".png",
			ContentType: contentType, Size: int64(len(data)), Status: model.StorageObjectStatusUploaded,
			DeleteAfter: common.GetTimestamp() + model.ImageStudioRetentionSeconds}
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

func TestCreateRunCallsRelayOncePerIdempotencyKey(t *testing.T) {
	db, token := setupImageStudio(t)
	execution := RelayContext{Context: context.Background()}
	input := RunInput{TokenID: token.Id, IdempotencyKey: "once", Request: RunRequest{Model: "dall-e-3", Prompt: "mountain"}}
	original := []byte(`{"token_id":1,"idempotency_key":"once","request":{"model":"dall-e-3","prompt":"mountain"}}`)
	var calls atomic.Int32
	var relay RelayExecutor
	relay = func(ctx RelayContext, selected *model.Token, path string, body []byte, format types.RelayFormat, snapshot *RequestSnapshot) RelayResult {
		calls.Add(1)
		assert.Equal(t, token.Id, selected.Id)
		assert.Equal(t, "/v1/images/generations", path)
		assert.EqualValues(t, types.RelayFormatOpenAIImage, format)
		assert.JSONEq(t, `{"model":"dall-e-3","prompt":"mountain"}`, string(body))
		require.NotNil(t, snapshot)
		assert.Equal(t, original, snapshot.Original)
		require.NoError(t, model.LinkImageStudioRunRequest(snapshot.RunID, "req-1", 17))
		require.NoError(t, model.LinkImageStudioRunRequest(snapshot.RunID, "req-1", 17))
		assert.Error(t, model.LinkImageStudioRunRequest(snapshot.RunID, "other", 17))
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
	assert.Equal(t, model.ImageStudioRunStatusRunning, view.Status)
	view = awaitStudioRun(t, view.PublicID)
	assert.Equal(t, model.ImageStudioRunStatusCompleted, view.Status)
	assert.Empty(t, view.RequestID, "owners do not see the archived request reference")
	assert.Equal(t, 42, view.Quota)
	assert.Error(t, model.LinkImageStudioRunRequest(view.ID, "req-1", 17))
	require.Len(t, view.Results, 2)

	again, replay, err := CreateRun(execution, 501, input, original, relay)
	require.NoError(t, err)
	assert.True(t, replay)
	assert.Equal(t, view.PublicID, again.PublicID)
	assert.EqualValues(t, 1, calls.Load())
	var assets int64
	require.NoError(t, db.Model(&model.MediaAsset{}).Count(&assets).Error)
	assert.Zero(t, assets, "generated images are not saved to the library by default")
}

func TestReferenceImagesUseSignedEditInputsAndExpireAfterSevenDays(t *testing.T) {
	db, token := setupImageStudio(t)
	require.NoError(t, db.Model(&model.Channel{}).Where("name = ?", "image").Update("models", "gpt-image-2-1K").Error)
	require.NoError(t, db.Model(&model.Ability{}).Where("model = ?", "dall-e-3").Update("model", "gpt-image-2-1K").Error)
	model.InvalidatePricingCache()
	policy, err := model.GetStoragePolicyByKey(model.StoragePolicyAssetLibrary)
	require.NoError(t, err)
	now := common.GetTimestamp()
	object := &model.StorageObject{ObjectID: "reference-owned", OwnerUserID: 501, CreatedAt: now,
		StorageProfileID: policy.StorageProfileID, StoragePolicyID: policy.ID,
		Purpose: model.StorageObjectPurposeImageStudioReference, ObjectKey: "references/source.png",
		ContentType: "image/png", Status: model.StorageObjectStatusUploaded,
		DeleteAfter: now + 7*24*60*60}
	require.NoError(t, db.Create(object).Error)
	assert.Equal(t, object.CreatedAt+7*24*60*60, object.DeleteAfter)

	input := RunInput{TokenID: token.Id, IdempotencyKey: "with-reference", Request: RunRequest{
		Model: "gpt-image-2-1K", Prompt: "Use the reference composition", ReferenceIDs: []string{object.ObjectID},
	}}
	view, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, input, nil,
		func(_ RelayContext, _ *model.Token, path string, body []byte, _ types.RelayFormat, _ *RequestSnapshot) RelayResult {
			assert.Equal(t, "/v1/images/edits", path)
			var request dto.ImageRequest
			require.NoError(t, common.Unmarshal(body, &request))
			assert.Empty(t, request.Image, "the legacy image field is ignored by the provider")
			var references []struct {
				ImageURL string `json:"image_url"`
			}
			require.NoError(t, common.Unmarshal(request.Images, &references))
			require.Len(t, references, 1)
			parsed, parseErr := url.Parse(references[0].ImageURL)
			require.NoError(t, parseErr)
			assert.Equal(t, "/references/source.png", parsed.Path)
			assert.NotEmpty(t, parsed.Query().Get("x-oss-signature"))
			return RelayResult{StatusCode: http.StatusOK, Body: pngResponse(t, 1)}
		})
	require.NoError(t, err)
	awaitStudioRun(t, view.PublicID)
	atExpiry, err := model.ClaimStorageObjectsForDeletion(object.DeleteAfter, object.DeleteAfter+60, "reference-expiry", 10)
	require.NoError(t, err)
	require.Len(t, atExpiry, 1)
	assert.Equal(t, object.ID, atExpiry[0].ID)
}

func TestUnavailableReferenceImagesFailBeforeCreatingOrBillingARun(t *testing.T) {
	db, token := setupImageStudio(t)
	now := common.GetTimestamp()
	for _, item := range []struct {
		name    string
		owner   int
		purpose string
		expires int64
		status  string
	}{
		{"another-owner", 502, model.StorageObjectPurposeImageStudioReference, now + 60, model.StorageObjectStatusUploaded},
		{"expired", 501, model.StorageObjectPurposeImageStudioReference, now - 1, model.StorageObjectStatusUploaded},
		{"deleted", 501, model.StorageObjectPurposeImageStudioReference, now + 60, model.StorageObjectStatusDeleted},
		{"another-purpose", 501, model.StorageObjectPurposeAssetLibrary, now + 60, model.StorageObjectStatusUploaded},
	} {
		t.Run(item.name, func(t *testing.T) {
			require.NoError(t, db.Create(&model.StorageObject{ObjectID: item.name, OwnerUserID: item.owner,
				Purpose: item.purpose, DeleteAfter: item.expires, Status: item.status}).Error)
			input := RunInput{TokenID: token.Id, IdempotencyKey: item.name, Request: RunRequest{
				Model: "dall-e-3", Prompt: "Use this image", ReferenceIDs: []string{item.name},
			}}
			_, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, input, nil,
				func(RelayContext, *model.Token, string, []byte, types.RelayFormat, *RequestSnapshot) RelayResult {
					t.Fatal("unavailable references must not reach the billed relay")
					return RelayResult{}
				})
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, StatusCode(err))
		})
	}
	var count int64
	_, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, RunInput{
		TokenID: token.Id, IdempotencyKey: "too-many-references",
		Request: RunRequest{Model: "dall-e-3", Prompt: "Use these images", ReferenceIDs: make([]string, 9)},
	}, nil, nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	require.NoError(t, db.Model(&model.ImageStudioRun{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestReferenceUploadRejectsInvalidImagesAndNonHTTPImports(t *testing.T) {
	for _, data := range [][]byte{[]byte("<svg xmlns='http://www.w3.org/2000/svg'></svg>"), make([]byte, MaxReferenceImageBytes+1)} {
		_, err := UploadReference(context.Background(), 501, data)
		require.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	}
	_, err := ImportReference(context.Background(), 501, "file:///etc/passwd")
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, StatusCode(err))
}

func TestFailedRunKeepsRequestIDForAudit(t *testing.T) {
	_, token := setupImageStudio(t)
	input := RunInput{TokenID: token.Id, IdempotencyKey: "fail", Request: RunRequest{Model: "dall-e-3", Prompt: "mountain"}}
	relay := func(RelayContext, *model.Token, string, []byte, types.RelayFormat, *RequestSnapshot) RelayResult {
		return RelayResult{StatusCode: http.StatusBadRequest, Body: []byte(`{"error":{"message":"content rejected by Bearer secret-key"}}`), RequestID: "req-fail", ChannelID: 17}
	}
	view, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, input, nil, relay)
	require.NoError(t, err)
	view = awaitStudioRun(t, view.PublicID)
	assert.Equal(t, model.ImageStudioRunStatusFailed, view.Status)
	assert.NotContains(t, view.ErrorMessage, "secret-key")
	run, err := model.FindImageStudioRunByIdempotencyKey(501, "fail")
	require.NoError(t, err)
	assert.Equal(t, model.ImageStudioRunStatusFailed, run.Status)
	assert.Equal(t, "req-fail", run.RequestID)
	assert.Equal(t, 17, run.ChannelID)
}

func TestHistoryIsOwnerScoped(t *testing.T) {
	db, token := setupImageStudio(t)
	relay := func(RelayContext, *model.Token, string, []byte, types.RelayFormat, *RequestSnapshot) RelayResult {
		return RelayResult{StatusCode: http.StatusOK, Body: pngResponse(t, 1), RequestID: "req"}
	}
	view, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, RunInput{TokenID: token.Id, IdempotencyKey: "own", Request: RunRequest{Model: "dall-e-3", Prompt: "lake"}}, nil, relay)
	require.NoError(t, err)
	view = awaitStudioRun(t, view.PublicID)
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

func TestDeletingSavedImageAsksAboutTheLibraryAsset(t *testing.T) {
	db, token := setupImageStudio(t)
	require.NoError(t, db.AutoMigrate(&model.AssetReplica{}))
	relay := func(RelayContext, *model.Token, string, []byte, types.RelayFormat, *RequestSnapshot) RelayResult {
		return RelayResult{StatusCode: http.StatusOK, Body: pngResponse(t, 2), RequestID: "req-saved"}
	}
	view, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, RunInput{TokenID: token.Id, IdempotencyKey: "saved", Request: RunRequest{Model: "dall-e-3", Prompt: "lake"}}, nil, relay)
	require.NoError(t, err)
	view = awaitStudioRun(t, view.PublicID)
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

// Wait for observable persisted completion, never assume CreateRun waits for relay.
func awaitStudioRun(t *testing.T, id string) *RunView {
	t.Helper()
	var view *RunView
	require.Eventually(t, func() bool {
		var err error
		view, err = GetRun(context.Background(), 501, id)
		return err == nil && view.Status != model.ImageStudioRunStatusRunning
	}, 5*time.Second, 10*time.Millisecond)
	return view
}

func TestRunningHistorySurvivesRefreshAndExpiredRunsStopPolling(t *testing.T) {
	db, _ := setupImageStudio(t)
	running := &model.ImageStudioRun{UserID: 501, IdempotencyKey: "running-history", Model: "dall-e-3", Prompt: "Draft", Status: model.ImageStudioRunStatusRunning}
	stale := &model.ImageStudioRun{UserID: 501, IdempotencyKey: "stale-history", Model: "dall-e-3", Prompt: "Old draft", Status: model.ImageStudioRunStatusRunning, CreatedAt: common.GetTimestamp() - model.ImageStudioRunStaleSeconds - 1}
	require.NoError(t, db.Create(running).Error)
	require.NoError(t, db.Create(stale).Error)
	list, err := ListRuns(context.Background(), ListFilter{OwnerUserID: 501, Page: &common.PageInfo{Page: 1, PageSize: 12}})
	require.NoError(t, err)
	require.Len(t, list.Items, 2)
	assert.Equal(t, model.ImageStudioRunStatusFailed, list.Items[0].Status)
	assert.Equal(t, model.ImageStudioRunStatusRunning, list.Items[1].Status)
}

func TestCreateRunRejectsUnboundedImageSizesBeforeRelay(t *testing.T) {
	db, token := setupImageStudio(t)
	_, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, RunInput{TokenID: token.Id, IdempotencyKey: "size", Request: RunRequest{Model: "dall-e-3", Prompt: "Lake", Size: "999999x999999"}}, nil, nil)
	assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	assert.Equal(t, "invalid_size", ErrorCode(err))
	var count int64
	require.NoError(t, db.Model(&model.ImageStudioRun{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestCreateRunReturnsWhileRelayIsBlockedAndIgnoresClientDisconnect(t *testing.T) {
	_, token := setupImageStudio(t)
	release := make(chan struct{})
	body := pngResponse(t, 1)
	relay := func(ctx RelayContext, _ *model.Token, _ string, _ []byte, _ types.RelayFormat, _ *RequestSnapshot) RelayResult {
		<-release
		assert.NoError(t, ctx.Context.Err())
		return RelayResult{StatusCode: http.StatusOK, Body: body}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	first, _, err := CreateRun(RelayContext{Context: ctx}, 501, RunInput{TokenID: token.Id, IdempotencyKey: "queued-first", Request: RunRequest{Model: "dall-e-3", Prompt: "first"}}, nil, relay)
	require.NoError(t, err)
	second, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, RunInput{TokenID: token.Id, IdempotencyKey: "queued-second", Request: RunRequest{Model: "dall-e-3", Prompt: "second", Size: "1792x1024"}}, nil, relay)
	cancel()
	close(release)
	require.NoError(t, err)
	assert.Equal(t, model.ImageStudioRunStatusRunning, first.Status)
	assert.Equal(t, model.ImageStudioRunStatusRunning, second.Status)
	assert.Equal(t, model.ImageStudioRunStatusCompleted, awaitStudioRun(t, first.PublicID).Status)
	assert.Equal(t, model.ImageStudioRunStatusCompleted, awaitStudioRun(t, second.PublicID).Status)
}

func TestBackgroundPanicFailsRunWithoutExposingPanicData(t *testing.T) {
	_, token := setupImageStudio(t)
	view, replay, err := CreateRun(RelayContext{Context: context.Background()}, 501,
		RunInput{TokenID: token.Id, IdempotencyKey: "panic", Request: RunRequest{Model: "dall-e-3", Prompt: "mountain"}}, nil,
		func(RelayContext, *model.Token, string, []byte, types.RelayFormat, *RequestSnapshot) RelayResult {
			panic("private provider details")
		})
	require.NoError(t, err)
	assert.False(t, replay)
	finished := awaitStudioRun(t, view.PublicID)
	assert.Equal(t, model.ImageStudioRunStatusFailed, finished.Status)
	assert.Equal(t, "image generation failed", finished.ErrorMessage)
	assert.Empty(t, finished.Results)
}

func TestCreateRunLimitsRunningGenerationsPerUser(t *testing.T) {
	_, token := setupImageStudio(t)
	release := make(chan struct{})
	relay := func(RelayContext, *model.Token, string, []byte, types.RelayFormat, *RequestSnapshot) RelayResult {
		<-release
		return RelayResult{StatusCode: http.StatusOK, Body: pngResponse(t, 1)}
	}
	submit := func(key string) (*RunView, error) {
		view, _, err := CreateRun(RelayContext{Context: context.Background()}, 501, RunInput{TokenID: token.Id, IdempotencyKey: key, Request: RunRequest{Model: "dall-e-3", Prompt: key}}, nil, relay)
		return view, err
	}
	var running []*RunView
	for i := 0; i < operation_setting.ImageStudioMaxRunningPerUser; i++ {
		view, err := submit(fmt.Sprintf("slot-%d", i))
		require.NoError(t, err)
		running = append(running, view)
	}
	_, err := submit("overflow")
	assert.Equal(t, http.StatusTooManyRequests, StatusCode(err))
	assert.Equal(t, "too_many_running_runs", ErrorCode(err))
	close(release)
	for _, view := range running {
		assert.Equal(t, model.ImageStudioRunStatusCompleted, awaitStudioRun(t, view.PublicID).Status)
	}
	after, err := submit("after-release")
	require.NoError(t, err)
	awaitStudioRun(t, after.PublicID)
}

func TestReferenceHistoryListsOnlyOwnersLiveUploadsWithPagination(t *testing.T) {
	db, _ := setupImageStudio(t)
	var profile model.StorageProfile
	require.NoError(t, db.First(&profile).Error)
	now := common.GetTimestamp()
	objects := []model.StorageObject{
		{ObjectID: "older-upload", OwnerUserID: 501, Purpose: model.StorageObjectPurposeImageStudioReference, Status: model.StorageObjectStatusUploaded, DeleteAfter: now + 3600},
		{ObjectID: "newer-upload", OwnerUserID: 501, Purpose: model.StorageObjectPurposeImageStudioReference, Status: model.StorageObjectStatusUploaded, DeleteAfter: now + 7200},
		{ObjectID: "other-owner", OwnerUserID: 502, Purpose: model.StorageObjectPurposeImageStudioReference, Status: model.StorageObjectStatusUploaded, DeleteAfter: now + 3600},
		{ObjectID: "expired", OwnerUserID: 501, Purpose: model.StorageObjectPurposeImageStudioReference, Status: model.StorageObjectStatusUploaded, DeleteAfter: now - 1},
		{ObjectID: "deleted", OwnerUserID: 501, Purpose: model.StorageObjectPurposeImageStudioReference, Status: model.StorageObjectStatusDeletePending, DeleteAfter: now + 3600},
		{ObjectID: "generated-image", OwnerUserID: 501, Purpose: model.StorageObjectPurposeImageStudio, Status: model.StorageObjectStatusUploaded, DeleteAfter: now + 3600},
	}
	for i := range objects {
		objects[i].StorageProfileID = profile.ID
		objects[i].ObjectKey = "references/" + objects[i].ObjectID + ".png"
		objects[i].ContentType = "image/png"
		require.NoError(t, db.Create(&objects[i]).Error)
	}
	page := &common.PageInfo{Page: 1, PageSize: 1}
	first, err := ListReferences(context.Background(), 501, page)
	require.NoError(t, err)
	assert.EqualValues(t, 2, first.Total)
	require.Len(t, first.Items, 1)
	assert.Equal(t, "newer-upload", first.Items[0].ID)
	assert.Equal(t, now+7200, first.Items[0].ExpiresAt)
	assert.NotZero(t, first.Items[0].CreatedAt)
	assert.Contains(t, first.Items[0].URL, "references/newer-upload.png")
	assert.Contains(t, first.Items[0].URL, "x-oss-signature=")
	page.Page = 2
	second, err := ListReferences(context.Background(), 501, page)
	require.NoError(t, err)
	require.Len(t, second.Items, 1)
	assert.Equal(t, "older-upload", second.Items[0].ID)
	empty, err := ListReferences(context.Background(), 503, page)
	require.NoError(t, err)
	assert.Empty(t, empty.Items)
	assert.Zero(t, empty.Total)
}
