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
import type { ReactNode } from 'react'
import { beforeEach, describe, expect, test, vi } from 'vitest'

import {
  getBillingAccount,
  getStatement,
  saveBillingAccount,
  statementAction,
} from '../api'
import { BillingProfileCard } from '../components/billing-profile-card'
import { StatementDetail } from '../components/statement-detail'
import type { BillingSnapshot, StatementDetail as Detail } from '../types'

vi.mock('../api', async (original) => ({
  ...(await original<typeof import('../api')>()),
  getBillingAccount: vi.fn(),
  getStatement: vi.fn(),
  saveBillingAccount: vi.fn(),
  statementAction: vi.fn(),
}))
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

function renderBilling(node: ReactNode) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return render(
    <QueryClientProvider client={client}>{node}</QueryClientProvider>
  )
}
const account = {
  user_id: 2,
  company_title: 'Customer Ltd',
  tax_id: 'TAX123',
  accounting_start_at: 100,
  profile_version: 3,
}
const snapshot: BillingSnapshot = {
  month: '2026-02',
  timezone: 'Asia/Shanghai',
  company_title: 'Frozen Customer',
  tax_id: 'FROZEN123',
  issuer: 'Service',
  accounting_start_at: 100,
  currency: { code: 'USD', symbol: '$', rate: '1', quota_per_unit: '500000' },
  days: [
    {
      label: '2026-02-01',
      charge: '1.230000',
      refund: '0.000000',
      amount: '1.230000',
      count: 1,
    },
  ],
  total: {
    label: '2026-02',
    charge: '1.230000',
    refund: '0.000000',
    amount: '1.230000',
    count: 1,
  },
}
const detail: Detail = {
  statement: {
    id: 'statement',
    user_id: 2,
    month: '2026-02',
    revision: 1,
    status: 'issued',
    snapshot: JSON.stringify(snapshot),
    manifest_sha256: 'frozen-manifest',
    pdf_sha256: 'frozen-pdf',
    issued_at: 200,
    confirmed_at: 0,
    due_at: 300,
    created_at: 150,
  },
  events: [],
  artifacts: [],
}

describe('Billing customer workflow', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getBillingAccount).mockResolvedValue(account)
    vi.mocked(getStatement).mockResolvedValue(detail)
    vi.mocked(saveBillingAccount).mockResolvedValue(account)
    vi.mocked(statementAction).mockResolvedValue({
      ...detail.statement,
      status: 'confirmed',
      confirmed_at: 400,
    })
  })
  test('customer can change company identity but not the accounting start', async () => {
    renderBilling(<BillingProfileCard userId={2} />)
    const title = await screen.findByLabelText('Company title')
    expect(
      screen.queryByLabelText('Enable formal accounting')
    ).not.toBeInTheDocument()
    fireEvent.change(title, { target: { value: 'New Customer' } })
    fireEvent.click(screen.getByRole('button', { name: 'Save' }))
    await waitFor(() =>
      expect(saveBillingAccount).toHaveBeenCalledWith(2, {
        company_title: 'New Customer',
        tax_id: 'TAX123',
        profile_version: 3,
      })
    )
  })
  test('owner must acknowledge and explicitly confirm the frozen version', async () => {
    renderBilling(
      <StatementDetail
        id='statement'
        onClose={vi.fn()}
        admin={false}
        currentUserId={2}
      />
    )
    const confirm = await screen.findByRole('button', {
      name: 'Confirm statement',
    })
    expect(confirm).toBeDisabled()
    expect(
      screen.getByText('Company title: Frozen Customer')
    ).toBeInTheDocument()
    expect(screen.getAllByText('$ 1.230000')).toHaveLength(4)
    fireEvent.click(screen.getByRole('checkbox'))
    fireEvent.click(confirm)
    expect(statementAction).not.toHaveBeenCalled()
    fireEvent.click(screen.getByRole('button', { name: 'Confirm' }))
    await waitFor(() =>
      expect(statementAction).toHaveBeenCalledWith(
        'statement',
        'confirm',
        'frozen-manifest',
        ''
      )
    )
  })
  test('administrator cannot confirm on behalf of another customer', async () => {
    renderBilling(
      <StatementDetail
        id='statement'
        onClose={vi.fn()}
        admin
        currentUserId={1}
      />
    )
    await screen.findByText('Company title: Frozen Customer')
    expect(
      screen.queryByRole('button', { name: 'Confirm statement' })
    ).not.toBeInTheDocument()
    expect(screen.queryByRole('checkbox')).not.toBeInTheDocument()
  })
  test('statement identifies the customer separately from the administrator who prepared it', async () => {
    vi.mocked(getStatement).mockResolvedValue({
      ...detail,
      customer: {
        id: 2,
        username: 'example_customer',
        display_name: 'Customer',
      },
      events: [
        {
          id: 1,
          actor_id: 1,
          actor_username: 'admin',
          action: 'prepare',
          note: '',
          created_at: 150,
        },
      ],
    })
    renderBilling(
      <StatementDetail
        id='statement'
        onClose={vi.fn()}
        admin
        currentUserId={1}
      />
    )
    expect(
      await screen.findByText('Statement customer: example_customer (#2)')
    ).toBeInTheDocument()
    expect(screen.getByText(/Operator: admin \(#1\)/)).toBeInTheDocument()
    expect(screen.queryByText(/Account 1/)).not.toBeInTheDocument()
  })
  test('an unreconciled empty draft cannot be issued and explains the historical-data gap', async () => {
    vi.mocked(getStatement).mockResolvedValue({
      ...detail,
      statement: { ...detail.statement, status: 'draft' },
      source_warning: 'historical_data_unreconciled',
    })
    renderBilling(
      <StatementDetail
        id='statement'
        onClose={vi.fn()}
        admin
        currentUserId={1}
      />
    )
    expect(
      await screen.findByRole('button', { name: 'Issue to customer' })
    ).toBeDisabled()
    expect(screen.getByRole('alert')).toHaveTextContent(
      'Historical usage exists but this statement has no formal entries.'
    )
    expect(statementAction).not.toHaveBeenCalled()
  })
})
