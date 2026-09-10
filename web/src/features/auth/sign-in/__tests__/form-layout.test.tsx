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
  createMemoryHistory,
  createRootRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, test } from 'vitest'

import type { SystemStatus } from '@/features/auth/types'

import { UserAuthForm } from '../components/user-auth-form'

let queryClient: QueryClient

async function renderForm(status: Partial<SystemStatus>) {
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  })
  queryClient.setQueryData(['status'], status)

  const rootRoute = createRootRoute({ component: UserAuthForm })
  const router = createRouter({
    routeTree: rootRoute,
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })

  await router.load()
  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
}

afterEach(() => {
  queryClient?.clear()
})

test('legal consent appears before sign-in actions and unlocks them when accepted', async () => {
  await renderForm({
    password_login_enabled: true,
    user_agreement_enabled: true,
  })

  const user = userEvent.setup()
  const password = screen.getByLabelText('Password')
  const legalConsent = screen.getByRole('checkbox')
  const submit = screen.getByRole('button', { name: 'Sign in' })

  expect(password.compareDocumentPosition(legalConsent)).toBe(
    Node.DOCUMENT_POSITION_FOLLOWING
  )
  expect(legalConsent.compareDocumentPosition(submit)).toBe(
    Node.DOCUMENT_POSITION_FOLLOWING
  )
  expect(submit).toBeDisabled()

  await user.click(legalConsent)
  expect(submit).toBeEnabled()
})

test('alternative providers follow the password action under one separator', async () => {
  await renderForm({
    github_oauth: true,
    password_login_enabled: true,
  })

  const submit = screen.getByRole('button', { name: 'Sign in' })
  const github = screen.getByRole('button', { name: 'Continue with GitHub' })

  expect(submit.compareDocumentPosition(github)).toBe(
    Node.DOCUMENT_POSITION_FOLLOWING
  )
  expect(screen.getAllByText('Or continue with')).toHaveLength(1)
})

test('password sign-in accepts a username, email address, or phone number', async () => {
  await renderForm({ password_login_enabled: true })

  expect(
    screen.getByRole('textbox', {
      name: 'Username, Email or Phone Number',
    })
  ).toHaveAttribute('placeholder', 'Enter your username, email or phone number')
})
