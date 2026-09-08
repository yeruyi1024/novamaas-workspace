package doubao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedance20AliasUsesPresetVideoPricing(t *testing.T) {
	assert.Contains(t, ModelList, "doubao-seedance-2-0")

	tests := []struct {
		name       string
		resolution string
		hasVideo   bool
		wantRatio  float64
	}{
		{name: "720p text or image input", resolution: "720p", wantRatio: 1},
		{name: "1080p text or image input", resolution: "1080p", wantRatio: 51.0 / 46.0},
		{name: "1080p video input", resolution: "1080p", hasVideo: true, wantRatio: 31.0 / 46.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratio, ok := GetVideoInputRatio("doubao-seedance-2-0", tt.resolution, tt.hasVideo)

			require.True(t, ok)
			assert.InDelta(t, tt.wantRatio, ratio, 1e-12)
		})
	}
}
