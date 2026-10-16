package authz

import "github.com/casbin/casbin/v2"

// Can reports effective access. Financial management includes accounting views,
// even when the independent accounting switch is off. Other resources keep
// their existing override and role-baseline rules.
func Can(userID int, systemRole int, permission Permission) bool {
	if directPermissionAllowed(userID, systemRole, permission) {
		return true
	}
	return permission == FinancialAccountingView && directPermissionAllowed(userID, systemRole, FinancialManagementView)
}

// directPermissionAllowed evaluates the editable grant without implications.
// Root grants are inherent; otherwise the user's override precedes its roles.
func directPermissionAllowed(userID int, systemRole int, permission Permission) bool {
	roles := resolveSubjectRoles(userID, systemRole)
	if len(roles) == 0 {
		return false
	}
	for _, role := range roles {
		if isSuperuserRole(role) {
			return true
		}
	}
	if !isKnownPermission(permission) {
		return false
	}

	e := currentEnforcer()
	if e == nil {
		return false
	}
	if effect, ok := explicitSubjectEffect(e, UserSubject(userID), permission); ok {
		return effect == EffectAllow
	}
	for _, role := range roles {
		if roleBaselineAllows(e, role, permission) {
			return true
		}
	}
	return false
}

// RawCapabilities is the editable role baseline plus user overrides. Keeping
// implied accounting access out of this matrix prevents an editor round trip
// from turning inherited access into a lasting independent accounting grant.
func RawCapabilities(userID int, systemRole int) PermissionsMap {
	result := make(PermissionsMap, len(registry))
	for _, resource := range registry {
		actions := make(map[string]bool, len(resource.Actions))
		for _, action := range resource.Actions {
			actions[action.Action] = directPermissionAllowed(userID, systemRole, Permission{Resource: resource.Resource, Action: action.Action})
		}
		result[resource.Resource] = actions
	}
	return result
}

// Capabilities returns the full resource/action matrix the subject is allowed.
func Capabilities(userID int, systemRole int) PermissionsMap {
	result := make(PermissionsMap, len(registry))
	for _, resource := range registry {
		actions := make(map[string]bool, len(resource.Actions))
		for _, action := range resource.Actions {
			actions[action.Action] = Can(userID, systemRole, Permission{
				Resource: resource.Resource,
				Action:   action.Action,
			})
		}
		result[resource.Resource] = actions
	}
	return result
}

func roleBaselineAllows(e *casbin.SyncedEnforcer, roleKey string, permission Permission) bool {
	effect, ok := explicitSubjectEffect(e, RoleSubject(roleKey), permission)
	return ok && effect == EffectAllow
}

func explicitSubjectEffect(e *casbin.SyncedEnforcer, subject string, permission Permission) (string, bool) {
	policies, err := e.GetFilteredPolicy(0, subject, permission.Resource, permission.Action)
	if err != nil {
		return "", false
	}
	hasAllow := false
	for _, policy := range policies {
		switch policyEffect(policy) {
		case EffectDeny:
			return EffectDeny, true
		case EffectAllow:
			hasAllow = true
		}
	}
	if hasAllow {
		return EffectAllow, true
	}
	return "", false
}

func policyEffect(policy []string) string {
	if len(policy) < 4 || policy[3] == "" {
		return EffectAllow
	}
	return policy[3]
}
