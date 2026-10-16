package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

const (
	ImageStudioMaxImages           = 4
	ImageStudioMaxPromptCharacters = 4000
)

// ImageStudioSetting controls the PC image generation studio. Generated images
// require the asset library storage policy to be enabled; their lifetime is 30 days.
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
