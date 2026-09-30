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
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import i18next from 'i18next'
import { afterEach, beforeAll, describe, expect, test, vi } from 'vitest'

import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import { getLogStats } from '../../api'
import { CommonLogsStats } from '../common-logs-stats'
import { UsageLogsProvider } from '../usage-logs-provider'

vi.mock('@tanstack/react-router', async (original) => ({
  ...(await original<typeof import('@tanstack/react-router')>()),
  getRouteApi: () => ({ useSearch: () => ({}) }),
}))

vi.mock('../../api', async (original) => ({
  ...(await original<typeof import('../../api')>()),
  getLogStats: vi.fn(),
}))

describe('common log financial summary', () => {
  beforeAll(async () => {
    i18next.addResourceBundle('en', 'translation', {
      Usage: 'Usage',
      Turnover: 'Turnover',
      'Cost amount': 'Cost amount',
      'Profit amount': 'Profit amount',
      'Revenue profit margin': 'Revenue profit margin',
      'Cost profit margin': 'Cost profit margin',
    })
    await i18next.changeLanguage('en')
  })

  afterEach(() => {
    vi.clearAllMocks()
    useAuthStore.getState().auth.setUser(null)
  })

  test('shows turnover immediately before cost and profit for finance users', async () => {
    useAuthStore.getState().auth.setUser({
      id: 1,
      username: 'finance-admin',
      role: ROLE.ADMIN,
      permissions: {
        admin_permissions: { financial_accounting: { view: true } },
      },
    })
    vi.mocked(getLogStats).mockResolvedValue({
      success: true,
      data: {
        quota: 11_573_731,
        rpm: 0,
        tpm: 0,
        revenue_quota: 10_177_610,
        cost_quota: 9_467_539,
        profit_quota: 710_071,
      },
    })
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })

    render(
      <QueryClientProvider client={client}>
        <UsageLogsProvider>
          <CommonLogsStats />
        </UsageLogsProvider>
      </QueryClientProvider>
    )

    const turnover = await screen.findByText('Turnover')
    const cost = screen.getByText('Cost amount')
    const profit = screen.getByText('Profit amount')

    expect(
      turnover.compareDocumentPosition(cost) & Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
    expect(
      cost.compareDocumentPosition(profit) & Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
  })

  test('profit hover shows revenue-based and cost-based rates', async () => {
    useAuthStore.getState().auth.setUser({
      id: 1,
      username: 'finance-admin',
      role: ROLE.ADMIN,
      permissions: {
        admin_permissions: { financial_accounting: { view: true } },
      },
    })
    vi.mocked(getLogStats).mockResolvedValue({
      success: true,
      data: {
        quota: 1000,
        rpm: 0,
        tpm: 0,
        revenue_quota: 1000,
        cost_quota: 800,
        profit_quota: 200,
      },
    })
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    render(
      <QueryClientProvider client={client}>
        <UsageLogsProvider>
          <CommonLogsStats />
        </UsageLogsProvider>
      </QueryClientProvider>
    )

    const trigger = await screen.findByRole('button', {
      name: /Profit amount:.*Revenue profit margin: 20.00%.*Cost profit margin: 25.00%/,
    })
    await userEvent.setup().hover(trigger)
    await screen.findByText('20.00%')
    const tooltip = document.querySelector('[data-slot="tooltip-content"]')
    expect(tooltip).toHaveTextContent('Revenue profit margin20.00%')
    expect(tooltip).toHaveTextContent('Cost profit margin25.00%')
  })

  test('an administrator without financial access sees no financial summary', async () => {
    useAuthStore.getState().auth.setUser({
      id: 2,
      username: 'operations-admin',
      role: ROLE.ADMIN,
      permissions: {
        admin_permissions: { financial_accounting: { view: false } },
      },
    })
    vi.mocked(getLogStats).mockResolvedValue({
      success: true,
      data: {
        quota: 1000,
        rpm: 0,
        tpm: 0,
        revenue_quota: 1000,
        cost_quota: 800,
        profit_quota: 200,
      },
    })
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    render(
      <QueryClientProvider client={client}>
        <UsageLogsProvider>
          <CommonLogsStats />
        </UsageLogsProvider>
      </QueryClientProvider>
    )

    expect(await screen.findByText('Usage')).toBeVisible()
    expect(screen.queryByText('Turnover')).toBeNull()
    expect(screen.queryByText('Profit amount')).toBeNull()
  })
})
