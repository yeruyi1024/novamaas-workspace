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
import { beforeEach, describe, expect, test, vi } from 'vitest'

import { api } from '@/lib/api'

import {
  canGetTaskInformation,
  getLogRequestBody,
  getTaskInformation,
} from '../task-content-api'

vi.mock('@/lib/api', () => ({
  api: { get: vi.fn() },
}))

describe('task information API', () => {
  beforeEach(() => {
    vi.mocked(api.get).mockReset()
    vi.mocked(api.get).mockResolvedValue({ data: { status: 'succeeded' } })
  })

  test.each([
    ['17', '/v1/video/generations/task%2F1'],
    ['61', '/api/v3/contents/generations/tasks/task%2F1'],
    ['54', '/v1/video/generations/task%2F1'],
  ])(
    'uses the platform endpoint for channel type %s',
    async (platform, path) => {
      await getTaskInformation('task/1', platform)

      expect(api.get).toHaveBeenCalledWith(path, {
        disableDuplicate: true,
        skipErrorHandler: true,
      })
    }
  )

  test('rejects unsupported task platforms before making a request', async () => {
    expect(canGetTaskInformation('other')).toBe(false)
    await expect(getTaskInformation('task_1', 'other')).rejects.toThrow(
      'task information unsupported'
    )
    expect(api.get).not.toHaveBeenCalled()
  })

  test('looks up archived log request bodies by task and request identifiers', async () => {
    await getLogRequestBody('task/1', 'request-1')

    expect(api.get).toHaveBeenCalledWith('/api/log/request-body', {
      params: { task_id: 'task/1', request_id: 'request-1' },
      disableDuplicate: true,
    })
  })
})
