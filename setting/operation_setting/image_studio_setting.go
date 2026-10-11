package operation_setting

import (
	"github.com/QuantumNous/new-api/setting/config"
	"strconv"
)

const (
	DefaultImageStudioMaxImages           = 15
	DefaultImageStudioMaxReferenceImages  = 10
	DefaultImageStudioMaxPromptCharacters = 32000
)

type ImageStudioSetting struct {
	Visible             bool `json:"visible"`
	AssetLibraryEnabled bool `json:"asset_library_enabled"`
	Enabled             bool `json:"enabled"`
}

var imageStudioSetting = ImageStudioSetting{Visible: true, AssetLibraryEnabled: true, Enabled: true}

func init()                                     { config.GlobalConfig.Register("image_studio_setting", &imageStudioSetting) }
func GetImageStudioSetting() ImageStudioSetting { return imageStudioSetting }

func ValidateImageStudioOption(key, value string) error {
	switch key {
	case "image_studio_setting.visible", "image_studio_setting.enabled", "image_studio_setting.asset_library_enabled":
		_, err := strconv.ParseBool(value)
		return err
	}
	return nil
}
