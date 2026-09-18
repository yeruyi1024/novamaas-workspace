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
import { afterEach, expect, test, vi } from 'vitest'

import { BASIC_CHECKS, CACHE_CHECKS } from './constants'

vi.mock('./hooks/use-supplier-test-run', () => ({
  useSupplierTestRun: () => ({
    runningModule: null,
    basicChecks: BASIC_CHECKS,
    cacheChecks: CACHE_CHECKS,
    streamText: '',
    basicStreamText: '',
    progress: { completed: 0, total: 0 },
    metrics: null,
    cacheMetrics: null,
    summaries: { basic: '', cache: '', stress: '' },
    errorMessage: '',
    start: vi.fn(),
    stop: vi.fn(),
  }),
}))

const { SupplierTest } = await import('./index')

let queryClient: QueryClient

function renderPage() {
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <SupplierTest />
    </QueryClientProvider>
  )
}

afterEach(() => {
  queryClient?.clear()
  sessionStorage.clear()
})

test('target sits above tests, modules are tabbed, and judgment starts collapsed', async () => {
  renderPage()
  const user = userEvent.setup()

  const target = screen.getByText('Target')
  const tests = screen.getByText('Tests')
  const judgment = screen.getByText('Judgment standard')
  expect(screen.getByText('Vendor')).toBeInTheDocument()
  expect(
    screen.getByText(
      'Using generic OpenAI-compatible checks. Missing vendor fields are skipped.'
    )
  ).toBeInTheDocument()
  expect(target.compareDocumentPosition(tests)).toBe(
    Node.DOCUMENT_POSITION_FOLLOWING
  )
  expect(tests.compareDocumentPosition(judgment)).toBe(
    Node.DOCUMENT_POSITION_FOLLOWING
  )

  expect(
    screen.getByRole('button', { name: 'Run all basic checks' })
  ).toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: 'Run cache test' })
  ).not.toBeInTheDocument()
  expect(screen.queryByLabelText('Error rate (%)')).not.toBeInTheDocument()

  await user.click(screen.getByRole('tab', { name: 'Cache test' }))
  expect(
    screen.getByRole('button', { name: 'Run cache test' })
  ).toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: 'Run all basic checks' })
  ).not.toBeInTheDocument()

  await user.click(screen.getByRole('button', { name: /Expand/ }))
  expect(screen.getByLabelText('Error rate (%)')).toBeInTheDocument()
})
