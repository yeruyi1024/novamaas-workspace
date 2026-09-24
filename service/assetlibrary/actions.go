package assetlibrary

import (
	"errors"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

type GroupListInput struct {
	OwnerUserID      int
	IncludeAllOwners bool
	GroupIDs         []string
	Search           string
	Page             int
	PageSize         int
	SortBy           string
	SortOrder        string
}

type GroupListView struct {
	Items    []model.AssetGroup `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

func GetGroup(publicID string, ownerUserID int) (*model.AssetGroup, error) {
	return model.FindAssetGroup(strings.TrimSpace(publicID), ownerUserID, false)
}

func ListGroupsPage(input GroupListInput) (*GroupListView, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}
	query := model.DB.Model(&model.AssetGroup{}).Where("status <> ?", model.AssetStatusDeleted)
	if !input.IncludeAllOwners {
		query = query.Where("owner_user_id = ?", input.OwnerUserID)
	}
	if len(input.GroupIDs) > 0 {
		query = query.Where("public_id IN ?", input.GroupIDs)
	}
	if keyword := strings.ToLower(strings.TrimSpace(input.Search)); keyword != "" {
		keyword = strings.ReplaceAll(keyword, "!", "!!")
		keyword = strings.ReplaceAll(keyword, "%", "!%")
		keyword = strings.ReplaceAll(keyword, "_", "!_")
		pattern := "%" + keyword + "%"
		query = query.Where("LOWER(name) LIKE ? ESCAPE '!'", pattern)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	orderColumn := "created_at"
	if strings.EqualFold(input.SortBy, "UpdateTime") {
		orderColumn = "updated_at"
	} else if strings.EqualFold(input.SortBy, "Name") {
		orderColumn = "LOWER(name)"
	}
	orderDirection := "desc"
	if strings.EqualFold(input.SortOrder, "Asc") {
		orderDirection = "asc"
	}
	var groups []model.AssetGroup
	secondaryOrder := "id desc"
	if orderColumn == "LOWER(name)" {
		secondaryOrder = "id asc"
	}
	if err := query.Order(orderColumn + " " + orderDirection).Order(secondaryOrder).
		Limit(input.PageSize).Offset((input.Page - 1) * input.PageSize).Find(&groups).Error; err != nil {
		return nil, err
	}
	return &GroupListView{Items: groups, Total: total, Page: input.Page, PageSize: input.PageSize}, nil
}

func GetAsset(publicID string, ownerUserID int) (*AssetView, error) {
	asset, err := model.FindMediaAsset(strings.TrimSpace(publicID), ownerUserID, false)
	if err != nil {
		return nil, err
	}
	group, err := model.FindAssetGroupByID(asset.GroupID)
	if err != nil {
		return nil, err
	}
	ownerName, err := model.GetUsernameById(asset.OwnerUserID, false)
	if err != nil {
		return nil, err
	}
	views := assetViews([]model.MediaAsset{*asset}, map[int64]string{asset.GroupID: group.PublicID}, map[int]string{asset.OwnerUserID: ownerName})
	views[0].GroupName = group.Name
	return &views[0], nil
}

func UpdateAsset(publicID string, ownerUserID int, name string) (*AssetView, error) {
	asset, err := model.FindMediaAsset(strings.TrimSpace(publicID), ownerUserID, false)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 255 {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("asset name must contain 1 to 255 characters")}
	}
	asset.Name = name
	asset.UpdatedAt = common.GetTimestamp()
	if err = model.DB.Model(asset).Updates(map[string]any{"name": name, "updated_at": asset.UpdatedAt}).Error; err != nil {
		return nil, err
	}
	return GetAsset(asset.PublicID, ownerUserID)
}
