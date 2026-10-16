package authz

const (
	ResourceFinancialManagement   = "financial_management"
	ActionFinancialManagementView = "view"
)

var FinancialManagementView = Permission{
	Resource: ResourceFinancialManagement,
	Action:   ActionFinancialManagementView,
}

func init() {
	RegisterResource(ResourceDefinition{
		Resource: ResourceFinancialManagement,
		LabelKey: "Financial management",
		Actions: []ActionDefinition{{
			Action:         ActionFinancialManagementView,
			LabelKey:       "Review upstream statements",
			DescriptionKey: "Review, confirm, and archive upstream statements, with access to financial accounting.",
		}},
	})
}
