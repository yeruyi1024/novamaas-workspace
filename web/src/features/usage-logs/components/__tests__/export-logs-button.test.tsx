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
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import i18next from 'i18next'
import { afterEach, beforeAll, describe, expect, test, vi } from 'vitest'

import { ExportLogsButton } from '../export-logs-button'
import { useLogsViewScope } from '../usage-logs-provider'

const current = vi.hoisted(() => ({ search: {} as Record<string, unknown> }))

vi.mock('@tanstack/react-router', async (original) => ({
  ...(await original<typeof import('@tanstack/react-router')>()),
  getRouteApi: () => ({ useSearch: () => current.search }),
}))

vi.mock('../usage-logs-provider', () => ({ useLogsViewScope: vi.fn() }))

const search = {
  startTime: new Date(2026, 9, 8, 0, 0, 0).getTime(),
  endTime: new Date(2026, 10, 6, 23, 59, 59).getTime(),
}

async function openConfirmation() {
  render(<ExportLogsButton />)
  await userEvent.click(screen.getByRole('button', { name: 'Export' }))
  return screen.getByRole('dialog')
}

describe('export logs confirmation', () => {
  beforeAll(async () => {
    i18next.addResourceBundle('en', 'translation', {
      Export: 'Export',
      'Export logs': 'Export logs',
      'View Scope': 'View Scope',
      All: 'All',
      'Only Mine': 'Only Mine',
      Unlimited: 'Unlimited',
      'All Types': 'All Types',
      Consume: 'Consume',
    })
    await i18next.changeLanguage('en')
  })

  afterEach(() => {
    vi.clearAllMocks()
    current.search = {}
  })

  test('lists every selected condition and marks the empty ones unlimited', async () => {
    current.search = {
      ...search,
      type: ['2'],
      model: 'kimi-k3',
      token: 'aigv-k3',
      group: 'spe_kimi_53',
      username: 'isv_co_lingchuan',
    }
    vi.mocked(useLogsViewScope).mockReturnValue({
      canManageScope: true,
      isAdminView: true,
      viewScope: 'all',
      setViewScope: vi.fn(),
    })

    const dialog = await openConfirmation()
    const rows = Object.fromEntries(
      [...dialog.querySelectorAll('dl > div')].map((row) => [
        row.querySelector('dt')?.textContent,
        row.querySelector('dd')?.textContent,
      ])
    )

    expect(rows).toEqual({
      'Date Range': '2026-10-08 00:00:00 ~ 2026-11-06 23:59:59',
      'View Scope': 'All',
      'Log Type': 'Consume',
      Username: 'isv_co_lingchuan',
      Token: 'aigv-k3',
      'Model Name': 'kimi-k3',
      Group: 'spe_kimi_53',
      Channel: 'Unlimited',
      'Request ID': 'Unlimited',
      'Upstream Request ID': 'Unlimited',
    })
    expect(
      within(dialog).getByText(/Only consumption and refund records/)
    ).toBeTruthy()
  })

  test('shows all types, and no admin-only conditions for a regular user', async () => {
    current.search = search
    vi.mocked(useLogsViewScope).mockReturnValue({
      canManageScope: false,
      isAdminView: false,
      viewScope: 'all',
      setViewScope: vi.fn(),
    })

    const dialog = await openConfirmation()
    const labels = [...dialog.querySelectorAll('dt')].map(
      (label) => label.textContent
    )

    expect(labels).toEqual([
      'Date Range',
      'Log Type',
      'Token',
      'Model Name',
      'Group',
      'Request ID',
      'Upstream Request ID',
    ])
    expect(within(dialog).getByText('All Types')).toBeTruthy()
  })

  test('an administrator viewing only their own logs sees that scope and no user or channel', async () => {
    current.search = search
    vi.mocked(useLogsViewScope).mockReturnValue({
      canManageScope: true,
      isAdminView: false,
      viewScope: 'self',
      setViewScope: vi.fn(),
    })

    const dialog = await openConfirmation()
    const labels = [...dialog.querySelectorAll('dt')].map(
      (label) => label.textContent
    )

    expect(within(dialog).getByText('Only Mine')).toBeTruthy()
    expect(labels).not.toContain('Username')
    expect(labels).not.toContain('Channel')
  })
})
