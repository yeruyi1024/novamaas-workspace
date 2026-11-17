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
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
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

test('copy progress waits for processing to finish and then confirms batch source deletion', async () => {
  let receipt = {
    ...result,
    ready_count: 1,
    processing_count: 1,
    failed_count: 0,
    can_delete_sources: false,
  }
  const get = vi.mocked(api.get).getMockImplementation()
  if (!get) throw new Error('Missing API fixture')
  vi.mocked(api.get).mockImplementation(async (...args) =>
    String(args[0]).includes('/admin/copies/')
      ? { data: { success: true, data: receipt } }
      : get(...args)
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
  expect(
    screen.getByRole('progressbar', { name: 'Copy progress' })
  ).toHaveAttribute('aria-valuenow', '50')
  expect(
    screen.getByRole('button', { name: 'Delete copied sources' })
  ).toBeDisabled()
  receipt = { ...result, ready_count: 1, failed_count: 1 }
  await userEvent.click(screen.getByRole('button', { name: 'Refresh' }))
  await waitFor(() =>
    expect(
      screen.getByRole('progressbar', { name: 'Copy progress' })
    ).toHaveAttribute('aria-valuenow', '100')
  )
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
  expect(api.post).toHaveBeenCalledTimes(1)
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

test('stale successful receipt cannot enable deletion after a refresh error', async () => {
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

test('an in-flight copy has indeterminate progress and no delete action', async () => {
  renderCopy()
  await selectSourceGroup()
  let resolve: (value: unknown) => void = () => {}
  vi.mocked(api.post).mockImplementationOnce(
    () =>
      new Promise((done) => {
        resolve = done
      })
  )
  await userEvent.click(
    screen.getByRole('button', { name: 'Copy asset groups' })
  )
  expect(
    screen.getByRole('progressbar', { name: 'Copying...' })
  ).not.toHaveAttribute('aria-valuenow')
  expect(
    screen.queryByRole('button', { name: 'Delete copied sources' })
  ).not.toBeInTheDocument()
  resolve({ data: { success: true, data: result } })
  await screen.findByText('Created 2 assets in 1 groups')
})

test('in-flight synchronization continues polling until the batch finishes', async () => {
  let receipt = {
    ...result,
    ready_count: 1,
    processing_count: 1,
    failed_count: 0,
    can_delete_sources: false,
  }
  const get = vi.mocked(api.get).getMockImplementation()
  if (!get) throw new Error('Missing API fixture')
  vi.mocked(api.get).mockImplementation(async (...args) =>
    String(args[0]).includes('/admin/copies/')
      ? { data: { success: true, data: receipt } }
      : get(...args)
  )
  vi.mocked(api.post).mockResolvedValue({
    data: { success: true, data: receipt },
  })
  renderCopy()
  await selectSourceGroup()
  vi.useFakeTimers()
  try {
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Copy asset groups' }))
    })
    expect(
      screen.getByRole('button', { name: 'Delete copied sources' })
    ).toBeDisabled()
    receipt = { ...result }
    await act(async () => {
      await vi.advanceTimersByTimeAsync(3001)
    })
    expect(
      screen.getByRole('progressbar', { name: 'Copy progress' })
    ).toHaveAttribute('aria-valuenow', '100')
    expect(
      screen.getByRole('button', { name: 'Delete copied sources' })
    ).toBeEnabled()
  } finally {
    vi.useRealTimers()
  }
})

test('select-all uses every group and discards a late response after changing source', async () => {
  renderCopy()
  await selectSourceGroup()
  let resolve: (value: unknown) => void = () => {}
  vi.mocked(api.get).mockImplementationOnce(
    () =>
      new Promise((done) => {
        resolve = done
      })
  )
  await userEvent.click(
    screen.getByRole('button', { name: 'Select all asset groups' })
  )
  await selectCustomer('Source user', 3)
  await act(async () => {
    resolve({
      data: {
        success: true,
        data: { items: [{ id: 'group-old', name: 'Old source' }], total: 1 },
      },
    })
  })
  expect(
    screen.queryByRole('button', { name: 'Remove Old source' })
  ).not.toBeInTheDocument()
  expect(
    screen.getByRole('button', { name: 'Copy asset groups' })
  ).toBeDisabled()
  vi.mocked(api.get).mockResolvedValueOnce({
    data: {
      success: true,
      data: {
        items: [
          { id: 'group-1', name: 'First' },
          { id: 'group-2', name: 'Second' },
        ],
        total: 2,
      },
    },
  })
  await userEvent.click(
    screen.getByRole('button', { name: 'Select all asset groups' })
  )
  await screen.findByRole('button', { name: 'Remove Second' })
  await userEvent.click(
    screen.getByRole('button', { name: 'Copy asset groups' })
  )
  await waitFor(() =>
    expect(api.post).toHaveBeenCalledWith(
      '/api/asset-library/admin/copies',
      expect.objectContaining({
        source_user_id: 3,
        group_ids: ['group-1', 'group-2'],
      }),
      expect.any(Object)
    )
  )
})

test('partial completion reaches 100 percent and reports the failure count', async () => {
  const partial = {
    ...result,
    ready_count: 1,
    failed_count: 1,
    can_delete_sources: true,
  }
  const get = vi.mocked(api.get).getMockImplementation()
  if (!get) throw new Error('Missing API fixture')
  vi.mocked(api.get).mockImplementation(async (...args) =>
    String(args[0]).includes('/admin/copies/')
      ? { data: { success: true, data: partial } }
      : get(...args)
  )
  vi.mocked(api.post).mockResolvedValue({
    data: { success: true, data: partial },
  })
  renderCopy()
  await selectSourceGroup()
  await userEvent.click(
    screen.getByRole('button', { name: 'Copy asset groups' })
  )
  await screen.findByText('Created 2 assets in 1 groups')
  expect(
    screen.getByRole('progressbar', { name: 'Copy progress' })
  ).toHaveAttribute('aria-valuenow', '100')
  expect(screen.getByRole('status')).toHaveTextContent(
    'Copy complete: 1 succeeded, 1 failed.'
  )
})

test('a wholly failed batch finishes processing and allows confirmed batch cleanup', async () => {
  const failed = {
    ...result,
    ready_count: 0,
    failed_count: 2,
    can_delete_sources: true,
  }
  const get = vi.mocked(api.get).getMockImplementation()
  if (!get) throw new Error('Missing API fixture')
  vi.mocked(api.get).mockImplementation(async (...args) =>
    String(args[0]).includes('/admin/copies/')
      ? { data: { success: true, data: failed } }
      : get(...args)
  )
  vi.mocked(api.post).mockResolvedValue({
    data: { success: true, data: failed },
  })
  renderCopy()
  await selectSourceGroup()
  await userEvent.click(
    screen.getByRole('button', { name: 'Copy asset groups' })
  )
  await screen.findByText('Created 2 assets in 1 groups')
  expect(
    screen.getByRole('progressbar', { name: 'Copy progress' })
  ).toHaveAttribute('aria-valuenow', '100')
  expect(screen.getByRole('status')).toHaveTextContent(
    'Copy complete: 0 succeeded, 2 failed.'
  )
  expect(
    screen.getByRole('button', { name: 'Delete copied sources' })
  ).toBeEnabled()
})
