package model

import "context"

// Billing views must never serialize a full User (credentials, email, balances,
// and administrator-only settings are deliberately excluded).
type BillingUserIdentity struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

func GetBillingUserIdentities(ctx context.Context, ids []int) (map[int]BillingUserIdentity, error) {
	unique := make(map[int]bool)
	userIDs := make([]int, 0)
	for _, id := range ids {
		if id > 0 && !unique[id] {
			unique[id] = true
			userIDs = append(userIDs, id)
		}
	}
	identities := make(map[int]BillingUserIdentity, len(userIDs))
	for offset := 0; offset < len(userIDs); offset += 500 {
		var users []BillingUserIdentity
		if err := DB.WithContext(ctx).Model(&User{}).Select("id", "username", "display_name").Where("id IN ?", userIDs[offset:min(offset+500, len(userIDs))]).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			identities[user.ID] = user
		}
	}
	return identities, nil
}
