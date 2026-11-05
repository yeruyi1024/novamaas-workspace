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
import { act, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'

import {
  getMediaAssetPreview,
  listAssetGroupsPage,
  listMediaAssets,
} from '../api'
import { AssetLibrary } from '../asset-library'
import type { MediaAsset } from '../types'

vi.mock('../api', () => ({
  createAssetGroup: vi.fn(),
  createAssetAccessKey: vi.fn(),
  createAssetWebhookEndpoint: vi.fn(),
  deleteAssetAccessKey: vi.fn(),
  deleteAssetWebhookEndpoint: vi.fn(),
  deleteAssetGroup: vi.fn(),
  deleteMediaAsset: vi.fn(),
  getMediaAssetPreview: vi.fn(),
  listAssetAccessKeys: vi.fn(),
  listAssetWebhookEndpoints: vi.fn(),
  listAssetGroups: vi.fn(),
  listAssetGroupsPage: vi.fn(),
  listMediaAssets: vi.fn(),
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
const asset: MediaAsset = {
  id: 'asset-polling',
  group_id: 'group-polling',
  owner_user_id: 1,
  owner_name: 'member',
  name: 'Pending upload',
  type: 'image',
  content_type: 'image/jpeg',
  size: 1024,
  sha256: 'checksum',
  status: 'processing',
  created_at: 1,
  updated_at: 1,
}

function assetsResponse(items: MediaAsset[]) {
  return {
    success: true,
    data: { items, total: items.length, page: 1, page_size: 40 },
  }
}

function renderLibrary() {
  render(
    <QueryClientProvider client={queryClient}>
      <AssetLibrary />
    </QueryClientProvider>
  )
}

beforeEach(() => {
  vi.useFakeTimers()
  vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('visible')
  useAuthStore.getState().auth.reset()
  useAuthStore.getState().auth.setUser({ id: 1, username: 'member', role: 1 })
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  vi.mocked(listAssetGroupsPage).mockResolvedValue({
    success: true,
    data: { items: [], total: 0, page: 1, page_size: 40 },
  })
  vi.mocked(getMediaAssetPreview).mockResolvedValue({
    success: true,
    data: { url: 'https://example.com/preview.jpg', expires_at: 100 },
  })
})

afterEach(() => {
  queryClient.clear()
  vi.useRealTimers()
})

test('processing items become ready after a three-second refresh and then use the idle interval', async () => {
  const ready = { ...asset, status: 'ready' }
  vi.mocked(listMediaAssets)
    .mockResolvedValueOnce(assetsResponse([ready, { ...asset, id: 'pending' }]))
    .mockResolvedValue(assetsResponse([ready]))
  renderLibrary()
  await act(() => vi.advanceTimersByTimeAsync(1))
  expect(screen.getByText('Under review')).toBeVisible()

  await act(() => vi.advanceTimersByTimeAsync(3_001))
  expect(listMediaAssets).toHaveBeenCalledTimes(2)
  expect(screen.queryByText('Under review')).not.toBeInTheDocument()

  await act(() => vi.advanceTimersByTimeAsync(29_000))
  expect(listMediaAssets).toHaveBeenCalledTimes(2)
  await act(() => vi.advanceTimersByTimeAsync(1_001))
  expect(listMediaAssets).toHaveBeenCalledTimes(3)
})

test.each(['ready', 'sync_failed', 'unavailable', 'empty'])(
  'a %s list keeps the thirty-second idle refresh interval',
  async (status) => {
    vi.mocked(listMediaAssets).mockResolvedValue(
      assetsResponse(status === 'empty' ? [] : [{ ...asset, status }])
    )
    renderLibrary()
    await act(() => vi.advanceTimersByTimeAsync(1))
    await act(() => vi.advanceTimersByTimeAsync(29_000))
    expect(listMediaAssets).toHaveBeenCalledTimes(1)
    await act(() => vi.advanceTimersByTimeAsync(1_001))
    expect(listMediaAssets).toHaveBeenCalledTimes(2)
  }
)

test('a failed refresh backs off to thirty seconds while stale processing data remains', async () => {
  vi.mocked(listMediaAssets)
    .mockResolvedValueOnce(assetsResponse([asset]))
    .mockRejectedValueOnce(new Error('Temporarily unavailable'))
    .mockResolvedValue(assetsResponse([{ ...asset, status: 'ready' }]))
  renderLibrary()
  await act(() => vi.advanceTimersByTimeAsync(1))
  await act(() => vi.advanceTimersByTimeAsync(3_001))
  expect(listMediaAssets).toHaveBeenCalledTimes(2)
  await act(() => vi.advanceTimersByTimeAsync(29_000))
  expect(listMediaAssets).toHaveBeenCalledTimes(2)
  await act(() => vi.advanceTimersByTimeAsync(1_001))
  expect(listMediaAssets).toHaveBeenCalledTimes(3)
  expect(screen.queryByText('Under review')).not.toBeInTheDocument()
})
