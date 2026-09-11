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
  getTaskInformation,
  getTaskVideoContentInfo,
  getTaskRequestSnapshots,
} from '../../../task-content-api'
import type { TaskLog } from '../../../types'
import { TaskLogDetailsCell } from '../task-log-details-cell'

vi.mock('../../../task-content-api', () => ({
  canGetTaskInformation: (platform: string) =>
    platform === '17' || platform === '54' || platform === '61',
  downloadTaskVideo: vi.fn(),
  getTaskInformation: vi.fn(),
  getTaskVideoContentInfo: vi.fn(),
  getTaskRequestSnapshots: vi.fn(),
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
    vi.mocked(getTaskRequestSnapshots).mockReset()
    vi.mocked(downloadTaskVideo).mockReset()
    vi.mocked(getTaskVideoContentInfo).mockReset()
    vi.mocked(getTaskInformation).mockReset()
  })

  test('shows the formatted request JSON regardless of task status', async () => {
    vi.mocked(getTaskRequestSnapshots).mockResolvedValue({
      success: true,
      data: {
        original: { prompt: 'hello', parameters: { duration: 5 } },
        upstream: {
          model: 'doubao-seedance-upstream',
          content: [
            {
              type: 'image_url',
              image_url: { url: 'https://storage.example.com/staged.webp' },
            },
          ],
        },
      },
    })

    render(
      <TaskLogDetailsCell
        isAdmin
        log={{
          ...successfulVideoLog,
          status: 'FAILURE',
          fail_reason: 'upstream failed',
        }}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: 'View request body' }))

    await waitFor(() =>
      expect(getTaskRequestSnapshots).toHaveBeenCalledWith('task_video')
    )
    expect((await screen.findByTestId('request-json')).textContent).toBe(`{
  "prompt": "hello",
  "parameters": {
    "duration": 5
  }
}`)
    fireEvent.click(
      screen.getByRole('tab', { name: 'Actual Upstream Request' })
    )
    expect((await screen.findByTestId('request-json')).textContent).toContain(
      'https://storage.example.com/staged.webp'
    )
    expect(screen.getByText('upstream failed')).toBeInTheDocument()
  })

  test('hides the request body action from non-administrators', () => {
    render(
      <TaskLogDetailsCell
        isAdmin={false}
        log={{
          ...successfulVideoLog,
          platform: 'other',
          result_url: undefined,
        }}
      />
    )

    expect(
      screen.queryByRole('button', { name: 'View request body' })
    ).not.toBeInTheDocument()
    expect(getTaskRequestSnapshots).not.toHaveBeenCalled()
    expect(screen.getByText('-')).toBeInTheDocument()
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

  test.each([
    { platform: '17', taskId: 'task_ali' },
    { platform: '54', taskId: 'task_doubao' },
    { platform: '61', taskId: 'task_volc_native' },
  ])(
    'queries and displays task information for platform $platform',
    async ({ platform, taskId }) => {
      vi.mocked(getTaskInformation).mockResolvedValue({
        id: taskId,
        status: 'succeeded',
      })

      render(
        <TaskLogDetailsCell
          isAdmin
          log={{
            ...successfulVideoLog,
            platform,
            task_id: taskId,
          }}
        />
      )

      fireEvent.click(screen.getByRole('button', { name: 'View information' }))

      await waitFor(() =>
        expect(getTaskInformation).toHaveBeenCalledWith(taskId, platform)
      )
      expect((await screen.findByTestId('request-json')).textContent).toContain(
        `"id": "${taskId}"`
      )
    }
  )

  test('keeps safe task information visible while hiding the request body from regular users', async () => {
    vi.mocked(getTaskInformation).mockResolvedValue({
      id: 'task_video',
      status: 'succeeded',
    })

    render(
      <TaskLogDetailsCell
        isAdmin={false}
        log={{
          ...successfulVideoLog,
          platform: '54',
          result_url: undefined,
        }}
      />
    )

    expect(
      screen.queryByRole('button', { name: 'Download video' })
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'View request body' })
    ).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'View information' }))

    await waitFor(() =>
      expect(getTaskInformation).toHaveBeenCalledWith('task_video', '54')
    )
    expect((await screen.findByTestId('request-json')).textContent).toContain(
      '"status": "succeeded"'
    )
  })
})
