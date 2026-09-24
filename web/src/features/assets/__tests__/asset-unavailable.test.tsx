import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'

import {
  getMediaAssetPreview,
  listAssetGroups,
  listAssetGroupsPage,
  listMediaAssets,
  uploadMediaAsset,
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
const rejectedAsset: MediaAsset = {
  id: 'asset-20260922120000-old',
  group_id: 'group-1',
  group_name: 'Campaign',
  owner_user_id: 1,
  owner_name: 'member',
  name: 'Hero image',
  type: 'image',
  content_type: 'image/png',
  size: 12,
  sha256: 'checksum',
  status: 'unavailable',
  unavailable_reason: 'sensitive_content',
  created_at: 1,
  updated_at: 2,
}

beforeEach(() => {
  vi.clearAllMocks()
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
    data: { items: [rejectedAsset], total: 1, page: 1, page_size: 40 },
  })
  vi.mocked(getMediaAssetPreview).mockResolvedValue({
    success: true,
    data: { url: 'https://example.com/hero.png', expires_at: 100 },
  })
  Object.defineProperty(URL, 'createObjectURL', {
    configurable: true,
    value: vi.fn(() => 'blob:new-asset'),
  })
  Object.defineProperty(URL, 'revokeObjectURL', {
    configurable: true,
    value: vi.fn(),
  })
})

afterEach(() => queryClient.clear())

test('a rejected asset in a mixed row keeps its reason out of the card and reveals it from the thumbnail status', async () => {
  vi.mocked(listMediaAssets).mockResolvedValue({
    success: true,
    data: {
      items: [
        rejectedAsset,
        {
          ...rejectedAsset,
          id: 'asset-20260922120001-ready',
          name: 'Ready image',
          status: 'ready',
          unavailable_reason: undefined,
        },
      ],
      total: 2,
      page: 1,
      page_size: 40,
    },
  })
  render(
    <QueryClientProvider client={queryClient}>
      <AssetLibrary />
    </QueryClientProvider>
  )
  const user = userEvent.setup()

  expect(await screen.findByText('Ready image')).toBeVisible()
  const grid = screen.getByTestId('asset-grid')
  expect(grid.querySelectorAll('[data-slot="card"]')).toHaveLength(2)
  expect(within(grid).queryByText(/Sensitive content was rejected/)).toBeNull()
  expect(grid.querySelector('[data-slot="alert"]')).toBeNull()
  const status = within(grid).getByRole('button', {
    name: 'Asset unavailable: View details',
  })
  expect(status).toHaveClass('absolute')

  await user.click(status)
  const dialog = screen.getByRole('dialog', { name: 'Hero image' })
  expect(
    within(dialog).getByText(/Sensitive content was rejected/)
  ).toBeVisible()
  expect(
    within(dialog).getByText(/Upload revised material and use its new asset ID/)
  ).toBeVisible()
  await user.click(
    within(dialog).getByRole('button', { name: 'Re-upload revised file' })
  )
  expect(
    screen.getByRole('dialog', { name: 'Re-upload revised file' })
  ).toBeVisible()
})

test('a non-owner can inspect the rejection reason but cannot re-upload the asset', async () => {
  vi.mocked(listMediaAssets).mockResolvedValue({
    success: true,
    data: {
      items: [{ ...rejectedAsset, owner_user_id: 2 }],
      total: 1,
      page: 1,
      page_size: 40,
    },
  })
  render(
    <QueryClientProvider client={queryClient}>
      <AssetLibrary />
    </QueryClientProvider>
  )
  const user = userEvent.setup()

  const status = await screen.findByRole('button', {
    name: 'Asset unavailable: View details',
  })
  await user.click(status)
  const dialog = screen.getByRole('dialog', { name: 'Hero image' })
  expect(
    within(dialog).getByText(/Sensitive content was rejected/)
  ).toBeVisible()
  expect(
    within(dialog).queryByRole('button', { name: 'Re-upload revised file' })
  ).toBeNull()
  expect(
    screen.queryByRole('button', { name: 'Re-upload revised file' })
  ).toBeNull()
})

test('rejected asset offers compact re-upload without copying its old ID', async () => {
  vi.mocked(uploadMediaAsset).mockResolvedValue({
    success: true,
    data: {
      ...rejectedAsset,
      id: 'asset-20260922120000-new',
      status: 'ready',
      unavailable_reason: undefined,
    },
  })
  render(
    <QueryClientProvider client={queryClient}>
      <AssetLibrary />
    </QueryClientProvider>
  )
  const user = userEvent.setup()

  expect(await screen.findByText('Hero image')).toBeVisible()
  expect(
    screen.queryByRole('button', { name: 'Copy asset reference' })
  ).not.toBeInTheDocument()
  const reupload = screen.getByRole('button', {
    name: 'Re-upload revised file',
  })
  expect(reupload).toHaveTextContent('Re-upload')
  await user.click(reupload)
  expect(
    screen.getByRole('dialog', { name: 'Re-upload revised file' })
  ).toBeVisible()
  expect(screen.getByText(/A new asset ID will be created/)).toBeVisible()
  expect(screen.getByRole('textbox', { name: 'Name' })).toHaveValue(
    'Hero image'
  )

  await user.upload(
    screen.getByLabelText('File'),
    new File(['image'], 'revised.png', { type: 'image/png' })
  )
  await user.click(screen.getByRole('button', { name: 'Upload' }))
  await waitFor(() => expect(uploadMediaAsset).toHaveBeenCalled())
  const formData = vi.mocked(uploadMediaAsset).mock.lastCall?.[0]
  expect(formData?.get('group_id')).toBe('group-1')
  expect(formData?.get('name')).toBe('Hero image')
})
