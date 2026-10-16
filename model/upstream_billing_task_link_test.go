package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type upstreamTaskLinkFixture struct {
	source     Log
	settlement Log
	task       Task
	manifest   map[string]any
	candidate  map[string]any
}

func upstreamTaskLinkReviewFixture(t *testing.T, failed bool) upstreamTaskLinkFixture {
	t.Helper()
	upstreamBillingModelFixture(t)
	require.NoError(t, DB.AutoMigrate(&Log{}, &Task{}, &UpstreamBillingTaskLink{}))
	start, _, err := BillingMonthBounds("2026-09")
	require.NoError(t, err)
	pricing := map[string]any{"group_ratio": 0.86, "model_price": -1, "model_ratio": 3.285714}
	sourceMetadata := map[string]any{"is_task": true}
	for key, value := range pricing {
		sourceMetadata[key] = value
	}
	sourceOther, err := common.Marshal(sourceMetadata)
	require.NoError(t, err)
	source := Log{Id: 1, UserId: 1, ChannelId: 19, TokenId: 7, Type: LogTypeConsume, CreatedAt: start + 1, ModelName: "review-video", Group: "original-group", Quota: 100, Content: "task submitted", RequestId: "review-original", UpstreamRequestId: "review-upstream-request", Other: string(sourceOther)}
	require.NoError(t, LOG_DB.Create(&source).Error)
	task := Task{TaskID: "review-public-task", UserId: source.UserId, ChannelId: source.ChannelId, Group: source.Group, Platform: "54", Status: TaskStatusSuccess, Quota: 40, SubmitTime: source.CreatedAt, FinishTime: source.CreatedAt + 20, Properties: Properties{OriginModelName: source.ModelName, UpstreamModelName: "provider-video"}, PrivateData: TaskPrivateData{TokenId: source.TokenId, UpstreamTaskID: "review-provider-task", BillingContext: &TaskBillingContext{OriginModelName: source.ModelName, GroupRatio: 0.86, ModelPrice: -1, ModelRatio: 3.285714}}}
	task.SetData(map[string]any{"id": task.PrivateData.UpstreamTaskID, "usage": map[string]any{"total_tokens": 123}, "duration": 5})
	settlementMetadata := map[string]any{"is_task": true, "task_id": task.TaskID, "pre_consumed_quota": 100, "actual_quota": 40}
	for key, value := range pricing {
		settlementMetadata[key] = value
	}
	settlement := Log{Id: 2, UserId: source.UserId, ChannelId: source.ChannelId, TokenId: source.TokenId, Type: LogTypeRefund, CreatedAt: source.CreatedAt + 10, ModelName: source.ModelName, Group: source.Group, Quota: 60, RequestId: "review-settlement"}
	chain := map[string]any{"log_id": settlement.Id, "type": settlement.Type, "signed_delta": -60, "occurred_at": settlement.CreatedAt, "pre_consumed_quota": 100, "actual_quota": 40, "anchor_matches": true}
	if failed {
		task.Status, task.Quota = TaskStatusFailure, 0
		task.SetData(map[string]any{"id": task.PrivateData.UpstreamTaskID, "status": "failed"})
		delete(settlementMetadata, "pre_consumed_quota")
		delete(settlementMetadata, "actual_quota")
		settlement.Quota = 100
		chain = map[string]any{"log_id": settlement.Id, "type": settlement.Type, "signed_delta": -100, "occurred_at": settlement.CreatedAt, "full_failure_refund": true}
	}
	settlementOther, err := common.Marshal(settlementMetadata)
	require.NoError(t, err)
	settlement.Other = string(settlementOther)
	require.NoError(t, LOG_DB.Create(&settlement).Error)
	require.NoError(t, DB.Create(&task).Error)
	properties, err := common.Marshal(task.Properties)
	require.NoError(t, err)
	hash := func(value []byte) string {
		digest := sha256.Sum256(value)
		return hex.EncodeToString(digest[:])
	}
	pricingChecks := make(map[string]any)
	for key, value := range pricing {
		pricingChecks[key] = map[string]any{"log": value, "task": value, "equal": true}
	}
	candidate := map[string]any{
		"log_id": source.Id, "user_id": source.UserId, "channel_id": source.ChannelId, "token_id": source.TokenId,
		"created_at": source.CreatedAt, "model": source.ModelName, "group": source.Group, "quota": source.Quota,
		"request_id": source.RequestId, "upstream_request_id": source.UpstreamRequestId, "content_sha256": hash([]byte(source.Content)),
		"task_row_id": task.ID, "task_id": task.TaskID, "upstream_task_id": task.PrivateData.UpstreamTaskID,
		"status": task.Status, "submit_time": task.SubmitTime, "initial_quota_from_final_and_linked_events": source.Quota,
		"data_sha256": hash(task.Data), "properties_sha256": hash(properties), "data_id_matches_private_upstream_id": true,
		"adjustment_chain": []any{chain}, "adjustment_chain_valid": true, "all_submission_reverse_match_count": 1,
		"pricing_checks": pricingChecks, "decision": "review_compound_exact",
	}
	manifest := map[string]any{"schema_version": 1, "rule": "compound_task_link_exact_second_v1", "read_only": true, "candidates": []any{candidate}}
	return upstreamTaskLinkFixture{source: source, settlement: settlement, task: task, manifest: manifest, candidate: candidate}
}

func upstreamTaskLinkManifestBytes(t *testing.T, manifest map[string]any) ([]byte, string) {
	t.Helper()
	raw, err := common.Marshal(manifest)
	require.NoError(t, err)
	digest := sha256.Sum256(raw)
	return raw, hex.EncodeToString(digest[:])
}

func TestUpstreamBillingTaskLinksRequireIndependentReviewedAnchors(t *testing.T) {
	for _, failed := range []bool{false, true} {
		name := "successful preconsume and actual anchor"
		if failed {
			name = "failed explicit task full refund"
		}
		t.Run(name, func(t *testing.T) {
			fixture := upstreamTaskLinkReviewFixture(t, failed)
			raw, digest := upstreamTaskLinkManifestBytes(t, fixture.manifest)
			actorID := 0
			if failed {
				actorID = 1
			}
			links, err := ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "review-batch", actorID)
			require.NoError(t, err)
			require.Len(t, links, 1)
			assert.Equal(t, actorID, links[0].CreatedBy, "system execution is audited as zero without impersonating an administrator")
			var original Log
			require.NoError(t, LOG_DB.First(&original, fixture.source.Id).Error)
			assert.Equal(t, fixture.source.Other, original.Other, "audited association never rewrites the historical log")
			assert.Equal(t, fixture.source.Quota, original.Quota)
			var task Task
			require.NoError(t, DB.First(&task, fixture.task.ID).Error)
			assert.Equal(t, fixture.task.Quota, task.Quota)
			assert.Equal(t, fixture.task.PrivateData, task.PrivateData)
		})
	}
}

func TestUpstreamBillingTaskLinksAcceptLegacyUnmarkedAccountingEvents(t *testing.T) {
	for _, failed := range []bool{false, true} {
		name := "independently anchored settlement"
		if failed {
			name = "explicit full failure refund"
		}
		t.Run(name, func(t *testing.T) {
			fixture := upstreamTaskLinkReviewFixture(t, failed)
			var metadata map[string]any
			require.NoError(t, common.UnmarshalJsonStr(fixture.settlement.Other, &metadata))
			delete(metadata, "is_task") // September writers retained task_id but did not emit this marker.
			other, err := common.Marshal(metadata)
			require.NoError(t, err)
			require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.settlement.Id).Update("other", string(other)).Error)
			raw, digest := upstreamTaskLinkManifestBytes(t, fixture.manifest)
			links, err := ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "legacy-accounting-schema", 0)
			require.NoError(t, err)
			require.Len(t, links, 1)
			assert.Equal(t, fixture.source.Id, links[0].SourceLogID)
			var original Log
			require.NoError(t, LOG_DB.First(&original, fixture.settlement.Id).Error)
			assert.Equal(t, string(other), original.Other, "legacy source metadata must remain unchanged")
		})
	}
}

func TestUpstreamBillingTaskLinksRejectChangedOrInsufficientEvidence(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*testing.T, *upstreamTaskLinkFixture)
	}{
		{name: "original source changed since review", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.source.Id).Update("content", "changed").Error)
		}},
		{name: "task response changed since review", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			require.NoError(t, DB.Model(&Task{}).Where("id = ?", fixture.task.ID).Update("data", []byte(`{"id":"review-provider-task","usage":{"total_tokens":124}}`)).Error)
		}},
		{name: "response id disagrees with the explicit private upstream identity", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			fixture.task.SetData(map[string]any{"id": "wrong-provider-task", "usage": map[string]any{"total_tokens": 123}})
			digest := sha256.Sum256(fixture.task.Data)
			fixture.candidate["data_sha256"] = hex.EncodeToString(digest[:])
			require.NoError(t, DB.Model(&Task{}).Where("id = ?", fixture.task.ID).Update("data", []byte(fixture.task.Data)).Error)
		}},
		{name: "public id fallback is not an explicit private upstream identity", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			fixture.task.PrivateData.UpstreamTaskID = ""
			fixture.task.SetData(map[string]any{"id": fixture.task.TaskID, "usage": map[string]any{"total_tokens": 123}})
			digest := sha256.Sum256(fixture.task.Data)
			fixture.candidate["data_sha256"] = hex.EncodeToString(digest[:])
			fixture.candidate["upstream_task_id"] = fixture.task.TaskID
			require.NoError(t, DB.Model(&Task{}).Where("id = ?", fixture.task.ID).Updates(map[string]any{"data": []byte(fixture.task.Data), "private_data": fixture.task.PrivateData}).Error)
		}},
		{name: "one second later is not an exact identity", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			fixture.candidate["submit_time"] = fixture.task.SubmitTime + 1
			require.NoError(t, DB.Model(&Task{}).Where("id = ?", fixture.task.ID).Update("submit_time", fixture.task.SubmitTime+1).Error)
		}},
		{name: "one second earlier is not an exact identity", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			fixture.candidate["submit_time"] = fixture.task.SubmitTime - 1
			require.NoError(t, DB.Model(&Task{}).Where("id = ?", fixture.task.ID).Update("submit_time", fixture.task.SubmitTime-1).Error)
		}},
		{name: "original already has a public task identity", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			var metadata map[string]any
			require.NoError(t, common.UnmarshalJsonStr(fixture.source.Other, &metadata))
			metadata["task_id"] = "another-public-task"
			other, err := common.Marshal(metadata)
			require.NoError(t, err)
			require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.source.Id).Update("other", string(other)).Error)
		}},
		{name: "billing parameter differs despite approved equal flag", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			fixture.task.PrivateData.BillingContext.ModelRatio = 6.571428
			require.NoError(t, DB.Model(&Task{}).Where("id = ?", fixture.task.ID).Update("private_data", fixture.task.PrivateData).Error)
		}},
		{name: "settlement has a different public task identity", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			var metadata map[string]any
			require.NoError(t, common.UnmarshalJsonStr(fixture.settlement.Other, &metadata))
			metadata["task_id"] = "different-task"
			other, err := common.Marshal(metadata)
			require.NoError(t, err)
			require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.settlement.Id).Update("other", string(other)).Error)
		}},
		{name: "accounting event explicitly denies task identity", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			var metadata map[string]any
			require.NoError(t, common.UnmarshalJsonStr(fixture.settlement.Other, &metadata))
			metadata["is_task"] = false
			other, err := common.Marshal(metadata)
			require.NoError(t, err)
			require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.settlement.Id).Update("other", string(other)).Error)
		}},
		{name: "null task marker is not a legacy omission", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			var metadata map[string]any
			require.NoError(t, common.UnmarshalJsonStr(fixture.settlement.Other, &metadata))
			metadata["is_task"] = nil
			other, err := common.Marshal(metadata)
			require.NoError(t, err)
			require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.settlement.Id).Update("other", string(other)).Error)
		}},
		{name: "string task marker is not a boolean", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			var metadata map[string]any
			require.NoError(t, common.UnmarshalJsonStr(fixture.settlement.Other, &metadata))
			metadata["is_task"] = "true"
			other, err := common.Marshal(metadata)
			require.NoError(t, err)
			require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.settlement.Id).Update("other", string(other)).Error)
		}},
		{name: "self closing algebra without raw preconsume anchor", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			var metadata map[string]any
			require.NoError(t, common.UnmarshalJsonStr(fixture.settlement.Other, &metadata))
			delete(metadata, "pre_consumed_quota")
			delete(metadata, "actual_quota")
			other, err := common.Marshal(metadata)
			require.NoError(t, err)
			require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.settlement.Id).Update("other", string(other)).Error)
			chain := fixture.candidate["adjustment_chain"].([]any)[0].(map[string]any)
			delete(chain, "pre_consumed_quota")
			delete(chain, "actual_quota")
		}},
		{name: "raw preconsume contradicts original amount", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			var metadata map[string]any
			require.NoError(t, common.UnmarshalJsonStr(fixture.settlement.Other, &metadata))
			metadata["pre_consumed_quota"] = 101
			metadata["actual_quota"] = 41
			other, err := common.Marshal(metadata)
			require.NoError(t, err)
			require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.settlement.Id).Update("other", string(other)).Error)
			chain := fixture.candidate["adjustment_chain"].([]any)[0].(map[string]any)
			chain["pre_consumed_quota"], chain["actual_quota"] = 101, 41
		}},
		{name: "manifest omits a real settlement", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			fixture.candidate["adjustment_chain"] = []any{}
		}},
		{name: "settlement predates submission", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			before := fixture.source.CreatedAt - 1
			require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.settlement.Id).Update("created_at", before).Error)
			fixture.candidate["adjustment_chain"].([]any)[0].(map[string]any)["occurred_at"] = before
		}},
		{name: "duplicate raw task identity cannot select a row", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			duplicate := fixture.task
			duplicate.ID = 0
			require.NoError(t, DB.Create(&duplicate).Error)
		}},
		{name: "another independently anchored task makes forward matching ambiguous", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			duplicate := fixture.task
			duplicate.ID, duplicate.TaskID = 0, "other-public-task"
			duplicate.PrivateData.UpstreamTaskID = "other-provider-task"
			duplicate.SetData(map[string]any{"id": duplicate.PrivateData.UpstreamTaskID, "usage": map[string]any{"total_tokens": 123}})
			require.NoError(t, DB.Create(&duplicate).Error)
			event := fixture.settlement
			event.Id, event.RequestId = 3, "other-settlement"
			var metadata map[string]any
			require.NoError(t, common.UnmarshalJsonStr(event.Other, &metadata))
			metadata["task_id"] = duplicate.TaskID
			other, err := common.Marshal(metadata)
			require.NoError(t, err)
			event.Other = string(other)
			require.NoError(t, LOG_DB.Create(&event).Error)
		}},
		{name: "another original submission makes reverse matching ambiguous", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			duplicate := fixture.source
			duplicate.Id, duplicate.RequestId, duplicate.UpstreamRequestId = 3, "other-original", "other-upstream"
			require.NoError(t, LOG_DB.Create(&duplicate).Error)
		}},
		{name: "explicitly linked original also participates in reverse uniqueness", mutate: func(t *testing.T, fixture *upstreamTaskLinkFixture) {
			duplicate := fixture.source
			duplicate.Id, duplicate.RequestId, duplicate.UpstreamRequestId = 3, "linked-original", "linked-upstream"
			var metadata map[string]any
			require.NoError(t, common.UnmarshalJsonStr(duplicate.Other, &metadata))
			metadata["task_id"] = "already-known-task"
			other, err := common.Marshal(metadata)
			require.NoError(t, err)
			duplicate.Other = string(other)
			require.NoError(t, LOG_DB.Create(&duplicate).Error)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := upstreamTaskLinkReviewFixture(t, false)
			test.mutate(t, &fixture)
			raw, digest := upstreamTaskLinkManifestBytes(t, fixture.manifest)
			_, err := ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "review-reject", 1)
			require.Error(t, err)
			var count int64
			require.NoError(t, DB.Model(&UpstreamBillingTaskLink{}).Count(&count).Error)
			assert.Zero(t, count, "failed evidence validation must not publish any association")
		})
	}
}

func TestUpstreamBillingTaskLinksBindRawManifestAndRemainAppendOnly(t *testing.T) {
	fixture := upstreamTaskLinkReviewFixture(t, false)
	raw, digest := upstreamTaskLinkManifestBytes(t, fixture.manifest)
	_, err := ApplyUpstreamBillingTaskLinks(context.Background(), append(raw, '\n'), digest, "review-batch", 1)
	require.Error(t, err, "a semantic equivalent manifest cannot use a different raw-byte digest")
	_, err = ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "review-batch", -1)
	require.Error(t, err, "negative actors cannot create audit evidence")
	links, err := ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "review-batch", 1)
	require.NoError(t, err)
	require.Len(t, links, 1)
	repeated, err := ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "review-batch", 1)
	require.NoError(t, err)
	assert.Equal(t, links, repeated, "an identical audited batch is idempotent")
	_, err = ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "different-review-batch", 1)
	require.Error(t, err, "a second batch cannot replace the original audit provenance")
	assert.Error(t, DB.Model(&links[0]).Update("batch_id", "rewritten").Error)
	assert.Error(t, DB.Delete(&links[0]).Error)
	var stored []UpstreamBillingTaskLink
	require.NoError(t, DB.Find(&stored).Error)
	require.Len(t, stored, 1)
	assert.Equal(t, links[0], stored[0])
}

func TestUpstreamBillingTaskLinkFailureRequiresExplicitFullRefund(t *testing.T) {
	for _, partial := range []bool{false, true} {
		name := "zero task quota without refund"
		if partial {
			name = "partial refund with zero task quota"
		}
		t.Run(name, func(t *testing.T) {
			fixture := upstreamTaskLinkReviewFixture(t, true)
			if partial {
				require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.settlement.Id).Update("quota", 60).Error)
				fixture.candidate["adjustment_chain"].([]any)[0].(map[string]any)["signed_delta"] = -60
			} else {
				require.NoError(t, LOG_DB.Delete(&Log{}, fixture.settlement.Id).Error)
				fixture.candidate["adjustment_chain"] = []any{}
			}
			raw, digest := upstreamTaskLinkManifestBytes(t, fixture.manifest)
			_, err := ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "invalid-failure", 1)
			require.Error(t, err)
			var count int64
			require.NoError(t, DB.Model(&UpstreamBillingTaskLink{}).Count(&count).Error)
			assert.Zero(t, count)
		})
	}
}
