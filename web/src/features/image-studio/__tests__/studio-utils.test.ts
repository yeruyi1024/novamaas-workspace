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
import { AxiosError, AxiosHeaders } from 'axios'
import { afterEach, expect, test, vi } from 'vitest'

import { api } from '@/lib/api'

import { pollStudioRuns } from '../api'
import { studioImageSize, studioPollInterval } from '../studio-utils'

vi.mock('@/lib/api', () => ({ api: { get: vi.fn() } }))
afterEach(() => {
  vi.resetAllMocks()
  vi.restoreAllMocks()
})

test.each([
  ['a landscape, 16:9', '1792x1024'],
  ['竖版海报，比例 9：16', '1024x1792'],
  ['square 1:1', '1024x1024'],
  ['portrait 2 / 3', '1024x1536'],
  ['landscape 3:2', '1536x1024'],
  ['change 16:9 to 9:16', '1024x1792'],
  ['unsupported 4:3', undefined],
  ['a numbered item 19:16', undefined],
  ['a landscape', undefined],
])('auto size parses supported ratios safely from %s', (prompt, size) => {
  expect(studioImageSize(prompt, 'auto')).toBe(size)
})

test('a manually selected size takes priority over the prompt ratio', () => {
  expect(studioImageSize('portrait 9:16', '1024x1024')).toBe('1024x1024')
})

test('normal batch polling leaves room for other API requests', () => {
  expect(studioPollInterval(null)).toBe(3000)
  expect(studioPollInterval(null, 4)).toBe(6000)
})

test.each([
  ['180', 180000],
  ['invalid', 30000],
  ['Mon, 19 Oct 2026 00:01:00 GMT', 60000],
])(
  '429 polling honors Retry-After %s with interval %s ms',
  (retryAfter, interval) => {
    vi.spyOn(Date, 'now').mockReturnValue(Date.UTC(2026, 9, 19, 0, 0, 0))
    const error = new AxiosError(
      'rate limited',
      undefined,
      undefined,
      undefined,
      {
        status: 429,
        statusText: '',
        data: '',
        headers: new AxiosHeaders({ 'retry-after': retryAfter }),
        config: { headers: new AxiosHeaders() },
      }
    )
    expect(studioPollInterval(error)).toBe(interval)
  }
)

test('a batch beyond 50 tasks uses bounded requests and no individual detail queries', async () => {
  vi.mocked(api.get).mockResolvedValue({
    data: { success: true, data: { items: [], total: 0 } },
  })
  const ids = Array.from({ length: 51 }, (_, index) => `run-${index}`)
  await pollStudioRuns(ids)
  expect(api.get).toHaveBeenCalledTimes(2)
  expect(api.get).toHaveBeenNthCalledWith(1, '/api/image-studio/runs/poll', {
    params: { ids: ids.slice(0, 50).join(',') },
    skipBusinessError: true,
    skipErrorHandler: true,
  })
  expect(api.get).toHaveBeenNthCalledWith(2, '/api/image-studio/runs/poll', {
    params: { ids: 'run-50' },
    skipBusinessError: true,
    skipErrorHandler: true,
  })
})
