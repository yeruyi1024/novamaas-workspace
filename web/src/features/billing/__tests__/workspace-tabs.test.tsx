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
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { getUpstreamGroups } from '@/features/upstream-billing/api'
import { Route as BillingRoute } from '@/routes/_authenticated/billing'
import { Route as LegacyUpstreamRoute } from '@/routes/_authenticated/upstream-billing'
import { useAuthStore } from '@/stores/auth-store'

import { getBillingAccount, getBillingDay, searchBillingAccounts } from '../api'
import { Billing } from '../index'

vi.mock('../api', async (original) => ({
  ...(await original<typeof import('../api')>()),
  getBillingAccount: vi.fn(),
  getBillingDay: vi.fn(),
  searchBillingAccounts: vi.fn(),
}))
vi.mock('@/features/upstream-billing/api', async (original) => ({
  ...(await original<typeof import('@/features/upstream-billing/api')>()),
  getUpstreamGroups: vi.fn(),
}))

beforeEach(() => {
  vi.resetAllMocks()
  vi.mocked(getBillingDay).mockRejectedValue(new Error('No usage'))
  vi.mocked(getBillingAccount).mockResolvedValue({
    user_id: 1,
    company_title: '',
    tax_id: '',
    profile_version: 1,
    accounting_start_at: 0,
  })
  vi.mocked(searchBillingAccounts).mockResolvedValue({ items: [] })
  vi.mocked(getUpstreamGroups).mockResolvedValue([])
})
afterEach(() => useAuthStore.getState().auth.reset())

test('financial administrators can switch from customer statements to upstream statements without preloading upstream data', async () => {
  useAuthStore.getState().auth.setUser({
    id: 1,
    username: 'accountant',
    role: 10,
    permissions: {
      admin_permissions: { financial_accounting: { view: true } },
    },
  })
  render(
    <QueryClientProvider
      client={
        new QueryClient({ defaultOptions: { queries: { retry: false } } })
      }
    >
      <Billing />
    </QueryClientProvider>
  )
  expect(
    screen.getByRole('tab', { name: 'Customer statements' })
  ).toHaveAttribute('aria-selected', 'true')
  expect(getUpstreamGroups).not.toHaveBeenCalled()
  fireEvent.click(screen.getByRole('tab', { name: 'Upstream statements' }))
  await waitFor(() => expect(getUpstreamGroups).toHaveBeenCalledOnce())
  expect(
    screen.getByRole('button', { name: 'Create upstream group' })
  ).toBeVisible()
})

test('ordinary accounts cannot expose or request upstream data even when the upstream area is selected', () => {
  useAuthStore.getState().auth.setUser({
    id: 2,
    username: 'customer',
    role: 1,
    permissions: {
      admin_permissions: { financial_management: { view: true } },
    },
  })
  render(
    <QueryClientProvider
      client={
        new QueryClient({ defaultOptions: { queries: { retry: false } } })
      }
    >
      <Billing area='upstream' />
    </QueryClientProvider>
  )
  expect(
    screen.queryByRole('tab', { name: 'Upstream statements' })
  ).not.toBeInTheDocument()
  expect(screen.getByRole('tab', { name: 'Consumption lookup' })).toBeVisible()
  expect(getUpstreamGroups).not.toHaveBeenCalled()
})

test('legacy upstream links redirect to the billing upstream tab', async () => {
  const load = LegacyUpstreamRoute.options.beforeLoad
  expect(typeof load).toBe('function')
  await expect(
    Promise.resolve().then(
      () => typeof load === 'function' && load({} as never)
    )
  ).rejects.toMatchObject({
    options: { to: '/billing', search: { tab: 'upstream' }, replace: true },
  })
})

test('the billing route redirects an ordinary upstream deep link back to customer statements', async () => {
  useAuthStore.getState().auth.setUser({ id: 2, username: 'customer', role: 1 })
  const load = BillingRoute.options.beforeLoad
  await expect(
    Promise.resolve().then(
      () =>
        typeof load === 'function' &&
        load({ search: { tab: 'upstream' } } as never)
    )
  ).rejects.toMatchObject({
    options: { to: '/billing', search: { tab: 'customer' }, replace: true },
  })
})
