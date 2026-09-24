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
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'

import {
  deleteAssetGroup,
  getMediaAssetPreview,
  listAssetGroups,
  listAssetGroupsPage,
  listMediaAssets,
} from '../api'
import { AssetLibrary } from '../asset-library'
import { AssetUploadDialog } from '../components/asset-upload-dialog'

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

function renderLibrary() {
  queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
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
    data: [
      {
        id: 'group-1',
        owner_user_id: 1,
        owner_name: 'member',
        name: 'Campaign assets',
        description: '',
        status: 'ready',
        created_at: 1,
        updated_at: 1,
      },
    ],
  })
  vi.mocked(listAssetGroupsPage).mockResolvedValue({
    success: true,
    data: {
      items: [
        {
          id: 'group-1',
          owner_user_id: 1,
          owner_name: 'member',
          name: 'Campaign assets',
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
  })
  vi.mocked(listMediaAssets).mockResolvedValue({
    success: true,
    data: {
      items: [
        {
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
        },
      ],
      total: 1,
      page: 1,
      page_size: 40,
    },
  })
  vi.mocked(getMediaAssetPreview).mockResolvedValue({
    success: true,
    data: { url: 'https://example.com/hero.png', expires_at: 100 },
  })
})

afterEach(() => {
  queryClient?.clear()
  vi.unstubAllGlobals()
})

describe('asset library business layout', () => {
  test('admin asset cards show group name and ID without creator metadata', async () => {
    useAuthStore
      .getState()
      .auth.setUser({ id: 99, username: 'admin', role: 10 })
    renderLibrary()

    const groupLine = await screen.findByText(/Campaign assets · ID: group-1/)
    expect(groupLine).toBeVisible()
    expect(groupLine).not.toHaveTextContent('Creator')
    expect(groupLine).not.toHaveTextContent('KiB')
  })

  test('uses a compact asset grid without exposing synchronization controls', async () => {
    renderLibrary()

    expect(await screen.findByText('Hero image')).toBeVisible()
    expect(screen.getByTestId('asset-grid')).toHaveAttribute(
      'data-density',
      'compact'
    )
    expect(
      screen.queryByRole('button', { name: /sync/i })
    ).not.toBeInTheDocument()
    expect(
      screen.queryByText('No asset channels are enabled')
    ).not.toBeInTheDocument()
    expect(screen.getByText(/asset-owner/)).toBeVisible()
    expect(screen.getByText(/Uploaded at/)).toBeVisible()
    expect(screen.getByRole('button', { name: 'AK/SK access' })).toBeVisible()
    expect(
      screen
        .getByRole('button', { name: 'Download asset' })
        .closest('[data-slot="card-footer"]')
    ).toHaveClass('flex-wrap')
    const groupSelect = screen.getByRole('combobox', { name: 'Asset group' })
    expect(groupSelect).toHaveTextContent('All groups')
    await userEvent.setup().click(groupSelect)
    expect(
      await screen.findByRole('option', {
        name: /Campaign assets.*ID: group-1/,
      })
    ).toBeVisible()
    expect(listAssetGroupsPage).toHaveBeenCalledWith(
      expect.objectContaining({
        includeAllOwners: false,
        page: 1,
        pageSize: 20,
      })
    )
    expect(listMediaAssets).toHaveBeenCalledWith(
      expect.objectContaining({ includeAllOwners: false })
    )
  })

  test('shows every creator to admins and loads all assets for the selected group', async () => {
    useAuthStore
      .getState()
      .auth.setUser({ id: 99, username: 'admin', role: 10 })
    vi.mocked(listAssetGroupsPage).mockImplementation(async (params) => ({
      success: true,
      data: params.includeAllOwners
        ? {
            items: [
              {
                id: 'group-1',
                owner_user_id: 1,
                owner_name: 'alice',
                name: 'Campaign assets',
                description: '',
                status: 'ready',
                created_at: 2,
                updated_at: 2,
              },
              {
                id: 'group-20260922095429-zkvzd',
                owner_user_id: 2,
                owner_name: 'admin',
                name: 'Campaign assets',
                description: '',
                status: 'ready',
                created_at: 1,
                updated_at: 1,
              },
            ],
            total: 2,
            page: 1,
            page_size: 20,
          }
        : {
            items: [
              {
                id: 'group-20260922095429-zkvzd',
                owner_user_id: 2,
                owner_name: 'admin',
                name: 'Campaign assets',
                description: '',
                status: 'ready',
                created_at: 1,
                updated_at: 1,
              },
            ],
            total: 1,
            page: 1,
            page_size: 1,
          },
    }))
    renderLibrary()

    const groupSelect = await screen.findByRole('combobox', {
      name: 'Asset group',
    })
    const user = userEvent.setup()
    await user.click(groupSelect)
    expect(
      await screen.findByRole('option', {
        name: /Campaign assets.*ID: group-1.*Creator: alice/,
      })
    ).toBeVisible()
    await user.click(
      screen.getByRole('option', {
        name: /Campaign assets.*ID: group-20260922095429-zkvzd.*Creator: admin/,
      })
    )

    expect(groupSelect).toHaveTextContent('Campaign assets')
    expect(groupSelect).toHaveTextContent('ID: group-20260922095429-zkvzd')
    expect(groupSelect).toHaveTextContent('Creator: admin')
    expect(screen.getByTestId('selected-asset-group-name')).toHaveClass(
      'truncate'
    )
    const selectedMetadata = screen.getByTestId('selected-asset-group-metadata')
    expect(groupSelect).toContainElement(selectedMetadata)
    expect(selectedMetadata).toHaveClass(
      'flex',
      'items-baseline',
      'overflow-hidden',
      'whitespace-nowrap'
    )
    expect(selectedMetadata).toHaveAttribute(
      'aria-label',
      'ID: group-20260922095429-zkvzd · Creator: admin'
    )
    expect(screen.getByTestId('selected-asset-group-id')).not.toHaveClass(
      'flex-1'
    )

    await waitFor(() =>
      expect(listMediaAssets).toHaveBeenLastCalledWith(
        expect.objectContaining({
          groupId: 'group-20260922095429-zkvzd',
          includeAllOwners: true,
        })
      )
    )
    expect(listAssetGroupsPage).toHaveBeenCalledWith(
      expect.objectContaining({ includeAllOwners: true })
    )
  })

  test('searches paginated groups and preserves a selection across pages', async () => {
    vi.mocked(listAssetGroupsPage).mockImplementation(async (params) => ({
      success: true,
      data: {
        items: [
          {
            id: params.search ? 'group-beta' : `group-page-${params.page}`,
            owner_user_id: 1,
            owner_name: 'member',
            name: params.search ? 'Beta' : `Group page ${params.page}`,
            description: '',
            status: 'ready',
            created_at: params.page,
            updated_at: params.page,
          },
        ],
        total: params.search ? 1 : 21,
        page: params.page,
        page_size: 20,
      },
    }))
    renderLibrary()

    const user = userEvent.setup()
    const groupSelect = screen.getByRole('combobox', { name: 'Asset group' })
    await user.click(groupSelect)
    await user.click(await screen.findByRole('button', { name: 'Next' }))
    expect(
      await screen.findByRole('option', { name: /Group page 2/ })
    ).toBeVisible()
    await user.click(screen.getByRole('option', { name: /Group page 2/ }))
    expect(groupSelect).toHaveTextContent('Group page 2')

    await user.click(groupSelect)
    await user.type(
      screen.getByRole('textbox', { name: 'Search groups...' }),
      'Be'
    )
    await waitFor(() =>
      expect(listAssetGroupsPage).toHaveBeenLastCalledWith(
        expect.objectContaining({ page: 1, search: 'Be' })
      )
    )
    expect(await screen.findByRole('option', { name: /Beta/ })).toBeVisible()
    expect(groupSelect).toHaveTextContent('Group page 2')
  })

  test('places empty-group deletion in the page actions and confirms it', async () => {
    vi.mocked(listMediaAssets).mockResolvedValue({
      success: true,
      data: { items: [], total: 0, page: 1, page_size: 40 },
    })
    vi.mocked(deleteAssetGroup).mockResolvedValue({
      success: true,
      data: undefined,
    })
    renderLibrary()

    const user = userEvent.setup()
    await user.click(
      await screen.findByRole('combobox', { name: 'Asset group' })
    )
    await user.click(
      await screen.findByRole('option', {
        name: /Campaign assets.*ID: group-1/,
      })
    )
    const deleteButton = await screen.findByRole('button', {
      name: 'Delete group',
    })
    expect(deleteButton).toBeVisible()
    expect(screen.queryByText('Delete empty group')).not.toBeInTheDocument()
    fireEvent.click(deleteButton)
    expect(
      await screen.findByText(
        'Are you sure you want to delete group "Campaign assets"? This action cannot be undone.'
      )
    ).toBeVisible()
    fireEvent.click(screen.getByRole('button', { name: 'Delete' }))

    await waitFor(() =>
      expect(deleteAssetGroup).toHaveBeenCalledWith('group-1')
    )
  })

  test('keeps the grid bounded with server pagination', async () => {
    vi.mocked(listMediaAssets).mockResolvedValueOnce({
      success: true,
      data: {
        items: [],
        total: 41,
        page: 1,
        page_size: 40,
      },
    })
    renderLibrary()

    const nextButton = await screen.findByRole('button', { name: 'Next' })
    fireEvent.click(nextButton)

    await waitFor(() =>
      expect(listMediaAssets).toHaveBeenLastCalledWith(
        expect.objectContaining({ page: 2, pageSize: 40 })
      )
    )
  })

  test('requests a preview only when its card approaches the viewport', async () => {
    let intersectionCallback: IntersectionObserverCallback | undefined
    const disconnect = vi.fn()
    vi.stubGlobal(
      'IntersectionObserver',
      class {
        constructor(callback: IntersectionObserverCallback) {
          intersectionCallback = callback
        }

        observe() {}
        disconnect() {
          disconnect()
        }
      }
    )
    renderLibrary()

    expect(await screen.findByText('Hero image')).toBeVisible()
    expect(getMediaAssetPreview).not.toHaveBeenCalled()

    act(() => {
      intersectionCallback?.(
        [{ isIntersecting: true } as IntersectionObserverEntry],
        {} as IntersectionObserver
      )
    })

    await waitFor(() => expect(getMediaAssetPreview).toHaveBeenCalledOnce())
    expect(disconnect).toHaveBeenCalled()
  })

  test('infers the asset type from the selected file before upload', async () => {
    const onSubmit = vi.fn()
    render(
      <AssetUploadDialog
        open
        onOpenChange={vi.fn()}
        groups={[
          {
            id: 'group-1',
            owner_user_id: 1,
            owner_name: 'member',
            name: 'Campaign assets',
            description: '',
            status: 'ready',
            created_at: 1,
            updated_at: 1,
          },
        ]}
        selectedGroup='group-1'
        onSubmit={onSubmit}
        pending={false}
        progress={0}
      />
    )

    const fileInput = await waitFor(() => {
      const input = document.querySelector('input[type="file"]')
      expect(input).toBeInstanceOf(HTMLInputElement)
      return input
    })
    expect(fileInput).toBeInstanceOf(HTMLInputElement)
    const file = new File(['png'], 'sample.png', { type: 'image/png' })
    fireEvent.change(fileInput as HTMLInputElement, {
      target: { files: [file] },
    })
    expect(await screen.findByText('sample.png')).toBeVisible()

    fireEvent.click(screen.getByRole('button', { name: 'Upload' }))
    await waitFor(() => expect(onSubmit).toHaveBeenCalledOnce())
    const formData = onSubmit.mock.calls[0][0] as FormData
    expect(formData.get('type')).toBe('image')
    expect(formData.get('file')).toBe(file)
  })
})
