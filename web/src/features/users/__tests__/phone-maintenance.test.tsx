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
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { afterEach, expect, test, vi } from 'vitest'

import { UsersMutateDrawer } from '../components/users-mutate-drawer'
import { UsersProvider } from '../components/users-provider'
import {
  transformFormDataToPayload,
  transformUserToFormDefaults,
  USER_FORM_DEFAULT_VALUES,
  userFormSchema,
} from '../lib/user-form'
import type { User } from '../types'

const apiMocks = vi.hoisted(() => ({
  adjustUserQuota: vi.fn(async () => ({ success: true })),
  createUser: vi.fn(async () => ({ success: true })),
  getGroups: vi.fn(async () => ({ success: true, data: ['default'] })),
  getPermissionCatalog: vi.fn(async () => ({ resources: [], roles: [] })),
  getUser: vi.fn(async () => ({ success: false })),
  updateUser: vi.fn(async () => ({ success: true })),
}))

vi.mock('../api', () => apiMocks)

let queryClient: QueryClient

afterEach(() => {
  queryClient?.clear()
})

test('create-user form exposes a phone number field', async () => {
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  render(
    <QueryClientProvider client={queryClient}>
      <UsersProvider>
        <UsersMutateDrawer open onOpenChange={vi.fn()} />
      </UsersProvider>
    </QueryClientProvider>
  )

  expect(await screen.findByLabelText('Phone Number')).toHaveAttribute(
    'type',
    'tel'
  )
})

test('create-user payload trims and preserves the phone number', () => {
  const values = userFormSchema.parse({
    ...USER_FORM_DEFAULT_VALUES,
    username: 'phone-create-user',
    password: 'NewPassword123',
    phone: ' 13800138000 ',
  })

  expect(transformFormDataToPayload(values)).toMatchObject({
    username: 'phone-create-user',
    phone: '13800138000',
  })
})

test('edit-user form loads the phone number and can clear it', () => {
  const user: User = {
    id: 1,
    username: 'phone-edit-user',
    display_name: 'Phone Edit User',
    phone: '13800138000',
    quota: 0,
    used_quota: 0,
    request_count: 0,
    group: 'default',
    status: 1,
    role: 1,
  }
  const defaults = transformUserToFormDefaults(user)

  expect(defaults.phone).toBe('13800138000')
  expect(
    transformFormDataToPayload({ ...defaults, phone: '' }, user.id)
  ).toMatchObject({ id: user.id, phone: '' })
})
