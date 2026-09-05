import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, beforeEach, describe, expect, test } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'
import { useSystemConfigStore } from '@/stores/system-config-store'

import { Footer } from '../components/footer'
import { PublicLayout } from '../components/public-layout'

const initialConfig = useSystemConfigStore.getState()
let queryClient: QueryClient

async function renderPage(
  children: ReactNode,
  status: Record<string, unknown> = {}
) {
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  })
  queryClient.setQueryData(['status'], {
    user_agreement_enabled: true,
    privacy_policy_enabled: true,
    ...status,
  })
  queryClient.setQueryData(['notice'], { success: true, data: '' })
  const root = createRootRoute({ component: () => children })
  const destinations = [
    '/profile',
    '/wallet',
    '/system-settings/site/$section',
    '/dashboard',
    '/sign-in',
  ].map((path) =>
    createRoute({ getParentRoute: () => root, path, component: () => null })
  )
  const router = createRouter({
    routeTree: root.addChildren(destinations),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  await router.load()
  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
  return router
}

beforeEach(() => {
  localStorage.clear()
  useAuthStore.getState().auth.reset()
  useSystemConfigStore.setState(initialConfig)
  useSystemConfigStore.getState().setConfig({
    systemName: 'Test MaaS',
    footerHtml: '',
    demoSiteEnabled: false,
  })
  useSystemConfigStore.getState().setLoading(false)
})
afterEach(() => {
  queryClient?.clear()
  useAuthStore.getState().auth.reset()
  useSystemConfigStore.setState(initialConfig)
  localStorage.clear()
})

describe('Public footer', () => {
  test.each(['', '<p>Operator footer</p>'])(
    'keeps configured legal links and fork source without the appended upstream copyright (%s)',
    async (footerHtml) => {
      useSystemConfigStore.getState().setConfig({ footerHtml })
      await renderPage(<Footer />)
      const footer = screen.getByRole('contentinfo')
      expect(
        within(footer).getByRole('link', { name: 'User Agreement' })
      ).toHaveAttribute('href', '/user-agreement')
      expect(
        within(footer).getByRole('link', { name: 'Privacy Policy' })
      ).toHaveAttribute('href', '/privacy-policy')
      expect(
        within(footer).getByRole('link', { name: 'Source Code' })
      ).toHaveAttribute(
        'href',
        'https://github.com/yeruyi1024/novamaas-workspace'
      )
      expect(
        within(footer).queryByRole('link', { name: 'New API' })
      ).not.toBeInTheDocument()
      expect(footer).not.toHaveTextContent('projectAttributionSuffix')
      expect(footer).toHaveTextContent(
        footerHtml ? 'Operator footer' : 'Test MaaS'
      )
    }
  )
})

describe('Public account navigation', () => {
  test('anonymous visitors have sign-in access and no account menu', async () => {
    await renderPage(<PublicLayout appearance='maas'>Content</PublicLayout>)
    expect(
      screen.queryByRole('button', { name: 'Account menu' })
    ).not.toBeInTheDocument()
    expect(
      screen.getAllByRole('button', { name: 'Sign in' })[0]
    ).toHaveAttribute('href', '/sign-in')
  })

  test.each([1, 10, 100])(
    'role %i keeps profile and wallet access while system settings remain super-admin only',
    async (role) => {
      useAuthStore.getState().auth.setUser({ id: 1, username: 'preview', role })
      const router = await renderPage(
        <PublicLayout appearance='maas'>Content</PublicLayout>
      )
      const user = userEvent.setup()
      // Both the desktop and compact navigation render the original dropdown.
      const triggers = screen.getAllByRole('button', { name: 'Account menu' })
      expect(triggers).toHaveLength(2)
      for (const trigger of triggers) {
        await user.click(trigger)
        const menu = await screen.findByRole('menu')
        expect(
          within(menu).getByRole('menuitem', { name: 'Profile' })
        ).toBeVisible()
        expect(
          within(menu).getByRole('menuitem', { name: 'Wallet' })
        ).toBeVisible()
        if (role === 100) {
          expect(
            within(menu).getByRole('menuitem', { name: 'System Settings' })
          ).toBeVisible()
        } else {
          expect(
            within(menu).queryByRole('menuitem', { name: 'System Settings' })
          ).not.toBeInTheDocument()
        }
        await user.keyboard('{Escape}')
        await waitFor(() =>
          expect(screen.queryByRole('menu')).not.toBeInTheDocument()
        )
        expect(trigger).toHaveFocus()
      }
      await user.click(triggers[0])
      await user.click(
        await screen.findByRole('menuitem', {
          name: role === 100 ? 'System Settings' : 'Profile',
        })
      )
      await waitFor(() =>
        expect(router.state.location.pathname).toBe(
          role === 100 ? '/system-settings/site/system-info' : '/profile'
        )
      )
    }
  )

  test('disabled wallet stays hidden in the account menu', async () => {
    useAuthStore
      .getState()
      .auth.setUser({ id: 1, username: 'preview', role: 1 })
    await renderPage(<PublicLayout appearance='maas'>Content</PublicLayout>, {
      SidebarModulesAdmin: JSON.stringify({
        personal: { enabled: true, topup: false },
      }),
    })
    const user = userEvent.setup()
    await user.click(screen.getAllByRole('button', { name: 'Account menu' })[0])
    const menu = await screen.findByRole('menu')
    expect(
      within(menu).queryByRole('menuitem', { name: 'Wallet' })
    ).not.toBeInTheDocument()
  })

  test('compact navigation exposes its expanded state and makes closed links inert', async () => {
    await renderPage(<PublicLayout appearance='maas'>Content</PublicLayout>)
    const user = userEvent.setup()
    const toggle = screen.getByRole('button', {
      name: 'Toggle navigation menu',
    })
    const panel = document.querySelector(
      `#${toggle.getAttribute('aria-controls')}`
    )
    expect(toggle).toHaveAttribute('aria-expanded', 'false')
    expect(panel).toHaveAttribute('inert')
    await user.click(toggle)
    expect(toggle).toHaveAttribute('aria-expanded', 'true')
    expect(panel).not.toHaveAttribute('inert')
    await user.click(toggle)
    expect(toggle).toHaveAttribute('aria-expanded', 'false')
    expect(panel).toHaveAttribute('inert')
  })
})
