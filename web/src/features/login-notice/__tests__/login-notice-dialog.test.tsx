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
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'

import { acknowledgeLoginNotice, getLoginNotice } from '../api'
import { createDeviceFingerprint } from '../device-fingerprint'
import { LoginNoticeDialog } from '../login-notice-dialog'

vi.mock('../api', () => ({
  acknowledgeLoginNotice: vi.fn(),
  getLoginNotice: vi.fn(),
}))

vi.mock('../device-fingerprint', () => ({
  createDeviceFingerprint: vi.fn(),
}))

vi.mock('@/components/rich-content', () => ({
  RichContent: ({ content }: { content: string }) => <div>{content}</div>,
}))

vi.mock('@/components/ui/scroll-area', () => ({
  ScrollArea: ({ children }: { children: React.ReactNode }) => (
    <div>{children}</div>
  ),
}))

function renderDialog() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <LoginNoticeDialog />
    </QueryClientProvider>
  )
}

describe('LoginNoticeDialog', () => {
  beforeEach(() => {
    vi.mocked(getLoginNotice).mockReset()
    vi.mocked(acknowledgeLoginNotice).mockReset()
    vi.mocked(createDeviceFingerprint).mockReset()
    vi.mocked(createDeviceFingerprint).mockResolvedValue('ab'.repeat(32))
    useAuthStore.getState().auth.setBundle({
      access_token: 'access-token',
      token_type: 'Bearer',
      access_expires_at: 1_800_000_000,
      user: { id: 12, username: 'notice-user', role: 1 },
      session: {
        sid: 'session-12',
        current: true,
        login_method: 'password',
        ip: '127.0.0.1',
        user_agent: 'test',
        created_at: 1_700_000_000,
        last_active_at: 1_700_000_000,
        expires_at: 1_800_000_000,
      },
    })
  })

  afterEach(() => {
    useAuthStore.getState().auth.reset('complete')
  })

  test('keeps a recent-violation notice open until acknowledgement succeeds', async () => {
    vi.mocked(getLoginNotice).mockResolvedValue({
      announcements: [{ id: 1, content: 'Scheduled maintenance' }],
      statistics: {
        today: { generated: 2, violations: 1 },
        seven_days: { generated: 8, violations: 1 },
        thirty_days: { generated: 20, violations: 3 },
      },
      requires_acknowledgement: true,
      acknowledged: false,
    })
    vi.mocked(acknowledgeLoginNotice).mockResolvedValue()

    renderDialog()

    expect(await screen.findByRole('alertdialog')).toBeInTheDocument()
    expect(screen.getByText('Scheduled maintenance')).toBeInTheDocument()
    expect(screen.getByText('Acknowledgement required')).toBeInTheDocument()

    fireEvent.keyDown(document, { key: 'Escape' })
    expect(screen.getByRole('alertdialog')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'I acknowledge' }))
    await waitFor(() =>
      expect(acknowledgeLoginNotice).toHaveBeenCalledWith('ab'.repeat(32))
    )
    await waitFor(() =>
      expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument()
    )
  })

  test('does not show a notice already acknowledged in this login session', async () => {
    vi.mocked(getLoginNotice).mockResolvedValue({
      announcements: [],
      statistics: {
        today: { generated: 0, violations: 0 },
        seven_days: { generated: 0, violations: 0 },
        thirty_days: { generated: 0, violations: 0 },
      },
      requires_acknowledgement: false,
      acknowledged: true,
    })

    renderDialog()

    await waitFor(() => expect(getLoginNotice).toHaveBeenCalledOnce())
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument()
  })
})
