package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

const (
	ImageStudioMaxImages           = 4
	ImageStudioMaxPromptCharacters = 4000
	// ImageStudioMaxRunningPerUser bounds simultaneous generations per user.
	ImageStudioMaxRunningPerUser = 5
)

// ImageStudioSetting controls the PC image generation studio. Generated images
// and references use independent, configurable business storage policies.
type ImageStudioSetting struct {
	Enabled bool `json:"enabled"`
}

var imageStudioSetting = ImageStudioSetting{Enabled: true}

func init() {
	config.GlobalConfig.Register("image_studio_setting", &imageStudioSetting)
}

func GetImageStudioSetting() *ImageStudioSetting {
	return &imageStudioSetting
}
