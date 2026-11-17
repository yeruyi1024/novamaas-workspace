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
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest'

import { api } from '@/lib/api'

import { AssetCopyDialog } from '../components/asset-copy-dialog'
import type { AssetCopyInput } from '../types'

vi.mock('@/lib/api', () => ({ api: { get: vi.fn(), post: vi.fn() } }))
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

let client: QueryClient
const customers = [1, 2, 3, 4].map((id) => ({
  id,
  username: `customer-${id}`,
  display_name: `Customer ${id}`,
  role: 1,
  status: 1,
  group: 'default',
}))
const result = {
  id: 'copy-request-000001',
  source_user_id: 1,
  target_user_id: 2,
  ready_count: 1,
  processing_count: 0,
  failed_count: 1,
  unavailable_count: 0,
  groups: [{ id: 'group-copy', name: 'Campaign', asset_count: 2 }],
  asset_count: 2,
  sources_deleted_at: 0,
}

beforeEach(() => {
  vi.clearAllMocks()
  vi.mocked(api.get).mockImplementation(async (url) => {
    if (String(url).startsWith('/api/asset-library/admin/copies/')) {
      return { data: { success: true, data: result } }
    }
    if (String(url).startsWith('/api/user/search?')) {
      const keyword =
        new URL(String(url), 'http://localhost').searchParams.get('keyword') ??
        ''
      const items = customers.filter((user) =>
        `${user.username} ${user.display_name} ${user.id}`
          .toLowerCase()
          .includes(keyword.toLowerCase())
      )
      return {
        data: {
          success: true,
          data: { items, total: items.length, page: 1, page_size: 20 },
        },
      }
    }
    if (String(url).startsWith('/api/user/')) {
      return {
        data: {
          success: true,
          data: customers.find(
            (user) => user.id === Number(String(url).split('/').pop())
          ),
        },
      }
    }

    return {
      data: {
        success: true,
        data: {
          items: [
            {
              id: 'group-source',
              owner_user_id: 1,
              owner_name: 'customer-1',
              name: 'Campaign',
              description: '',
              status: 'ready',
              created_at: 1,
              updated_at: 1,
            },
          ],
          total: 1,
          page: 1,
          page_size: 20,
        },
      },
    }
  })
  vi.mocked(api.post).mockImplementation(async (url) => ({
    data: {
      success: true,
      data: String(url).endsWith('/delete-sources')
        ? { ...result, sources_deleted_at: 123 }
        : result,
    },
  }))
  client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
})

afterEach(() => client.clear())

function renderCopy() {
  render(
    <QueryClientProvider client={client}>
      <AssetCopyDialog open onOpenChange={vi.fn()} />
    </QueryClientProvider>
  )
}

async function selectCustomer(
  label: 'Source user' | 'Target user',
  id: number
) {
  await userEvent.click(screen.getByRole('combobox', { name: label }))
  await userEvent.click(
    await screen.findByRole('option', {
      name: new RegExp(`customer-${id} .*Customer ${id}`),
    })
  )
}

async function selectSourceGroup() {
  await selectCustomer('Source user', 1)
  await selectCustomer('Target user', 2)
  await screen.findByText('customer-1')
  await userEvent.click(screen.getByRole('combobox', { name: 'Asset group' }))
  await userEvent.click(
    await screen.findByRole('option', { name: /Campaign · ID: group-source/ })
  )
}

describe('customer asset copying', () => {
  test('copies between selected users without a channel field and only deletes sources after confirmation', async () => {
    renderCopy()
    expect(
      screen.getByRole('button', { name: 'Copy asset groups' })
    ).toBeDisabled()
    await selectSourceGroup()
    expect(api.get).toHaveBeenCalledWith(
      '/api/asset-library/groups',
      expect.objectContaining({
        params: expect.objectContaining({ owner_user_id: 1, scope: 'all' }),
      })
    )
    await userEvent.click(
      screen.getByRole('button', { name: 'Copy asset groups' })
    )
    await screen.findByText('Created 2 assets in 1 groups')
    expect(
      screen.getByText(
        'Synchronization: 1 ready, 0 processing, 1 failed, 0 unavailable'
      )
    ).toBeInTheDocument()
    expect(api.post).toHaveBeenCalledWith(
      '/api/asset-library/admin/copies',
      expect.objectContaining({
        source_user_id: 1,
        target_user_id: 2,
        group_ids: ['group-source'],
        request_id: expect.any(String),
      }),
      expect.any(Object)
    )
    expect(api.post).toHaveBeenCalledTimes(1)
    await userEvent.click(
      screen.getByRole('button', { name: 'Delete copied sources' })
    )
    const confirmation = await screen.findByRole('alertdialog')
    expect(api.post).toHaveBeenCalledTimes(1)
    await userEvent.click(
      within(confirmation).getByRole('button', {
        name: 'Delete copied sources',
      })
    )
    await waitFor(() =>
      expect(
        screen.getByRole('button', { name: 'Copied source assets deleted' })
      ).toBeDisabled()
    )
    expect(api.post).toHaveBeenLastCalledWith(
      '/api/asset-library/admin/copies/copy-request-000001/delete-sources',
      undefined,
      expect.any(Object)
    )
  })

  test('selecting a source keeps both user pickers visible without grid size containment', async () => {
    renderCopy()
    await selectCustomer('Source user', 1)
    const source = screen.getByRole('combobox', { name: 'Source user' })
    expect(source).toBeVisible()
    expect(screen.getByRole('combobox', { name: 'Target user' })).toBeVisible()
    // WebKit collapses this dynamic grid inside the fieldset when the shared
    // FieldGroup applies inline-size containment after closing the picker.
    expect(source.closest('[data-slot="field-group"]')).toHaveClass(
      '[container-type:normal]'
    )
  })

  test('changing the source clears selected groups before another copy', async () => {
    renderCopy()
    await selectSourceGroup()
    await waitFor(() =>
      expect(
        screen.getByRole('button', { name: 'Copy asset groups' })
      ).toBeEnabled()
    )
    await selectCustomer('Source user', 3)
    await screen.findByText('customer-3')
    expect(screen.getByText(/^0 groups selected/)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Copy asset groups' })
    ).toBeDisabled()
  })

  test('copying without synchronization configuration creates ready assets without a channel choice', async () => {
    const pending = {
      ...result,
      ready_count: 2,
      failed_count: 0,
    }
    vi.mocked(api.post).mockResolvedValue({
      data: { success: true, data: pending },
    })
    const original = vi.mocked(api.get).getMockImplementation()
    if (!original) {
      throw new Error('API fixture missing')
    }
    vi.mocked(api.get).mockImplementation(async (...args) => {
      if (String(args[0]).startsWith('/api/asset-library/admin/copies/')) {
        return { data: { success: true, data: pending } }
      }
      return original(...args)
    })
    renderCopy()
    await selectSourceGroup()
    expect(
      screen.queryByRole('combobox', { name: 'Target channel' })
    ).not.toBeInTheDocument()
    await userEvent.click(
      screen.getByRole('button', { name: 'Copy asset groups' })
    )
    await screen.findByText('Created 2 assets in 1 groups')
    await screen.findByText(
      'Synchronization: 2 ready, 0 processing, 0 failed, 0 unavailable'
    )
    expect(api.post).toHaveBeenCalledWith(
      '/api/asset-library/admin/copies',
      expect.not.objectContaining({ channel_id: expect.anything() }),
      expect.any(Object)
    )
  })

  test('a failed copy retries with the same request ID and never offers source deletion', async () => {
    vi.mocked(api.post).mockRejectedValue(new Error('connection lost'))
    renderCopy()
    await selectSourceGroup()
    await userEvent.click(
      screen.getByRole('button', { name: 'Copy asset groups' })
    )
    await waitFor(() =>
      expect(
        screen.getByRole('button', { name: 'Copy asset groups' })
      ).toBeEnabled()
    )
    expect(
      screen.queryByRole('button', { name: 'Delete copied sources' })
    ).not.toBeInTheDocument()
    await userEvent.click(
      screen.getByRole('button', { name: 'Copy asset groups' })
    )
    await waitFor(() => expect(api.post).toHaveBeenCalledTimes(2))
    expect(vi.mocked(api.post).mock.calls[0][1]).toEqual(
      vi.mocked(api.post).mock.calls[1][1]
    )
  })

  test('editing input after a failed submission generates a new request ID', async () => {
    vi.mocked(api.post).mockRejectedValue(new Error('connection lost'))
    renderCopy()
    await selectSourceGroup()
    await userEvent.click(
      screen.getByRole('button', { name: 'Copy asset groups' })
    )
    await waitFor(() =>
      expect(
        screen.getByRole('button', { name: 'Copy asset groups' })
      ).toBeEnabled()
    )
    await selectCustomer('Target user', 3)
    await screen.findByText('customer-3')
    await userEvent.click(
      screen.getByRole('button', { name: 'Copy asset groups' })
    )
    await waitFor(() => expect(api.post).toHaveBeenCalledTimes(2))
    expect(
      (vi.mocked(api.post).mock.calls[0][1] as AssetCopyInput).request_id
    ).not.toEqual(
      (vi.mocked(api.post).mock.calls[1][1] as AssetCopyInput).request_id
    )
  })

  test('refreshing a receipt shows current target readiness without creating another copy', async () => {
    renderCopy()
    await selectSourceGroup()
    await userEvent.click(
      screen.getByRole('button', { name: 'Copy asset groups' })
    )
    await screen.findByText(
      'Synchronization: 1 ready, 0 processing, 1 failed, 0 unavailable'
    )
    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'Refresh' })).toBeEnabled()
    )
    const original = vi.mocked(api.get).getMockImplementation()
    if (!original) {
      throw new Error('API fixture is missing')
    }
    vi.mocked(api.get).mockImplementation(async (...args) => {
      if (String(args[0]).startsWith('/api/asset-library/admin/copies/')) {
        return {
          data: {
            success: true,
            data: { ...result, ready_count: 2, failed_count: 0 },
          },
        }
      }
      return original(...args)
    })
    await userEvent.click(screen.getByRole('button', { name: 'Refresh' }))
    await screen.findByText(
      'Synchronization: 2 ready, 0 processing, 0 failed, 0 unavailable'
    )
    expect(api.post).toHaveBeenCalledTimes(1)
  })

  test('select all replaces the selection with every source group across normal picker pages', async () => {
    const original = vi.mocked(api.get).getMockImplementation()
    if (!original) {
      throw new Error('API fixture missing')
    }
    const allGroups = Array.from({ length: 23 }, (_, index) => ({
      id: `group-${index + 1}`,
      owner_user_id: 1,
      name: `Group ${index + 1}`,
      description: '',
      status: 'ready',
      created_at: 1,
      updated_at: 1,
    }))
    vi.mocked(api.get).mockImplementation(async (url, options) => {
      if (
        url === '/api/asset-library/groups' &&
        options?.params?.page_size === 100
      ) {
        return {
          data: {
            success: true,
            data: { items: allGroups, total: 23, page: 1, page_size: 100 },
          },
        }
      }
      return original(url, options)
    })
    renderCopy()
    await selectSourceGroup()
    await userEvent.click(
      screen.getByRole('button', { name: 'Select all asset groups' })
    )
    await screen.findByText(/^23 groups selected/)
    expect(
      screen.queryByRole('button', { name: 'Remove Campaign' })
    ).not.toBeInTheDocument()
    expect(api.get).toHaveBeenCalledWith(
      '/api/asset-library/groups',
      expect.objectContaining({
        params: expect.objectContaining({
          owner_user_id: 1,
          page_size: 100,
          scope: 'all',
        }),
      })
    )
    await userEvent.click(
      screen.getByRole('button', { name: 'Copy asset groups' })
    )
    await waitFor(() =>
      expect(api.post).toHaveBeenCalledWith(
        '/api/asset-library/admin/copies',
        expect.objectContaining({
          group_ids: allGroups.map((group) => group.id),
        }),
        expect.any(Object)
      )
    )
  })

  test('select all rejects more than 100 groups and preserves the current selection', async () => {
    const original = vi.mocked(api.get).getMockImplementation()
    if (!original) {
      throw new Error('API fixture missing')
    }
    vi.mocked(api.get).mockImplementation(async (url, options) => {
      if (
        url === '/api/asset-library/groups' &&
        options?.params?.page_size === 100
      ) {
        return {
          data: {
            success: true,
            data: { items: [], total: 101, page: 1, page_size: 100 },
          },
        }
      }
      return original(url, options)
    })
    renderCopy()
    await selectSourceGroup()
    await userEvent.click(
      screen.getByRole('button', { name: 'Select all asset groups' })
    )
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'This user has more than 100 groups. Select up to 100 groups per copy.'
    )
    expect(screen.getByText(/^1 groups selected/)).toBeInTheDocument()
    expect(api.post).not.toHaveBeenCalled()
  })

  test('an old select-all response cannot restore groups after changing the source user', async () => {
    const original = vi.mocked(api.get).getMockImplementation()
    if (!original) {
      throw new Error('API fixture missing')
    }
    let resolveGroups!: (value: unknown) => void
    vi.mocked(api.get).mockImplementation(async (url, options) => {
      if (
        url === '/api/asset-library/groups' &&
        options?.params?.page_size === 100
      ) {
        return new Promise((resolve) => {
          resolveGroups = resolve
        })
      }
      return original(url, options)
    })
    renderCopy()
    await selectSourceGroup()
    await userEvent.click(
      screen.getByRole('button', { name: 'Select all asset groups' })
    )
    expect(
      screen.getByRole('button', { name: 'Copy asset groups' })
    ).toBeDisabled()
    await selectCustomer('Source user', 3)
    resolveGroups({
      data: {
        success: true,
        data: {
          items: [
            {
              id: 'old-source-group',
              owner_user_id: 1,
              name: 'Old source group',
            },
          ],
          total: 1,
          page: 1,
          page_size: 100,
        },
      },
    })
    await waitFor(() => expect(client.isMutating()).toBe(0))
    expect(screen.getByText(/^0 groups selected/)).toBeInTheDocument()
    expect(screen.queryByText('Old source group')).not.toBeInTheDocument()
  })

  test.each([
    {
      name: 'empty',
      total: 0,
      items: [],
      message: 'This user has no asset groups.',
      error: false,
    },
    {
      name: 'incomplete',
      total: 2,
      items: [{ id: 'group-incomplete', owner_user_id: 1, name: 'Incomplete' }],
      message: 'The group list changed. Please select all groups again.',
      error: true,
    },
    {
      name: 'request failure',
      total: 0,
      items: [],
      message: 'group lookup failed',
      error: true,
    },
  ])(
    'select all handles $name without creating a partial copy',
    async (scenario) => {
      const original = vi.mocked(api.get).getMockImplementation()
      if (!original) {
        throw new Error('API fixture missing')
      }
      vi.mocked(api.get).mockImplementation(async (url, options) => {
        if (
          url === '/api/asset-library/groups' &&
          options?.params?.page_size === 100
        ) {
          if (scenario.name === 'request failure') {
            throw new Error(scenario.message)
          }
          return {
            data: {
              success: true,
              data: {
                items: scenario.items,
                total: scenario.total,
                page: 1,
                page_size: 100,
              },
            },
          }
        }
        return original(url, options)
      })
      renderCopy()
      await selectSourceGroup()
      await userEvent.click(
        screen.getByRole('button', { name: 'Select all asset groups' })
      )
      expect(
        await screen.findByRole(scenario.error ? 'alert' : 'status')
      ).toHaveTextContent(scenario.message)
      expect(
        screen.getByText(
          scenario.error ? /^1 groups selected/ : /^0 groups selected/
        )
      ).toBeInTheDocument()
      expect(api.post).not.toHaveBeenCalled()
    }
  )
})
