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
import { render, screen, waitFor, within } from '@testing-library/react'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'
import { useSystemConfigStore } from '@/stores/system-config-store'

import { Home } from '../index'

const initialConfig = useSystemConfigStore.getState()
const initialAdapter = api.defaults.adapter
let client: QueryClient

beforeEach(() => {
  localStorage.clear()
  useAuthStore.getState().auth.reset()
  useSystemConfigStore.setState(initialConfig)
  useSystemConfigStore
    .getState()
    .setConfig({ systemName: 'Test MaaS', footerHtml: '' })
  useSystemConfigStore.getState().setLoading(false)
  const matchMedia = window.matchMedia
  vi.spyOn(window, 'matchMedia').mockImplementation((query) => ({
    ...matchMedia(query),
    matches: query.includes('prefers-reduced-motion'),
  }))
})
afterEach(() => {
  client?.clear()
  api.defaults.adapter = initialAdapter
  useAuthStore.getState().auth.reset()
  useSystemConfigStore.setState(initialConfig)
  localStorage.clear()
})

async function renderHome(content: string) {
  api.defaults.adapter = async (config) => ({
    data: { success: true, data: content },
    status: 200,
    statusText: 'OK',
    headers: {},
    config,
  })
  client = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  })
  client.setQueryData(['status'], { docs_link: 'https://docs.example.com' })
  client.setQueryData(['notice'], { success: true, data: '' })
  const root = createRootRoute({ component: Home })
  const router = createRouter({
    routeTree: root,
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  await router.load()
  return render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
}

test('empty administrator content presents the AI supply value chain and configured documentation link', async () => {
  await renderHome('')
  const heroTitle = await screen.findByRole('heading', {
    level: 1,
    name: /Turn fragmented AI supply into one programmable market/,
  })
  const main = screen.getByRole('main')
  expect(heroTitle).toBeVisible()
  expect(screen.getByRole('button', { name: 'Docs' })).toHaveAttribute(
    'href',
    'https://docs.example.com'
  )
  expect(
    within(main).getByRole('heading', {
      level: 2,
      name: 'One platform, three layers of value',
    })
  ).toBeVisible()
  expect(
    within(main).getByRole('group', { name: 'AI supply network' })
  ).toBeVisible()
  expect(within(main).getAllByText('Roadmap')).not.toHaveLength(0)
})

test('default homepage applies the balanced typography treatment to its primary value statement', async () => {
  await renderHome('')
  expect(
    await screen.findByRole('heading', {
      level: 1,
      name: /Turn fragmented AI supply into one programmable market/,
    })
  ).toHaveClass('maas-hero-title')
})

test('default homepage stacks the value statement above the centered supply network', async () => {
  await renderHome('')
  const layout = await screen.findByTestId('home-hero-layout')
  const heroTitle = await screen.findByRole('heading', {
    level: 1,
    name: /Turn fragmented AI supply into one programmable market/,
  })
  const network = screen.getByRole('group', { name: 'AI supply network' })

  expect(layout).toHaveClass('flex-col', 'items-center')
  expect(heroTitle.compareDocumentPosition(network)).toBe(
    Node.DOCUMENT_POSITION_FOLLOWING
  )
})

test('default homepage omits the project source code entry', async () => {
  await renderHome('')
  const footer = await screen.findByRole('contentinfo')
  expect(
    within(footer).queryByRole('link', { name: 'Source Code' })
  ).not.toBeInTheDocument()
})

test('authenticated visitors keep dashboard access from the homepage', async () => {
  useAuthStore.getState().auth.setUser({ id: 1, username: 'member', role: 1 })
  await renderHome('')
  expect(
    await screen.findByRole('button', { name: 'Go to Dashboard' })
  ).toHaveAttribute('href', '/dashboard')
  expect(
    screen.queryByRole('button', { name: 'Start building' })
  ).not.toBeInTheDocument()
})

test('administrator URL takes precedence over the default home while keeping its sandbox and header', async () => {
  await renderHome('https://example.com/company')
  const frame = await screen.findByTitle('Custom Home Page')
  expect(frame).toHaveAttribute('src', 'https://example.com/company')
  expect(frame.getAttribute('sandbox')).toContain(
    'allow-top-navigation-by-user-activation'
  )
  expect(frame.getAttribute('sandbox')).not.toContain('allow-same-origin')
  expect(
    screen.queryByRole('heading', { name: /Turn fragmented AI supply/ })
  ).not.toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Sign in' })).toBeVisible()
})

test('administrator Markdown remains visible instead of the default presentation', async () => {
  await renderHome('# Company welcome')
  expect(
    await screen.findByRole('heading', { name: 'Company welcome' })
  ).toBeVisible()
  expect(
    screen.queryByRole('heading', { name: /Turn fragmented AI supply/ })
  ).not.toBeInTheDocument()
})

test('administrator HTML remains isolated and takes precedence over the default presentation', async () => {
  const view = await renderHome(
    '<section><h1>Custom company</h1><script>window.compromised = true</script></section>'
  )

  const shadowRoot = await waitFor(() => {
    const root = view.container.querySelector<HTMLElement>(
      '.custom-home-content'
    )?.shadowRoot
    if (!root) throw new Error('Custom home shadow root is not ready')
    return root
  })
  expect(shadowRoot.textContent).toContain('Custom company')
  expect(shadowRoot.querySelector('script')).toBeNull()
  expect(
    screen.queryByRole('heading', { name: /Turn fragmented AI supply/ })
  ).not.toBeInTheDocument()
})
