package assetlibrary

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	coreService "github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"gorm.io/gorm"
)

const (
	assetWebhookDeliveryLease  = time.Minute
	assetWebhookLeaseMargin    = 5 * time.Second
	assetWebhookDeliveryBatch  = 20
	assetWebhookDeliveryMaxAge = 72 * time.Hour
)

func runWebhookDeliveryPass(runnerID string) {
	for range assetWebhookDeliveryBatch {
		// Acquire the lease immediately before sending, not while earlier
		// callbacks are still using up the batch's lease time.
		now := common.GetTimestamp()
		attemptID := runnerID + "-" + common.GetRandomString(16)
		deliveries, err := model.ClaimAssetWebhookDeliveries(now, now+int64(assetWebhookDeliveryLease/time.Second), attemptID, 1)
		if err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("claim asset webhook deliveries failed: %v", err))
			return
		}
		if len(deliveries) == 0 {
			return
		}
		deliverAssetWebhook(attemptID, deliveries[0])
	}
}

func deliverAssetWebhook(runnerID string, delivery *model.AssetWebhookDelivery) {
	// Include database lookups in the lease budget. A paused worker must not
	// resume with a fresh HTTP timeout after another node has taken over.
	leaseDeadline := time.Unix(delivery.LeaseUntil, 0).Add(-assetWebhookLeaseMargin)
	if !time.Now().Before(leaseDeadline) {
		return
	}
	leaseCtx, cancelLease := context.WithDeadline(context.Background(), leaseDeadline)
	defer cancelLease()
	deadline := time.Unix(delivery.CreatedAt, 0).Add(assetWebhookDeliveryMaxAge)
	if !time.Now().Before(deadline) {
		finishWebhookDeliveryFailure(leaseCtx, runnerID, delivery, 0, errors.New("webhook delivery expired after 72 hours"), true)
		return
	}
	if delivery.EventType == WebhookEventAssetActive && delivery.AssetID > 0 {
		var asset model.MediaAsset
		err := model.DB.WithContext(leaseCtx).Select("status").First(&asset, delivery.AssetID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && asset.Status != model.AssetStatusReady) {
			if finishErr := model.SupersedeAssetWebhookDelivery(leaseCtx, delivery.ID, runnerID, "superseded by current asset status"); finishErr != nil {
				logger.LogWarn(context.Background(), fmt.Sprintf("supersede stale asset webhook delivery failed: delivery_id=%d error=%v", delivery.ID, finishErr))
			}
			return
		}
		if err != nil {
			finishWebhookDeliveryFailure(leaseCtx, runnerID, delivery, 0, err, false)
			return
		}
	}
	var endpoint model.AssetWebhookEndpoint
	if err := model.DB.WithContext(leaseCtx).First(&endpoint, delivery.EndpointID).Error; err != nil {
		finishWebhookDeliveryFailure(leaseCtx, runnerID, delivery, 0, err, errors.Is(err, gorm.ErrRecordNotFound))
		return
	}
	if endpoint.Status != model.AssetWebhookEndpointStatusEnabled {
		finishWebhookDeliveryFailure(leaseCtx, runnerID, delivery, 0, errors.New("webhook endpoint is disabled"), true)
		return
	}
	owned, err := model.OwnsAssetWebhookDeliveryLease(leaseCtx, delivery.ID, runnerID, common.GetTimestamp())
	if err != nil {
		finishWebhookDeliveryFailure(leaseCtx, runnerID, delivery, 0, err, false)
		return
	}
	if !owned {
		return
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	if requestDeadline := time.Now().Add(10 * time.Second); requestDeadline.Before(deadline) {
		deadline = requestDeadline
	}
	ctx, cancel := context.WithDeadline(leaseCtx, deadline)
	defer cancel()
	headers := map[string]string{
		"Content-Type":      "application/json",
		"User-Agent":        "New-API/1.0",
		"webhook-id":        delivery.WebhookID,
		"webhook-timestamp": timestamp,
	}
	var response *http.Response
	if system_setting.EnableWorker() {
		response, err = coreService.DoWorkerRequestWithContext(ctx, &coreService.WorkerRequest{
			URL: endpoint.URL, Key: system_setting.WorkerValidKey, Method: http.MethodPost,
			Headers: headers, Body: []byte(delivery.Payload),
		})
	} else {
		if err = coreService.ValidateSSRFProtectedFetchURL(endpoint.URL); err != nil {
			finishWebhookDeliveryFailure(leaseCtx, runnerID, delivery, 0, fmt.Errorf("webhook URL rejected: %w", err), false)
			return
		}
		var request *http.Request
		request, err = http.NewRequestWithContext(ctx, http.MethodPost, endpoint.URL, bytes.NewBufferString(delivery.Payload))
		if err == nil {
			for key, value := range headers {
				request.Header.Set(key, value)
			}
			client := *coreService.GetSSRFProtectedHTTPClient()
			client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
			response, err = client.Do(request)
		}
	}
	if err != nil {
		finishWebhookDeliveryFailure(leaseCtx, runnerID, delivery, 0, err, false)
		return
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64*1024))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		finishWebhookDeliveryFailure(leaseCtx, runnerID, delivery, response.StatusCode, fmt.Errorf("webhook endpoint returned HTTP %d", response.StatusCode), false)
		return
	}
	if err = model.FinishAssetWebhookDelivery(leaseCtx, delivery.ID, runnerID, model.AssetWebhookDeliveryStatusSucceeded, response.StatusCode, 0, ""); err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("finish asset webhook delivery failed: delivery_id=%d error=%v", delivery.ID, err))
	}
}

func finishWebhookDeliveryFailure(ctx context.Context, runnerID string, delivery *model.AssetWebhookDelivery, responseStatus int, deliveryErr error, permanent bool) {
	now := common.GetTimestamp()
	status := model.AssetWebhookDeliveryStatusFailed
	nextAttemptAt := now + int64(assetWebhookRetryDelay(delivery.Attempts)/time.Second)
	nextAttemptAt = min(nextAttemptAt, delivery.CreatedAt+int64(assetWebhookDeliveryMaxAge/time.Second))
	if permanent || now-delivery.CreatedAt >= int64(assetWebhookDeliveryMaxAge/time.Second) {
		status = model.AssetWebhookDeliveryStatusExhausted
		nextAttemptAt = 0
	}
	if err := model.FinishAssetWebhookDelivery(ctx, delivery.ID, runnerID, status, responseStatus, nextAttemptAt, truncateError(deliveryErr)); err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("record asset webhook delivery failure: delivery_id=%d error=%v", delivery.ID, err))
		return
	}
	logger.LogWarn(context.Background(), fmt.Sprintf("asset webhook delivery failed: delivery_id=%d endpoint_id=%s error=%v", delivery.ID, delivery.EndpointPublicID, deliveryErr))
}

func assetWebhookRetryDelay(attempts int) time.Duration {
	if attempts < 0 {
		attempts = 0
	}
	delay := time.Minute * time.Duration(1<<min(attempts, 10))
	if delay > 12*time.Hour {
		return 12 * time.Hour
	}
	return delay
}
