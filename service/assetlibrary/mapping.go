package assetlibrary

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/model"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/gorm"
)

// ResolveRequestAssetIDs translates gateway-owned asset:// identifiers only
// after channel selection. Native upstream asset identifiers are left intact
// for backwards compatibility.
func ResolveRequestAssetIDs(body []byte, ownerUserID int, channelID int) ([]byte, int, error) {
	if !gjson.ValidBytes(body) {
		return nil, 0, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("request body must be valid JSON")}
	}
	content := gjson.GetBytes(body, "content")
	if !content.IsArray() {
		return body, 0, nil
	}
	paths := make(map[string]string)
	for index := range content.Array() {
		for _, mediaType := range []string{"image_url", "video_url", "audio_url"} {
			path := fmt.Sprintf("content.%d.%s.url", index, mediaType)
			value := gjson.GetBytes(body, path)
			if !value.Exists() || value.Type != gjson.String {
				continue
			}
			publicID, ok := assetReferenceID(value.String())
			if ok {
				paths[path] = publicID
			}
		}
	}
	if len(paths) == 0 {
		return body, 0, nil
	}
	resolved := make(map[string]string)
	localAssets := make(map[string]model.MediaAsset)
	for _, publicID := range paths {
		if _, exists := resolved[publicID]; exists {
			continue
		}
		var asset model.MediaAsset
		err := model.DB.Where("public_id = ?", publicID).First(&asset).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Native upstream asset IDs remain valid. Only identifiers found
				// in the local ownership ledger participate in gateway mapping.
				continue
			}
			return nil, 0, err
		}
		if asset.OwnerUserID != ownerUserID || asset.Status == model.AssetStatusDeleted {
			return nil, 0, &RequestError{StatusCode: http.StatusNotFound, Err: fmt.Errorf("asset %s does not exist or is not accessible", publicID)}
		}
		if asset.Status == model.AssetStatusUnavailable {
			reason := "upstream rejected this material"
			switch asset.UnavailableReason {
			case model.AssetUnavailableRealPerson:
				reason = "upstream rejected real-person content"
			case model.AssetUnavailableSensitiveContent:
				reason = "upstream rejected sensitive content"
			}
			return nil, 0, &RequestError{
				StatusCode: http.StatusUnprocessableEntity, Code: "asset_unavailable",
				Err: fmt.Errorf("asset %s is unavailable: %s; upload revised material and use its new asset ID", publicID, reason),
			}
		}
		resolved[publicID] = ""
		localAssets[publicID] = asset
	}
	if len(resolved) == 0 {
		return body, 0, nil
	}
	var configCount int64
	if err := model.DB.Model(&model.AssetChannelConfig{}).Where("channel_id = ? AND enabled = ?", channelID, true).Count(&configCount).Error; err != nil {
		return nil, 0, err
	}
	if configCount != 1 {
		return nil, 0, &RequestError{StatusCode: http.StatusConflict, Err: errors.New("the selected channel has no enabled asset-library integration")}
	}
	for publicID := range resolved {
		asset := localAssets[publicID]
		var replica model.AssetReplica
		if err := model.DB.Where("asset_id = ? AND channel_id = ?", asset.ID, channelID).First(&replica).Error; err != nil {
			return nil, 0, &RequestError{StatusCode: http.StatusConflict, Err: fmt.Errorf("asset %s has not been synchronized to the selected channel", publicID)}
		}
		if replica.Status != model.AssetReplicaStatusActive || replica.UpstreamAssetID == "" {
			return nil, 0, &RequestError{StatusCode: http.StatusConflict, Err: fmt.Errorf("asset %s is not ready on the selected channel (status: %s)", publicID, replica.Status)}
		}
		resolved[publicID] = replica.UpstreamAssetID
	}
	result := append([]byte(nil), body...)
	mappedCount := 0
	for path, publicID := range paths {
		upstreamID, mapped := resolved[publicID]
		if !mapped {
			continue
		}
		var err error
		result, err = sjson.SetBytes(result, path, "asset://"+upstreamID)
		if err != nil {
			return nil, 0, fmt.Errorf("replace asset ID at %s: %w", path, err)
		}
		mappedCount++
	}
	return result, mappedCount, nil
}

func assetReferenceID(value string) (string, bool) {
	if !strings.HasPrefix(value, "asset://") {
		return "", false
	}
	publicID := strings.TrimPrefix(value, "asset://")
	return publicID, strings.HasPrefix(publicID, "asset-") || strings.HasPrefix(publicID, "asset_local_")
}
