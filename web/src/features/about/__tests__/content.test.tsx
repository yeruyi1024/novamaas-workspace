import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  createMemoryHistory,
  createRootRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen, within } from '@testing-library/react'
import { afterEach, beforeEach, expect, test } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'
import { useSystemConfigStore } from '@/stores/system-config-store'

import { About } from '../index'

const initialConfig = useSystemConfigStore.getState()
let client: QueryClient
beforeEach(() => {
  localStorage.clear()
  useAuthStore.getState().auth.reset()
  useSystemConfigStore.setState(initialConfig)
  useSystemConfigStore.getState().setConfig({ systemName: 'Test MaaS' })
  useSystemConfigStore.getState().setLoading(false)
})
afterEach(() => {
  client?.clear()
  useAuthStore.getState().auth.reset()
  useSystemConfigStore.setState(initialConfig)
  localStorage.clear()
})

test.each(['', '# Company story', 'https://example.com/about'])(
  'about content keeps visible upstream notices with administrator content %s',
  async (content) => {
    client = new QueryClient({
      defaultOptions: { queries: { retry: false, staleTime: Infinity } },
    })
    client.setQueryData(['about-content'], { success: true, data: content })
    client.setQueryData(['status'], {})
    client.setQueryData(['notice'], { success: true, data: '' })
    const root = createRootRoute({ component: About })
    const router = createRouter({
      routeTree: root,
      history: createMemoryHistory({ initialEntries: ['/'] }),
    })
    await router.load()
    render(
      <QueryClientProvider client={client}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    )
    const notice = screen.getByRole('region', {
      name: 'Open-source information',
    })
    expect(
      within(notice).getByText(
        'Frontend design and development by New API contributors.'
      )
    ).toBeVisible()
    expect(
      within(notice).getByRole('link', { name: 'New API · QuantumNous' })
    ).toHaveAttribute('href', 'https://github.com/QuantumNous/new-api')
    if (content.startsWith('https')) {
      expect(screen.getByTitle('About')).toHaveAttribute('src', content)
    } else {
      expect(
        screen.getByRole('heading', {
          level: 1,
          name: content ? 'Company story' : 'Test MaaS',
        })
      ).toBeVisible()
    }
  }
)
