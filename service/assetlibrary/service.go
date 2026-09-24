package assetlibrary

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	storageService "github.com/QuantumNous/new-api/service/storage"

	"gorm.io/gorm"
)

type RequestError struct {
	StatusCode int
	Code       string
	Err        error
}

func (err *RequestError) Error() string {
	if err == nil || err.Err == nil {
		return "asset library request failed"
	}
	return err.Err.Error()
}

func (err *RequestError) Unwrap() error { return err.Err }

func (err *RequestError) HTTPStatusCode() int {
	if err == nil || err.StatusCode == 0 {
		return http.StatusInternalServerError
	}
	return err.StatusCode
}

func (err *RequestError) ErrorCode() string {
	if err == nil {
		return ""
	}
	return err.Code
}

type GroupInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GroupView struct {
	ID          string `json:"id"`
	OwnerUserID int    `json:"owner_user_id"`
	OwnerName   string `json:"owner_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type AssetInput struct {
	GroupPublicID string
	Name          string
	AssetType     string
	FileName      string
	ContentType   string
	Size          int64
	Body          io.Reader
}

type AssetView struct {
	ID                string `json:"id"`
	GroupID           string `json:"group_id"`
	GroupName         string `json:"group_name"`
	OwnerUserID       int    `json:"owner_user_id"`
	OwnerName         string `json:"owner_name"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	ContentType       string `json:"content_type"`
	Size              int64  `json:"size"`
	SHA256            string `json:"sha256"`
	Status            string `json:"status"`
	UnavailableReason string `json:"unavailable_reason,omitempty"`
	CreatedAt         int64  `json:"created_at"`
	UpdatedAt         int64  `json:"updated_at"`
}

type AssetListInput struct {
	OwnerUserID      int
	IncludeAllOwners bool
	GroupPublicID    string
	GroupPublicIDs   []string
	Statuses         []string
	Search           string
	Page             int
	PageSize         int
	SortBy           string
	SortOrder        string
}

type AssetListView struct {
	Items    []AssetView `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

type ChannelConfigInput struct {
	Enabled     bool   `json:"enabled"`
	Protocol    string `json:"protocol"`
	AuthType    string `json:"auth_type"`
	BaseURL     string `json:"base_url"`
	Region      string `json:"region"`
	Service     string `json:"service"`
	APIVersion  string `json:"api_version"`
	ProjectName string `json:"project_name"`
	QPM         int    `json:"qpm"`
	AccessKeyID string `json:"access_key_id"`
	Credential  string `json:"credential"`
}

type ChannelConfigView struct {
	ChannelID            int    `json:"channel_id"`
	ChannelName          string `json:"channel_name"`
	ChannelType          int    `json:"channel_type"`
	Enabled              bool   `json:"enabled"`
	Protocol             string `json:"protocol"`
	AuthType             string `json:"auth_type"`
	BaseURL              string `json:"base_url"`
	Region               string `json:"region"`
	Service              string `json:"service"`
	APIVersion           string `json:"api_version"`
	ProjectName          string `json:"project_name"`
	QPM                  int    `json:"qpm"`
	AccessKeyHint        string `json:"access_key_hint"`
	CredentialConfigured bool   `json:"credential_configured"`
	UpdatedAt            int64  `json:"updated_at"`
}

type SyncJobView struct {
	ID           int64  `json:"id"`
	AssetID      string `json:"asset_id"`
	AssetName    string `json:"asset_name"`
	AssetType    string `json:"asset_type"`
	GroupID      string `json:"group_id"`
	GroupName    string `json:"group_name"`
	OwnerUserID  int    `json:"owner_user_id"`
	OwnerName    string `json:"owner_name"`
	ChannelID    int    `json:"channel_id"`
	ChannelName  string `json:"channel_name"`
	Operation    string `json:"operation"`
	Status       string `json:"status"`
	Progress     int    `json:"progress"`
	Attempts     int    `json:"attempts"`
	LastError    string `json:"last_error,omitempty"`
	LastSyncedAt int64  `json:"last_synced_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

type SyncJobSummary struct {
	Total      int64 `json:"total"`
	Pending    int64 `json:"pending"`
	Processing int64 `json:"processing"`
	Active     int64 `json:"active"`
	Failed     int64 `json:"failed"`
}

type SyncJobListView struct {
	Items    []SyncJobView  `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Summary  SyncJobSummary `json:"summary"`
}

func CreateGroup(ownerUserID int, input GroupInput) (*model.AssetGroup, error) {
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)
	if ownerUserID <= 0 {
		return nil, &RequestError{StatusCode: http.StatusUnauthorized, Err: errors.New("authentication is required")}
	}
	if name == "" || len([]rune(name)) > 64 {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("asset group name must contain 1 to 64 characters")}
	}
	if len([]rune(description)) > 300 {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("asset group description must not exceed 300 characters")}
	}
	publicID, err := newPublicID("group-")
	if err != nil {
		return nil, err
	}
	group := &model.AssetGroup{PublicID: publicID, OwnerUserID: ownerUserID, Name: name, Description: description, Status: model.AssetStatusReady}
	if err = model.DB.Create(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

func UpdateGroup(publicID string, ownerUserID int, isAdmin bool, input GroupInput) (*model.AssetGroup, error) {
	group, err := model.FindAssetGroup(publicID, ownerUserID, isAdmin)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)
	if name == "" || len([]rune(name)) > 64 || len([]rune(description)) > 300 {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset group name or description")}
	}
	group.Name = name
	group.Description = description
	group.UpdatedAt = common.GetTimestamp()
	if err = model.DB.Model(group).Updates(map[string]any{"name": name, "description": description, "updated_at": group.UpdatedAt}).Error; err != nil {
		return nil, err
	}
	return group, nil
}

func DeleteGroup(publicID string, ownerUserID int, isAdmin bool) error {
	group, err := model.FindAssetGroup(publicID, ownerUserID, isAdmin)
	if err != nil {
		return err
	}
	var count int64
	if err = model.DB.Model(&model.MediaAsset{}).Where("group_id = ? AND status <> ?", group.ID, model.AssetStatusDeleted).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return &RequestError{StatusCode: http.StatusConflict, Err: errors.New("delete all assets in the group before deleting the group")}
	}
	return model.DB.Model(group).Updates(map[string]any{"status": model.AssetStatusDeleted, "updated_at": common.GetTimestamp()}).Error
}

func ListGroups(ownerUserID int, includeAllOwners bool) ([]GroupView, error) {
	query := model.DB.Where("status <> ?", model.AssetStatusDeleted)
	if !includeAllOwners {
		query = query.Where("owner_user_id = ?", ownerUserID)
	}
	var groups []model.AssetGroup
	if err := query.Order("created_at desc").Order("id desc").Find(&groups).Error; err != nil {
		return nil, err
	}
	ownerIDs := make([]int, 0, len(groups))
	for _, group := range groups {
		ownerIDs = append(ownerIDs, group.OwnerUserID)
	}
	ownerNames, err := model.GetUsernamesByIDs(ownerIDs)
	if err != nil {
		return nil, err
	}
	views := make([]GroupView, 0, len(groups))
	for _, group := range groups {
		views = append(views, GroupView{
			ID: group.PublicID, OwnerUserID: group.OwnerUserID, OwnerName: ownerNames[group.OwnerUserID],
			Name: group.Name, Description: group.Description, Status: group.Status,
			CreatedAt: group.CreatedAt, UpdatedAt: group.UpdatedAt,
		})
	}
	return views, nil
}

func CreateAsset(ctx context.Context, ownerUserID int, input AssetInput) (*AssetView, error) {
	group, err := model.FindAssetGroup(input.GroupPublicID, ownerUserID, false)
	if err != nil {
		return nil, err
	}
	assetType := strings.ToLower(strings.TrimSpace(input.AssetType))
	if assetType != model.AssetTypeImage && assetType != model.AssetTypeVideo && assetType != model.AssetTypeAudio {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("asset type must be image, video, or audio")}
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = strings.TrimSpace(input.FileName)
	}
	if name == "" || len([]rune(name)) > 255 {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("asset name must contain 1 to 255 characters")}
	}
	if !strings.HasPrefix(strings.ToLower(input.ContentType), assetType+"/") {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("asset type does not match the uploaded file")}
	}
	policy, err := storageService.GetAssetLibraryPolicy()
	if err != nil {
		return nil, err
	}
	var count int64
	if err = model.DB.Model(&model.MediaAsset{}).Where("group_id = ? AND status <> ?", group.ID, model.AssetStatusDeleted).Count(&count).Error; err != nil {
		return nil, err
	}
	if count >= int64(policy.MaxFiles) {
		return nil, &RequestError{StatusCode: http.StatusConflict, Err: fmt.Errorf("asset group already contains the maximum of %d assets", policy.MaxFiles)}
	}
	publicID, err := newPublicID("asset-")
	if err != nil {
		return nil, err
	}
	object, err := storageService.UploadAssetObject(ctx, ownerUserID, group.PublicID, publicID, input.FileName, input.ContentType, input.Size, input.Body)
	if err != nil {
		return nil, err
	}
	asset := &model.MediaAsset{
		PublicID:        publicID,
		OwnerUserID:     ownerUserID,
		GroupID:         group.ID,
		StorageObjectID: object.ID,
		Name:            name,
		AssetType:       assetType,
		ContentType:     object.ContentType,
		Size:            object.Size,
		SHA256:          object.SHA256,
		Status:          model.AssetStatusReady,
	}
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(asset).Error; err != nil {
			return err
		}
		channelIDs, err := enabledChannelIDs(tx)
		if err != nil {
			return err
		}
		return model.QueueAssetReplicas(tx, asset.ID, channelIDs)
	})
	if err != nil {
		_ = storageService.DeleteAssetObject(object.ID, "asset metadata transaction failed")
		return nil, err
	}
	views := assetViews([]model.MediaAsset{*asset}, map[int64]string{group.ID: group.PublicID})
	views[0].GroupName = group.Name
	return &views[0], nil
}

func ListAssets(input AssetListInput) (*AssetListView, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 40
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}
	query := model.DB.Model(&model.MediaAsset{}).Where("status <> ?", model.AssetStatusDeleted)
	if len(input.Statuses) > 0 {
		query = query.Where("status IN ?", input.Statuses)
	}
	if !input.IncludeAllOwners {
		query = query.Where("owner_user_id = ?", input.OwnerUserID)
	}
	if input.GroupPublicID != "" {
		group, err := model.FindAssetGroup(input.GroupPublicID, input.OwnerUserID, input.IncludeAllOwners)
		if err != nil {
			return nil, err
		}
		query = query.Where("group_id = ?", group.ID)
	}
	if len(input.GroupPublicIDs) > 0 {
		groupQuery := model.DB.Model(&model.AssetGroup{}).
			Select("id").
			Where("public_id IN ? AND status <> ?", input.GroupPublicIDs, model.AssetStatusDeleted)
		if !input.IncludeAllOwners {
			groupQuery = groupQuery.Where("owner_user_id = ?", input.OwnerUserID)
		}
		query = query.Where("group_id IN (?)", groupQuery)
	}
	if keyword := strings.ToLower(strings.TrimSpace(input.Search)); keyword != "" {
		keyword = strings.ReplaceAll(keyword, "!", "!!")
		keyword = strings.ReplaceAll(keyword, "%", "!%")
		keyword = strings.ReplaceAll(keyword, "_", "!_")
		pattern := "%" + keyword + "%"
		ownerIDs := model.DB.Model(&model.User{}).
			Select("id").
			Where("LOWER(username) LIKE ? ESCAPE '!' OR LOWER(display_name) LIKE ? ESCAPE '!'", pattern, pattern)
		query = query.Where(
			"LOWER(name) LIKE ? ESCAPE '!' OR LOWER(public_id) LIKE ? ESCAPE '!' OR LOWER(content_type) LIKE ? ESCAPE '!' OR owner_user_id IN (?)",
			pattern, pattern, pattern, ownerIDs,
		)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var assets []model.MediaAsset
	orderColumn := "created_at"
	if strings.EqualFold(input.SortBy, "UpdateTime") {
		orderColumn = "updated_at"
	} else if strings.EqualFold(input.SortBy, "GroupId") {
		orderColumn = "group_id"
	}
	orderDirection := "desc"
	if strings.EqualFold(input.SortOrder, "Asc") {
		orderDirection = "asc"
	}
	if err := query.Order(orderColumn + " " + orderDirection).Order("id desc").
		Limit(input.PageSize).Offset((input.Page - 1) * input.PageSize).Find(&assets).Error; err != nil {
		return nil, err
	}
	groupIDs := make([]int64, 0, len(assets))
	ownerIDs := make([]int, 0, len(assets))
	for _, asset := range assets {
		groupIDs = append(groupIDs, asset.GroupID)
		ownerIDs = append(ownerIDs, asset.OwnerUserID)
	}
	groups := make(map[int64]string)
	groupNames := make(map[int64]string)
	if len(groupIDs) > 0 {
		var values []model.AssetGroup
		if err := model.DB.Where("id IN ?", groupIDs).Find(&values).Error; err != nil {
			return nil, err
		}
		for _, group := range values {
			groups[group.ID] = group.PublicID
			groupNames[group.ID] = group.Name
		}
	}
	ownerNames, err := model.GetUsernamesByIDs(ownerIDs)
	if err != nil {
		return nil, err
	}
	views := assetViews(assets, groups, ownerNames)
	for index := range views {
		views[index].GroupName = groupNames[assets[index].GroupID]
	}
	return &AssetListView{Items: views, Total: total, Page: input.Page, PageSize: input.PageSize}, nil
}

func PreviewURL(ctx context.Context, publicID string, ownerUserID int, isAdmin bool) (string, int64, error) {
	return PreviewURLVariant(ctx, publicID, ownerUserID, isAdmin, "original")
}

func PreviewURLVariant(ctx context.Context, publicID string, ownerUserID int, isAdmin bool, variant string) (string, int64, error) {
	mode := storageService.AssetObjectURLOriginal
	switch variant {
	case "original":
	case "thumbnail":
		mode = storageService.AssetObjectURLThumbnail
	case "download":
		mode = storageService.AssetObjectURLDownload
	default:
		return "", 0, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset preview variant")}
	}
	asset, err := model.FindMediaAsset(publicID, ownerUserID, isAdmin)
	if err != nil {
		return "", 0, err
	}
	policy, err := storageService.GetAssetLibraryPolicy()
	if err != nil {
		return "", 0, err
	}
	ttl := time.Duration(policy.SignedURLTTLSeconds) * time.Second
	value, err := storageService.PresignAssetObject(ctx, asset.StorageObjectID, ttl, mode, asset.PublicID)
	return value, common.GetTimestamp() + policy.SignedURLTTLSeconds, err
}

func DeleteAsset(publicID string, ownerUserID int, isAdmin bool) error {
	asset, err := model.FindMediaAsset(publicID, ownerUserID, isAdmin)
	if err != nil {
		return err
	}
	now := common.GetTimestamp()
	return model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.MediaAsset{}).Where("id = ?", asset.ID).Updates(map[string]any{
			"status": model.AssetStatusDeleted, "updated_at": now,
		}).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.AssetReplica{}).Where("asset_id = ? AND upstream_asset_id <> '' AND status <> ?", asset.ID, model.AssetReplicaStatusDeleted).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return tx.Model(&model.StorageObject{}).
				Where("id = ? AND status <> ?", asset.StorageObjectID, model.StorageObjectStatusDeleted).
				Updates(map[string]any{
					"status":       model.StorageObjectStatusDeletePending,
					"delete_after": now,
					"last_error":   "asset deleted",
					"updated_at":   now,
				}).Error
		}
		return tx.Model(&model.AssetReplica{}).Where("asset_id = ? AND status <> ?", asset.ID, model.AssetReplicaStatusDeleted).Updates(map[string]any{
			"operation":    model.AssetReplicaOperationDelete,
			"status":       model.AssetReplicaStatusDeleting,
			"next_sync_at": now,
			"last_error":   "",
			"updated_at":   now,
		}).Error
	})
}

func ListChannelConfigs() ([]ChannelConfigView, error) {
	var channels []model.Channel
	if err := model.DB.Where("type IN ?", []int{constant.ChannelTypeDoubaoVideo, constant.ChannelTypeVolcNative}).Order("id asc").Find(&channels).Error; err != nil {
		return nil, err
	}
	var configs []model.AssetChannelConfig
	if err := model.DB.Find(&configs).Error; err != nil {
		return nil, err
	}
	byChannel := make(map[int]model.AssetChannelConfig, len(configs))
	for _, config := range configs {
		byChannel[config.ChannelID] = config
	}
	views := make([]ChannelConfigView, 0, len(channels))
	for _, channel := range channels {
		config, ok := byChannel[channel.Id]
		if !ok {
			config = model.AssetChannelConfig{ChannelID: channel.Id, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthAKSK, BaseURL: DefaultBaseURL, Region: DefaultRegion, Service: DefaultService, APIVersion: DefaultAPIVersion, QPM: 60}
		}
		views = append(views, channelConfigView(channel, config))
	}
	return views, nil
}

func SaveChannelConfig(channelID int, input ChannelConfigInput) (*ChannelConfigView, error) {
	channel, err := eligibleChannel(channelID)
	if err != nil {
		return nil, err
	}
	var existing model.AssetChannelConfig
	hasExisting := model.DB.Where("channel_id = ?", channelID).First(&existing).Error == nil
	config := model.AssetChannelConfig{
		ChannelID:   channelID,
		Enabled:     input.Enabled,
		Protocol:    normalizeProtocol(input.Protocol),
		AuthType:    strings.TrimSpace(input.AuthType),
		BaseURL:     strings.TrimSpace(input.BaseURL),
		Region:      strings.TrimSpace(input.Region),
		Service:     strings.TrimSpace(input.Service),
		APIVersion:  strings.TrimSpace(input.APIVersion),
		ProjectName: strings.TrimSpace(input.ProjectName),
		QPM:         input.QPM,
		AccessKeyID: strings.TrimSpace(input.AccessKeyID),
	}
	applyProtocolDefaults(&config)
	if config.QPM == 0 {
		config.QPM = 60
	}
	credential := strings.TrimSpace(input.Credential)
	if credential == "" && hasExisting && normalizeProtocol(existing.Protocol) == config.Protocol && existing.AuthType == config.AuthType {
		config.EncryptedCredential = existing.EncryptedCredential
		config.CredentialKeyVersion = existing.CredentialKeyVersion
		config.AccessKeyHint = existing.AccessKeyHint
		if config.AccessKeyID == "" {
			config.AccessKeyID = existing.AccessKeyID
		}
	} else if credential != "" {
		config.EncryptedCredential, err = encryptChannelCredential(credential)
		if err != nil {
			return nil, &RequestError{StatusCode: http.StatusServiceUnavailable, Err: err}
		}
		config.CredentialKeyVersion = channelCredentialVersion
		if config.AuthType == model.AssetChannelAuthBearer {
			config.AccessKeyHint = credentialHint(credential)
		} else {
			config.AccessKeyHint = credentialHint(config.AccessKeyID)
		}
	}
	if err = validateChannelConfig(&config); err != nil {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: err}
	}
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if hasExisting {
			config.ID = existing.ID
			config.CreatedAt = existing.CreatedAt
			config.UpdatedAt = common.GetTimestamp()
			if err := tx.Save(&config).Error; err != nil {
				return err
			}
		} else if err := tx.Create(&config).Error; err != nil {
			return err
		}
		if config.Enabled {
			var assetIDs []int64
			if err := tx.Model(&model.MediaAsset{}).Where("status = ?", model.AssetStatusReady).Pluck("id", &assetIDs).Error; err != nil {
				return err
			}
			for _, assetID := range assetIDs {
				if err := model.QueueAssetReplicas(tx, assetID, []int{channelID}); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	view := channelConfigView(*channel, config)
	return &view, nil
}

func TestChannelConfig(ctx context.Context, channelID int, input *ChannelConfigInput) error {
	var config *model.AssetChannelConfig
	if input != nil {
		view, err := buildUnsavedConfig(channelID, *input)
		if err != nil {
			return err
		}
		config = view
	} else {
		if _, err := eligibleChannel(channelID); err != nil {
			return err
		}
		var stored model.AssetChannelConfig
		if err := model.DB.Where("channel_id = ?", channelID).First(&stored).Error; err != nil {
			return err
		}
		config = &stored
	}
	client, err := newProviderClient(config)
	if err != nil {
		return err
	}
	return client.test(withRequestLogContext(ctx, "test", 0, 0))
}

func QueueChannelSync(channelID int) error {
	config := model.AssetChannelConfig{}
	if err := model.DB.Where("channel_id = ? AND enabled = ?", channelID, true).First(&config).Error; err != nil {
		return err
	}
	var assetIDs []int64
	if err := model.DB.Model(&model.MediaAsset{}).Where("status = ?", model.AssetStatusReady).Pluck("id", &assetIDs).Error; err != nil {
		return err
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		for _, assetID := range assetIDs {
			if err := model.QueueAssetReplicas(tx, assetID, []int{channelID}); err != nil {
				return err
			}
		}
		return nil
	})
}

func ListSyncJobs(page int, pageSize int, channelID int, status string) (*SyncJobListView, error) {
	status = strings.TrimSpace(status)
	if status != "" && !validReplicaStatus(status) {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid synchronization status")}
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	listQuery := model.DB.Model(&model.AssetReplica{})
	if channelID > 0 {
		listQuery = listQuery.Where("channel_id = ?", channelID)
	}
	if status == "" {
		listQuery = listQuery.Where("status <> ?", model.AssetReplicaStatusDeleted)
	} else {
		listQuery = listQuery.Where("status = ?", status)
	}
	var total int64
	if err := listQuery.Count(&total).Error; err != nil {
		return nil, err
	}
	var replicas []model.AssetReplica
	if err := listQuery.Order("updated_at desc").Order("id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&replicas).Error; err != nil {
		return nil, err
	}

	summaryQuery := model.DB.Model(&model.AssetReplica{}).Where("status <> ?", model.AssetReplicaStatusDeleted)
	if channelID > 0 {
		summaryQuery = summaryQuery.Where("channel_id = ?", channelID)
	}
	var statusCounts []struct {
		Status string
		Count  int64
	}
	if err := summaryQuery.Select("status, count(*) as count").Group("status").Find(&statusCounts).Error; err != nil {
		return nil, err
	}
	summary := SyncJobSummary{}
	for _, value := range statusCounts {
		summary.Total += value.Count
		switch value.Status {
		case model.AssetReplicaStatusPending:
			summary.Pending += value.Count
		case model.AssetReplicaStatusSyncing, model.AssetReplicaStatusProcessing, model.AssetReplicaStatusDeleting:
			summary.Processing += value.Count
		case model.AssetReplicaStatusActive:
			summary.Active += value.Count
		case model.AssetReplicaStatusFailed, model.AssetReplicaStatusRejected:
			summary.Failed += value.Count
		}
	}

	assetIDs := make([]int64, 0, len(replicas))
	channelIDs := make([]int, 0, len(replicas))
	for _, replica := range replicas {
		assetIDs = append(assetIDs, replica.AssetID)
		channelIDs = append(channelIDs, replica.ChannelID)
	}
	assets := make(map[int64]model.MediaAsset, len(assetIDs))
	groupIDs := make([]int64, 0, len(assetIDs))
	ownerIDs := make([]int, 0, len(assetIDs))
	if len(assetIDs) > 0 {
		var values []model.MediaAsset
		if err := model.DB.Where("id IN ?", assetIDs).Find(&values).Error; err != nil {
			return nil, err
		}
		for _, asset := range values {
			assets[asset.ID] = asset
			groupIDs = append(groupIDs, asset.GroupID)
			ownerIDs = append(ownerIDs, asset.OwnerUserID)
		}
	}
	groups := make(map[int64]model.AssetGroup, len(groupIDs))
	if len(groupIDs) > 0 {
		var values []model.AssetGroup
		if err := model.DB.Where("id IN ?", groupIDs).Find(&values).Error; err != nil {
			return nil, err
		}
		for _, group := range values {
			groups[group.ID] = group
		}
	}
	owners := make(map[int]string, len(ownerIDs))
	if len(ownerIDs) > 0 {
		var values []model.User
		if err := model.DB.Select("id", "username", "display_name").Where("id IN ?", ownerIDs).Find(&values).Error; err != nil {
			return nil, err
		}
		for _, owner := range values {
			owners[owner.Id] = owner.DisplayName
			if owners[owner.Id] == "" {
				owners[owner.Id] = owner.Username
			}
		}
	}
	channels := make(map[int]string, len(channelIDs))
	if len(channelIDs) > 0 {
		var values []model.Channel
		if err := model.DB.Select("id", "name").Where("id IN ?", channelIDs).Find(&values).Error; err != nil {
			return nil, err
		}
		for _, channel := range values {
			channels[channel.Id] = channel.Name
		}
	}

	items := make([]SyncJobView, 0, len(replicas))
	for _, replica := range replicas {
		asset := assets[replica.AssetID]
		group := groups[asset.GroupID]
		items = append(items, SyncJobView{
			ID: replica.ID, AssetID: asset.PublicID, AssetName: asset.Name, AssetType: asset.AssetType,
			GroupID: group.PublicID, GroupName: group.Name, OwnerUserID: asset.OwnerUserID, OwnerName: owners[asset.OwnerUserID],
			ChannelID: replica.ChannelID, ChannelName: channels[replica.ChannelID], Operation: replica.Operation,
			Status: replica.Status, Progress: replica.Progress, Attempts: replica.Attempts, LastError: replica.LastError,
			LastSyncedAt: replica.LastSyncedAt, UpdatedAt: replica.UpdatedAt,
		})
	}
	return &SyncJobListView{Items: items, Total: total, Page: page, PageSize: pageSize, Summary: summary}, nil
}

func RetrySyncJob(replicaID int64) error {
	readyAssets := model.DB.Model(&model.MediaAsset{}).Select("id").Where("status = ?", model.AssetStatusReady)
	result := model.DB.Model(&model.AssetReplica{}).
		Where("id = ? AND status = ? AND asset_id IN (?)", replicaID, model.AssetReplicaStatusFailed, readyAssets).
		Updates(map[string]any{
			"status":       model.AssetReplicaStatusPending,
			"attempts":     0,
			"next_sync_at": common.GetTimestamp(),
			"locked_by":    "",
			"lease_until":  0,
			"last_error":   "",
			"updated_at":   common.GetTimestamp(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return &RequestError{StatusCode: http.StatusConflict, Err: errors.New("only failed synchronization jobs can be retried")}
	}
	return nil
}

func validReplicaStatus(status string) bool {
	switch status {
	case model.AssetReplicaStatusPending, model.AssetReplicaStatusSyncing, model.AssetReplicaStatusProcessing,
		model.AssetReplicaStatusActive, model.AssetReplicaStatusFailed, model.AssetReplicaStatusRejected, model.AssetReplicaStatusDeleting,
		model.AssetReplicaStatusDeleted:
		return true
	default:
		return false
	}
}

func newPublicID(prefix string) (string, error) {
	value, err := common.GenerateRandomCharsKey(5)
	if err != nil {
		return "", err
	}
	return prefix + time.Now().UTC().Format("20060102150405") + "-" + strings.ToLower(value), nil
}

func assetViews(assets []model.MediaAsset, groups map[int64]string, ownerNames ...map[int]string) []AssetView {
	names := map[int]string{}
	if len(ownerNames) > 0 && ownerNames[0] != nil {
		names = ownerNames[0]
	}
	views := make([]AssetView, 0, len(assets))
	for _, asset := range assets {
		views = append(views, AssetView{
			ID: asset.PublicID, GroupID: groups[asset.GroupID], OwnerUserID: asset.OwnerUserID, OwnerName: names[asset.OwnerUserID], Name: asset.Name, Type: asset.AssetType,
			ContentType: asset.ContentType, Size: asset.Size, SHA256: asset.SHA256, Status: asset.Status,
			UnavailableReason: asset.UnavailableReason,
			CreatedAt:         asset.CreatedAt, UpdatedAt: asset.UpdatedAt,
		})
	}
	return views
}

func enabledChannelIDs(tx *gorm.DB) ([]int, error) {
	var ids []int
	err := tx.Model(&model.AssetChannelConfig{}).Where("enabled = ?", true).Pluck("channel_id", &ids).Error
	return ids, err
}

func eligibleChannel(channelID int) (*model.Channel, error) {
	channel, err := model.GetChannelById(channelID, true)
	if err != nil {
		return nil, err
	}
	if channel == nil || (channel.Type != constant.ChannelTypeDoubaoVideo && channel.Type != constant.ChannelTypeVolcNative) {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("only DoubaoVideo and Volc Native channels support asset-library mapping")}
	}
	return channel, nil
}

func normalizeProtocol(protocol string) string {
	protocol = strings.TrimSpace(protocol)
	if protocol == "" {
		return model.AssetChannelProtocolVolcAction
	}
	return protocol
}

func applyProtocolDefaults(config *model.AssetChannelConfig) {
	config.Protocol = normalizeProtocol(config.Protocol)
	if config.Protocol == model.AssetChannelProtocolYoufangREST {
		if config.AuthType == "" {
			config.AuthType = model.AssetChannelAuthBearer
		}
		if config.BaseURL == "" {
			config.BaseURL = DefaultYoufangBaseURL
		}
		return
	}
	if config.AuthType == "" {
		config.AuthType = model.AssetChannelAuthAKSK
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}
	if config.Region == "" {
		config.Region = DefaultRegion
	}
	if config.Service == "" {
		config.Service = DefaultService
	}
	if config.APIVersion == "" {
		config.APIVersion = DefaultAPIVersion
	}
}

func validateChannelConfig(config *model.AssetChannelConfig) error {
	config.Protocol = normalizeProtocol(config.Protocol)
	switch config.Protocol {
	case model.AssetChannelProtocolVolcAction:
		if config.AuthType != model.AssetChannelAuthAKSK && config.AuthType != model.AssetChannelAuthBearer {
			return errors.New("Volcengine Action authentication must be ak_sk or bearer")
		}
		if config.AuthType == model.AssetChannelAuthAKSK && strings.TrimSpace(config.AccessKeyID) == "" {
			return errors.New("access key ID is required for AK/SK authentication")
		}
		if config.Region == "" || config.Service == "" || config.APIVersion == "" {
			return errors.New("Volcengine Action region, service, and API version are required")
		}
	case model.AssetChannelProtocolYoufangREST:
		if config.AuthType != model.AssetChannelAuthBearer {
			return errors.New("YooFang REST requires bearer authentication with the supplied sk key")
		}
	default:
		return errors.New("unsupported asset provider protocol")
	}
	if strings.TrimSpace(config.EncryptedCredential) == "" {
		return errors.New("asset channel credential is required")
	}
	endpoint, err := url.Parse(config.BaseURL)
	if err != nil || endpoint.Hostname() == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return errors.New("asset channel base URL is invalid")
	}
	if endpoint.Scheme != "https" && !(endpoint.Scheme == "http" && isLoopbackHost(endpoint.Hostname())) {
		return errors.New("asset channel base URL must use HTTPS; HTTP is allowed only for loopback testing")
	}
	if config.QPM < 1 || config.QPM > 1000 {
		return errors.New("asset channel QPM must be between 1 and 1000")
	}
	return nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func channelConfigView(channel model.Channel, config model.AssetChannelConfig) ChannelConfigView {
	return ChannelConfigView{
		ChannelID: channel.Id, ChannelName: channel.Name, ChannelType: channel.Type, Enabled: config.Enabled,
		Protocol: normalizeProtocol(config.Protocol), AuthType: config.AuthType, BaseURL: config.BaseURL, Region: config.Region, Service: config.Service,
		APIVersion: config.APIVersion, ProjectName: config.ProjectName, AccessKeyHint: config.AccessKeyHint,
		QPM:                  config.QPM,
		CredentialConfigured: config.EncryptedCredential != "", UpdatedAt: config.UpdatedAt,
	}
}

func buildUnsavedConfig(channelID int, input ChannelConfigInput) (*model.AssetChannelConfig, error) {
	if _, err := eligibleChannel(channelID); err != nil {
		return nil, err
	}
	config := &model.AssetChannelConfig{
		ChannelID: channelID, Enabled: input.Enabled, Protocol: normalizeProtocol(input.Protocol), AuthType: strings.TrimSpace(input.AuthType),
		BaseURL: strings.TrimSpace(input.BaseURL), Region: strings.TrimSpace(input.Region),
		Service: strings.TrimSpace(input.Service), APIVersion: strings.TrimSpace(input.APIVersion),
		ProjectName: strings.TrimSpace(input.ProjectName), AccessKeyID: strings.TrimSpace(input.AccessKeyID),
		QPM: input.QPM,
	}
	applyProtocolDefaults(config)
	if config.QPM == 0 {
		config.QPM = 60
	}
	credential := strings.TrimSpace(input.Credential)
	if credential == "" {
		var stored model.AssetChannelConfig
		if err := model.DB.Where("channel_id = ?", channelID).First(&stored).Error; err != nil {
			return nil, errors.New("credential is required when testing an unsaved channel configuration")
		}
		if normalizeProtocol(stored.Protocol) != config.Protocol || stored.AuthType != config.AuthType {
			return nil, errors.New("credential is required after changing the provider protocol or authentication type")
		}
		config.EncryptedCredential = stored.EncryptedCredential
		config.CredentialKeyVersion = stored.CredentialKeyVersion
		if config.AccessKeyID == "" {
			config.AccessKeyID = stored.AccessKeyID
		}
	} else {
		encrypted, err := encryptChannelCredential(credential)
		if err != nil {
			return nil, err
		}
		config.EncryptedCredential = encrypted
		config.CredentialKeyVersion = channelCredentialVersion
	}
	if err := validateChannelConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}
