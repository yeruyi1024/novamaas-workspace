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
import { describe, expect, test, vi } from 'vitest'

import { getAllTaskLogs, getUserTaskLogs } from '../../api'
import { buildSearchParams } from '../filter'
import { fetchLogsByCategory } from '../utils'

vi.mock('../../api', () => ({
  getAllTaskLogs: vi.fn(),
  getUserTaskLogs: vi.fn(),
}))

describe('task log filters', () => {
  test('passes username and model through the admin query', async () => {
    const search = buildSearchParams(
      { taskId: 'task-1', username: 'alice', model: 'video-model' },
      'task'
    )
    vi.mocked(getAllTaskLogs).mockResolvedValue({ success: true })

    await fetchLogsByCategory({
      logCategory: 'task',
      isAdmin: true,
      page: 2,
      pageSize: 20,
      searchParams: search,
      columnFilters: [],
    })

    expect(getAllTaskLogs).toHaveBeenCalledWith(
      expect.objectContaining({
        p: 2,
        task_id: 'task-1',
        username: 'alice',
        model_name: 'video-model',
      })
    )
  })

  test('keeps username outside the self query while filtering its models', async () => {
    vi.mocked(getUserTaskLogs).mockResolvedValue({ success: true })

    await fetchLogsByCategory({
      logCategory: 'task',
      isAdmin: false,
      page: 1,
      pageSize: 20,
      searchParams: { username: 'another-user', model: 'video-model' },
      columnFilters: [],
    })

    expect(getUserTaskLogs).toHaveBeenCalledWith(
      expect.objectContaining({
        model_name: 'video-model',
        username: undefined,
      })
    )
  })
})
