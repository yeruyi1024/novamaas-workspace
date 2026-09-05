import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  createMemoryHistory,
  createRootRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen } from '@testing-library/react'
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

test('empty administrator content shows the MaaS home and configured documentation link', async () => {
  await renderHome('')
  expect(
    await screen.findByRole('heading', {
      level: 1,
      name: /Unified model services/,
    })
  ).toBeVisible()
  expect(screen.getByRole('button', { name: 'Docs' })).toHaveAttribute(
    'href',
    'https://docs.example.com'
  )
  expect(screen.getByText('One model service catalog')).toBeVisible()
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
    screen.queryByRole('heading', { name: /Unified model services/ })
  ).not.toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Sign in' })).toBeVisible()
})

test('administrator Markdown remains visible instead of the default presentation', async () => {
  await renderHome('# Company welcome')
  expect(
    await screen.findByRole('heading', { name: 'Company welcome' })
  ).toBeVisible()
  expect(
    screen.queryByRole('heading', { name: /Unified model services/ })
  ).not.toBeInTheDocument()
})
