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
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, test, vi } from 'vitest'

import { getBillingUsageDetails } from '../api'
import { BillingRows } from '../components/billing-rows'
import { UsageDetailsDialog } from '../components/usage-details-dialog'

vi.mock('../api', async (original) => ({
  ...(await original<typeof import('../api')>()),
  getBillingUsageDetails: vi.fn(),
}))
const currency = {
  code: 'USD',
  symbol: '$',
  rate: '1',
  quota_per_unit: '500000',
}
const empty = {
  label: '00:00',
  charge: '0.000000',
  refund: '0.000000',
  amount: '0.000000',
  count: 0,
}

describe('Consumption drill-down', () => {
  beforeEach(() => vi.clearAllMocks())
  test('only populated hours are actionable; future hours do not display zero charges', async () => {
    const onSelect = vi.fn()
    render(
      <BillingRows
        rows={[
          empty,
          {
            ...empty,
            label: '09:00',
            count: 1,
            charge: '1.000000',
            amount: '1.000000',
          },
          { ...empty, label: '23:00', state: 'future' },
        ]}
        total={empty}
        symbol='$'
        onSelectRow={onSelect}
      />
    )
    expect(
      screen.queryByRole('button', { name: '00:00' })
    ).not.toBeInTheDocument()
    const future = screen.getByRole('row', { name: /23:00/ })
    expect(within(future).getByText(/Not yet occurred/)).toBeVisible()
    expect(within(future).queryByText(/0.000000/)).not.toBeInTheDocument()
    const user = userEvent.setup()
    screen.getByRole('button', { name: '09:00' }).focus()
    await user.keyboard('{Enter}')
    expect(onSelect).toHaveBeenCalledWith(
      expect.objectContaining({ label: '09:00', count: 1 })
    )
  })
  test('next page uses the selected account and hour; a failed page offers retry without stale rows', async () => {
    const longModel = 'example-model-'.repeat(30)
    vi.mocked(getBillingUsageDetails).mockResolvedValueOnce({
      currency,
      items: [
        {
          id: 1,
          created_at: 1580691600,
          request_id: 'request-one',
          model_name: longModel,
          kind: 'consumption',
          amount: '1.000000',
        },
      ],
      next_cursor: 'next-page',
    })
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    render(
      <QueryClientProvider client={client}>
        <UsageDetailsDialog
          userId={4}
          date='2020-02-03'
          hour={9}
          onClose={vi.fn()}
        />
      </QueryClientProvider>
    )
    expect(await screen.findByText(longModel)).toBeVisible()
    expect(getBillingUsageDetails).toHaveBeenCalledWith(4, '2020-02-03', 9, '')
    vi.mocked(getBillingUsageDetails).mockRejectedValueOnce(
      new Error('Unavailable')
    )
    fireEvent.click(screen.getByRole('button', { name: 'Next' }))
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Failed to load data'
    )
    expect(screen.queryByText('request-one')).not.toBeInTheDocument()
    expect(getBillingUsageDetails).toHaveBeenLastCalledWith(
      4,
      '2020-02-03',
      9,
      'next-page'
    )
    expect(screen.getByRole('button', { name: 'Next' })).toBeDisabled()
    vi.mocked(getBillingUsageDetails).mockResolvedValueOnce({
      currency,
      items: [],
      next_cursor: '',
    })
    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }))
    expect(
      await screen.findByText('No usage records for this hour.')
    ).toBeVisible()
  })
})
