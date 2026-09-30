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

test('empty administrator content presents the gateway and planned sourcing and mobile sections', async () => {
  await renderHome('')
  const heroTitle = await screen.findByRole('heading', {
    level: 1,
    name: /One gateway to diverse AI supply Choose every token with purpose/,
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
      name: 'Access, source, operate',
    })
  ).toBeVisible()
  const sourcing = within(main).getByRole('region', {
    name: 'Make token supply visible before you buy',
  })
  const enterprise = within(main).getByRole('region', {
    name: 'Enterprise operations at every layer',
  })
  const miniProgram = within(main).getByRole('region', {
    name: 'AI results and usage, always close',
  })
  expect(sourcing).toHaveTextContent('Roadmap')
  expect(sourcing).toHaveTextContent('Review what matters')
  expect(sourcing).toHaveTextContent('Match the need')
  expect(enterprise).toHaveTextContent('security and compliance governance')
  expect(enterprise).toHaveTextContent('Compute leasing is next')
  expect(miniProgram).toHaveTextContent('Coming soon')
  expect(miniProgram).toHaveTextContent('Generation results')
  expect(miniProgram).toHaveTextContent('Token usage')
  expect(
    within(main).getByRole('heading', {
      level: 2,
      name: 'Use safer compute tokens',
    })
  ).toBeVisible()
  expect(
    within(miniProgram).getByRole('img', { name: 'WeChat mini program' })
  ).toBeVisible()
  expect(within(miniProgram).queryByText('Mini program concept')).toBeNull()
  expect(sourcing.compareDocumentPosition(enterprise)).toBe(
    Node.DOCUMENT_POSITION_FOLLOWING
  )
  expect(enterprise.compareDocumentPosition(miniProgram)).toBe(
    Node.DOCUMENT_POSITION_FOLLOWING
  )
  expect(within(main).getByTestId('home-hero-artwork')).toHaveAttribute(
    'aria-hidden',
    'true'
  )
  expect(
    within(main).getAllByRole('button', { name: 'Get Started' })
  ).toHaveLength(2)
  expect(within(main).getAllByText('Roadmap')).not.toHaveLength(0)
})

test('default homepage keeps the primary value statement compact over the hero image', async () => {
  await renderHome('')
  const heroTitle = await screen.findByRole('heading', {
    level: 1,
    name: /One gateway to diverse AI supply Choose every token with purpose/,
  })
  const description = screen.getByText(
    /Connect models and token channels through one gateway/
  )

  expect(heroTitle).toHaveClass('maas-hero-title')
  expect(heroTitle.parentElement).toHaveClass(
    'max-w-xl',
    'xl:max-w-[700px]',
    '2xl:max-w-[780px]'
  )
  expect(heroTitle.closest('[data-testid="home-hero-layout"]')).toHaveClass(
    'max-w-7xl',
    '2xl:max-w-[clamp(80rem,80vw,112rem)]'
  )
  expect(screen.getByRole('banner').firstElementChild).toHaveClass(
    '2xl:max-w-[clamp(80rem,80vw,112rem)]'
  )
  expect(description).toHaveClass('text-base', 'max-w-lg')
})

test('default homepage layers a decorative full-width image behind the value statement', async () => {
  await renderHome('')
  const layout = await screen.findByTestId('home-hero-layout')
  const heroTitle = await screen.findByRole('heading', {
    level: 1,
    name: /One gateway to diverse AI supply Choose every token with purpose/,
  })
  const hero = heroTitle.closest('section')
  const artwork = screen.getByTestId('home-hero-artwork')

  expect(hero).toHaveClass('maas-hero', 'overflow-hidden', 'md:min-h-[720px]')
  expect(artwork).toHaveClass('maas-hero-image', 'absolute', 'inset-0')
  expect(artwork).toHaveAttribute('aria-hidden', 'true')
  expect(layout).toContainElement(heroTitle)
  expect(screen.queryByRole('group', { name: 'Gateway workbench' })).toBeNull()
})

test('control plane diagrams explain supply convergence, policy routing and itemized settlement', async () => {
  await renderHome('')
  const controlPlane = await screen.findByRole('region', {
    name: 'Enterprise operations at every layer',
  })

  const supplyDiagram = within(controlPlane).getByRole('img', {
    name: 'Unify heterogeneous supply',
  })
  const routingDiagram = within(controlPlane).getByRole('img', {
    name: 'Route by policy',
  })
  const settlementDiagram = within(controlPlane).getByRole('img', {
    name: 'Settle with confidence',
  })

  expect(supplyDiagram).toHaveTextContent('Unified token pool')
  expect(supplyDiagram).toHaveTextContent('POST /v1')
  expect(routingDiagram).toHaveTextContent('Request')
  expect(
    routingDiagram.querySelector('[data-selected="true"]')
  ).toHaveTextContent('Claude')
  expect(settlementDiagram).toHaveTextContent('req_84C2')
  expect(settlementDiagram).toHaveTextContent('Metered settlement')
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
    screen.queryByRole('button', { name: 'Get Started' })
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
    screen.queryByRole('heading', { name: /One gateway to diverse AI supply/ })
  ).not.toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Sign in' })).toBeVisible()
})

test('administrator Markdown remains visible instead of the default presentation', async () => {
  await renderHome('# Company welcome')
  expect(
    await screen.findByRole('heading', { name: 'Company welcome' })
  ).toBeVisible()
  expect(
    screen.queryByRole('heading', { name: /One gateway to diverse AI supply/ })
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
    screen.queryByRole('heading', { name: /One gateway to diverse AI supply/ })
  ).not.toBeInTheDocument()
})
