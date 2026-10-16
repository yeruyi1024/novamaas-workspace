import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, test, vi } from 'vitest'

import { ApplicationCenter } from '../components/application-center'

let client: QueryClient
afterEach(() => client?.clear())

async function renderMenu(enabled = true, onNavigate?: () => void) {
  client = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  })
  client.setQueryData(['status'], { ImageStudioEnabled: enabled })
  const root = createRootRoute({ component: Outlet })
  const home = createRoute({
    getParentRoute: () => root,
    path: '/',
    component: () => <ApplicationCenter onNavigate={onNavigate} />,
  })
  const studio = createRoute({
    getParentRoute: () => root,
    path: '/image-studio',
    component: () => <main>Image editor</main>,
  })
  const router = createRouter({
    routeTree: root.addChildren([home, studio]),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  await router.load()
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
  return userEvent.setup()
}

test('hovering application center exposes image creation and clicking enters the editor directly', async () => {
  const onNavigate = vi.fn()
  const user = await renderMenu(true, onNavigate)
  const trigger = screen.getByRole('button', { name: 'Application center' })
  expect(trigger).toHaveClass('max-lg:size-8')
  expect(screen.getByText('Application center')).toHaveClass(
    'sr-only',
    'lg:not-sr-only'
  )
  expect(trigger).toHaveAttribute('aria-expanded', 'false')
  await user.hover(trigger)
  const link = await screen.findByRole('link', { name: 'Image creation' })
  expect(trigger).toHaveAttribute('aria-expanded', 'true')
  expect(link).toHaveAttribute('href', '/image-studio')
  await user.click(link)
  expect(await screen.findByText('Image editor')).toBeVisible()
  expect(onNavigate).toHaveBeenCalledOnce()
})

test('keyboard users can open application center and dismiss it with Escape', async () => {
  const user = await renderMenu()
  const trigger = screen.getByRole('button', { name: 'Application center' })
  await user.tab()
  expect(trigger).toHaveFocus()
  await user.keyboard('{Enter}')
  expect(
    await screen.findByRole('link', { name: 'Image creation' })
  ).toBeVisible()
  await user.keyboard('{Escape}')
  await waitFor(() => expect(trigger).toHaveAttribute('aria-expanded', 'false'))
  expect(trigger).toHaveFocus()
})

test('disabling image creation keeps application center but removes its image creation link', async () => {
  const user = await renderMenu(false)
  const trigger = screen.getByRole('button', { name: 'Application center' })
  expect(trigger).toBeVisible()
  await user.click(trigger)
  expect(
    screen.queryByRole('link', { name: 'Image creation' })
  ).not.toBeInTheDocument()
})
