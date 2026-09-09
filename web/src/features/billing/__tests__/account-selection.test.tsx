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
import userEvent from '@testing-library/user-event'
import { afterEach, expect, test, vi } from 'vitest'

import { searchUsers } from '@/features/users/api'
import { useAuthStore } from '@/stores/auth-store'

import { getBillingAccount, getBillingDay } from '../api'
import { Billing } from '../index'

vi.mock('@/features/users/api', async (original) => ({
  ...(await original<typeof import('@/features/users/api')>()),
  searchUsers: vi.fn(),
}))
vi.mock('../api', async (original) => ({
  ...(await original<typeof import('../api')>()),
  getBillingAccount: vi.fn(),
  getBillingDay: vi.fn(),
}))

afterEach(() => {
  useAuthStore.getState().auth.reset()
  vi.clearAllMocks()
})

test('a selected customer keeps its username and ownership after account search results change', async () => {
  useAuthStore.getState().auth.setUser({ id: 1, username: 'admin', role: 100 })
  const customer = {
    id: 4,
    username: 'example_customer',
    display_name: 'Customer',
    quota: 0,
    used_quota: 0,
    request_count: 0,
    group: 'default',
    status: 1,
    role: 1,
  }
  vi.mocked(searchUsers).mockResolvedValue({
    success: true,
    data: { items: [customer], page: 1, page_size: 30, total: 1 },
  })
  vi.mocked(getBillingDay).mockRejectedValue(new Error('No usage fixture'))
  vi.mocked(getBillingAccount).mockResolvedValue({
    user_id: 4,
    accounting_start_at: 0,
    company_title: 'Customer',
    tax_id: 'TAX',
    profile_version: 1,
  })
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <QueryClientProvider client={client}>
      <Billing />
    </QueryClientProvider>
  )
  const user = userEvent.setup()
  await user.click(screen.getByRole('combobox', { name: 'Account' }))
  await user.click(
    await screen.findByRole('option', { name: 'example_customer (#4)' })
  )
  expect(
    screen.getByText('Selected billing customer: example_customer (#4)')
  ).toBeVisible()
  vi.mocked(searchUsers).mockResolvedValue({
    success: true,
    data: {
      items: [{ ...customer, id: 5, username: 'another_customer' }],
      page: 1,
      page_size: 30,
      total: 1,
    },
  })
  fireEvent.change(screen.getByLabelText('Search accounts'), {
    target: { value: 'another' },
  })
  await waitFor(() =>
    expect(searchUsers).toHaveBeenCalledWith({
      keyword: 'another',
      page_size: 30,
    })
  )
  await user.click(screen.getByRole('combobox', { name: 'Account' }))
  expect(
    await screen.findByRole('option', { name: 'another_customer (#5)' })
  ).toBeVisible()
  await user.keyboard('{Escape}')
  expect(screen.getByRole('combobox', { name: 'Account' })).toHaveTextContent(
    'example_customer (#4)'
  )
  await user.click(screen.getByRole('tab', { name: 'Billing identity' }))
  await screen.findByLabelText('Company title')
  expect(getBillingAccount).toHaveBeenCalledWith(4)
  expect(
    screen.getByText('Selected billing customer: example_customer (#4)')
  ).toBeVisible()
})
