package authz

const (
	ResourceImageStudio = "image_studio"
	ActionContentRead   = "content_read"
)

var (
	ImageStudioRead        = Permission{Resource: ResourceImageStudio, Action: ActionRead}
	ImageStudioContentRead = Permission{Resource: ResourceImageStudio, Action: ActionContentRead}
)

func init() {
	RegisterResource(ResourceDefinition{
		Resource: ResourceImageStudio,
		LabelKey: "Image review",
		Actions: []ActionDefinition{
			{
				Action:         ActionRead,
				LabelKey:       "Read AI studio metadata",
				DescriptionKey: "View image generation records without prompts or generated content.",
			},
			{
				Action:         ActionContentRead,
				LabelKey:       "Read AI studio content",
				DescriptionKey: "View prompts, reference assets, and generated images across users.",
			},
		},
	})
}
