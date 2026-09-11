package constant

type TaskPlatform string

const (
	TaskPlatformSuno       TaskPlatform = "suno"
	TaskPlatformMidjourney              = "mj"
)

// MaxVideoTaskRequestBodyBytes bounds each independently persisted request
// snapshot. Keeping the limit in constant avoids coupling provider adaptors to
// the model package, which imports relay/common for task utilities.
const MaxVideoTaskRequestBodyBytes = 2 * 1024 * 1024

const (
	SunoActionMusic  = "MUSIC"
	SunoActionLyrics = "LYRICS"

	TaskActionGenerate          = "generate"
	TaskActionTextGenerate      = "textGenerate"
	TaskActionFirstTailGenerate = "firstTailGenerate"
	TaskActionReferenceGenerate = "referenceGenerate"
	TaskActionRemix             = "remixGenerate"
)

var SunoModel2Action = map[string]string{
	"suno_music":  SunoActionMusic,
	"suno_lyrics": SunoActionLyrics,
}

func ShouldStoreVideoTaskRequestBody(channelType int) bool {
	switch channelType {
	case ChannelTypeAli, ChannelTypeDoubaoVideo, ChannelTypeVolcNative:
		return true
	default:
		return false
	}
}
