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
import {
  fireEvent,
  render,
  screen,
  within,
  waitFor,
} from '@testing-library/react'
import { expect, test, vi } from 'vitest'

import { collectReviewedDebit, reviewDebits, type ReviewedDebit } from '../api'
import { PendingDebitDialog } from '../components/pending-debit-dialog'

vi.mock('../api', () => ({
  collectReviewedDebit: vi.fn(),
  reviewDebits: vi.fn(),
}))

test('preview never charges and confirmation applies only the selected evidence row', async () => {
  const first: ReviewedDebit = {
    user_id: 12,
    username: 'example',
    request_id: 'request-a',
    source: 'usage_log',
    quota: 100,
    amount: '0.001400',
    status: 'ready',
  }
  const second: ReviewedDebit = {
    ...first,
    request_id: 'request-b',
    source: 'supplier_statement',
    upstream_request_id: 'supplier-b',
    quota: 200,
    amount: '0.002800',
  }
  vi.mocked(reviewDebits).mockResolvedValue([first, second])
  vi.mocked(collectReviewedDebit).mockResolvedValue({
    ...second,
    status: 'applied',
    entry_id: 7,
  })
  render(
    <QueryClientProvider client={new QueryClient()}>
      <PendingDebitDialog account={12} month='2025-10' />
    </QueryClientProvider>
  )
  fireEvent.click(
    screen.getByRole('button', { name: 'Collect outstanding charges' })
  )
  const source = await screen.findByText('Supplier statement')
  const section = source.closest('section')
  expect(section).not.toBeNull()
  expect(collectReviewedDebit).not.toHaveBeenCalled()
  if (section) {
    fireEvent.click(
      within(section).getByRole('button', { name: 'Confirm wallet debit' })
    )
  }
  await waitFor(() =>
    expect(collectReviewedDebit).toHaveBeenCalledWith(
      { account: 12, month: '2025-10' },
      second
    )
  )
  expect(await screen.findByText(/entry 7/)).toBeVisible()
  expect(
    screen.getAllByRole('button', { name: 'Confirm wallet debit' })
  ).toHaveLength(1)
})
