package authz

const ResourceImageStudio = "image_studio"

var (
	ImageStudioRead  = Permission{Resource: ResourceImageStudio, Action: ActionRead}
	ImageStudioWrite = Permission{Resource: ResourceImageStudio, Action: ActionWrite}
)

func init() {
	RegisterResource(ResourceDefinition{
		Resource: ResourceImageStudio,
		LabelKey: "Image review",
		Actions: []ActionDefinition{
			{
				Action:         ActionRead,
				LabelKey:       "Review generated images",
				DescriptionKey: "Review generated images across users. Content access is audited.",
			},
			{
				Action:         ActionWrite,
				LabelKey:       "Delete reviewed images",
				DescriptionKey: "Delete other users' generated images from image review. Deletions are audited.",
			},
		},
	})
}
