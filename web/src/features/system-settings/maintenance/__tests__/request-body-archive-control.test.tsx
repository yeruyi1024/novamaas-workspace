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
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, test, vi } from 'vitest'

import {
  getCurrentRequestBodyArchiveTask,
  getRequestBodyArchiveTask,
  startRequestBodyArchiveTask,
} from '../../api'
import type { RequestBodyArchiveTask } from '../../types'
import { RequestBodyArchiveControl } from '../request-body-archive-control'

vi.mock('../../api', () => ({
  getCurrentRequestBodyArchiveTask: vi.fn(),
  getRequestBodyArchiveTask: vi.fn(),
  startRequestBodyArchiveTask: vi.fn(),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

const runningTask: RequestBodyArchiveTask = {
  id: 1,
  task_id: 'systask_archive',
  type: 'request_body_archive',
  status: 'running',
  created_at: 100,
  updated_at: 100,
  payload: { batch_size: 100 },
  state: {
    initialized: true,
    phase: 'tasks',
    total: 10,
    processed: 2,
    progress: 20,
    archived_count: 2,
    task_rows_cleaned: 1,
    log_rows_cleaned: 1,
    invalid_rows_skipped: 0,
  },
}

describe('RequestBodyArchiveControl', () => {
  beforeEach(() => {
    vi.mocked(getCurrentRequestBodyArchiveTask).mockReset()
    vi.mocked(getRequestBodyArchiveTask).mockReset()
    vi.mocked(startRequestBodyArchiveTask).mockReset()
    vi.mocked(getCurrentRequestBodyArchiveTask).mockResolvedValue({
      success: true,
      message: '',
      data: null,
    })
  })

  test('requires confirmation and disables the action after starting', async () => {
    vi.mocked(startRequestBodyArchiveTask).mockResolvedValue({
      success: true,
      message: '',
      data: runningTask,
    })

    render(<RequestBodyArchiveControl />)
    fireEvent.click(
      screen.getByRole('button', { name: 'Scan and archive history' })
    )
    expect(
      screen.getByRole('heading', {
        name: 'Archive historical request bodies?',
      })
    ).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Start archive' }))

    await waitFor(() =>
      expect(startRequestBodyArchiveTask).toHaveBeenCalledOnce()
    )
    expect(screen.getByRole('button', { name: 'Archiving...' })).toBeDisabled()
    expect(screen.getByText('20%')).toBeInTheDocument()
    expect(screen.getByText('2 of 10 rows scanned.')).toBeInTheDocument()
  })

  test('restores an active task and prevents a duplicate start', async () => {
    vi.mocked(getCurrentRequestBodyArchiveTask).mockResolvedValue({
      success: true,
      message: '',
      data: runningTask,
    })

    render(<RequestBodyArchiveControl />)

    const button = await screen.findByRole('button', { name: 'Archiving...' })
    expect(button).toBeDisabled()
    expect(startRequestBodyArchiveTask).not.toHaveBeenCalled()
  })
})
