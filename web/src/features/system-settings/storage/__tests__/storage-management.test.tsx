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

import { StorageManagement } from '../storage-management'

vi.mock('../api', () => ({
  archiveStorageProfile: vi.fn(),
  createStorageProfile: vi.fn(),
  getRelayMediaStoragePolicy: vi.fn(async () => ({
    success: true,
    data: {
      id: 0,
      key: 'relay_media_temp',
      name: 'Relay media temporary storage',
      purpose: 'relay_media_temp',
      storage_profile_id: 0,
      object_prefix: 'temporary/relay-media',
      signed_url_ttl_seconds: 259200,
      retention_seconds: 259200,
      max_file_bytes: 10485760,
      max_total_bytes: 20971520,
      max_files: 10,
      allowed_mime_types:
        'image/jpeg,image/png,image/webp,video/mp4,video/webm,video/quicktime',
      enabled: false,
      created_at: 0,
      updated_at: 0,
    },
  })),
  listStorageProfiles: vi.fn(async () => ({ success: true, data: [] })),
  testSavedStorageProfile: vi.fn(),
  testStorageProfile: vi.fn(),
  updateRelayMediaStoragePolicy: vi.fn(),
  updateStorageProfile: vi.fn(),
}))

let queryClient: QueryClient

afterEach(() => queryClient?.clear())

describe('storage management', () => {
  test('opens the create profile dialog from the page action', async () => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    })

    render(
      <QueryClientProvider client={queryClient}>
        <StorageManagement />
      </QueryClientProvider>
    )

    fireEvent.click(await screen.findByRole('button', { name: 'Add profile' }))

    expect(await screen.findByRole('dialog')).toBeVisible()
    expect(screen.getByText('Add storage profile')).toBeVisible()
  })
})
