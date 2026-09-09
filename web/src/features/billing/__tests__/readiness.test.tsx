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
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, test, vi } from 'vitest'

import {
  getBillingStorageProfiles,
  getStatements,
  previewStatement,
} from '../api'
import { MonthlyPreviewCard } from '../components/monthly-preview-card'
import { StatementsPanel } from '../components/statements-panel'

vi.mock('../api', async (original) => ({
  ...(await original<typeof import('../api')>()),
  getBillingStorageProfiles: vi.fn(),
  getStatements: vi.fn(),
  previewStatement: vi.fn(),
}))

const reference = {
  month: '2020-02',
  timezone: 'Asia/Shanghai',
  company_title: '',
  tax_id: '',
  issuer: '',
  accounting_start_at: 0,
  currency: { code: 'USD', symbol: '$', rate: '1', quota_per_unit: '500000' },
  days: [
    {
      label: '2020-02-03',
      charge: '1.000000',
      refund: '0.000000',
      amount: '1.000000',
      count: 1,
    },
  ],
  total: {
    label: '2020-02',
    charge: '1.000000',
    refund: '0.000000',
    amount: '1.000000',
    count: 1,
  },
  source: 'usage_logs' as const,
  updated_at: 100,
  last_log_at: 90,
}

describe('Monthly billing readiness', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getStatements).mockResolvedValue([])
    vi.mocked(getBillingStorageProfiles).mockResolvedValue([])
    vi.mocked(previewStatement).mockResolvedValue({
      ...reference,
      reference,
      formal: null,
      readiness: {
        ready: false,
        status: 'not_configured',
        accounting_start_at: 0,
        period_start_at: 1580486400,
        period_end_at: 1582992000,
        prepare_at: 1583078400,
        pending: 0,
        existing_statement: '',
        checks: [
          { code: 'accounting_configured', passed: false },
          { code: 'storage_missing', passed: false },
        ],
      },
    })
  })

  test('unconfigured account shows reference usage and a setup action, not an official zero bill', async () => {
    const configure = vi.fn()
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    render(
      <QueryClientProvider client={client}>
        <StatementsPanel
          userId={4}
          currentUserId={1}
          admin
          onConfigureIdentity={configure}
        />
      </QueryClientProvider>
    )
    fireEvent.change(screen.getByLabelText('Month'), {
      target: { value: '2020-02' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'Preview' }))
    expect(
      await screen.findByRole('button', { name: 'Configure billing identity' })
    ).toBeVisible()
    expect(
      screen.getByText('Historical consumption (reference only)')
    ).toBeVisible()
    expect(
      screen.getByText(
        'Formal accounting is not enabled. Historical usage is shown below for reference.'
      )
    ).toBeVisible()
    expect(
      screen.queryByText('Formal statement preview')
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Create archived draft' })
    ).toBeDisabled()
    fireEvent.click(
      screen.getByRole('button', { name: 'Configure billing identity' })
    )
    expect(configure).toHaveBeenCalledOnce()
  })

  test('refreshing the same month rechecks readiness; a failed refresh cannot create a stale draft', async () => {
    vi.mocked(getBillingStorageProfiles).mockResolvedValue([
      { id: 1, name: 'Test OSS' },
    ])
    vi.mocked(previewStatement).mockResolvedValue({
      reference,
      formal: reference,
      readiness: {
        ready: true,
        status: 'ready',
        accounting_start_at: 1580486400,
        period_start_at: 1580486400,
        period_end_at: 1582992000,
        prepare_at: 1583078400,
        pending: 0,
        existing_statement: '',
        checks: [],
      },
    })
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    render(
      <QueryClientProvider client={client}>
        <MonthlyPreviewCard userId={4} onViewStatement={vi.fn()} />
      </QueryClientProvider>
    )
    await screen.findByText('Formal statement preview')
    const user = userEvent.setup()
    await user.click(screen.getByRole('combobox', { name: 'Archive storage' }))
    await user.click(await screen.findByRole('option', { name: 'Test OSS' }))
    expect(
      await screen.findByRole('button', { name: 'Create archived draft' })
    ).toBeEnabled()
    vi.mocked(previewStatement).mockRejectedValue(new Error('Unavailable'))
    await user.click(screen.getByRole('button', { name: 'Preview' }))
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Failed to load data'
    )
    expect(
      screen.getByRole('button', { name: 'Create archived draft' })
    ).toBeDisabled()
    expect(
      screen.queryByText('Formal statement preview')
    ).not.toBeInTheDocument()
  })

  test('a month with historical usage and an empty ledger exposes review instead of a formal zero draft', async () => {
    vi.mocked(previewStatement).mockResolvedValue({
      reference,
      formal: null,
      readiness: {
        ready: false,
        status: 'historical_data_unreconciled',
        accounting_start_at: 1580486400,
        period_start_at: 1580486400,
        period_end_at: 1582992000,
        prepare_at: 1583078400,
        pending: 0,
        existing_statement: '',
        checks: [{ code: 'history_reconciled', passed: false }],
      },
    })
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    render(
      <QueryClientProvider client={client}>
        <MonthlyPreviewCard userId={4} onViewStatement={vi.fn()} />
      </QueryClientProvider>
    )
    expect(await screen.findByRole('status')).toHaveTextContent(
      'Historical usage exists without formal entries.'
    )
    expect(
      screen.getByRole('button', { name: 'Review historical import' })
    ).toBeVisible()
    expect(
      screen.getByRole('button', { name: 'Create archived draft' })
    ).toBeDisabled()
    expect(
      screen.queryByText('Formal statement preview')
    ).not.toBeInTheDocument()
  })
})
