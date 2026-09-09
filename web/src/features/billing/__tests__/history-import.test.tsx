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
import { beforeEach, describe, expect, test, vi } from 'vitest'

import {
  confirmBillingHistoryImport,
  getBillingHistoryImports,
  reviewBillingHistory,
} from '../api'
import { HistoryImportDialog } from '../components/history-import-dialog'
import type { BillingHistoryReview } from '../types'

vi.mock('../api', async (original) => ({
  ...(await original<typeof import('../api')>()),
  reviewBillingHistory: vi.fn(),
  confirmBillingHistoryImport: vi.fn(),
  getBillingHistoryImports: vi.fn(),
}))
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

const review: BillingHistoryReview = {
  user_id: 4,
  month: '2026-08',
  customer: { id: 4, username: 'example_customer', display_name: 'Customer' },
  ready: true,
  checks: [],
  source_sha256: 'reviewed-digest',
  source_count: 3,
  record_limit: 10000,
  existing_statement: '',
  old_start_at: 100,
  new_start_at: 100,
  snapshot: {
    month: '2026-08',
    timezone: 'Asia/Shanghai',
    company_title: 'Customer',
    tax_id: 'TAX',
    issuer: 'Platform',
    accounting_start_at: 100,
    currency: { code: 'CNY', symbol: '¥', rate: '7', quota_per_unit: '500000' },
    days: [],
    total: {
      label: '2026-08',
      charge: '250.000000',
      refund: '10.000000',
      amount: '240.000000',
      count: 3,
    },
  },
}

function renderHistory() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  render(
    <QueryClientProvider client={client}>
      <HistoryImportDialog
        userId={4}
        month='2026-08'
        profileId={1}
        onViewStatement={vi.fn()}
      />
    </QueryClientProvider>
  )
}

describe('Historical import consent', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(reviewBillingHistory).mockResolvedValue(review)
    vi.mocked(getBillingHistoryImports).mockResolvedValue([])
    vi.mocked(confirmBillingHistoryImport).mockResolvedValue({
      id: 'imported',
      user_id: 4,
      month: '2026-08',
      source_sha256: 'reviewed-digest',
      records: 3,
      confirmed_by: 1,
      confirmed_at: 200,
      note: 'Verified source records',
    })
  })

  test('opening a review never imports; notes, acknowledgement and a second confirmation are required', async () => {
    renderHistory()
    expect(reviewBillingHistory).not.toHaveBeenCalled()
    fireEvent.click(
      screen.getByRole('button', { name: 'Review historical import' })
    )
    const confirm = await screen.findByRole('button', {
      name: 'Confirm historical import',
    })
    expect(confirm).toBeDisabled()
    expect(
      screen.getByText(/Statement customer: example_customer \(#4\)/)
    ).toBeVisible()
    fireEvent.change(screen.getByLabelText('Verification basis'), {
      target: { value: 'Verified source records' },
    })
    expect(confirm).toBeDisabled()
    fireEvent.click(screen.getByRole('checkbox'))
    fireEvent.click(confirm)
    expect(confirmBillingHistoryImport).not.toHaveBeenCalled()
    expect(screen.getByRole('alertdialog')).toHaveTextContent('240.000000')
    fireEvent.click(
      screen.getByRole('button', { name: 'Import verified records' })
    )
    await waitFor(() =>
      expect(confirmBillingHistoryImport).toHaveBeenCalledWith(
        4,
        expect.objectContaining({
          month: '2026-08',
          source_sha256: 'reviewed-digest',
          storage_profile_id: 1,
          acknowledged: true,
          note: 'Verified source records',
        })
      )
    )
  })

  test('an active draft blocks importing even after acknowledging the records', async () => {
    vi.mocked(reviewBillingHistory).mockResolvedValue({
      ...review,
      ready: false,
      existing_statement: 'old-empty-draft',
      checks: [{ code: 'no_active_statement', passed: false }],
    })
    renderHistory()
    fireEvent.click(
      screen.getByRole('button', { name: 'Review historical import' })
    )
    await screen.findByRole('button', { name: 'View existing statement' })
    expect(screen.getByRole('alert')).toHaveTextContent(
      'Void active drafts explicitly'
    )
    fireEvent.change(screen.getByLabelText('Verification basis'), {
      target: { value: 'Verified' },
    })
    fireEvent.click(screen.getByRole('checkbox'))
    expect(
      screen.getByRole('button', { name: 'Confirm historical import' })
    ).toBeDisabled()
    expect(confirmBillingHistoryImport).not.toHaveBeenCalled()
  })

  test('a failed review refresh hides stale confirmation controls', async () => {
    renderHistory()
    fireEvent.click(
      screen.getByRole('button', { name: 'Review historical import' })
    )
    await screen.findByLabelText('Verification basis')
    vi.mocked(reviewBillingHistory).mockRejectedValue(new Error('Unavailable'))
    fireEvent.click(screen.getByRole('button', { name: 'Refresh review' }))
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Failed to load data'
    )
    expect(
      screen.queryByRole('button', { name: 'Confirm historical import' })
    ).not.toBeInTheDocument()
    expect(confirmBillingHistoryImport).not.toHaveBeenCalled()
  })
})
