package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSidebarSupplierTestIsAdminOnly(t *testing.T) {
	t.Parallel()

	assert.True(t, sidebarHasSupplierTest(t, generateDefaultSidebarConfigForRole(common.RoleAdminUser)))
	assert.True(t, sidebarHasSupplierTest(t, generateDefaultSidebarConfigForRole(common.RoleRootUser)))
	assert.False(t, sidebarHasSupplierTest(t, generateDefaultSidebarConfigForRole(common.RoleCommonUser)))
	assert.False(t, sidebarHasSupplierTest(t, generateDefaultSidebarConfigForRole(common.RoleGuestUser)))
}

func sidebarHasSupplierTest(t *testing.T, raw string) bool {
	t.Helper()
	var cfg map[string]any
	require.NoError(t, common.UnmarshalJsonStr(raw, &cfg))
	admin, _ := cfg["admin"].(map[string]any)
	if admin == nil {
		return false
	}
	enabled, _ := admin["supplier_test"].(bool)
	return enabled
}
