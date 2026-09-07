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
  downloadTaskVideo,
  getTaskVideoContentInfo,
  getTaskRequestBody,
} from '../../../task-content-api'
import type { TaskLog } from '../../../types'
import { TaskLogDetailsCell } from '../task-log-details-cell'

vi.mock('../../../task-content-api', () => ({
  downloadTaskVideo: vi.fn(),
  getTaskVideoContentInfo: vi.fn(),
  getTaskRequestBody: vi.fn(),
}))

vi.mock('@/components/ai-elements/code-block', () => ({
  CodeBlock: ({ code }: { code: string }) => (
    <pre data-testid='request-json'>{code}</pre>
  ),
  CodeBlockCopyButton: () => null,
}))

const successfulVideoLog: TaskLog = {
  id: 1,
  user_id: 2,
  platform: '54',
  task_id: 'task_video',
  action: 'generate',
  channel_id: 3,
  submit_time: 1,
  status: 'SUCCESS',
  result_url: 'https://cdn.example.com/video.mp4',
  request_body_available: true,
}

describe('TaskLogDetailsCell', () => {
  beforeEach(() => {
    vi.mocked(getTaskRequestBody).mockReset()
    vi.mocked(downloadTaskVideo).mockReset()
    vi.mocked(getTaskVideoContentInfo).mockReset()
  })

  test('shows the formatted request JSON regardless of task status', async () => {
    vi.mocked(getTaskRequestBody).mockResolvedValue({
      success: true,
      data: { prompt: 'hello', parameters: { duration: 5 } },
    })

    render(
      <TaskLogDetailsCell
        isAdmin={false}
        log={{
          ...successfulVideoLog,
          status: 'FAILURE',
          fail_reason: 'upstream failed',
        }}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: 'View request body' }))

    await waitFor(() =>
      expect(getTaskRequestBody).toHaveBeenCalledWith('task_video', false)
    )
    expect((await screen.findByTestId('request-json')).textContent).toBe(`{
  "prompt": "hello",
  "parameters": {
    "duration": 5
  }
}`)
    expect(screen.getByText('upstream failed')).toBeInTheDocument()
  })

  test('downloads successful video content through the authenticated API', async () => {
    const createObjectURL = vi
      .spyOn(URL, 'createObjectURL')
      .mockReturnValue('blob:video')
    const revokeObjectURL = vi
      .spyOn(URL, 'revokeObjectURL')
      .mockImplementation(() => undefined)
    const click = vi
      .spyOn(HTMLAnchorElement.prototype, 'click')
      .mockImplementation(() => undefined)
    vi.mocked(downloadTaskVideo).mockResolvedValue({
      data: new Blob(['video'], { type: 'video/mp4' }),
      headers: { 'content-type': 'video/mp4' },
    })
    vi.mocked(getTaskVideoContentInfo).mockResolvedValue({
      success: true,
      data: { delivery_mode: 'proxy' },
    })

    render(<TaskLogDetailsCell isAdmin={false} log={successfulVideoLog} />)

    fireEvent.click(screen.getByRole('button', { name: 'Download video' }))

    await waitFor(() =>
      expect(getTaskVideoContentInfo).toHaveBeenCalledWith('task_video')
    )
    expect(downloadTaskVideo).toHaveBeenCalledWith('task_video')
    expect(createObjectURL).toHaveBeenCalledOnce()
    expect(click).toHaveBeenCalledOnce()
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:video')
  })

  test('uses browser navigation for redirect delivery without fetching a blob', async () => {
    const redirectURL =
      'https://cdn.example.com/video.mp4?signature=authorized-download'
    let clickedHref = ''
    const click = vi
      .spyOn(HTMLAnchorElement.prototype, 'click')
      .mockImplementation(function (this: HTMLAnchorElement) {
        clickedHref = this.href
      })
    vi.mocked(getTaskVideoContentInfo).mockResolvedValue({
      success: true,
      data: { delivery_mode: 'redirect', url: redirectURL },
    })

    render(<TaskLogDetailsCell isAdmin={false} log={successfulVideoLog} />)

    fireEvent.click(screen.getByRole('button', { name: 'Download video' }))

    await waitFor(() => expect(click).toHaveBeenCalledOnce())
    expect(clickedHref).toBe(redirectURL)
    expect(downloadTaskVideo).not.toHaveBeenCalled()
  })

  test('hides unavailable request body and video actions for legacy tasks', () => {
    render(
      <TaskLogDetailsCell
        isAdmin={false}
        log={{
          ...successfulVideoLog,
          result_url: undefined,
          request_body_available: false,
        }}
      />
    )

    expect(
      screen.queryByRole('button', { name: 'Download video' })
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'View request body' })
    ).not.toBeInTheDocument()
    expect(screen.getByText('-')).toBeInTheDocument()
  })
})
