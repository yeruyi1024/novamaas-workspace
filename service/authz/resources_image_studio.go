package authz

const ResourceImageStudio = "image_studio"

var ImageStudioRead = Permission{Resource: ResourceImageStudio, Action: ActionRead}

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
		},
	})
}
