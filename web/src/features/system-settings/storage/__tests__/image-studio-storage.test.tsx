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
import { expect, test, vi } from 'vitest'

import * as api from '../api'
import { ImageStudioStorageSettings } from '../image-studio-storage-settings'
import type { StoragePolicy, StorageProfile } from '../types'

vi.mock('../api')

test('each image policy saves its profile and retention independently', async () => {
  const policies: StoragePolicy[] = [
    'image_studio',
    'image_studio_reference',
  ].map((purpose, index) => ({
    id: index + 1,
    key: purpose,
    purpose,
    name: purpose,
    enabled: true,
    storage_profile_id: 1,
    object_prefix: `images/${purpose}`,
    signed_url_ttl_seconds: 3600,
    retention_seconds: (index ? 7 : 30) * 86400,
    max_file_bytes: 20 * 1024 * 1024,
    max_total_bytes: 20 * 1024 * 1024,
    max_files: 1,
    allowed_mime_types: 'image/jpeg,image/png,image/webp',
    created_at: 1,
    updated_at: 1,
  }))
  vi.mocked(api.getImageStudioStoragePolicies).mockResolvedValue({
    success: true,
    data: policies,
  })
  vi.mocked(api.listStorageProfiles).mockResolvedValue({
    success: true,
    data: [
      { id: 1, name: 'Original OSS', status: 1 },
      { id: 2, name: 'Studio OSS', status: 1 },
    ] as StorageProfile[],
  })
  vi.mocked(api.updateImageStudioStoragePolicy).mockResolvedValue({
    success: true,
    data: policies[0],
  })
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  const view = render(
    <QueryClientProvider client={client}>
      <ImageStudioStorageSettings />
    </QueryClientProvider>
  )
  try {
    const generated = within(
      await screen.findByRole('form', { name: 'Generated image' })
    )
    const reference = within(
      screen.getByRole('form', { name: 'Reference images' })
    )
    fireEvent.change(generated.getByLabelText('Storage profile'), {
      target: { value: '2' },
    })
    fireEvent.change(generated.getByLabelText('Maximum retention (hours)'), {
      target: { value: '1440' },
    })
    fireEvent.click(
      generated.getByRole('button', { name: 'Save storage policy' })
    )
    await waitFor(() =>
      expect(api.updateImageStudioStoragePolicy).toHaveBeenCalledWith(
        'image_studio',
        expect.objectContaining({
          storage_profile_id: 2,
          retention_seconds: 60 * 86400,
          signed_url_ttl_seconds: 3600,
        })
      )
    )
    expect(reference.getByLabelText('Storage profile')).toHaveValue('1')
    expect(reference.getByLabelText('Maximum retention (hours)')).toHaveValue(
      168
    )
  } finally {
    view.unmount()
    client.clear()
    vi.resetAllMocks()
  }
})
