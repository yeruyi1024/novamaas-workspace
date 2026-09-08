package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"gorm.io/gorm"
)

const (
	minSignedURLTTLSeconds = 60 * 60
	maxSignedURLTTLSeconds = 7 * 24 * 60 * 60
	maxRetentionSeconds    = 30 * 24 * 60 * 60
	maxStorageFiles        = 100
	maxStorageBytes        = 512 * 1024 * 1024
)

type ProfileInput struct {
	Name            string `json:"name"`
	ProviderType    string `json:"provider_type"`
	Status          int    `json:"status"`
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	AuthType        string `json:"auth_type"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	SecurityToken   string `json:"security_token"`
}

type ProfileView struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	ProviderType         string `json:"provider_type"`
	Status               int    `json:"status"`
	Endpoint             string `json:"endpoint"`
	Region               string `json:"region"`
	Bucket               string `json:"bucket"`
	AuthType             string `json:"auth_type"`
	CredentialConfigured bool   `json:"credential_configured"`
	AccessKeyHint        string `json:"access_key_hint"`
	CreatedAt            int64  `json:"created_at"`
	UpdatedAt            int64  `json:"updated_at"`
}

type PolicyInput struct {
	StorageProfileID    int    `json:"storage_profile_id"`
	ObjectPrefix        string `json:"object_prefix"`
	SignedURLTTLSeconds int64  `json:"signed_url_ttl_seconds"`
	RetentionSeconds    int64  `json:"retention_seconds"`
	MaxFileBytes        int64  `json:"max_file_bytes"`
	MaxTotalBytes       int64  `json:"max_total_bytes"`
	MaxFiles            int    `json:"max_files"`
	AllowedMIMETypes    string `json:"allowed_mime_types"`
	Enabled             bool   `json:"enabled"`
}

func ListProfiles() ([]ProfileView, error) {
	profiles, err := model.ListStorageProfiles()
	if err != nil {
		return nil, err
	}
	views := make([]ProfileView, 0, len(profiles))
	for _, profile := range profiles {
		credential, credentialErr := model.GetActiveStorageCredential(profile.ID)
		if credentialErr != nil {
			return nil, credentialErr
		}
		views = append(views, profileView(profile, credential))
	}
	return views, nil
}

func GetProfile(id int) (*ProfileView, error) {
	profile, err := model.GetStorageProfileByID(id)
	if err != nil || profile == nil || profile.Status == model.StorageProfileStatusArchived {
		return nil, err
	}
	credential, err := model.GetActiveStorageCredential(id)
	if err != nil {
		return nil, err
	}
	view := profileView(profile, credential)
	return &view, nil
}

func SaveProfile(id int, input ProfileInput) (*ProfileView, error) {
	profile := &model.StorageProfile{
		ID:           id,
		Name:         strings.TrimSpace(input.Name),
		ProviderType: strings.TrimSpace(input.ProviderType),
		Status:       input.Status,
		Endpoint:     strings.TrimRight(strings.TrimSpace(input.Endpoint), "/"),
		Region:       strings.TrimSpace(input.Region),
		Bucket:       strings.TrimSpace(input.Bucket),
		Config:       "{}",
	}
	if err := validateProfileInput(profile, input.AuthType); err != nil {
		return nil, err
	}

	var savedCredential *model.StorageCredential
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var existing model.StorageProfile
		if id > 0 {
			if err := tx.First(&existing, id).Error; err != nil {
				return err
			}
			if existing.Status == model.StorageProfileStatusArchived {
				return errors.New("storage profile is archived")
			}
			hasObjects, err := storageProfileHasObjects(tx, id)
			if err != nil {
				return err
			}
			if hasObjects && (existing.ProviderType != profile.ProviderType || existing.Endpoint != profile.Endpoint || existing.Region != profile.Region || existing.Bucket != profile.Bucket) {
				return errors.New("a storage profile with objects cannot change provider, endpoint, region, or bucket; create a new profile instead")
			}
			profile.CreatedAt = existing.CreatedAt
			profile.UpdatedAt = common.GetTimestamp()
			if err = tx.Model(&model.StorageProfile{}).Where("id = ?", id).Updates(map[string]any{
				"name":          profile.Name,
				"provider_type": profile.ProviderType,
				"status":        profile.Status,
				"endpoint":      profile.Endpoint,
				"region":        profile.Region,
				"bucket":        profile.Bucket,
				"config":        profile.Config,
				"updated_at":    profile.UpdatedAt,
			}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(profile).Error; err != nil {
				return err
			}
		}

		var active model.StorageCredential
		activeErr := tx.Where("storage_profile_id = ? AND status = ?", profile.ID, model.StorageCredentialStatusActive).
			Order("version desc").First(&active).Error
		if activeErr != nil && !errors.Is(activeErr, gorm.ErrRecordNotFound) {
			return activeErr
		}
		credential, changed, err := buildStorageCredential(profile.ID, input, &active, activeErr == nil)
		if err != nil {
			return err
		}
		if !changed {
			savedCredential = &active
			return nil
		}
		if activeErr == nil {
			if err = tx.Model(&model.StorageCredential{}).Where("id = ?", active.ID).Updates(map[string]any{
				"status":     model.StorageCredentialStatusInactive,
				"updated_at": common.GetTimestamp(),
			}).Error; err != nil {
				return err
			}
		}
		if err = tx.Create(credential).Error; err != nil {
			return err
		}
		savedCredential = credential
		return nil
	})
	if err != nil {
		return nil, err
	}
	view := profileView(profile, savedCredential)
	return &view, nil
}

func ArchiveProfile(id int) error {
	profile, err := model.GetStorageProfileByID(id)
	if err != nil {
		return err
	}
	if profile == nil || profile.Status == model.StorageProfileStatusArchived {
		return gorm.ErrRecordNotFound
	}
	referenced, err := model.StorageProfileIsReferenced(id)
	if err != nil {
		return err
	}
	if referenced {
		return errors.New("storage profile is still referenced by a storage policy")
	}
	hasObjects, err := model.StorageProfileHasObjects(id)
	if err != nil {
		return err
	}
	if hasObjects {
		return errors.New("storage profile still has objects pending retention or deletion")
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.StorageProfile{}).Where("id = ?", id).Updates(map[string]any{
			"status":     model.StorageProfileStatusArchived,
			"updated_at": common.GetTimestamp(),
		}).Error; err != nil {
			return err
		}
		return tx.Model(&model.StorageCredential{}).Where("storage_profile_id = ?", id).Updates(map[string]any{
			"status":     model.StorageCredentialStatusInactive,
			"updated_at": common.GetTimestamp(),
		}).Error
	})
}

func TestProfile(ctx context.Context, input ProfileInput) error {
	profile := &model.StorageProfile{
		Name:         strings.TrimSpace(input.Name),
		ProviderType: strings.TrimSpace(input.ProviderType),
		Status:       model.StorageProfileStatusEnabled,
		Endpoint:     strings.TrimRight(strings.TrimSpace(input.Endpoint), "/"),
		Region:       strings.TrimSpace(input.Region),
		Bucket:       strings.TrimSpace(input.Bucket),
		Config:       "{}",
	}
	if err := validateProfileInput(profile, input.AuthType); err != nil {
		return err
	}
	credential, _, err := buildStorageCredential(0, input, nil, false)
	if err != nil {
		return err
	}
	driver, err := driverForProfile(profile, credential)
	if err != nil {
		return err
	}
	return testObjectDriver(ctx, driver)
}

func TestSavedProfile(ctx context.Context, id int) error {
	profile, err := model.GetStorageProfileByID(id)
	if err != nil {
		return err
	}
	if profile == nil || profile.Status == model.StorageProfileStatusArchived {
		return gorm.ErrRecordNotFound
	}
	credential, err := model.GetActiveStorageCredential(id)
	if err != nil {
		return err
	}
	driver, err := driverForProfile(profile, credential)
	if err != nil {
		return err
	}
	return testObjectDriver(ctx, driver)
}

func GetRelayMediaPolicy() (*model.StoragePolicy, error) {
	policy, err := model.GetStoragePolicyByKey(model.StoragePolicyRelayMediaTemp)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return model.DefaultRelayMediaStoragePolicy(), nil
	}
	return policy, nil
}

func SaveRelayMediaPolicy(input PolicyInput) (*model.StoragePolicy, error) {
	policy := &model.StoragePolicy{
		Key:                 model.StoragePolicyRelayMediaTemp,
		Name:                "Relay media temporary storage",
		Purpose:             model.StorageObjectPurposeRelayMediaTemp,
		StorageProfileID:    input.StorageProfileID,
		ObjectPrefix:        strings.Trim(strings.TrimSpace(input.ObjectPrefix), "/"),
		SignedURLTTLSeconds: input.SignedURLTTLSeconds,
		RetentionSeconds:    input.RetentionSeconds,
		MaxFileBytes:        input.MaxFileBytes,
		MaxTotalBytes:       input.MaxTotalBytes,
		MaxFiles:            input.MaxFiles,
		AllowedMIMETypes:    normalizeMIMETypes(input.AllowedMIMETypes),
		Enabled:             input.Enabled,
	}
	if err := validatePolicy(policy); err != nil {
		return nil, err
	}
	if err := model.SaveStoragePolicy(policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func validateProfileInput(profile *model.StorageProfile, authType string) error {
	if profile.Name == "" || len(profile.Name) > 128 {
		return errors.New("storage profile name is required and must not exceed 128 characters")
	}
	if profile.Status != model.StorageProfileStatusDisabled && profile.Status != model.StorageProfileStatusEnabled {
		return errors.New("invalid storage profile status")
	}
	if authType != model.StorageCredentialAuthStatic && authType != model.StorageCredentialAuthEnvironment {
		return errors.New("invalid storage credential type")
	}
	switch profile.ProviderType {
	case model.StorageProviderAliyunOSS:
		return validateAliyunProfile(profile)
	case model.StorageProviderTencentCOS, model.StorageProviderS3:
		return errors.New("the selected storage provider is reserved but not implemented")
	default:
		return errors.New("invalid storage provider type")
	}
}

func buildStorageCredential(profileID int, input ProfileInput, active *model.StorageCredential, hasActive bool) (*model.StorageCredential, bool, error) {
	authType := strings.TrimSpace(input.AuthType)
	accessKeyID := strings.TrimSpace(input.AccessKeyID)
	accessKeySecret := strings.TrimSpace(input.AccessKeySecret)
	securityToken := strings.TrimSpace(input.SecurityToken)
	if hasActive && active.AuthType == authType && accessKeyID == "" && accessKeySecret == "" && securityToken == "" {
		return active, false, nil
	}
	version := 1
	if hasActive {
		version = active.Version + 1
	}
	credential := &model.StorageCredential{
		StorageProfileID: profileID,
		Version:          version,
		AuthType:         authType,
		KeyVersion:       credentialEncryptionVersion,
		Status:           model.StorageCredentialStatusActive,
	}
	if authType == model.StorageCredentialAuthEnvironment {
		return credential, true, nil
	}
	if accessKeyID == "" || accessKeySecret == "" {
		return nil, false, errors.New("access key ID and secret are required for static credentials")
	}
	payload, err := encryptCredential(CredentialSecret{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		SecurityToken:   securityToken,
	})
	if err != nil {
		return nil, false, err
	}
	credential.EncryptedPayload = payload
	credential.AccessKeyHint = accessKeyHint(accessKeyID)
	return credential, true, nil
}

func validatePolicy(policy *model.StoragePolicy) error {
	if policy.StorageProfileID <= 0 {
		if policy.Enabled {
			return errors.New("an enabled storage policy requires a storage profile")
		}
	} else {
		profile, err := model.GetStorageProfileByID(policy.StorageProfileID)
		if err != nil {
			return err
		}
		if profile == nil || profile.Status == model.StorageProfileStatusArchived {
			return errors.New("storage profile does not exist")
		}
		if policy.Enabled && profile.Status != model.StorageProfileStatusEnabled {
			return errors.New("storage profile must be enabled before enabling the policy")
		}
	}
	if policy.ObjectPrefix == "" || strings.Contains(policy.ObjectPrefix, "..") || strings.ContainsAny(policy.ObjectPrefix, "\\\x00") {
		return errors.New("invalid storage object prefix")
	}
	if policy.SignedURLTTLSeconds < minSignedURLTTLSeconds || policy.SignedURLTTLSeconds > maxSignedURLTTLSeconds {
		return fmt.Errorf("signed URL TTL must be between %d and %d seconds", minSignedURLTTLSeconds, maxSignedURLTTLSeconds)
	}
	if policy.RetentionSeconds < policy.SignedURLTTLSeconds || policy.RetentionSeconds > maxRetentionSeconds {
		return errors.New("retention must be at least the signed URL TTL and no more than 30 days")
	}
	if policy.MaxFileBytes <= 0 || policy.MaxFileBytes > maxStorageBytes {
		return errors.New("invalid per-file size limit")
	}
	if policy.MaxTotalBytes < policy.MaxFileBytes || policy.MaxTotalBytes > maxStorageBytes {
		return errors.New("total size limit must be at least the per-file limit and no more than 512 MiB")
	}
	if policy.MaxFiles <= 0 || policy.MaxFiles > maxStorageFiles {
		return errors.New("file count limit must be between 1 and 100")
	}
	if policy.AllowedMIMETypes == "" {
		return errors.New("at least one allowed MIME type is required")
	}
	for _, mimeType := range strings.Split(policy.AllowedMIMETypes, ",") {
		if _, supported := mediaExtension(mimeType); !supported {
			return fmt.Errorf("unsupported storage policy MIME type: %s", mimeType)
		}
	}
	return nil
}

func normalizeMIMETypes(value string) string {
	seen := map[string]struct{}{}
	ordered := make([]string, 0)
	for _, item := range strings.Split(value, ",") {
		mimeType := strings.ToLower(strings.TrimSpace(item))
		if mimeType == "" {
			continue
		}
		if _, exists := seen[mimeType]; exists {
			continue
		}
		seen[mimeType] = struct{}{}
		ordered = append(ordered, mimeType)
	}
	return strings.Join(ordered, ",")
}

func profileView(profile *model.StorageProfile, credential *model.StorageCredential) ProfileView {
	view := ProfileView{
		ID:           profile.ID,
		Name:         profile.Name,
		ProviderType: profile.ProviderType,
		Status:       profile.Status,
		Endpoint:     profile.Endpoint,
		Region:       profile.Region,
		Bucket:       profile.Bucket,
		CreatedAt:    profile.CreatedAt,
		UpdatedAt:    profile.UpdatedAt,
	}
	if credential != nil {
		view.AuthType = credential.AuthType
		view.CredentialConfigured = credential.AuthType == model.StorageCredentialAuthEnvironment || credential.EncryptedPayload != ""
		view.AccessKeyHint = credential.AccessKeyHint
	}
	return view
}

func storageProfileHasObjects(tx *gorm.DB, profileID int) (bool, error) {
	var count int64
	err := tx.Model(&model.StorageObject{}).
		Where("storage_profile_id = ? AND status <> ?", profileID, model.StorageObjectStatusDeleted).
		Count(&count).Error
	return count > 0, err
}

func ResolveEnabledPolicy(key string) (*model.StoragePolicy, *model.StorageProfile, *model.StorageCredential, error) {
	policy, err := model.GetStoragePolicyByKey(key)
	if err != nil {
		return nil, nil, nil, err
	}
	if policy == nil || !policy.Enabled {
		return nil, nil, nil, errors.New("storage policy is not enabled")
	}
	profile, err := model.GetStorageProfileByID(policy.StorageProfileID)
	if err != nil {
		return nil, nil, nil, err
	}
	if profile == nil || profile.Status != model.StorageProfileStatusEnabled {
		return nil, nil, nil, errors.New("storage profile is not enabled")
	}
	credential, err := model.GetActiveStorageCredential(profile.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	if credential == nil {
		return nil, nil, nil, errors.New("storage credential is not configured")
	}
	return policy, profile, credential, nil
}

func policySignedURLTTL(policy *model.StoragePolicy) time.Duration {
	return time.Duration(policy.SignedURLTTLSeconds) * time.Second
}
