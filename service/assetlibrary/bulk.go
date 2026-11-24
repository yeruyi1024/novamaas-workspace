package assetlibrary

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/model"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"gorm.io/gorm"
)

// MaxAssetBatchSize bounds batch deletes and batch previews to one list page.
const MaxAssetBatchSize = 100

// AssetRef names one asset. Public IDs are unique per owner only, so admins
// acting across owners must say whose asset they mean.
type AssetRef struct {
	ID          string `json:"id"`
	OwnerUserID int    `json:"owner_user_id"`
}

type AssetPreviewView struct {
	URL       string `json:"url"`
	PosterURL string `json:"poster_url,omitempty"`
	ExpiresAt int64  `json:"expires_at"`
}

type AssetFilterDeleteResult struct {
	DeletedCount int   `json:"deleted_count"`
	OwnerUserIDs []int `json:"owner_user_ids"`
}

type AssetSummaryView struct {
	Total    int64            `json:"total"`
	ByType   map[string]int64 `json:"by_type"`
	ByStatus map[string]int64 `json:"by_status"`
}

// AssetRefKey is the stable map key the web client uses for an asset.
func AssetRefKey(ownerUserID int, publicID string) string {
	return fmt.Sprintf("%d:%s", ownerUserID, publicID)
}

func validateAssetRefs(refs []AssetRef) error {
	if len(refs) == 0 {
		return &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("select at least one asset")}
	}
	if len(refs) > MaxAssetBatchSize {
		return &RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("select at most %d assets at a time", MaxAssetBatchSize)}
	}
	return nil
}

// resolveAssetRefs finds referenced assets in a single batched query using the
// same ownership rules as single-asset requests: non-admins only ever see their
// own assets. When skipMissing is true, concurrently deleted or missing entries
// are skipped so a batch preview still renders the remaining items.
func resolveAssetRefs(refs []AssetRef, operatorID int, isAdmin bool, skipMissing bool) ([]*model.MediaAsset, error) {
	publicIDs := make([]string, 0, len(refs))
	idSeen := make(map[string]bool, len(refs))
	for _, ref := range refs {
		if ref.ID == "" || idSeen[ref.ID] {
			continue
		}
		idSeen[ref.ID] = true
		publicIDs = append(publicIDs, ref.ID)
	}
	var found []model.MediaAsset
	if len(publicIDs) > 0 {
		query := model.DB.Where("public_id IN ? AND status <> ?", publicIDs, model.AssetStatusDeleted)
		if !isAdmin {
			query = query.Where("owner_user_id = ?", operatorID)
		}
		if err := query.Find(&found).Error; err != nil {
			return nil, err
		}
	}
	byOwnerAndID := make(map[string]*model.MediaAsset, len(found))
	byPublicID := make(map[string][]*model.MediaAsset, len(found))
	for index := range found {
		asset := &found[index]
		byOwnerAndID[AssetRefKey(asset.OwnerUserID, asset.PublicID)] = asset
		byPublicID[asset.PublicID] = append(byPublicID[asset.PublicID], asset)
	}

	assets := make([]*model.MediaAsset, 0, len(refs))
	seen := make(map[int64]bool, len(refs))
	for _, ref := range refs {
		var asset *model.MediaAsset
		if !isAdmin {
			asset = byOwnerAndID[AssetRefKey(operatorID, ref.ID)]
		} else if ref.OwnerUserID > 0 {
			asset = byOwnerAndID[AssetRefKey(ref.OwnerUserID, ref.ID)]
		} else {
			matches := byPublicID[ref.ID]
			if len(matches) > 1 {
				if skipMissing {
					continue
				}
				return nil, model.ErrAssetAmbiguousPublicID
			}
			if len(matches) == 1 {
				asset = matches[0]
			}
		}
		if asset == nil {
			if skipMissing {
				continue
			}
			return nil, gorm.ErrRecordNotFound
		}
		if seen[asset.ID] {
			continue
		}
		seen[asset.ID] = true
		assets = append(assets, asset)
	}
	return assets, nil
}

// DeleteAssets deletes every referenced asset or none of them. Assets are
// locked in ascending ID order, matching the user purge, to avoid deadlocks.
// Returns the resolved refs that were deleted.
func DeleteAssets(refs []AssetRef, operatorID int, isAdmin bool) ([]AssetRef, error) {
	if err := validateAssetRefs(refs); err != nil {
		return nil, err
	}
	assets, err := resolveAssetRefs(refs, operatorID, isAdmin, false)
	if err != nil {
		return nil, err
	}
	otherOwners := make([]int, 0)
	otherOwnerSeen := make(map[int]bool)
	for _, asset := range assets {
		if asset.OwnerUserID != operatorID && !otherOwnerSeen[asset.OwnerUserID] {
			otherOwnerSeen[asset.OwnerUserID] = true
			otherOwners = append(otherOwners, asset.OwnerUserID)
		}
	}
	for _, ownerID := range otherOwners {
		if err := model.CheckUserAssetPurgePermission(model.DB, ownerID, operatorID); err != nil {
			return nil, err
		}
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].ID < assets[j].ID })
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		for _, ownerID := range otherOwners {
			if err := model.CheckUserAssetPurgePermission(tx, ownerID, operatorID); err != nil {
				return err
			}
		}
		for _, asset := range assets {
			if err := model.RequestAssetDeletion(tx, asset.ID); err != nil {
				return err
			}
		}
		return nil
	}, model.AssetDeletionTxOptions())
	if err != nil {
		return nil, err
	}
	deleted := make([]AssetRef, 0, len(assets))
	for _, asset := range assets {
		deleted = append(deleted, AssetRef{ID: asset.PublicID, OwnerUserID: asset.OwnerUserID})
	}
	return deleted, nil
}

// DeleteAssetsByFilter deletes every active asset matching the list filter
// scope minus any explicitly excluded references, using bounded 100-asset
// transactions. Cross-owner deletions enforce the same role hierarchy as user
// purges so an admin cannot clear a root or peer admin's assets.
func DeleteAssetsByFilter(input AssetListInput, exclude []AssetRef, operatorID int, isAdmin bool) (*AssetFilterDeleteResult, error) {
	if len(exclude) > MaxAssetBatchSize*5 {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("too many excluded assets")}
	}
	input.OwnerUserID = operatorID
	if !isAdmin {
		input.IncludeAllOwners = false
		input.TargetOwnerUserID = 0
	}
	var excludedIDs []int64
	if len(exclude) > 0 {
		excludedAssets, err := resolveAssetRefs(exclude, operatorID, isAdmin, true)
		if err != nil {
			return nil, err
		}
		excludedIDs = make([]int64, 0, len(excludedAssets))
		for _, asset := range excludedAssets {
			excludedIDs = append(excludedIDs, asset.ID)
		}
	}
	baseQuery, err := assetScopeQuery(input)
	if err != nil {
		return nil, err
	}
	if len(excludedIDs) > 0 {
		baseQuery = baseQuery.Where("id NOT IN ?", excludedIDs)
	}
	var candidateOwners []int
	if err := baseQuery.Session(&gorm.Session{}).Distinct("owner_user_id").Pluck("owner_user_id", &candidateOwners).Error; err != nil {
		return nil, err
	}
	for _, ownerID := range candidateOwners {
		if err := model.CheckUserAssetPurgePermission(model.DB, ownerID, operatorID); err != nil {
			return nil, err
		}
	}

	result := &AssetFilterDeleteResult{OwnerUserIDs: make([]int, 0, len(candidateOwners))}
	ownerSeen := make(map[int]bool, len(candidateOwners))
	lastID := int64(0)
	for {
		batchScanned := 0
		batchDeleted := 0
		batchOwners := make([]int, 0)
		err := model.DB.Transaction(func(tx *gorm.DB) error {
			var batch []model.MediaAsset
			txQuery, err := assetScopeQueryOnDB(tx, input)
			if err != nil {
				return err
			}
			if lastID > 0 {
				txQuery = txQuery.Where("id > ?", lastID)
			}
			if len(excludedIDs) > 0 {
				txQuery = txQuery.Where("id NOT IN ?", excludedIDs)
			}
			if err := txQuery.Select("id", "owner_user_id").Order("id asc").Limit(MaxAssetBatchSize).Find(&batch).Error; err != nil {
				return err
			}
			batchScanned = len(batch)
			if batchScanned == 0 {
				return nil
			}
			lastID = batch[batchScanned-1].ID
			checked := make(map[int]bool)
			for _, asset := range batch {
				if !checked[asset.OwnerUserID] {
					if err := model.CheckUserAssetPurgePermission(tx, asset.OwnerUserID, operatorID); err != nil {
						return err
					}
					checked[asset.OwnerUserID] = true
				}
			}
			for _, asset := range batch {
				deleted, err := model.RequestAssetDeletionIfStatusIn(tx, asset.ID, input.Statuses)
				if err != nil {
					return err
				}
				if deleted {
					batchDeleted++
					batchOwners = append(batchOwners, asset.OwnerUserID)
				}
			}
			return nil
		}, model.AssetDeletionTxOptions())
		if err != nil {
			return result, err
		}
		result.DeletedCount += batchDeleted
		for _, ownerID := range batchOwners {
			if !ownerSeen[ownerID] {
				ownerSeen[ownerID] = true
				result.OwnerUserIDs = append(result.OwnerUserIDs, ownerID)
			}
		}
		if batchScanned < MaxAssetBatchSize {
			break
		}
	}
	return result, nil
}

// PreviewThumbnails signs thumbnail URLs for one page of assets in a single
// request. Assets without an uploaded file or concurrently removed are left out.
func PreviewThumbnails(ctx context.Context, refs []AssetRef, operatorID int, isAdmin bool) (map[string]AssetPreviewView, error) {
	if err := validateAssetRefs(refs); err != nil {
		return nil, err
	}
	assets, err := resolveAssetRefs(refs, operatorID, isAdmin, true)
	if err != nil {
		return nil, err
	}
	previews := make(map[string]AssetPreviewView, len(assets))
	if len(assets) == 0 {
		return previews, nil
	}
	policy, err := storageService.GetAssetLibraryPolicy()
	if err != nil {
		return nil, err
	}
	objectIDs := make([]int64, 0, len(assets))
	for _, asset := range assets {
		objectIDs = append(objectIDs, asset.StorageObjectID)
	}
	ttl := time.Duration(policy.SignedURLTTLSeconds) * time.Second
	signed, err := storageService.PresignAssetThumbnailsWithExpiry(ctx, objectIDs, ttl)
	if err != nil {
		return nil, err
	}
	for _, asset := range assets {
		if item, ok := signed[asset.StorageObjectID]; ok {
			previews[AssetRefKey(asset.OwnerUserID, asset.PublicID)] = AssetPreviewView{
				URL:       item.URL,
				PosterURL: item.PosterURL,
				ExpiresAt: item.ExpiresAt,
			}
		}
	}
	return previews, nil
}

// SummarizeAssets counts assets by type and status within the list scope,
// ignoring the type and status filters so every tab can show its own count.
func SummarizeAssets(input AssetListInput) (*AssetSummaryView, error) {
	input.AssetTypes, input.Statuses = nil, nil
	query, err := assetScopeQuery(input)
	if err != nil {
		return nil, err
	}
	var rows []struct {
		AssetType string
		Status    string
		Count     int64
	}
	if err := query.Select("asset_type, status, COUNT(*) AS count").Group("asset_type, status").Scan(&rows).Error; err != nil {
		return nil, err
	}
	summary := &AssetSummaryView{
		ByType: map[string]int64{model.AssetTypeImage: 0, model.AssetTypeVideo: 0, model.AssetTypeAudio: 0},
		ByStatus: map[string]int64{
			model.AssetStatusReady: 0, model.AssetStatusProcessing: 0, model.AssetStatusUnavailable: 0,
		},
	}
	for _, row := range rows {
		summary.Total += row.Count
		summary.ByType[row.AssetType] += row.Count
		status := row.Status
		if status == model.AssetStatusSyncFailed {
			status = model.AssetStatusUnavailable
		}
		summary.ByStatus[status] += row.Count
	}
	return summary, nil
}

// CountGroupAssets returns the number of active assets in each group.
func CountGroupAssets(groupIDs []int64) (map[int64]int64, error) {
	counts := make(map[int64]int64, len(groupIDs))
	if len(groupIDs) == 0 {
		return counts, nil
	}
	var rows []struct {
		GroupID int64
		Count   int64
	}
	if err := model.DB.Model(&model.MediaAsset{}).Select("group_id, COUNT(*) AS count").
		Where("group_id IN ? AND status <> ?", groupIDs, model.AssetStatusDeleted).
		Group("group_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.GroupID] = row.Count
	}
	return counts, nil
}

type AssetOwnerView struct {
	UserID     int    `json:"user_id"`
	Username   string `json:"username"`
	GroupCount int64  `json:"group_count"`
	AssetCount int64  `json:"asset_count"`
}

type AssetOwnerListView struct {
	Items    []AssetOwnerView `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// ListAssetOwners lists the users who own asset groups, most active first,
// so administrators can narrow the library to one user before a group.
func ListAssetOwners(search string, page, pageSize int) (*AssetOwnerListView, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > MaxAssetBatchSize {
		pageSize = 20
	}
	owners := model.DB.Model(&model.AssetGroup{}).
		Select("owner_user_id, COUNT(*) AS group_count").
		Where("status <> ?", model.AssetStatusDeleted).
		Group("owner_user_id")
	if keyword := strings.ToLower(strings.TrimSpace(search)); keyword != "" {
		keyword = strings.ReplaceAll(keyword, "!", "!!")
		keyword = strings.ReplaceAll(keyword, "%", "!%")
		keyword = strings.ReplaceAll(keyword, "_", "!_")
		pattern := "%" + keyword + "%"
		users := model.DB.Model(&model.User{}).Select("id").
			Where("LOWER(username) LIKE ? ESCAPE '!' OR LOWER(display_name) LIKE ? ESCAPE '!'", pattern, pattern)
		if id, err := strconv.Atoi(keyword); err == nil {
			owners = owners.Where("owner_user_id IN (?) OR owner_user_id = ?", users, id)
		} else {
			owners = owners.Where("owner_user_id IN (?)", users)
		}
	}
	owners = owners.Session(&gorm.Session{})
	var total int64
	if err := model.DB.Table("(?) AS asset_owners", owners).Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []struct {
		OwnerUserID int
		GroupCount  int64
	}
	if err := owners.Order("group_count DESC").Order("owner_user_id ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).Scan(&rows).Error; err != nil {
		return nil, err
	}
	ownerIDs := make([]int, 0, len(rows))
	for _, row := range rows {
		ownerIDs = append(ownerIDs, row.OwnerUserID)
	}
	names, err := model.GetUsernamesByIDs(ownerIDs)
	if err != nil {
		return nil, err
	}
	assetCounts := make(map[int]int64, len(rows))
	if len(ownerIDs) > 0 {
		var counts []struct {
			OwnerUserID int
			Count       int64
		}
		if err := model.DB.Model(&model.MediaAsset{}).Select("owner_user_id, COUNT(*) AS count").
			Where("owner_user_id IN ? AND status <> ?", ownerIDs, model.AssetStatusDeleted).
			Group("owner_user_id").Scan(&counts).Error; err != nil {
			return nil, err
		}
		for _, count := range counts {
			assetCounts[count.OwnerUserID] = count.Count
		}
	}
	items := make([]AssetOwnerView, 0, len(rows))
	for _, row := range rows {
		items = append(items, AssetOwnerView{
			UserID: row.OwnerUserID, Username: names[row.OwnerUserID],
			GroupCount: row.GroupCount, AssetCount: assetCounts[row.OwnerUserID],
		})
	}
	return &AssetOwnerListView{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}
