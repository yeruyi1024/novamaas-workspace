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
import { afterEach, describe, expect, test, vi } from 'vitest'

import {
  listAssetLibraryConnections,
  listAssetChannelConfigs,
  saveAssetLibraryConnection,
  bindAssetChannelLibrary,
  mergeAssetLibraries,
  syncAssetLibraryMissing,
  deleteAssetLibraryConnection,
} from '../api'
import { AssetLibrarySettings } from '../asset-library-settings'

const { libraryFixture } = vi.hoisted(() => ({
  libraryFixture: {
    id: 17,
    name: 'Production account',
    provider_name: 'Example provider',
    channel_id: 0,
    channel_name: '',
    channel_type: 0,
    enabled: true,
    protocol: 'volc_action' as const,
    auth_type: 'ak_sk' as const,
    base_url: 'https://ark.cn-beijing.volcengineapi.com',
    region: 'cn-beijing',
    service: 'ark',
    api_version: '2024-01-01',
    project_name: 'default',
    qpm: 60,
    access_key_hint: '****test',
    credential_configured: true,
    updated_at: 1,
    channels: [{ id: 7, name: 'Seedance upstream', enabled: true }],
    summary: { total: 1, active: 1, pending: 0, processing: 0, failed: 0 },
    missing: 0,
  },
}))

vi.mock('@/features/assets/api', () => ({
  listAssetRequestLogs: vi.fn(async () => ({
    success: true,
    data: { items: [], next_cursor: '', dropped_on_this_node: 0 },
  })),
}))
vi.mock('../api', () => ({
  listAssetLibraryConnections: vi.fn(async () => ({
    success: true,
    data: [libraryFixture],
  })),
  saveAssetLibraryConnection: vi.fn(async () => ({
    success: true,
    data: libraryFixture,
  })),
  bindAssetChannelLibrary: vi.fn(async () => ({ success: true, data: {} })),
  mergeAssetLibraries: vi.fn(async () => ({
    success: true,
    data: { merged_groups: 2, deferred_groups: 0, unreadable_accounts: 0 },
  })),
  syncAssetLibraryMissing: vi.fn(async () => ({
    success: true,
    data: undefined,
  })),
  deleteAssetLibraryConnection: vi.fn(async () => ({
    success: true,
    data: undefined,
  })),
  testAssetLibraryConnection: vi.fn(),
  getAssetLibraryStoragePolicy: vi.fn(async () => ({
    success: true,
    data: {
      id: 1,
      key: 'asset_library',
      name: 'Asset library permanent storage',
      purpose: 'asset_library',
      storage_profile_id: 1,
      object_prefix: 'assets/library',
      signed_url_ttl_seconds: 86400,
      retention_seconds: 0,
      max_file_bytes: 536870912,
      max_total_bytes: 536870912,
      max_files: 100,
      allowed_mime_types: 'image/png',
      enabled: true,
      created_at: 1,
      updated_at: 1,
    },
  })),
  listAssetChannelConfigs: vi.fn(async () => ({
    success: true,
    data: [
      {
        channel_id: 7,
        channel_name: 'Seedance upstream',
        channel_type: 61,
        library_id: 17,
        library_name: 'Production account',
        enabled: false,
        protocol: 'volc_action',
        auth_type: 'ak_sk',
        base_url: 'https://ark.cn-beijing.volcengineapi.com',
        region: 'cn-beijing',
        service: 'ark',
        api_version: '2024-01-01',
        project_name: 'default',
        qpm: 60,
        access_key_hint: '',
        credential_configured: false,
        updated_at: 1,
      },
    ],
  })),
  listStorageProfiles: vi.fn(async () => ({
    success: true,
    data: [
      {
        id: 1,
        name: 'Primary OSS',
        provider_type: 'aliyun_oss',
        status: 1,
        endpoint: 'https://oss-cn-hangzhou.aliyuncs.com',
        region: 'oss-cn-hangzhou',
        bucket: 'private-assets',
        auth_type: 'static_access_key',
        credential_configured: true,
        access_key_hint: '****1234',
        created_at: 1,
        updated_at: 1,
      },
    ],
  })),
  testAssetChannelConfig: vi.fn(),
  updateAssetChannelConfig: vi.fn(),
  updateAssetLibraryStoragePolicy: vi.fn(),
}))
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}))

let queryClient: QueryClient

afterEach(() => {
  queryClient?.clear()
  vi.clearAllMocks()
})

function renderSettings() {
  queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
  render(
    <QueryClientProvider client={queryClient}>
      <AssetLibrarySettings />
    </QueryClientProvider>
  )
}

describe('asset channel protocol configuration', () => {
  test('linked accounts do not expose account deletion', async () => {
    renderSettings()
    await screen.findByRole('button', { name: 'Manage channels' })
    expect(
      screen.queryByRole('button', { name: 'Delete account and copies' })
    ).not.toBeInTheDocument()
  })

  test.each([0, 1])(
    'deleting accounts freeze configuration and allow retry only after a failure (%i)',
    async (failed) => {
      vi.mocked(listAssetLibraryConnections).mockResolvedValueOnce({
        success: true,
        data: [
          {
            ...libraryFixture,
            channels: [],
            delete_requested_at: 1,
            summary: { ...libraryFixture.summary, failed },
          },
        ],
      })
      renderSettings()
      expect(
        await screen.findByRole('button', { name: 'Manage channels' })
      ).toBeDisabled()
      expect(screen.getByRole('button', { name: 'Configure' })).toBeDisabled()
      const deletion = screen.getByRole('button', {
        name: 'Delete account and copies',
      })
      if (failed > 0) {
        expect(deletion).toBeEnabled()
      } else {
        expect(deletion).toBeDisabled()
      }
    }
  )

  test('deleting an unlinked account requires confirmation and submits only its account ID', async () => {
    vi.mocked(listAssetLibraryConnections).mockResolvedValueOnce({
      success: true,
      data: [{ ...libraryFixture, channels: [] }],
    })
    renderSettings()
    fireEvent.click(
      await screen.findByRole('button', { name: 'Delete account and copies' })
    )
    const dialog = await screen.findByRole('dialog', {
      name: 'Delete account and copies',
    })
    expect(
      within(dialog).getByText(
        /Platform assets and upstream groups are retained/
      )
    ).toBeInTheDocument()
    expect(deleteAssetLibraryConnection).not.toHaveBeenCalled()
    fireEvent.click(within(dialog).getByRole('button', { name: 'Confirm' }))
    await waitFor(() =>
      expect(deleteAssetLibraryConnection).toHaveBeenCalledWith(17)
    )
    expect(deleteAssetLibraryConnection).toHaveBeenCalledTimes(1)
    await waitFor(() =>
      expect(
        screen.queryByRole('dialog', { name: 'Delete account and copies' })
      ).not.toBeInTheDocument()
    )
  })

  test('failed deletion requests keep the confirmation open for retry', async () => {
    vi.mocked(listAssetLibraryConnections).mockResolvedValueOnce({
      success: true,
      data: [{ ...libraryFixture, channels: [] }],
    })
    vi.mocked(deleteAssetLibraryConnection).mockRejectedValueOnce(
      new Error('deletion rejected')
    )
    renderSettings()
    fireEvent.click(
      await screen.findByRole('button', { name: 'Delete account and copies' })
    )
    const dialog = await screen.findByRole('dialog', {
      name: 'Delete account and copies',
    })
    fireEvent.click(within(dialog).getByRole('button', { name: 'Confirm' }))
    await waitFor(() =>
      expect(deleteAssetLibraryConnection).toHaveBeenCalledTimes(1)
    )
    expect(dialog).toBeInTheDocument()
    await waitFor(() =>
      expect(
        within(dialog).getByRole('button', { name: 'Confirm' })
      ).toBeEnabled()
    )
  })

  test('removes an association independently of unreadable account credentials', async () => {
    renderSettings()
    fireEvent.click(
      await screen.findByRole('button', { name: 'Manage channels' })
    )
    const dialog = await screen.findByRole('dialog', {
      name: 'Manage channels',
    })
    const channel = within(dialog).getByRole('checkbox', {
      name: 'Seedance upstream · ID 7',
    })
    expect(channel).toBeChecked()
    fireEvent.click(channel)
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save' }))
    await waitFor(() =>
      expect(bindAssetChannelLibrary).toHaveBeenCalledWith(7, 0, false)
    )
    expect(saveAssetLibraryConnection).not.toHaveBeenCalled()
  })
  test('disables channels owned by another account while keeping its own and unbound channels selectable', async () => {
    const channelDefaults = {
      channel_name: 'video channel',
      channel_type: 61,
      enabled: true,
      protocol: 'volc_action' as const,
      auth_type: 'ak_sk' as const,
      base_url: 'https://provider.example',
      region: 'cn-beijing',
      service: 'ark',
      api_version: '2024-01-01',
      project_name: 'default',
      qpm: 60,
      access_key_hint: '',
      credential_configured: false,
      updated_at: 1,
    }
    vi.mocked(listAssetChannelConfigs).mockResolvedValueOnce({
      success: true,
      data: [
        {
          ...channelDefaults,
          channel_id: 7,
          channel_name: 'Own channel',
          library_id: 17,
          library_name: 'Production account',
        },
        {
          ...channelDefaults,
          channel_id: 8,
          channel_name: 'Other account channel',
          library_id: 18,
          library_name: 'Other account',
          enabled: false,
          library_enabled: false,
        },
        {
          ...channelDefaults,
          channel_id: 9,
          channel_name: 'Unbound channel',
          library_id: 0,
        },
      ],
    })
    renderSettings()
    fireEvent.click(
      await screen.findByRole('button', { name: 'Manage channels' })
    )
    const dialog = await screen.findByRole('dialog', {
      name: 'Manage channels',
    })
    expect(
      within(dialog).getByRole('checkbox', { name: 'Own channel · ID 7' })
    ).toBeEnabled()
    const occupied = within(dialog).getByRole('checkbox', {
      name: /Other account channel/,
    })
    expect(occupied).toHaveAttribute('aria-disabled', 'true')
    fireEvent.click(occupied)
    expect(occupied).toHaveAttribute('aria-checked', 'false')
    const unbound = within(dialog).getByRole('checkbox', {
      name: 'Unbound channel · ID 9',
    })
    expect(unbound).not.toHaveAttribute('aria-disabled', 'true')
    fireEvent.click(unbound)
    expect(unbound).toHaveAttribute('aria-checked', 'true')
    expect(
      within(dialog).getByText('Currently linked to Other account')
    ).toBeVisible()
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save' }))
    await waitFor(() =>
      expect(bindAssetChannelLibrary).toHaveBeenCalledWith(9, 17)
    )
    expect(bindAssetChannelLibrary).toHaveBeenCalledTimes(1)
  })

  test('lays out independent resource accounts as compact squares with scrollable channel lists', async () => {
    const longChannelName =
      'Production asset channel with a long account and pricing group name'
    vi.mocked(listAssetLibraryConnections).mockResolvedValueOnce({
      success: true,
      data: [
        {
          ...libraryFixture,
          channels: [
            { id: 7, name: longChannelName, enabled: true },
            { id: 8, name: 'Second pricing channel', enabled: true },
            { id: 9, name: 'Third pricing channel', enabled: true },
            { id: 10, name: 'Fourth pricing channel', enabled: true },
            { id: 11, name: 'Fifth pricing channel', enabled: true },
            { id: 12, name: 'Sixth pricing channel', enabled: false },
          ],
        },
        {
          ...libraryFixture,
          id: 18,
          name: 'Second account',
          provider_name: 'Second provider',
        },
      ],
    })
    renderSettings()
    const provider = await screen.findByRole('region', {
      name: 'Production account',
    })
    expect(provider).toHaveClass('aspect-square', 'max-w-md', 'min-h-0')
    expect(
      screen.getByRole('group', { name: 'Provider asset libraries' })
    ).toHaveClass(
      'grid',
      'grid-cols-[repeat(auto-fill,minmax(min(100%,22rem),1fr))]'
    )
    const content = within(provider).getByRole('group', {
      name: 'Linked channels',
    })
    expect(content).toHaveClass('min-h-0', 'flex-1', 'overflow-hidden')
    const channels = within(provider).getByRole('list', {
      name: 'Linked channels',
    })
    expect(channels).toHaveClass(
      'flex-1',
      'min-h-0',
      'overflow-y-auto',
      'overscroll-contain'
    )
    expect(within(channels).getAllByRole('listitem')).toHaveLength(6)
    expect(
      within(channels).getByTitle(`${longChannelName} · ID 7`)
    ).toHaveClass('truncate')
    channels.focus()
    expect(channels).toHaveFocus()
    expect(
      within(provider).queryByRole('button', { name: 'Add resource account' })
    ).not.toBeInTheDocument()
    expect(
      screen.getAllByRole('button', { name: 'Add resource account' })
    ).toHaveLength(1)
  })

  test('shows the channel ID next to the channel name', async () => {
    renderSettings()

    expect(await screen.findByText('Example provider')).toBeVisible()
    expect(screen.getByText('Seedance upstream · ID 7')).toBeVisible()
  })

  test('opens request logs scoped to the selected asset channel', async () => {
    renderSettings()

    fireEvent.click(await screen.findByRole('button', { name: 'View logs' }))
    expect(
      await screen.findByRole('dialog', { name: 'Request logs' })
    ).toBeVisible()
    expect(screen.getByText('Upstream asset library ID · 17')).toBeVisible()
  })

  test('switching to YooFang REST selects Bearer sk authentication and hides Action fields', async () => {
    renderSettings()

    fireEvent.click(await screen.findByRole('button', { name: 'Configure' }))
    fireEvent.change(await screen.findByLabelText('Provider protocol'), {
      target: { value: 'youfang_rest' },
    })

    expect(screen.getByLabelText('Authentication')).toBeDisabled()
    expect(screen.getByLabelText('Authentication')).toHaveValue('bearer')
    expect(screen.getByLabelText('YooFang sk key')).toBeVisible()
    expect(screen.getByLabelText('Asset API base URL')).toHaveValue(
      'https://asset-inference-doubao.yoofang.com'
    )
    expect(screen.queryByLabelText('Access Key ID')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Region')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('API version')).not.toBeInTheDocument()
  })

  test('shows separate cards for two accounts from the same provider', async () => {
    vi.mocked(listAssetLibraryConnections).mockResolvedValueOnce({
      success: true,
      data: [
        libraryFixture,
        {
          ...libraryFixture,
          id: 18,
          name: 'Second account',
          channels: [],
          summary: {
            total: 0,
            active: 0,
            pending: 0,
            processing: 0,
            failed: 0,
          },
        },
      ],
    })
    renderSettings()
    const first = await screen.findByRole('region', {
      name: 'Production account',
    })
    const second = screen.getByRole('region', { name: 'Second account' })
    expect(within(first).getByText('Example provider')).toBeVisible()
    expect(within(second).getByText('Example provider')).toBeVisible()
    expect(
      within(first).getByRole('heading', { name: 'Production account' })
    ).toBeVisible()
    expect(
      within(second).getByRole('heading', { name: 'Second account' })
    ).toBeVisible()
  })

  test('opens a blank independent resource account from the global add action', async () => {
    renderSettings()
    fireEvent.click(
      await screen.findByRole('button', { name: 'Add resource account' })
    )
    const dialog = await screen.findByRole('dialog', {
      name: 'Configure upstream asset library',
    })
    expect(within(dialog).getByLabelText('Provider name')).toHaveValue('')
    expect(
      within(dialog).getByLabelText('Resource account / library name')
    ).toHaveValue('')
    expect(within(dialog).queryByRole('checkbox')).not.toBeInTheDocument()
  })

  test('saving account parameters does not change associations or start synchronization', async () => {
    renderSettings()
    fireEvent.click(await screen.findByRole('button', { name: 'Configure' }))
    fireEvent.click(await screen.findByRole('button', { name: 'Save' }))
    await waitFor(() =>
      expect(saveAssetLibraryConnection).toHaveBeenCalledWith(
        17,
        expect.objectContaining({
          provider_name: 'Example provider',
          name: 'Production account',
        })
      )
    )
    expect(
      vi.mocked(saveAssetLibraryConnection).mock.calls[0]?.[1]
    ).not.toHaveProperty('channel_ids')
    expect(syncAssetLibraryMissing).not.toHaveBeenCalled()
  })

  test('saving unchanged associations keeps a disabled channel linked', async () => {
    vi.mocked(listAssetLibraryConnections).mockResolvedValueOnce({
      success: true,
      data: [
        {
          ...libraryFixture,
          channels: [{ id: 7, name: 'Seedance upstream', enabled: false }],
        },
      ],
    })
    renderSettings()
    fireEvent.click(
      await screen.findByRole('button', { name: 'Manage channels' })
    )
    expect(await screen.findByRole('checkbox')).toBeChecked()
    fireEvent.click(screen.getByRole('button', { name: 'Save' }))
    await waitFor(() =>
      expect(
        screen.queryByRole('dialog', { name: 'Manage channels' })
      ).not.toBeInTheDocument()
    )
    expect(bindAssetChannelLibrary).not.toHaveBeenCalled()
  })

  test('missing synchronization requires reviewing the count before submitting once', async () => {
    vi.mocked(listAssetLibraryConnections).mockResolvedValueOnce({
      success: true,
      data: [{ ...libraryFixture, missing: 2 }],
    })
    renderSettings()
    fireEvent.click(
      await screen.findByRole('button', { name: 'Sync missing assets' })
    )
    expect(syncAssetLibraryMissing).not.toHaveBeenCalled()
    const dialog = await screen.findByRole('dialog', {
      name: 'Sync missing assets',
    })
    expect(within(dialog).getByText(/Sync 2 missing assets/)).toBeVisible()
    fireEvent.click(within(dialog).getByRole('button', { name: 'Confirm' }))
    await waitFor(() =>
      expect(syncAssetLibraryMissing).toHaveBeenCalledWith(17)
    )
  })

  test('merging identical accounts needs the administrator to confirm', async () => {
    renderSettings()
    fireEvent.click(
      await screen.findByRole('button', { name: 'Merge identical accounts' })
    )
    expect(mergeAssetLibraries).not.toHaveBeenCalled()
    const dialog = await screen.findByRole('dialog', {
      name: 'Merge identical accounts',
    })
    expect(within(dialog).getByText(/nothing is deleted/)).toBeVisible()
    fireEvent.click(within(dialog).getByRole('button', { name: 'Confirm' }))
    await waitFor(() => expect(mergeAssetLibraries).toHaveBeenCalledTimes(1))
  })

  test('accounts retired by a merge are collapsed and stay reachable', async () => {
    vi.mocked(listAssetLibraryConnections).mockResolvedValueOnce({
      success: true,
      data: [{ ...libraryFixture, channels: [], legacy_channel_id: 7 }],
    })
    renderSettings()
    const summary = await screen.findByText('Unlinked / historical library · 1')
    fireEvent.click(summary)
    expect(
      screen.getByText(
        'Historical assets and logs are retained. Link channels to use this library.'
      )
    ).toBeVisible()
  })
})
