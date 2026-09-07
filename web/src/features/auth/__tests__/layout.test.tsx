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
import {
  createMemoryHistory,
  createRootRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, test } from 'vitest'

import { useSystemConfigStore } from '@/stores/system-config-store'

import { AuthLayout } from '../auth-layout'

const initialConfig = useSystemConfigStore.getState()

async function renderAuthLayout(showcase = false) {
  const rootRoute = createRootRoute({
    component: () => (
      <AuthLayout showcase={showcase ? <div>Brand story</div> : undefined}>
        <div>Authentication content</div>
      </AuthLayout>
    ),
  })
  const router = createRouter({
    routeTree: rootRoute,
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })

  await router.load()
  return render(<RouterProvider router={router} />)
}

beforeEach(() => {
  useSystemConfigStore.setState(initialConfig)
  useSystemConfigStore.getState().setConfig({
    systemName: 'Test MaaS',
    logo: '/logo.svg',
  })
  useSystemConfigStore.getState().setLoading(false)
})

afterEach(() => {
  useSystemConfigStore.setState(initialConfig)
})

describe('Auth layout presentation', () => {
  test('showcase scales within the viewport and reveals the brand story at desktop width', async () => {
    await renderAuthLayout(true)

    const layout = screen.getByTestId('auth-layout')
    const shell = screen.getByTestId('auth-shell')
    const showcase = screen.getByRole('complementary', {
      name: 'AI supply infrastructure',
    })

    expect(layout).toHaveClass(
      'items-center',
      'py-20',
      'lg:py-[clamp(6rem,10vh,9rem)]'
    )
    expect(shell).toHaveClass(
      'max-w-[30rem]',
      'lg:max-w-[76rem]',
      'lg:min-h-[clamp(30rem,65svh,42rem)]',
      'lg:gap-[clamp(3rem,6vw,6rem)]',
      'lg:grid-cols-[minmax(0,1.08fr)_minmax(25rem,0.92fr)]'
    )
    expect(shell).not.toHaveClass(
      'border',
      'rounded-[2rem]',
      'shadow-xl',
      'bg-card/80'
    )
    expect(showcase).toHaveClass('hidden', 'lg:flex')
    expect(showcase).toHaveTextContent('Brand story')
    expect(screen.getByTestId('auth-content')).toHaveTextContent(
      'Authentication content'
    )
    expect(screen.getByRole('link', { name: /Test MaaS/ })).toHaveAttribute(
      'href',
      '/'
    )
    expect(
      screen.getByRole('button', { name: 'Change language' })
    ).toBeVisible()
    expect(screen.getByRole('button', { name: 'Toggle theme' })).toBeVisible()
  })

  test('default auth pages remain single-column without the desktop showcase', async () => {
    await renderAuthLayout()

    expect(screen.queryByRole('complementary')).not.toBeInTheDocument()
    expect(screen.queryByTestId('auth-shell')).not.toBeInTheDocument()
    expect(screen.getByTestId('auth-layout')).toHaveClass('items-center')
    expect(screen.getByText('Authentication content')).toBeVisible()
  })
})
