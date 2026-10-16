/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { expect, test } from 'vitest'

import { hasPermission } from '@/lib/admin-permissions'

import {
  canManageUpstreamBilling,
  canViewUpstreamBilling,
} from '../permissions'

test('financial management inherits accounting view without granting unrelated resources or actions', () => {
  const manager = {
    id: 1,
    username: 'manager',
    role: 10,
    permissions: {
      admin_permissions: {
        financial_management: { view: true },
        financial_accounting: { view: false },
      },
    },
  }
  expect(hasPermission(manager, 'financial_accounting', 'view')).toBe(true)
  expect(canViewUpstreamBilling(manager)).toBe(true)
  expect(canManageUpstreamBilling(manager)).toBe(true)
  expect(hasPermission(manager, 'channel', 'write')).toBe(false)
  expect(hasPermission(manager, 'financial_accounting', 'write')).toBe(false)
  expect(
    hasPermission({ ...manager, role: 1 }, 'financial_management', 'view')
  ).toBe(true)
  expect(canViewUpstreamBilling({ ...manager, role: 1 })).toBe(false)
  expect(canManageUpstreamBilling({ ...manager, role: 1 })).toBe(false)
  const accountant = {
    ...manager,
    permissions: {
      admin_permissions: { financial_accounting: { view: true } },
    },
  }
  expect(canViewUpstreamBilling(accountant)).toBe(true)
  expect(canManageUpstreamBilling(accountant)).toBe(false)
})
