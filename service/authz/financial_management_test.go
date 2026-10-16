package authz

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestFinancialManagementDefaultsPreserveAccountingAndChannelBoundaries(t *testing.T) {
	db := newAuthzTestDB(t)
	require.NoError(t, Init(db))
	assert.False(t, Can(42, common.RoleAdminUser, FinancialManagementView))
	assert.True(t, Can(42, common.RoleAdminUser, FinancialAccountingView))
	assert.True(t, Can(1, common.RoleRootUser, FinancialManagementView))
	assert.False(t, Can(43, common.RoleCommonUser, FinancialManagementView))
	var definition *ResourceDefinition
	for _, resource := range Catalog() {
		if resource.Resource == ResourceFinancialManagement {
			copy := resource
			definition = &copy
			break
		}
	}
	require.NotNil(t, definition)
	assert.Equal(t, "Financial management", definition.LabelKey)
	require.Len(t, definition.Actions, 1)
	assert.Equal(t, ActionFinancialManagementView, definition.Actions[0].Action)
	assert.Equal(t, "Review upstream statements", definition.Actions[0].LabelKey)
	for _, role := range Roles() {
		assert.Equal(t, role.Superuser, role.Grants[ResourceFinancialManagement][ActionFinancialManagementView])
	}
	require.NoError(t, SetUserPermissions(42, PermissionsMap{
		ResourceChannel:             {ActionRead: false, ActionOperate: false, ActionWrite: false, ActionSensitiveWrite: false, ActionSecretView: false},
		ResourceFinancialAccounting: {ActionFinancialView: false},
		ResourceFinancialManagement: {ActionFinancialManagementView: true},
	}))
	assert.True(t, Can(42, common.RoleAdminUser, FinancialManagementView))
	assert.True(t, Can(42, common.RoleAdminUser, FinancialAccountingView), "management includes accounting even when the independent switch is off")
	for _, permission := range []Permission{ChannelRead, ChannelOperate, ChannelWrite, ChannelSensitiveWrite, ChannelSecretView} {
		assert.False(t, Can(42, common.RoleAdminUser, permission))
	}
	assert.False(t, Can(42, common.RoleCommonUser, FinancialManagementView), "stored grants cannot elevate a nonadministrator system role")
	assert.False(t, Can(42, common.RoleCommonUser, FinancialAccountingView))
}

func TestFinancialManagementEditorRoundTripPreservesIndependentAccountingOffState(t *testing.T) {
	db := newAuthzTestDB(t)
	require.NoError(t, Init(db))
	require.NoError(t, SetUserPermissions(42, PermissionsMap{
		ResourceFinancialAccounting: {ActionFinancialView: false},
		ResourceFinancialManagement: {ActionFinancialManagementView: true},
	}))
	raw := RawCapabilities(42, common.RoleAdminUser)
	effective := Capabilities(42, common.RoleAdminUser)
	assert.False(t, raw[ResourceFinancialAccounting][ActionFinancialView])
	assert.True(t, raw[ResourceFinancialManagement][ActionFinancialManagementView])
	assert.True(t, effective[ResourceFinancialAccounting][ActionFinancialView])
	assert.True(t, effective[ResourceFinancialManagement][ActionFinancialManagementView])
	assert.Equal(t, raw, ExplicitUserPermissions(42))
	require.NoError(t, SetUserPermissions(42, raw))
	assert.False(t, ExplicitUserOverrides(42)[ResourceFinancialAccounting][ActionFinancialView])
	assert.True(t, Can(42, common.RoleAdminUser, FinancialAccountingView))
	raw[ResourceFinancialManagement][ActionFinancialManagementView] = false
	require.NoError(t, SetUserPermissions(42, raw))
	assert.False(t, Can(42, common.RoleAdminUser, FinancialManagementView))
	assert.False(t, Can(42, common.RoleAdminUser, FinancialAccountingView), "revoking management must not retain an implied grant from a prior editor save")
	assert.True(t, Can(42, common.RoleAdminUser, ChannelRead))
	assert.False(t, Can(42, common.RoleAdminUser, ChannelSensitiveWrite))
	require.NoError(t, ClearUserAuthorization(42))
	assert.True(t, Can(42, common.RoleAdminUser, FinancialAccountingView))
	assert.False(t, Can(42, common.RoleAdminUser, FinancialManagementView))
}

func TestFinancialManagementTransactionReloadKeepsRawDenialAndEffectiveInheritanceSeparate(t *testing.T) {
	db := newAuthzTestDB(t)
	require.NoError(t, Init(db))
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return SetUserPermissionsInTx(tx, 42, PermissionsMap{
			ResourceFinancialAccounting: {ActionFinancialView: false},
			ResourceFinancialManagement: {ActionFinancialManagementView: true},
		})
	}))
	assert.False(t, Can(42, common.RoleAdminUser, FinancialManagementView), "the local authorization snapshot changes only after reload")
	require.NoError(t, ReloadPolicy())
	assert.True(t, Can(42, common.RoleAdminUser, FinancialManagementView))
	assert.True(t, Can(42, common.RoleAdminUser, FinancialAccountingView))
	raw := RawCapabilities(42, common.RoleAdminUser)
	assert.False(t, raw[ResourceFinancialAccounting][ActionFinancialView])
	raw[ResourceFinancialManagement][ActionFinancialManagementView] = false
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error { return SetUserPermissionsInTx(tx, 42, raw) }))
	require.NoError(t, ReloadPolicy())
	assert.False(t, Can(42, common.RoleAdminUser, FinancialManagementView))
	assert.False(t, Can(42, common.RoleAdminUser, FinancialAccountingView))
	root := RawCapabilities(1, common.RoleRootUser)
	assert.True(t, root[ResourceFinancialManagement][ActionFinancialManagementView])
	assert.True(t, root[ResourceFinancialAccounting][ActionFinancialView])
}
