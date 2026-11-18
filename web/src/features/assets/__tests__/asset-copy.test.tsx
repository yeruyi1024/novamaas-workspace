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
import {
  QueryClient,
  QueryClientProvider,
  QueryObserver,
} from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

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
  ready_count: 2,
  processing_count: 0,
  failed_count: 0,
  unavailable_count: 0,
  groups: [{ id: 'group-copy', name: 'Campaign', asset_count: 2 }],
  asset_count: 2,
  sources_deleted_at: 0,
  can_delete_sources: true,
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

test('a retry reuses its request ID and a changed selection gets a new one', async () => {
  vi.mocked(api.post).mockRejectedValue(new Error('connection lost'))
  renderCopy()
  await selectSourceGroup()
  const submit = () =>
    userEvent.click(screen.getByRole('button', { name: 'Copy asset groups' }))
  const requestId = (call: number) =>
    (vi.mocked(api.post).mock.calls[call][1] as AssetCopyInput).request_id
  await submit()
  await waitFor(() => expect(api.post).toHaveBeenCalledTimes(1))
  await waitFor(() =>
    expect(
      screen.getByRole('button', { name: 'Copy asset groups' })
    ).toBeEnabled()
  )
  await submit()
  await waitFor(() => expect(api.post).toHaveBeenCalledTimes(2))
  expect(requestId(1)).toEqual(requestId(0))
  expect(
    screen.queryByRole('button', { name: 'Delete copied sources' })
  ).not.toBeInTheDocument()
  await selectCustomer('Target user', 3)
  await submit()
  await waitFor(() => expect(api.post).toHaveBeenCalledTimes(3))
  expect(requestId(2)).not.toEqual(requestId(1))
})

test('changing the source clears the selected groups', async () => {
  renderCopy()
  await selectSourceGroup()
  await selectCustomer('Source user', 3)
  await screen.findByText('customer-3')
  expect(screen.getByText(/^0 groups selected/)).toBeInTheDocument()
  expect(
    screen.getByRole('button', { name: 'Copy asset groups' })
  ).toBeDisabled()
})

test('select all rejects more than 100 groups and keeps the current selection', async () => {
  const original = vi.mocked(api.get).getMockImplementation()
  vi.mocked(api.get).mockImplementation(async (url, options) =>
    url === '/api/asset-library/groups' && options?.params?.page_size === 100
      ? {
          data: {
            success: true,
            data: { items: [], total: 101, page: 1, page_size: 100 },
          },
        }
      : original?.(url, options)
  )
  renderCopy()
  await selectSourceGroup()
  await userEvent.click(
    screen.getByRole('button', { name: 'Select all asset groups' })
  )
  expect(await screen.findByRole('alert')).toHaveTextContent(
    'This user has more than 100 groups. Select up to 100 groups per copy.'
  )
  expect(screen.getByText(/^1 groups selected/)).toBeInTheDocument()
})

test('progress waits for processing, then deletes the batch sources without refetching previews', async () => {
  let receipt = {
    ...result,
    ready_count: 1,
    processing_count: 1,
    can_delete_sources: false,
  }
  const get = vi.mocked(api.get).getMockImplementation()
  vi.mocked(api.get).mockImplementation(async (...args) =>
    String(args[0]).includes('/admin/copies/')
      ? { data: { success: true, data: receipt } }
      : get?.(...args)
  )
  vi.mocked(api.post).mockImplementation(async (url) => ({
    data: {
      success: true,
      data: String(url).endsWith('/delete-sources')
        ? { ...receipt, sources_deleted_at: 123, can_delete_sources: false }
        : receipt,
    },
  }))
  renderCopy()
  await selectSourceGroup()
  await userEvent.click(
    screen.getByRole('button', { name: 'Copy asset groups' })
  )
  await screen.findByText('Created 2 assets in 1 groups')
  const progress = () =>
    screen.getByRole('progressbar', { name: 'Copy progress' })
  expect(progress()).toHaveAttribute('aria-valuenow', '50')
  expect(
    screen.getByRole('button', { name: 'Delete copied sources' })
  ).toBeDisabled()

  receipt = { ...result, ready_count: 1, failed_count: 1 }
  await userEvent.click(screen.getByRole('button', { name: 'Refresh' }))
  await waitFor(() =>
    expect(progress()).toHaveAttribute('aria-valuenow', '100')
  )
  expect(
    screen.getByText('Copy complete: 1 succeeded, 1 failed.')
  ).toBeInTheDocument()

  const previewRequest = vi.fn(async () => ({ url: '/source-preview.jpg' }))
  const preview = new QueryObserver(client, {
    queryKey: ['asset-library', 'preview', 'source', 'thumbnail'],
    queryFn: previewRequest,
    staleTime: Infinity,
  })
  const unsubscribe = preview.subscribe(() => {})
  await waitFor(() => expect(previewRequest).toHaveBeenCalledOnce())
  await userEvent.click(
    screen.getByRole('button', { name: 'Delete copied sources' })
  )
  await userEvent.click(
    within(screen.getByRole('alertdialog')).getByRole('button', {
      name: 'Delete copied sources',
    })
  )
  await waitFor(() =>
    expect(
      screen.getByRole('button', { name: 'Copied source assets deleted' })
    ).toBeDisabled()
  )
  unsubscribe()
  expect(previewRequest).toHaveBeenCalledOnce()
})

test('a failed refresh disables source deletion', async () => {
  renderCopy()
  await selectSourceGroup()
  await userEvent.click(
    screen.getByRole('button', { name: 'Copy asset groups' })
  )
  await screen.findByText('Created 2 assets in 1 groups')
  vi.mocked(api.get).mockRejectedValue(new Error('Receipt unavailable'))
  await userEvent.click(screen.getByRole('button', { name: 'Refresh' }))
  await screen.findByText('Receipt unavailable')
  expect(
    screen.getByRole('button', { name: 'Delete copied sources' })
  ).toBeDisabled()
})
