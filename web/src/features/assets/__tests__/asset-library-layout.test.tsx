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
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { toast } from 'sonner'
import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'

import {
  deleteAssetGroup,
  deleteMediaAssets,
  deleteMediaAssetsByFilter,
  getMediaAssetPreview,
  listAssetGroups,
  listAssetGroupsPage,
  listAssetOwners,
  listMediaAssets,
  previewMediaAssets,
  summarizeMediaAssets,
  uploadMediaAsset,
} from '../api'
import { AssetLibrary } from '../asset-library'
import { AssetUploadDialog } from '../components/asset-upload-dialog'
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
  purgeUserAssets: vi.fn(),
  rotateAssetWebhookEndpointSecret: vi.fn(),
  testAssetWebhookEndpoint: vi.fn(),
  uploadMediaAsset: vi.fn(),
}))
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))
vi.mock('@/features/webhooks/api', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/features/webhooks/api')>()),
  getWebhookCapabilities: vi.fn(async () => ({
    success: true,
    data: {
      asset_library_enabled: false,
      media_tasks_enabled: false,
      manual_enabled: false,
    },
  })),
}))

let queryClient: QueryClient

function renderLibrary(props?: ComponentProps<typeof AssetLibrary>) {
  queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
  render(
    <QueryClientProvider client={queryClient}>
      <AssetLibrary {...props} />
    </QueryClientProvider>
  )
}

function asset(overrides: Partial<MediaAsset>): MediaAsset {
  return {
    id: 'asset-20260922120000-abcde',
    group_id: 'group-1',
    group_name: 'Campaign assets',
    owner_user_id: 1,
    owner_name: 'asset-owner',
    name: 'Hero image',
    type: 'image',
    content_type: 'image/png',
    size: 2048,
    sha256: 'checksum',
    status: 'ready',
    created_at: 1,
    updated_at: 1,
    ...overrides,
  }
}

function assetPage(items: MediaAsset[], total = items.length) {
  return {
    success: true,
    data: { items, total, page: 1, page_size: 24 },
  }
}

const defaultGroup = {
  id: 'group-1',
  owner_user_id: 1,
  owner_name: 'member',
  name: 'Campaign assets',
  description: '',
  status: 'ready',
  created_at: 1,
  updated_at: 1,
}

beforeEach(() => {
  vi.clearAllMocks()
  mockAssetPageApis()
  useAuthStore.getState().auth.reset()
  useAuthStore.getState().auth.setUser({ id: 1, username: 'member', role: 1 })
  Object.defineProperty(URL, 'createObjectURL', {
    configurable: true,
    value: vi.fn(() => 'blob:test-preview'),
  })
  Object.defineProperty(URL, 'revokeObjectURL', {
    configurable: true,
    value: vi.fn(),
  })
  vi.mocked(listAssetGroups).mockResolvedValue({
    success: true,
    data: [defaultGroup],
  })
  vi.mocked(listAssetGroupsPage).mockResolvedValue({
    success: true,
    data: { items: [defaultGroup], total: 1, page: 1, page_size: 20 },
  })
  vi.mocked(listMediaAssets).mockResolvedValue(assetPage([asset({})]))
  vi.mocked(getMediaAssetPreview).mockResolvedValue({
    success: true,
    data: { url: 'https://example.com/hero.png', expires_at: 100 },
  })
})

afterEach(() => {
  queryClient?.clear()
  vi.unstubAllGlobals()
})

describe('asset library layout, filters, and group management', () => {
  test.each([
    { role: 1, visible: false },
    { role: 10, visible: true },
  ])(
    'header exposes flat actions and role $role admin tools ($visible)',
    async ({ role, visible }) => {
      useAuthStore
        .getState()
        .auth.setUser({ id: 99, username: 'operator', role })
      renderLibrary()

      expect(
        await screen.findByRole('button', { name: 'AK/SK access' })
      ).toBeVisible()
      expect(
        screen.queryByRole('button', { name: 'More actions' })
      ).not.toBeInTheDocument()
      if (!visible) {
        expect(
          screen.queryByRole('button', { name: 'Copy asset groups' })
        ).not.toBeInTheDocument()
        expect(
          screen.queryByRole('button', { name: 'Clear user assets' })
        ).not.toBeInTheDocument()
        return
      }
      expect(
        screen.getByRole('button', { name: 'Clear user assets' })
      ).toBeVisible()
      await userEvent.click(
        screen.getByRole('button', { name: 'Copy asset groups' })
      )
      expect(
        await screen.findByRole('dialog', { name: 'Copy asset groups' })
      ).toBeVisible()
    }
  )

  test('renders card audit metadata and inline thumbnails/video posters without batch preview calls', async () => {
    useAuthStore
      .getState()
      .auth.setUser({ id: 99, username: 'admin', role: 100 })
    vi.mocked(listMediaAssets).mockResolvedValue(
      assetPage([
        asset({
          id: 'asset-fast-1',
          name: 'Fast Hero',
          owner_user_id: 7,
          owner_name: 'alice',
          thumbnail_url: 'https://oss.example.com/fast-1.webp',
          thumbnail_expires_at: 4102444800,
        }),
        asset({
          id: 'asset-fast-video',
          name: 'Fast Video',
          type: 'video',
          thumbnail_url: 'https://oss.example.com/fast-video.mp4',
          poster_url: 'https://oss.example.com/fast-video-poster.jpg',
          thumbnail_expires_at: 4102444800,
        }),
      ])
    )
    renderLibrary()

    const card = await screen.findByRole('article', { name: 'Fast Hero' })
    expect(within(card).getByText('asset-fast-1')).toBeVisible()
    expect(within(card).getByText('Campaign assets')).toBeVisible()
    expect(within(card).getByText('alice')).toBeVisible()
    expect(within(card).getByText('ID: 7')).toBeVisible()
    const img = within(card).getByRole('img', { name: 'Fast Hero' })
    expect(img).toHaveAttribute('src', 'https://oss.example.com/fast-1.webp')
    expect(img).toHaveAttribute('loading', 'eager')
    expect(previewMediaAssets).not.toHaveBeenCalled()

    const videoCard = screen.getByRole('article', { name: 'Fast Video' })
    expect(
      within(videoCard).getByRole('img', { name: 'Fast Video' })
    ).toHaveAttribute('src', 'https://oss.example.com/fast-video-poster.jpg')
    expect(videoCard.querySelector('video')).toBeNull()
  })

  test('applies type/status/sort filters, displays removable chips, and clears all filters', async () => {
    renderLibrary()
    const user = userEvent.setup()
    await screen.findByText('Hero image')

    await user.click(screen.getByRole('radio', { name: /Videos/ }))
    await waitFor(() =>
      expect(listMediaAssets).toHaveBeenLastCalledWith(
        expect.objectContaining({ type: 'video', page: 1 })
      )
    )
    expect(
      screen.getByRole('button', { name: 'Remove filter Videos' })
    ).toBeVisible()

    await user.click(screen.getByRole('button', { name: 'Clear filters' }))
    await waitFor(() =>
      expect(
        screen.queryByRole('button', { name: 'Remove filter Videos' })
      ).not.toBeInTheDocument()
    )
  })

  test('admins disambiguate same-named groups and identical asset IDs across owners', async () => {
    useAuthStore
      .getState()
      .auth.setUser({ id: 99, username: 'admin', role: 100 })
    vi.mocked(listAssetGroupsPage).mockResolvedValue({
      success: true,
      data: {
        items: [
          {
            ...defaultGroup,
            id: 'group-shared',
            owner_user_id: 2,
            owner_name: 'alice',
          },
          {
            ...defaultGroup,
            id: 'group-shared',
            owner_user_id: 3,
            owner_name: 'bob',
          },
        ],
        total: 2,
        page: 1,
        page_size: 20,
      },
    })
    vi.mocked(listMediaAssets).mockResolvedValue(
      assetPage([
        asset({
          id: 'asset-same',
          owner_user_id: 2,
          owner_name: 'alice',
          name: 'Alice Photo',
        }),
        asset({
          id: 'asset-same',
          owner_user_id: 3,
          owner_name: 'bob',
          name: 'Bob Photo',
        }),
      ])
    )
    vi.mocked(deleteMediaAssets).mockResolvedValue({
      success: true,
      data: { deleted_count: 1 },
    })
    renderLibrary()
    const user = userEvent.setup()

    await user.click(
      await screen.findByRole('combobox', { name: 'Asset group' })
    )
    await user.click(
      await screen.findByRole('option', {
        name: /Campaign assets .* Creator: bob/,
      })
    )
    await waitFor(() =>
      expect(listMediaAssets).toHaveBeenLastCalledWith(
        expect.objectContaining({ groupId: 'group-shared', ownerUserId: 3 })
      )
    )

    await user.click(screen.getByRole('checkbox', { name: 'Select Bob Photo' }))
    const toolbar = await screen.findByRole('toolbar', {
      name: 'Selected assets',
    })
    await user.click(within(toolbar).getByRole('button', { name: 'Delete' }))
    const dialog = await screen.findByRole('alertdialog')
    await user.click(within(dialog).getByRole('button', { name: 'Delete' }))

    await waitFor(() =>
      expect(deleteMediaAssets).toHaveBeenCalledWith([
        { id: 'asset-same', owner_user_id: 3 },
      ])
    )
  })

  test('deleting a non-empty group requires explicit acknowledgement', async () => {
    vi.mocked(listAssetGroupsPage).mockResolvedValue({
      success: true,
      data: {
        items: [{ ...defaultGroup, asset_count: 2 }],
        total: 1,
        page: 1,
        page_size: 20,
      },
    })
    vi.mocked(deleteAssetGroup).mockResolvedValue({
      success: true,
      data: undefined,
    })
    renderLibrary({ filters: { group: 'group-1' } })
    const user = userEvent.setup()

    await user.click(
      await screen.findByRole('button', { name: 'Delete group' })
    )

    const dialog = await screen.findByRole('alertdialog', {
      name: 'Delete group',
    })
    const confirm = within(dialog).getByRole('button', { name: 'Delete' })
    expect(confirm).toBeDisabled()

    await user.click(
      within(dialog).getByRole('checkbox', {
        name: 'Also delete the 2 assets in this group.',
      })
    )
    expect(confirm).toBeEnabled()
    await user.click(confirm)

    await waitFor(() =>
      expect(deleteAssetGroup).toHaveBeenCalledWith('group-1', undefined, {
        withAssets: true,
      })
    )
  })

  test('uploads queued files sequentially and keeps selected files when groups arrive asynchronously', async () => {
    const fileA = new File(['a'], 'a.png', { type: 'image/png' })
    const fileB = new File(['b'], 'b.png', { type: 'image/png' })
    const onSubmit = vi.fn()
    const { rerender } = render(
      <AssetUploadDialog
        open
        onOpenChange={vi.fn()}
        groups={[]}
        selectedGroup=''
        onSubmit={onSubmit}
        pending={false}
        progress={0}
      />
    )
    const input = screen.getByLabelText('File') as HTMLInputElement
    fireEvent.change(input, { target: { files: [fileA, fileB] } })
    expect(screen.getByText('a.png')).toBeVisible()
    expect(screen.getByText('b.png')).toBeVisible()

    rerender(
      <AssetUploadDialog
        open
        onOpenChange={vi.fn()}
        groups={[defaultGroup]}
        selectedGroup=''
        onSubmit={onSubmit}
        pending={false}
        progress={0}
      />
    )
    await userEvent.click(
      screen.getByRole('button', { name: 'Upload 2 files' })
    )
    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1))
    expect(onSubmit.mock.calls[0][0]).toHaveLength(2)

    vi.mocked(uploadMediaAsset)
      .mockResolvedValueOnce({ success: true, data: asset({ id: 'a' }) })
      .mockRejectedValueOnce(new Error('network'))
    renderLibrary()
    const user = userEvent.setup()
    await user.click(
      await screen.findByRole('button', { name: 'Upload asset' })
    )
    const dialog = await screen.findByRole('dialog', { name: 'Upload asset' })
    fireEvent.change(within(dialog).getByLabelText('File'), {
      target: { files: [fileA, fileB] },
    })
    await user.click(
      within(dialog).getByRole('button', { name: 'Upload 2 files' })
    )
    await waitFor(() => expect(toast.error).toHaveBeenCalled())
    expect(toast.success).toHaveBeenCalled()
  })
})

describe('netdisk selection, unified purge, and review regressions', () => {
  test('supports Shift range select, Ctrl+A, and Escape shortcuts', async () => {
    vi.mocked(listMediaAssets).mockResolvedValue(
      assetPage([
        asset({ id: 'asset-1', name: 'One' }),
        asset({ id: 'asset-2', name: 'Two' }),
        asset({ id: 'asset-3', name: 'Three' }),
      ])
    )
    renderLibrary()
    await screen.findByText('One')

    fireEvent.click(screen.getByRole('checkbox', { name: 'Select One' }))
    fireEvent.click(screen.getByRole('checkbox', { name: 'Select Three' }), {
      shiftKey: true,
    })
    expect(screen.getByRole('checkbox', { name: 'Select Two' })).toBeChecked()

    fireEvent.keyDown(window, { key: 'Escape' })
    expect(
      screen.queryByRole('toolbar', { name: 'Selected assets' })
    ).not.toBeInTheDocument()

    fireEvent.keyDown(window, { key: 'a', ctrlKey: true })
    expect(
      await screen.findByRole('toolbar', { name: 'Selected assets' })
    ).toBeVisible()
  })

  test('netdisk selection strip promotes to all matching assets with exclusions and deletes via filter endpoint on Delete key', async () => {
    vi.mocked(deleteMediaAssetsByFilter).mockResolvedValue({
      success: true,
      data: { deleted_count: 59, owner_user_ids: [1] },
    })
    vi.mocked(listMediaAssets).mockResolvedValue(
      assetPage(
        [
          asset({ id: 'asset-1', name: 'First' }),
          asset({ id: 'asset-2', name: 'Second' }),
        ],
        60
      )
    )
    renderLibrary()
    const user = userEvent.setup()

    const strip = await screen.findByTestId('asset-selection-strip')
    await user.click(
      within(strip).getByRole('checkbox', { name: 'Select page' })
    )
    await user.click(
      await within(strip).findByRole('button', {
        name: 'Select all 60 matching assets',
      })
    )
    expect(
      within(strip).getByText('All 60 matching assets selected')
    ).toBeVisible()

    await user.click(screen.getByRole('checkbox', { name: 'Select First' }))
    expect(
      within(strip).getByText('All 59 matching assets selected')
    ).toBeVisible()

    fireEvent.keyDown(window, { key: 'Delete' })
    const dialog = await screen.findByRole('alertdialog', {
      name: 'Delete all 59 matching assets?',
    })
    const confirmButton = within(dialog).getByRole('button', { name: 'Delete' })
    expect(confirmButton).toBeDisabled()
    await user.click(
      within(dialog).getByRole('checkbox', {
        name: 'I confirm permanently deleting these 59 assets.',
      })
    )
    await user.click(confirmButton)

    await waitFor(() =>
      expect(deleteMediaAssetsByFilter).toHaveBeenCalledWith(
        expect.objectContaining({
          exclude: [{ id: 'asset-1', owner_user_id: 1 }],
        })
      )
    )
  })

  test('selecting an owner shows user context bar and pre-fills unified user cleanup dialog with unavailable-only mode', async () => {
    useAuthStore
      .getState()
      .auth.setUser({ id: 99, username: 'admin', role: 100 })
    vi.mocked(listAssetOwners).mockResolvedValue({
      success: true,
      data: {
        items: [
          { user_id: 7, username: 'alice', group_count: 3, asset_count: 12 },
        ],
        total: 1,
        page: 1,
        page_size: 20,
      },
    })
    vi.mocked(summarizeMediaAssets).mockResolvedValue({
      success: true,
      data: {
        total: 12,
        by_type: { image: 10, video: 2, audio: 0 },
        by_status: { ready: 9, processing: 0, unavailable: 3 },
      },
    })
    vi.mocked(deleteMediaAssetsByFilter).mockResolvedValue({
      success: true,
      data: { deleted_count: 3, owner_user_ids: [7] },
    })
    renderLibrary()
    const user = userEvent.setup()

    await user.click(
      await screen.findByRole('combobox', { name: 'Asset owner' })
    )
    await user.click(await screen.findByRole('option', { name: /alice/ }))

    const ownerHeader = await screen.findByRole('region', {
      name: 'User asset summary',
    })
    expect(within(ownerHeader).getByText('alice')).toBeVisible()
    expect(within(ownerHeader).getByText('3 unavailable')).toBeVisible()

    await user.click(
      within(ownerHeader).getByRole('button', {
        name: 'Clear this user’s assets',
      })
    )
    const dialog = await screen.findByRole('dialog', {
      name: 'Clear user assets',
    })
    expect(
      await within(dialog).findByTestId('purge-impact-summary')
    ).toHaveTextContent('12 assets')
    await user.click(
      within(dialog).getByRole('radio', {
        name: /Clear unavailable assets only/,
      })
    )
    await user.click(
      within(dialog).getByRole('button', { name: 'Clear unavailable assets' })
    )
    const confirm = await screen.findByRole('alertdialog', {
      name: 'Clear unavailable user assets?',
    })
    await user.click(
      within(confirm).getByRole('button', { name: 'Clear unavailable assets' })
    )
    await waitFor(() =>
      expect(deleteMediaAssetsByFilter).toHaveBeenCalledWith({
        scope: 'all',
        owner_user_id: 7,
        status: 'unavailable',
      })
    )
  })

  test('switching filters hides previous cards and clears their selection even when returning to the original filter', async () => {
    let resolveVideos:
      | ((value: ReturnType<typeof assetPage>) => void)
      | undefined
    vi.mocked(listMediaAssets).mockImplementation((params) => {
      if (params?.type === 'video') {
        return new Promise((resolve) => {
          resolveVideos = resolve
        })
      }
      return Promise.resolve(
        assetPage([asset({ id: 'asset-img-1', name: 'Stale Image' })])
      )
    })
    renderLibrary()
    const user = userEvent.setup()

    expect(
      await screen.findByRole('article', { name: 'Stale Image' })
    ).toBeVisible()
    await user.click(
      screen.getByRole('checkbox', { name: 'Select Stale Image' })
    )

    await user.click(screen.getByRole('radio', { name: /Videos/ }))
    await waitFor(() =>
      expect(
        screen.queryByRole('article', { name: 'Stale Image' })
      ).not.toBeInTheDocument()
    )
    expect(
      screen.queryByRole('checkbox', { name: 'Select Stale Image' })
    ).not.toBeInTheDocument()

    resolveVideos?.(
      assetPage([
        asset({ id: 'asset-vid-1', name: 'Fresh Video', type: 'video' }),
      ])
    )
    expect(
      await screen.findByRole('article', { name: 'Fresh Video' })
    ).toBeVisible()
    await user.click(
      screen.getByRole('button', { name: 'Remove filter Videos' })
    )
    expect(
      await screen.findByRole('checkbox', { name: 'Select Stale Image' })
    ).not.toBeChecked()
  })

  test('deselecting current page after selecting all matching assets preserves selections on other pages', async () => {
    vi.mocked(listMediaAssets).mockResolvedValue(
      assetPage(
        [
          asset({ id: 'asset-1', name: 'First' }),
          asset({ id: 'asset-2', name: 'Second' }),
        ],
        60
      )
    )
    renderLibrary()
    const user = userEvent.setup()

    const strip = await screen.findByTestId('asset-selection-strip')
    await user.click(
      within(strip).getByRole('checkbox', { name: 'Select page' })
    )
    await user.click(
      await within(strip).findByRole('button', {
        name: 'Select all 60 matching assets',
      })
    )
    await user.click(
      within(strip).getByRole('checkbox', {
        name: 'All 60 matching assets selected',
      })
    )
    expect(
      within(strip).getByText('All 58 matching assets selected')
    ).toBeVisible()
    expect(
      screen.getByRole('checkbox', { name: 'Select First' })
    ).not.toBeChecked()

    await user.click(
      within(strip).getByRole('checkbox', {
        name: 'All 58 matching assets selected',
      })
    )
    expect(
      within(strip).getByText('All 60 matching assets selected')
    ).toBeVisible()
  })

  test('retrying a partially failed chunked batch delete only submits remaining assets', async () => {
    const items = Array.from({ length: 101 }, (_, index) =>
      asset({ id: `asset-chunk-${index + 1}`, name: `Item ${index + 1}` })
    )
    vi.mocked(listMediaAssets).mockResolvedValue(assetPage(items, 101))
    vi.mocked(deleteMediaAssets)
      .mockResolvedValueOnce({ success: true, data: { deleted_count: 100 } })
      .mockRejectedValueOnce(new Error('temporary error'))
      .mockResolvedValueOnce({ success: true, data: { deleted_count: 1 } })
    renderLibrary()
    const user = userEvent.setup()

    const strip = await screen.findByTestId('asset-selection-strip')
    await user.click(
      within(strip).getByRole('checkbox', { name: 'Select page' })
    )
    const toolbar = await screen.findByRole('toolbar', {
      name: 'Selected assets',
    })
    await user.click(within(toolbar).getByRole('button', { name: 'Delete' }))
    const dialog = await screen.findByRole('alertdialog', {
      name: 'Delete 101 assets?',
    })
    await user.click(
      within(dialog).getByRole('checkbox', {
        name: 'I confirm permanently deleting these 101 assets.',
      })
    )
    await user.click(within(dialog).getByRole('button', { name: 'Delete' }))

    await waitFor(() =>
      expect(within(strip).getByText('1 selected')).toBeVisible()
    )
    expect(deleteMediaAssets).toHaveBeenCalledTimes(2)

    await user.click(within(dialog).getByRole('button', { name: 'Delete' }))
    await waitFor(() => expect(deleteMediaAssets).toHaveBeenCalledTimes(3))
    expect(vi.mocked(deleteMediaAssets).mock.calls[2][0]).toEqual([
      { id: 'asset-chunk-101', owner_user_id: 1 },
    ])
  })
})
