package assetlibrary

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlogger "gorm.io/gorm/logger"
)

type LibraryChannelView struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type LibraryConfigInput struct {
	ChannelConfigInput
	Name         string `json:"name"`
	ProviderName string `json:"provider_name"`
}

type LibraryConnectionView struct {
	ChannelConfigView
	ID               int                  `json:"id"`
	Name             string               `json:"name"`
	ProviderName     string               `json:"provider_name"`
	LegacyChannelID  *int                 `json:"legacy_channel_id,omitempty"`
	ReusedExisting   bool                 `json:"reused_existing,omitempty"`
	Channels         []LibraryChannelView `json:"channels"`
	Summary          SyncJobSummary       `json:"summary"`
	Missing          int64                `json:"missing"`
	CredentialStatus string               `json:"credential_status"`
}

func libraryChannelViews(bindings []model.AssetChannelConfig, channels []model.Channel) map[int][]LibraryChannelView {
	names := make(map[int]string, len(channels))
	for _, channel := range channels {
		names[channel.Id] = channel.Name
	}
	result := make(map[int][]LibraryChannelView)
	for _, binding := range bindings {
		if binding.LibraryID > 0 {
			result[binding.LibraryID] = append(result[binding.LibraryID], LibraryChannelView{ID: binding.ChannelID, Name: names[binding.ChannelID], Enabled: binding.Enabled})
		}
	}
	return result
}

// linkedLibraryChannels lists the existing channels of each account in one
// round trip per table, however many accounts are requested.
func linkedLibraryChannels(libraryIDs []int) (map[int][]LibraryChannelView, error) {
	var bindings []model.AssetChannelConfig
	if err := linkedAssetLibraryBindings(model.DB).Select("channel_id", "library_id", "enabled").Where("library_id IN ?", libraryIDs).Order("channel_id asc").Find(&bindings).Error; err != nil {
		return nil, err
	}
	var channels []model.Channel
	if len(bindings) > 0 {
		channelIDs := make([]int, 0, len(bindings))
		for _, binding := range bindings {
			channelIDs = append(channelIDs, binding.ChannelID)
		}
		if err := model.DB.Select("id", "name").Where("id IN ?", channelIDs).Find(&channels).Error; err != nil {
			return nil, err
		}
	}
	return libraryChannelViews(bindings, channels), nil
}

func ListLibraryConnections() ([]LibraryConnectionView, error) {
	var libraries []model.AssetLibraryConnection
	if err := model.DB.Order("id asc").Find(&libraries).Error; err != nil {
		return nil, err
	}
	return buildLibraryViews(libraries)
}

func libraryConnectionView(libraryID int) (*LibraryConnectionView, error) {
	var library model.AssetLibraryConnection
	if err := model.DB.First(&library, libraryID).Error; err != nil {
		return nil, err
	}
	views, err := buildLibraryViews([]model.AssetLibraryConnection{library})
	if err != nil {
		return nil, err
	}
	return &views[0], nil
}

func buildLibraryViews(libraries []model.AssetLibraryConnection) ([]LibraryConnectionView, error) {
	result := make([]LibraryConnectionView, 0, len(libraries))
	if len(libraries) == 0 {
		return result, nil
	}
	ids := make([]int, 0, len(libraries))
	for _, library := range libraries {
		ids = append(ids, library.ID)
	}
	linked, err := linkedLibraryChannels(ids)
	if err != nil {
		return nil, err
	}
	var counts []struct {
		LibraryID int
		Status    string
		Count     int64
	}
	if err := model.DB.Model(&model.AssetReplica{}).Where("library_id IN ? AND status <> ?", ids, model.AssetReplicaStatusDeleted).
		Select("library_id, status, count(*) as count").Group("library_id, status").Find(&counts).Error; err != nil {
		return nil, err
	}
	summaries := make(map[int]SyncJobSummary, len(libraries))
	for _, count := range counts {
		summary := summaries[count.LibraryID]
		summary.Total += count.Count
		switch count.Status {
		case model.AssetReplicaStatusPending:
			summary.Pending += count.Count
		case model.AssetReplicaStatusSyncing, model.AssetReplicaStatusProcessing, model.AssetReplicaStatusDeleting:
			summary.Processing += count.Count
		case model.AssetReplicaStatusActive:
			summary.Active += count.Count
		case model.AssetReplicaStatusFailed, model.AssetReplicaStatusRejected:
			summary.Failed += count.Count
		}
		summaries[count.LibraryID] = summary
	}
	liveAssets := model.DB.Model(&model.MediaAsset{}).Select("id").Where("status IN ?", []string{model.AssetStatusReady, model.AssetStatusProcessing})
	var total int64
	if err := model.DB.Model(&model.MediaAsset{}).Where("status IN ?", []string{model.AssetStatusReady, model.AssetStatusProcessing}).Count(&total).Error; err != nil {
		return nil, err
	}
	var covered []struct {
		LibraryID int
		Count     int64
	}
	if err := model.DB.Model(&model.AssetReplica{}).Where("library_id IN ? AND asset_id IN (?) AND status <> ?", ids, liveAssets, model.AssetReplicaStatusDeleted).
		Select("library_id, count(*) as count").Group("library_id").Find(&covered).Error; err != nil {
		return nil, err
	}
	coverage := make(map[int]int64, len(covered))
	for _, count := range covered {
		coverage[count.LibraryID] = count.Count
	}
	for _, library := range libraries {
		libraryChannels := linked[library.ID]
		if libraryChannels == nil {
			libraryChannels = []LibraryChannelView{}
		}
		view := channelConfigView(model.Channel{}, library.ProviderConfig(0))
		view.LibraryName, view.LibraryEnabled = library.Name, library.Enabled
		missing := max(0, total-coverage[library.ID])
		if len(libraryChannels) == 0 {
			missing = 0
		}
		credentialStatus := "available"
		if _, err := decryptChannelCredential(library.EncryptedCredential); err != nil {
			credentialStatus = "unavailable"
		}
		result = append(result, LibraryConnectionView{
			ChannelConfigView: view, ID: library.ID, Name: library.Name, ProviderName: library.ProviderName,
			LegacyChannelID: library.LegacyChannelID, Channels: libraryChannels, Summary: summaries[library.ID],
			Missing: missing, CredentialStatus: credentialStatus,
		})
	}
	return result, nil
}

func prepareLibraryConfig(input ChannelConfigInput, stored *model.AssetLibraryConnection) (model.AssetChannelConfig, error) {
	config := model.AssetChannelConfig{
		Enabled: input.Enabled, Protocol: normalizeProtocol(input.Protocol), AuthType: strings.TrimSpace(input.AuthType),
		BaseURL: strings.TrimSpace(input.BaseURL), Region: strings.TrimSpace(input.Region), Service: strings.TrimSpace(input.Service),
		APIVersion: strings.TrimSpace(input.APIVersion), ProjectName: strings.TrimSpace(input.ProjectName), QPM: input.QPM,
		AccessKeyID: strings.TrimSpace(input.AccessKeyID),
	}
	applyProtocolDefaults(&config)
	if config.QPM == 0 {
		config.QPM = 60
	}
	credential := strings.TrimSpace(input.Credential)
	if credential == "" && stored != nil && normalizeProtocol(stored.Protocol) == config.Protocol && stored.AuthType == config.AuthType {
		config.EncryptedCredential, config.CredentialKeyVersion = stored.EncryptedCredential, stored.CredentialKeyVersion
		config.AccessKeyHint = stored.AccessKeyHint
		if config.AccessKeyID == "" {
			config.AccessKeyID = stored.AccessKeyID
		}
	} else if credential != "" {
		var err error
		config.EncryptedCredential, err = encryptChannelCredential(credential)
		if err != nil {
			return config, err
		}
		config.CredentialKeyVersion = channelCredentialVersion
	}
	if config.AuthType == model.AssetChannelAuthBearer {
		if credential != "" {
			config.AccessKeyHint = credentialHint(credential)
		}
	} else {
		config.AccessKeyHint = credentialHint(config.AccessKeyID)
	}
	if err := validateChannelConfig(&config); err != nil {
		return config, &RequestError{StatusCode: http.StatusBadRequest, Err: err}
	}
	return config, nil
}

// libraryResourceKey is an opaque digest of the exact endpoint, resource scope
// and decrypted keys. A provider label alone never establishes access to the
// same upstream resources. Plaintext keys are neither returned nor persisted.
func libraryResourceKey(config model.AssetChannelConfig) (string, error) {
	applyProtocolDefaults(&config)
	secret, err := decryptChannelCredential(config.EncryptedCredential)
	if err != nil {
		return "", err
	}
	accessKey := config.AccessKeyID
	if config.AuthType == model.AssetChannelAuthBearer {
		accessKey = ""
	}
	project := config.ProjectName
	if project == "" && config.Protocol == model.AssetChannelProtocolVolcAction {
		project = "default"
	}
	payload, err := common.Marshal([]string{config.Protocol, config.AuthType, normalizeLibraryBaseURL(config.BaseURL), config.Region, config.Service, config.APIVersion, project, accessKey, secret})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}

func normalizeLibraryBaseURL(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return strings.TrimRight(value, "/")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return parsed.String()
}

// libraryIdentityGroups groups the accounts that still serve channels by the
// upstream account they point at. An account whose credential cannot be
// decrypted is never guessed into a group and keeps working on its own. Accounts
// left without channels by an earlier merge are retired and are not grouped.
func libraryIdentityGroups() (groups map[string][]int, order []string, unreadable int, err error) {
	var libraries []model.AssetLibraryConnection
	serving := model.DB.Model(&model.AssetChannelConfig{}).Select("library_id").Where("library_id > ?", 0)
	if err = model.DB.Where("id IN (?)", serving).Order("id asc").Find(&libraries).Error; err != nil {
		return nil, nil, 0, err
	}
	groups = make(map[string][]int)
	for _, library := range libraries {
		key, keyErr := libraryResourceKey(library.ProviderConfig(0))
		if keyErr != nil {
			unreadable++
			continue
		}
		if groups[key] == nil {
			order = append(order, key)
		}
		groups[key] = append(groups[key], library.ID)
	}
	return groups, order, unreadable, nil
}

// RecordLibraryIdentities stores the upstream identity of every readable account
// so that creating an account with the same credentials reuses it. It never
// merges anything.
func RecordLibraryIdentities() error {
	groups, order, _, err := libraryIdentityGroups()
	if err != nil {
		return err
	}
	for _, key := range order {
		var holders int64
		if err := model.DB.Model(&model.AssetLibraryConnection{}).Where("resource_key = ?", key).Count(&holders).Error; err != nil {
			return err
		}
		if holders > 0 {
			continue
		}
		if err := model.DB.Model(&model.AssetLibraryConnection{}).Where("id = ? AND resource_key IS NULL", groups[key][0]).Update("resource_key", key).Error; err != nil {
			return err
		}
	}
	return nil
}

type LibraryMergeReport struct {
	MergedGroups   int `json:"merged_groups"`
	DeferredGroups int `json:"deferred_groups"`
	// UnreadableAccounts counts accounts whose stored credential cannot be
	// decrypted on this server. They are left alone, so a merge that reports
	// nothing may only mean the credentials could not be compared.
	UnreadableAccounts int `json:"unreadable_accounts"`
}

// MergeIdenticalLibraries folds accounts that point at exactly the same upstream
// account into one, so channels that were configured separately with the same
// credentials stop uploading the same image once each. The administrator starts
// it; nothing is merged automatically.
func MergeIdenticalLibraries() (*LibraryMergeReport, error) {
	groups, order, unreadable, err := libraryIdentityGroups()
	if err != nil {
		return nil, err
	}
	report := &LibraryMergeReport{UnreadableAccounts: unreadable}
	for _, key := range order {
		ids := groups[key]
		if len(ids) < 2 {
			continue
		}
		keptID, err := model.MergeAssetLibraryGroup(ids, key)
		if err != nil {
			return report, err
		}
		if keptID == 0 {
			report.DeferredGroups++
			logger.LogWarn(context.Background(), fmt.Sprintf("asset-library accounts %v share one upstream account; merge deferred while requests are in flight", ids))
			continue
		}
		report.MergedGroups++
		logger.LogInfo(context.Background(), fmt.Sprintf("asset-library accounts %v share one upstream account and were merged into %d", ids, keptID))
		backfillMissingAssets(keptID)
	}
	return report, nil
}

func reusedLibraryView(resourceKey string) (*LibraryConnectionView, error) {
	var duplicate model.AssetLibraryConnection
	result := model.DB.Select("id").Where("resource_key = ?", resourceKey).Limit(1).Find(&duplicate)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil, result.Error
	}
	view, err := libraryConnectionView(duplicate.ID)
	if view != nil {
		view.ReusedExisting = true
	}
	return view, err
}

func SaveLibraryConnection(libraryID int, input LibraryConfigInput) (*LibraryConnectionView, error) {
	name := strings.TrimSpace(input.Name)
	providerName := strings.TrimSpace(input.ProviderName)
	if providerName == "" {
		providerName = name
	}
	if len([]rune(providerName)) > 128 {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("provider name must not exceed 128 characters")}
	}
	if name == "" || len([]rune(name)) > 128 {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("asset library name must contain 1 to 128 characters")}
	}
	var existing *model.AssetLibraryConnection
	if libraryID > 0 {
		existing = &model.AssetLibraryConnection{}
		if err := model.DB.First(existing, libraryID).Error; err != nil {
			return nil, err
		}
	}
	config, err := prepareLibraryConfig(input.ChannelConfigInput, existing)
	if err != nil {
		return nil, err
	}
	// Name, QPM and the enabled flag never need the stored credential, so they
	// remain editable even when it cannot be decrypted on this server.
	if existing != nil && strings.TrimSpace(input.Credential) == "" && sameLibraryIdentityConfig(existing.ProviderConfig(0), config) {
		err := model.DB.Model(&model.AssetLibraryConnection{}).Where("id = ?", existing.ID).
			Updates(map[string]any{"name": name, "provider_name": providerName, "enabled": config.Enabled, "qpm": config.QPM, "updated_at": common.GetTimestamp()}).Error
		if err != nil {
			return nil, err
		}
		if config.Enabled && !existing.Enabled {
			backfillMissingAssets(existing.ID)
		}
		return libraryConnectionView(existing.ID)
	}
	resourceKey, err := libraryResourceKey(config)
	if err != nil {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("the stored credential cannot be read on this server; enter the credential again")}
	}
	library := model.AssetLibraryFromConfig(name, config)
	library.ProviderName, library.ResourceKey = providerName, &resourceKey
	credentialWriter := model.DB.Session(&gorm.Session{Logger: model.DB.Logger.LogMode(gormlogger.Silent)})
	if existing == nil {
		if reused, err := reusedLibraryView(resourceKey); reused != nil || err != nil {
			return reused, err
		}
		if err := credentialWriter.Create(&library).Error; err != nil {
			// A concurrent request may have created the same upstream identity.
			if reused, lookupErr := reusedLibraryView(resourceKey); reused != nil && lookupErr == nil {
				return reused, nil
			}
			return nil, err
		}
		return libraryConnectionView(library.ID)
	}
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := model.LockAssetLibraryConnection(tx, existing.ID); err != nil {
			return err
		}
		// A different identity is a different upstream namespace. Existing
		// upstream copies cannot be carried across, so the account must be new.
		if priorKey, priorErr := libraryResourceKey(existing.ProviderConfig(0)); priorErr != nil || priorKey != resourceKey {
			var owned int64
			if err := tx.Model(&model.AssetReplica{}).Where("library_id = ? AND status <> ?", existing.ID, model.AssetReplicaStatusDeleted).Count(&owned).Error; err != nil {
				return err
			}
			if owned == 0 {
				if err := tx.Model(&model.AssetGroupReplica{}).Where("library_id = ?", existing.ID).Count(&owned).Error; err != nil {
					return err
				}
			}
			if owned > 0 {
				return &RequestError{StatusCode: http.StatusConflict, Code: "library_identity_locked", Err: errors.New("this account already owns upstream assets; create a separate account for another endpoint, project or credential")}
			}
			var other int64
			if err := tx.Model(&model.AssetLibraryConnection{}).Where("resource_key = ? AND id <> ?", resourceKey, existing.ID).Count(&other).Error; err != nil {
				return err
			}
			if other > 0 {
				return &RequestError{StatusCode: http.StatusConflict, Code: "library_identity_exists", Err: errors.New("another resource account already uses these credentials; link channels to that account instead")}
			}
		}
		library.ID, library.CreatedAt, library.LegacyChannelID = existing.ID, existing.CreatedAt, existing.LegacyChannelID
		library.UpdatedAt = common.GetTimestamp()
		return tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(gormlogger.Silent)}).Save(&library).Error
	})
	if err != nil {
		return nil, err
	}
	if config.Enabled && !existing.Enabled {
		backfillMissingAssets(existing.ID)
	}
	return libraryConnectionView(existing.ID)
}

func sameLibraryIdentityConfig(first, second model.AssetChannelConfig) bool {
	for _, config := range []*model.AssetChannelConfig{&first, &second} {
		applyProtocolDefaults(config)
		config.ID = 0
		config.ChannelID = 0
		config.LibraryID = 0
		config.Enabled = false
		config.QPM = 0
		config.AccessKeyHint = ""
		config.CreatedAt = 0
		config.UpdatedAt = 0
		config.BaseURL = normalizeLibraryBaseURL(config.BaseURL)
		if config.Protocol == model.AssetChannelProtocolVolcAction && config.ProjectName == "" {
			config.ProjectName = "default"
		}
		if config.AuthType == model.AssetChannelAuthBearer {
			config.AccessKeyID = ""
		}
	}
	return first == second
}

func lockEligibleLibraryChannel(tx *gorm.DB, channelID int) (*model.Channel, error) {
	if channelID <= 0 {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid channel ID")}
	}
	channels, err := model.LockAssetLibraryChannels(tx, []int{channelID})
	if err != nil {
		return nil, err
	}
	if len(channels) != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	if channels[0].Type != constant.ChannelTypeDoubaoVideo && channels[0].Type != constant.ChannelTypeVolcNative {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("only DoubaoVideo and Volc Native channels support asset-library mapping")}
	}
	return &channels[0], nil
}

// BindChannelLibrary links one channel to a resource account, or unlinks it
// when libraryID is 0. A channel belongs to at most one account.
func BindChannelLibrary(channelID, libraryID int, enabled bool) (*ChannelConfigView, error) {
	if libraryID < 0 {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset library ID")}
	}
	var channel *model.Channel
	var library *model.AssetLibraryConnection
	alreadyActive := false
	config := model.AssetChannelConfig{ChannelID: channelID}
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		if libraryID > 0 {
			var err error
			if library, err = model.LockAssetLibraryConnection(tx, libraryID); err != nil {
				return err
			}
			config = library.ProviderConfig(channelID)
			config.Enabled = enabled
		}
		var err error
		if channel, err = lockEligibleLibraryChannel(tx, channelID); err != nil {
			return err
		}
		var current model.AssetChannelConfig
		if err := tx.Where("channel_id = ?", channelID).Limit(1).Find(&current).Error; err != nil {
			return err
		}
		alreadyActive = current.LibraryID == libraryID && current.Enabled
		if libraryID > 0 && current.LibraryID > 0 && current.LibraryID != libraryID {
			return &RequestError{StatusCode: http.StatusConflict, Err: fmt.Errorf("channel %d is already linked to resource account %d; unlink it from that account before linking another", channelID, current.LibraryID)}
		}
		binding := model.AssetChannelConfig{ChannelID: channelID, LibraryID: libraryID, Enabled: config.Enabled}
		// Provider columns are cleared so a restart never rebuilds an account
		// from a stale per-channel configuration.
		updates := map[string]any{
			"library_id": libraryID, "enabled": config.Enabled, "protocol": "", "auth_type": "", "base_url": "", "region": "", "service": "",
			"api_version": "", "project_name": "", "qpm": 0, "access_key_id": "", "access_key_hint": "", "encrypted_credential": "",
			"credential_key_version": "", "updated_at": common.GetTimestamp(),
		}
		return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "channel_id"}}, DoUpdates: clause.Assignments(updates)}).Create(&binding).Error
	})
	if err != nil {
		return nil, err
	}
	if library != nil && enabled && !alreadyActive {
		backfillMissingAssets(library.ID)
	}
	view := channelConfigView(*channel, config)
	if library != nil {
		view.LibraryName, view.LibraryEnabled = library.Name, library.Enabled
	}
	return &view, nil
}

func TestLibraryConnection(ctx context.Context, libraryID int, input *LibraryConfigInput) error {
	var library *model.AssetLibraryConnection
	if libraryID > 0 {
		library = &model.AssetLibraryConnection{}
		if err := model.DB.First(library, libraryID).Error; err != nil {
			return err
		}
	}
	var config model.AssetChannelConfig
	if input != nil {
		var err error
		config, err = prepareLibraryConfig(input.ChannelConfigInput, library)
		if err != nil {
			return err
		}
		config.LibraryID = libraryID
	} else if library != nil {
		config = library.ProviderConfig(0)
	} else {
		return &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("asset library configuration is required")}
	}
	client, err := newProviderClient(&config)
	if err != nil {
		return err
	}
	return client.test(withRequestLogContext(ctx, "test", 0, 0))
}

// QueueLibrarySync only fills missing replicas. Existing state, upstream IDs and
// failed attempts remain untouched.
func QueueLibrarySync(libraryID int) error {
	var library model.AssetLibraryConnection
	if err := model.DB.Select("id", "enabled").First(&library, libraryID).Error; err != nil {
		return err
	}
	if !library.Enabled {
		return &RequestError{StatusCode: http.StatusConflict, Err: errors.New("asset library is disabled")}
	}
	var channelIDs []int
	if err := linkedAssetLibraryBindings(model.DB).Where("library_id = ? AND enabled = ?", libraryID, true).Pluck("channel_id", &channelIDs).Error; err != nil {
		return err
	}
	if len(channelIDs) == 0 {
		return &RequestError{StatusCode: http.StatusConflict, Err: errors.New("link an enabled channel before synchronizing this asset library")}
	}
	return queueMissingAssets(channelIDs[0])
}

// queueMissingAssets queues a copy of every live asset for the account the
// channel belongs to. Copies that already exist, failed ones included, are kept.
func queueMissingAssets(channelID int) error {
	var assetIDs []int64
	if err := model.DB.Model(&model.MediaAsset{}).Where("status IN ?", []string{model.AssetStatusReady, model.AssetStatusProcessing}).Pluck("id", &assetIDs).Error; err != nil {
		return err
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		for _, assetID := range assetIDs {
			if err := model.QueueMissingAssetReplicas(tx, assetID, []int{channelID}); err != nil {
				return err
			}
		}
		return nil
	})
}

// backfillMissingAssets gives an account that has just become active the assets
// uploaded before it did, as enabling a channel did before shared accounts. It
// does nothing for a disabled account or one without an enabled channel, and a
// failure only logs: "Sync missing assets" stays available as a manual retry.
func backfillMissingAssets(libraryID int) {
	var library model.AssetLibraryConnection
	if err := model.DB.Select("id", "enabled").First(&library, libraryID).Error; err != nil || !library.Enabled {
		return
	}
	var channelIDs []int
	if err := linkedAssetLibraryBindings(model.DB).Where("library_id = ? AND enabled = ?", libraryID, true).Pluck("channel_id", &channelIDs).Error; err != nil || len(channelIDs) == 0 {
		return
	}
	if err := queueMissingAssets(channelIDs[0]); err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("backfill of existing assets to asset library %d failed: %v", libraryID, err))
	}
}

// assetSyncConfigs feeds the scheduler once per scan interval, so it reads
// scheduling fields only and never touches stored credentials.
func assetSyncConfigs() ([]model.AssetChannelConfig, error) {
	var libraries []model.AssetLibraryConnection
	if err := model.DB.Select("id", "enabled", "qpm").Find(&libraries).Error; err != nil {
		return nil, err
	}
	var linkedIDs []int
	if err := linkedAssetLibraryBindings(model.DB).Where("enabled = ?", true).Distinct().Pluck("library_id", &linkedIDs).Error; err != nil {
		return nil, err
	}
	linked := make(map[int]bool, len(linkedIDs))
	for _, id := range linkedIDs {
		linked[id] = true
	}
	configs := make([]model.AssetChannelConfig, 0, len(libraries))
	for _, library := range libraries {
		configs = append(configs, model.AssetChannelConfig{LibraryID: library.ID, QPM: library.QPM, Enabled: library.Enabled && linked[library.ID]})
	}
	// Configurations that predate shared libraries stay usable until migrated.
	var legacy []model.AssetChannelConfig
	if err := model.DB.Select("channel_id", "qpm").Where("(library_id = ? OR library_id IS NULL) AND protocol <> ?", 0, "").Find(&legacy).Error; err != nil {
		return nil, err
	}
	return append(configs, legacy...), nil
}

func linkedAssetLibraryBindings(db *gorm.DB) *gorm.DB {
	return db.Model(&model.AssetChannelConfig{}).Where("library_id > ? AND channel_id IN (?)", 0, db.Model(&model.Channel{}).Select("id"))
}
