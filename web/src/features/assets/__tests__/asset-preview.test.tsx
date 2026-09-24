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
import { toast } from 'sonner'
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
  deleteAssetAccessKey: vi.fn(),
  deleteAssetGroup: vi.fn(),
  deleteMediaAsset: vi.fn(),
  getMediaAssetPreview: vi.fn(),
  listAssetAccessKeys: vi.fn(),
  listAssetGroups: vi.fn(),
  listAssetGroupsPage: vi.fn(),
  listMediaAssets: vi.fn(),
  uploadMediaAsset: vi.fn(),
}))
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

let queryClient: QueryClient
const asset: MediaAsset = {
  id: 'asset-image',
  group_id: 'group-1',
  group_name: 'Campaign',
  owner_user_id: 1,
  owner_name: 'member',
  name: 'Hero image',
  type: 'image',
  content_type: 'image/jpeg',
  size: 1024,
  sha256: 'checksum',
  status: 'ready',
  created_at: 1,
  updated_at: 1,
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
  useAuthStore.getState().auth.reset()
  useAuthStore.getState().auth.setUser({ id: 1, username: 'member', role: 1 })
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  vi.mocked(listAssetGroupsPage).mockResolvedValue({
    success: true,
    data: { items: [], total: 0, page: 1, page_size: 1 },
  })
  vi.mocked(listMediaAssets).mockResolvedValue({
    success: true,
    data: { items: [asset], total: 1, page: 1, page_size: 40 },
  })
  vi.mocked(getMediaAssetPreview).mockImplementation(async (_id, variant) => ({
    success: true,
    data: {
      url: `https://example.com/${variant || 'original'}.jpg`,
      expires_at: 4102444800,
    },
  }))
})

afterEach(() => {
  queryClient.clear()
  vi.restoreAllMocks()
})

test('clicking an image thumbnail opens the original image while the card uses a small OSS preview', async () => {
  renderLibrary()
  const user = userEvent.setup()

  const thumbnail = await screen.findByRole('button', {
    name: 'Preview Hero image',
  })
  await waitFor(() =>
    expect(thumbnail.querySelector('img')).toHaveAttribute(
      'src',
      'https://example.com/thumbnail.jpg'
    )
  )
  expect(getMediaAssetPreview).toHaveBeenCalledWith('asset-image', 'thumbnail')

  await user.click(thumbnail)
  const dialog = await screen.findByRole('dialog', { name: 'Hero image' })
  expect(dialog).toBeVisible()
  expect(
    await screen.findByRole('img', { name: 'Hero image Preview' })
  ).toHaveAttribute('src', 'https://example.com/original.jpg')
  expect(getMediaAssetPreview).toHaveBeenCalledWith('asset-image', 'original')
})

test('focused image thumbnail opens the preview with Enter', async () => {
  renderLibrary()
  const user = userEvent.setup()
  const thumbnail = await screen.findByRole('button', {
    name: 'Preview Hero image',
  })

  thumbnail.focus()
  await user.keyboard('{Enter}')

  expect(
    await screen.findByRole('dialog', { name: 'Hero image' })
  ).toBeVisible()
})

test('a failed OSS thumbnail falls back to the original signed image', async () => {
  renderLibrary()

  const thumbnail = await screen.findByRole('button', {
    name: 'Preview Hero image',
  })
  await waitFor(() =>
    expect(thumbnail.querySelector('img')).toHaveAttribute(
      'src',
      'https://example.com/thumbnail.jpg'
    )
  )
  fireEvent.error(thumbnail.querySelector('img') as HTMLImageElement)

  await waitFor(() =>
    expect(
      screen
        .getByRole('button', { name: 'Preview Hero image' })
        .querySelector('img')
    ).toHaveAttribute('src', 'https://example.com/original.jpg')
  )
  expect(getMediaAssetPreview).toHaveBeenCalledWith('asset-image', 'original')
})

test.each([
  { type: 'video' as const, name: 'Demo clip', element: 'video' },
  { type: 'audio' as const, name: 'Narration', element: 'audio' },
])(
  'opening a $type asset preview shows full-size media controls',
  async ({ type, name, element }) => {
    vi.mocked(listMediaAssets).mockResolvedValue({
      success: true,
      data: {
        items: [{ ...asset, id: `asset-${type}`, name, type }],
        total: 1,
        page: 1,
        page_size: 40,
      },
    })
    renderLibrary()

    await userEvent
      .setup()
      .click(await screen.findByRole('button', { name: 'Preview asset' }))

    const dialog = await screen.findByRole('dialog', { name })
    await waitFor(() =>
      expect(dialog.querySelector(element)).toHaveAttribute(
        'src',
        'https://example.com/original.jpg'
      )
    )
    expect(dialog.querySelector(element)).toHaveAttribute('controls')
    expect(getMediaAssetPreview).toHaveBeenCalledWith(
      `asset-${type}`,
      'original'
    )
  }
)

test('downloading an asset obtains a separate original-byte attachment URL', async () => {
  let openedURL = ''
  vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(
    function (this: HTMLAnchorElement) {
      openedURL = this.href
    }
  )
  renderLibrary()

  await userEvent
    .setup()
    .click(await screen.findByRole('button', { name: 'Download asset' }))

  await waitFor(() =>
    expect(getMediaAssetPreview).toHaveBeenCalledWith('asset-image', 'download')
  )
  await waitFor(() =>
    expect(openedURL).toBe('https://example.com/download.jpg')
  )
})

test('a failed download does not navigate and shows an error', async () => {
  vi.mocked(getMediaAssetPreview).mockImplementation(async (_id, variant) => {
    if (variant === 'download') throw new Error('Signed URL unavailable')
    return {
      success: true,
      data: { url: 'https://example.com/preview.jpg', expires_at: 4102444800 },
    }
  })
  const click = vi
    .spyOn(HTMLAnchorElement.prototype, 'click')
    .mockImplementation(() => {})
  renderLibrary()

  await userEvent
    .setup()
    .click(await screen.findByRole('button', { name: 'Download asset' }))

  await waitFor(() =>
    expect(toast.error).toHaveBeenCalledWith('Signed URL unavailable')
  )
  expect(click).not.toHaveBeenCalled()
})
