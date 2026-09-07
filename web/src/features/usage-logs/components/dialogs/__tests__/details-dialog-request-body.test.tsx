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
import { render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, test, vi } from 'vitest'

import type { UsageLog } from '../../../data/schema'
import { getTaskRequestBody } from '../../../task-content-api'
import { DetailsDialog } from '../details-dialog'

vi.mock('../../../task-content-api', () => ({
  getTaskRequestBody: vi.fn(),
}))

vi.mock('@/components/ai-elements/code-block', () => ({
  CodeBlock: ({ code }: { code: string }) => (
    <pre data-testid='request-json'>{code}</pre>
  ),
  CodeBlockCopyButton: () => null,
}))

const videoUsageLog: UsageLog = {
  id: 1,
  user_id: 2,
  created_at: 100,
  type: 2,
  content: 'generate',
  username: 'video-user',
  token_name: 'video-token',
  model_name: 'video-model',
  quota: 100,
  prompt_tokens: 0,
  completion_tokens: 0,
  use_time: 5,
  is_stream: false,
  channel: 3,
  channel_name: 'video-channel',
  token_id: 4,
  group: 'default',
  ip: '',
  request_id: 'request_video',
  upstream_request_id: 'upstream_video',
  other: JSON.stringify({
    is_task: true,
    task_id: 'task_video',
    request_body_available: true,
    request_path: '/api/v3/contents/generations/tasks',
  }),
}

function renderDialog(isAdmin: boolean, log: UsageLog = videoUsageLog) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <DetailsDialog
        isAdmin={isAdmin}
        log={log}
        open
        onOpenChange={() => undefined}
      />
    </QueryClientProvider>
  )
}

describe('DetailsDialog request body', () => {
  beforeEach(() => {
    vi.mocked(getTaskRequestBody).mockReset()
  })

  test.each([
    { isAdmin: false, role: 'user' },
    { isAdmin: true, role: 'admin' },
  ])(
    'shows formatted video request JSON for $role logs',
    async ({ isAdmin }) => {
      vi.mocked(getTaskRequestBody).mockResolvedValue({
        success: true,
        data: { prompt: 'hello', parameters: { duration: 5 } },
      })

      renderDialog(isAdmin)

      await waitFor(() =>
        expect(getTaskRequestBody).toHaveBeenCalledWith('task_video', isAdmin)
      )
      expect(await screen.findByText('Request Body')).toBeInTheDocument()
      expect((await screen.findByTestId('request-json')).textContent).toBe(`{
  "prompt": "hello",
  "parameters": {
    "duration": 5
  }
}`)
    }
  )

  test('does not request or show a body for unrelated usage logs', () => {
    renderDialog(false, {
      ...videoUsageLog,
      other: JSON.stringify({ request_path: '/v1/chat/completions' }),
    })

    expect(getTaskRequestBody).not.toHaveBeenCalled()
    expect(screen.queryByText('Request Body')).not.toBeInTheDocument()
  })

  test('shows an embedded request body when task submission fails', async () => {
    renderDialog(false, {
      ...videoUsageLog,
      type: 5,
      content: 'upstream failed',
      other: JSON.stringify({
        is_task: true,
        request_body: JSON.stringify({ prompt: 'failed request' }),
      }),
    })

    expect(getTaskRequestBody).not.toHaveBeenCalled()
    expect((await screen.findByTestId('request-json')).textContent).toBe(`{
  "prompt": "failed request"
}`)
  })

  test('shows the unavailable state when the request body lookup fails', async () => {
    vi.mocked(getTaskRequestBody).mockResolvedValue({
      success: false,
      message: 'not found',
    })

    renderDialog(false)

    expect(
      await screen.findByText('Request body unavailable')
    ).toBeInTheDocument()
    expect(screen.getByText('Request failed')).toBeInTheDocument()
  })
})
