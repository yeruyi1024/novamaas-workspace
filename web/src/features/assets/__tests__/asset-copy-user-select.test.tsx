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
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import type { User } from '@/features/users/types'
import { api } from '@/lib/api'

import { AssetCopyUserSelect } from '../components/asset-copy-user-select'

vi.mock('@/lib/api', () => ({ api: { get: vi.fn() } }))

const users: User[] = [
  {
    id: 1,
    username: 'customer-a',
    display_name: 'Alice',
    role: 1,
    status: 1,
    group: 'default',
    quota: 0,
    used_quota: 0,
    request_count: 0,
  },
  {
    id: 2,
    username: 'customer-b',
    display_name: 'Bob',
    role: 1,
    status: 1,
    group: 'default',
    quota: 0,
    used_quota: 0,
    request_count: 0,
  },
]
let client: QueryClient

beforeEach(() => {
  vi.resetAllMocks()
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  vi.mocked(api.get).mockImplementation(async (url) => {
    const params = new URL(String(url), 'http://localhost').searchParams
    const keyword = params.get('keyword')?.toLowerCase() ?? ''
    const items = users.filter((user) =>
      `${user.username} ${user.display_name} ${user.id}`
        .toLowerCase()
        .includes(keyword)
    )
    return {
      data: {
        success: true,
        data: { items, total: items.length, page: 1, page_size: 20 },
      },
    }
  })
})
afterEach(() => client.clear())

function renderPicker(excludeUserId = 0) {
  const onValueChange = vi.fn()
  render(
    <QueryClientProvider client={client}>
      <AssetCopyUserSelect
        id='test-user'
        label='Target user'
        excludeUserId={excludeUserId}
        enabledOnly
        onValueChange={onValueChange}
      />
    </QueryClientProvider>
  )
  return onValueChange
}

test('user list shows names and IDs, searches nicknames and supports keyboard selection', async () => {
  const select = renderPicker()
  await userEvent.click(screen.getByRole('combobox', { name: 'Target user' }))
  expect(
    await screen.findByRole('option', { name: 'customer-a · Alice · ID: 1' })
  ).toBeInTheDocument()
  await userEvent.type(
    screen.getByRole('combobox', {
      name: 'Search users by name or ID',
    }),
    'Bob'
  )
  await screen.findByRole('option', { name: 'customer-b · Bob · ID: 2' })
  expect(
    screen.queryByRole('option', { name: 'customer-a · Alice · ID: 1' })
  ).not.toBeInTheDocument()
  await userEvent.keyboard('{ArrowDown}{Enter}')
  expect(select).toHaveBeenCalledWith(users[1])
  expect(screen.getByRole('combobox', { name: 'Target user' })).toHaveAttribute(
    'aria-expanded',
    'false'
  )
})

test('pagination loads the next user page and keeps the selected user identity', async () => {
  vi.mocked(api.get).mockImplementation(async (url) => {
    const page = Number(
      new URL(String(url), 'http://localhost').searchParams.get('p')
    )
    return {
      data: {
        success: true,
        data: { items: [users[page - 1]], total: 21, page, page_size: 20 },
      },
    }
  })
  const select = renderPicker()
  await userEvent.click(screen.getByRole('combobox', { name: 'Target user' }))
  await screen.findByRole('option', { name: 'customer-a · Alice · ID: 1' })
  await userEvent.click(screen.getByRole('button', { name: 'Next' }))
  await userEvent.click(
    await screen.findByRole('option', { name: 'customer-b · Bob · ID: 2' })
  )
  expect(select).toHaveBeenCalledWith(users[1])
})

test('target selection prevents choosing the source and hides disabled, deleted or protected users', async () => {
  vi.mocked(api.get).mockResolvedValue({
    data: {
      success: true,
      data: {
        items: [
          users[0],
          users[1],
          { ...users[1], id: 3, username: 'disabled', status: 2 },
          {
            ...users[1],
            id: 4,
            username: 'deleted',
            DeletedAt: '2026-11-17T00:00:00Z',
          },
          { ...users[1], id: 5, username: 'admin', role: 10 },
        ],
        total: 5,
        page: 1,
        page_size: 20,
      },
    },
  })
  renderPicker(1)
  await userEvent.click(screen.getByRole('combobox', { name: 'Target user' }))
  expect(
    await screen.findByRole('option', { name: 'customer-a · Alice · ID: 1' })
  ).toHaveAttribute('aria-disabled', 'true')
  expect(
    screen.getByRole('option', { name: 'customer-b · Bob · ID: 2' })
  ).not.toHaveAttribute('aria-disabled', 'true')
  expect(screen.getAllByRole('option')).toHaveLength(2)
})

test('a failed user lookup shows a retry action and can recover', async () => {
  const original = vi.mocked(api.get).getMockImplementation()
  if (!original) {
    throw new Error('API fixture missing')
  }
  vi.mocked(api.get).mockRejectedValueOnce(new Error('users unavailable'))
  renderPicker()
  await userEvent.click(screen.getByRole('combobox', { name: 'Target user' }))
  expect(await screen.findByRole('alert')).toHaveTextContent(
    'users unavailable'
  )
  expect(screen.queryByRole('option')).not.toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: 'Retry' }))
  await screen.findByRole('option', { name: 'customer-a · Alice · ID: 1' })
})

test('an older search response cannot replace the current user list', async () => {
  let resolveOld!: (value: unknown) => void
  vi.mocked(api.get).mockImplementation(async (url) => {
    const keyword = new URL(String(url), 'http://localhost').searchParams.get(
      'keyword'
    )
    if (keyword === 'old') {
      return new Promise((resolve) => {
        resolveOld = resolve
      })
    }
    return {
      data: {
        success: true,
        data: {
          items: keyword === 'new' ? [users[1]] : users,
          total: 2,
          page: 1,
          page_size: 20,
        },
      },
    }
  })
  renderPicker()
  await userEvent.click(screen.getByRole('combobox', { name: 'Target user' }))
  await screen.findByRole('option', { name: 'customer-a · Alice · ID: 1' })
  const search = screen.getByRole('combobox', {
    name: 'Search users by name or ID',
  })
  fireEvent.change(search, { target: { value: 'old' } })
  await waitFor(() => expect(resolveOld).toBeDefined())
  fireEvent.change(search, { target: { value: 'new' } })
  await screen.findByRole('option', { name: 'customer-b · Bob · ID: 2' })
  resolveOld({
    data: {
      success: true,
      data: { items: [users[0]], total: 1, page: 1, page_size: 20 },
    },
  })
  await waitFor(() =>
    expect(
      client.getQueryData(['asset-copy-users', false, true, 'old', 1])
    ).toBeDefined()
  )
  expect(
    screen.queryByRole('option', { name: 'customer-a · Alice · ID: 1' })
  ).not.toBeInTheDocument()
  expect(
    screen.getByRole('option', { name: 'customer-b · Bob · ID: 2' })
  ).toBeInTheDocument()
})
