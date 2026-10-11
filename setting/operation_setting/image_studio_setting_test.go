package operation_setting

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStudioSwitchesRejectInvalidOptionValues(t *testing.T) {
	for _, field := range []string{"enabled", "visible", "asset_library_enabled"} {
		t.Run(field, func(t *testing.T) {
			key := "image_studio_setting." + field
			require.Error(t, ValidateImageStudioOption(key, "sometimes"))
			require.NoError(t, ValidateImageStudioOption(key, "true"))
			require.NoError(t, ValidateImageStudioOption(key, "false"))
		})
	}
}
