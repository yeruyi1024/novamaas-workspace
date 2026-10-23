package model

import "context"

const AdminManualEnabledOptionKey = "AdminManualEnabled"

// AdminManualEnabled reads the committed policy for every request. A node's
// periodically refreshed OptionMap must not keep granting access after closure.
func AdminManualEnabled(ctx context.Context) (bool, error) {
	var option Option
	err := DB.WithContext(ctx).Where(&Option{Key: AdminManualEnabledOptionKey}).Limit(1).Find(&option).Error
	return err == nil && option.Value == "true", err
}
