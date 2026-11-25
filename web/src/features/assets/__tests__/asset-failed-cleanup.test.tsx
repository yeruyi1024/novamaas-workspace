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
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'

import {
  deleteMediaAssetsByFilter,
  listAssetGroups,
  listAssetGroupsPage,
  listMediaAssets,
  summarizeMediaAssets,
} from '../api'
import { AssetLibrary } from '../asset-library'
import type { MediaAsset } from '../types'
import { mockAssetPageApis } from './asset-api-fixtures'

vi.mock('../api', () => ({
  createAssetGroup: vi.fn(),
  createAssetAccessKey: vi.fn(),
  createAssetWebhookEndpoint: vi.fn(),
  deleteAssetAccessKey: vi.fn(),
  deleteAssetWebhookEndpoint: vi.fn(),
  deleteAssetGroup: vi.fn(),
  deleteMediaAsset: vi.fn(),
  deleteMediaAssets: vi.fn(),
  deleteMediaAssetsByFilter: vi.fn(),
  getMediaAssetPreview: vi.fn(),
  listAssetAccessKeys: vi.fn(),
  listAssetWebhookEndpoints: vi.fn(),
  listAssetGroups: vi.fn(),
  listAssetGroupsPage: vi.fn(),
  listAssetOwners: vi.fn(),
  listMediaAssets: vi.fn(),
  previewMediaAssets: vi.fn(),
  summarizeMediaAssets: vi.fn(),
  updateAssetGroup: vi.fn(),
  rotateAssetWebhookEndpointSecret: vi.fn(),
  testAssetWebhookEndpoint: vi.fn(),
  uploadMediaAsset: vi.fn(),
}))
vi.mock('@/features/webhooks/api', async (importOriginal) => {
  const api = await importOriginal<typeof import('@/features/webhooks/api')>()
  return {
    ...api,
    getWebhookCapabilities: vi.fn().mockResolvedValue({
      success: true,
      data: {
        asset_library_enabled: false,
        media_tasks_enabled: false,
        manual_enabled: false,
      },
    }),
  }
})
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

let queryClient: QueryClient
const failedAsset: MediaAsset = {
  id: 'asset-20261124120000-fail',
  group_id: 'group-1',
  group_name: 'Campaign',
  owner_user_id: 1,
  owner_name: 'member',
  name: 'Failed image',
  type: 'image',
  content_type: 'image/png',
  size: 12,
  sha256: 'checksum',
  status: 'failed',
  created_at: 1,
  updated_at: 2,
}

function mockFailedCount(unavailable: number) {
  vi.mocked(summarizeMediaAssets).mockResolvedValue({
    success: true,
    data: {
      total: 2 + unavailable,
      by_type: { image: 2 + unavailable, video: 0, audio: 0 },
      by_status: { ready: 2, processing: 0, unavailable },
    },
  })
}

function renderLibrary() {
  render(
    <QueryClientProvider client={queryClient}>
      <AssetLibrary />
    </QueryClientProvider>
  )
}

beforeEach(() => {
  vi.clearAllMocks()
  mockAssetPageApis()
  useAuthStore.getState().auth.reset()
  useAuthStore.getState().auth.setUser({ id: 1, username: 'member', role: 1 })
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  const group = {
    id: 'group-1',
    owner_user_id: 1,
    name: 'Campaign',
    description: '',
    status: 'ready',
    created_at: 1,
    updated_at: 1,
  }
  vi.mocked(listAssetGroups).mockResolvedValue({ success: true, data: [group] })
  vi.mocked(listAssetGroupsPage).mockResolvedValue({
    success: true,
    data: { items: [group], total: 1, page: 1, page_size: 1 },
  })
  vi.mocked(listMediaAssets).mockResolvedValue({
    success: true,
    data: { items: [failedAsset], total: 1, page: 1, page_size: 40 },
  })
  vi.mocked(deleteMediaAssetsByFilter).mockResolvedValue({
    success: true,
    data: { deleted_count: 3, owner_user_ids: [1] },
  })
})

afterEach(() => {
  queryClient.clear()
})

test('no failed-asset notice is shown when the scope has no failed assets', async () => {
  mockFailedCount(0)
  renderLibrary()
  expect(await screen.findByText('Failed image')).toBeVisible()
  await waitFor(() => expect(summarizeMediaAssets).toHaveBeenCalled())
  expect(
    screen.queryByRole('button', { name: 'Clear failed assets' })
  ).not.toBeInTheDocument()
})

test('failed assets in the scope can be viewed and cleared in one confirmed step', async () => {
  mockFailedCount(3)
  renderLibrary()
  const user = userEvent.setup()
  const notice = await screen.findByText(
    '3 failed assets cannot be used (rejected by review or failed to synchronize).'
  )
  expect(notice).toBeVisible()

  await user.click(screen.getByRole('button', { name: 'View' }))
  await waitFor(() =>
    expect(listMediaAssets).toHaveBeenLastCalledWith(
      expect.objectContaining({ status: 'unavailable' })
    )
  )
  expect(screen.queryByRole('button', { name: 'View' })).not.toBeInTheDocument()

  await user.click(screen.getByRole('button', { name: 'Clear failed assets' }))
  const dialog = await screen.findByRole('alertdialog', {
    name: 'Clear 3 failed assets?',
  })
  expect(deleteMediaAssetsByFilter).not.toHaveBeenCalled()
  await user.click(within(dialog).getByRole('button', { name: 'Delete' }))
  await waitFor(() =>
    expect(deleteMediaAssetsByFilter).toHaveBeenCalledWith({
      scope: undefined,
      owner_user_id: undefined,
      group_id: undefined,
      status: 'unavailable',
      search: undefined,
    })
  )
  expect(deleteMediaAssetsByFilter).toHaveBeenCalledTimes(1)
})

test("an administrator clearing every user's failed assets must acknowledge the scope", async () => {
  useAuthStore.getState().auth.setUser({ id: 99, username: 'admin', role: 100 })
  mockFailedCount(5)
  renderLibrary()
  const user = userEvent.setup()
  await user.click(
    await screen.findByRole('button', { name: 'Clear failed assets' })
  )
  const dialog = await screen.findByRole('alertdialog', {
    name: 'Clear 5 failed assets?',
  })
  const confirm = within(dialog).getByRole('button', { name: 'Delete' })
  expect(confirm).toBeDisabled()
  await user.click(
    within(dialog).getByRole('checkbox', {
      name: 'I confirm clearing failed assets of all users in this view.',
    })
  )
  await user.click(confirm)
  await waitFor(() =>
    expect(deleteMediaAssetsByFilter).toHaveBeenCalledWith(
      expect.objectContaining({ scope: 'all', status: 'unavailable' })
    )
  )
})

test('viewing failed assets from a type tab drops the type filter so the list matches the count', async () => {
  mockFailedCount(3)
  const onFiltersChange = vi.fn()
  render(
    <QueryClientProvider client={queryClient}>
      <AssetLibrary
        filters={{ type: 'video' }}
        onFiltersChange={onFiltersChange}
      />
    </QueryClientProvider>
  )
  const user = userEvent.setup()
  await user.click(await screen.findByRole('button', { name: 'View' }))
  const patch = onFiltersChange.mock.calls.at(-1)?.[0]
  expect(patch).toMatchObject({ status: 'unavailable' })
  expect(patch).toHaveProperty('type', undefined)
})
