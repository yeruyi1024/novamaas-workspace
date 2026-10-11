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
import { afterEach, expect, test, vi } from 'vitest'

import { api } from '@/lib/api'

import { createStudioRun, getStudioSettings } from '../api'

vi.mock('@/lib/api', () => ({ api: { get: vi.fn(), post: vi.fn() } }))
afterEach(() => vi.resetAllMocks())

test('settings return data and reject unsuccessful business responses', async () => {
  const settings = {
    enabled: true,
    max_images: 4,
    max_reference_images: 2,
    max_prompt_characters: 1000,
  }
  vi.mocked(api.get)
    .mockResolvedValueOnce({ data: { success: true, data: settings } })
    .mockResolvedValueOnce({
      data: { success: false, message: 'Studio unavailable' },
    })
  await expect(getStudioSettings()).resolves.toEqual(settings)
  await expect(getStudioSettings()).rejects.toThrow('Studio unavailable')
})

test('generation preserves the default request and idempotency header and rejects business failures', async () => {
  const input = {
    token_id: 7,
    idempotency_key: 'retry-key',
    request: { model: 'gpt-image-1', prompt: 'mountain' },
    input_asset_ids: [],
  }
  vi.mocked(api.post).mockResolvedValue({
    data: { success: false, message: 'Provider unavailable' },
  })
  await expect(createStudioRun(input)).rejects.toThrow('Provider unavailable')
  expect(api.post).toHaveBeenCalledWith('/api/image-studio/runs', input, {
    timeout: 300_000,
    headers: { 'Idempotency-Key': 'retry-key' },
    skipBusinessError: true,
    skipErrorHandler: true,
  })
})
