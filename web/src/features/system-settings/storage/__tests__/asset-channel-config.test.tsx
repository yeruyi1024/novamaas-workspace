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
import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, test, vi } from 'vitest'

import { AssetLibrarySettings } from '../asset-library-settings'

vi.mock('@/features/assets/api', () => ({
  listAssetRequestLogs: vi.fn(async () => ({
    success: true,
    data: { items: [], next_cursor: '', dropped_on_this_node: 0 },
  })),
}))
vi.mock('../api', () => ({
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
        channel_type: 47,
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
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

let queryClient: QueryClient

afterEach(() => queryClient?.clear())

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
  test('shows the channel ID next to the channel name', async () => {
    renderSettings()

    expect(await screen.findByText('Seedance upstream')).toBeVisible()
    expect(screen.getByText('ID 7')).toBeVisible()
  })

  test('opens request logs scoped to the selected asset channel', async () => {
    renderSettings()

    fireEvent.click(await screen.findByRole('button', { name: 'View logs' }))
    expect(
      await screen.findByRole('dialog', { name: 'Request logs' })
    ).toBeVisible()
    expect(screen.getByText('Channel ID · 7')).toBeVisible()
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
})
