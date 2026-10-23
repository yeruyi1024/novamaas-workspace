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
  createRoute,
  createRouter,
  getRouteApi,
  Outlet,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, test, vi } from 'vitest'

import { useIsAdmin } from '@/hooks/use-admin'
import { api } from '@/lib/api'

import { StudioTaskLogs } from '../components/studio-task-logs'
import { UsageLogsProvider } from '../components/usage-logs-provider'

vi.mock('@/lib/api', () => ({ api: { get: vi.fn() } }))
vi.mock('@/hooks/use-admin', () => ({ useIsAdmin: vi.fn(() => false) }))
const routeApi = getRouteApi('/_authenticated/usage-logs/$section')
const client = new QueryClient()
afterEach(() => {
  client.clear()
  vi.mocked(useIsAdmin).mockReturnValue(false)
})

function TestPage() {
  const { section } = routeApi.useParams()
  if (section === 'common') return <p>Consumption log</p>
  return (
    <UsageLogsProvider>
      <StudioTaskLogs />
    </UsageLogsProvider>
  )
}

async function openLogs() {
  const root = createRootRoute({ component: Outlet })
  const authenticated = createRoute({
    getParentRoute: () => root,
    id: '_authenticated',
    component: Outlet,
  })
  const logs = createRoute({
    getParentRoute: () => authenticated,
    path: 'usage-logs/$section',
    validateSearch: (search) => search,
    component: TestPage,
  })
  const router = createRouter({
    routeTree: root.addChildren([authenticated.addChildren([logs])]),
    history: createMemoryHistory({
      initialEntries: [
        '/usage-logs/drawing?drawingSource=studio&startTime=100000&endTime=200000&page=1&pageSize=20',
      ],
    }),
  })
  await router.load()
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
  return router
}

test('task/request filtering keeps the source and the request link opens its consumption log', async () => {
  vi.mocked(api.get).mockResolvedValue({
    data: {
      success: true,
      data: {
        total: 1,
        items: [
          {
            id: 'run-1',
            owner_user_id: 7,
            model: 'image2',
            status: 'completed',
            request_id: 'request-1',
            created_at: 150,
            updated_at: 160,
            duration_ms: 1500,
          },
        ],
      },
    },
  })
  const router = await openLogs()
  const user = userEvent.setup()
  await screen.findByRole('link', { name: 'request-1' })
  expect(screen.getByRole('textbox', { name: 'Task ID' })).toHaveAttribute(
    'placeholder',
    'Filter by task or request ID'
  )
  expect(api.get).toHaveBeenCalledWith('/api/image-studio/logs', {
    params: expect.objectContaining({
      p: 1,
      page_size: 20,
      scope: 'self',
      start_timestamp: 100,
      end_timestamp: 200,
    }),
  })
  await user.type(
    screen.getByRole('textbox', { name: 'Task ID' }),
    'request-1{Enter}'
  )
  await waitFor(() =>
    expect(router.state.location.search).toMatchObject({
      drawingSource: 'studio',
      filter: 'request-1',
      page: 1,
    })
  )
  await waitFor(() =>
    expect(api.get).toHaveBeenLastCalledWith('/api/image-studio/logs', {
      params: expect.objectContaining({ filter: 'request-1' }),
    })
  )
  await user.click(screen.getByRole('button', { name: 'Reset' }))
  await waitFor(() =>
    expect(router.state.location.search).toMatchObject({
      drawingSource: 'studio',
      page: 1,
    })
  )
  expect(router.state.location.search).not.toHaveProperty('filter')
  expect(screen.getByRole('textbox', { name: 'Task ID' })).toHaveValue('')
  await user.click(await screen.findByRole('link', { name: 'request-1' }))
  await screen.findByText('Consumption log')
  expect(router.state.location.pathname).toBe('/usage-logs/common')
  expect(router.state.location.search).toMatchObject({
    requestId: 'request-1',
    startTime: 90000,
    endTime: 220000,
  })
})

test('a failed task-log request displays an error instead of silently appearing as successful empty data', async () => {
  vi.mocked(api.get).mockRejectedValue(new Error('unavailable'))
  await openLogs()
  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Failed to load logs'
  )
})

test('admin channel filtering uses the selected channel and retains the image source', async () => {
  vi.mocked(useIsAdmin).mockReturnValue(true)
  vi.mocked(api.get).mockResolvedValue({
    data: {
      success: true,
      data: {
        total: 1,
        items: [
          {
            id: 'run-channel',
            owner_user_id: 7,
            channel_id: 23,
            model: 'image2',
            status: 'failed',
            created_at: 150,
            updated_at: 160,
            duration_ms: 1500,
            error_message: 'provider unavailable',
          },
        ],
      },
    },
  })
  const router = await openLogs()
  const user = userEvent.setup()
  await screen.findByText('#23')
  expect(screen.getByText('provider unavailable')).toBeVisible()
  await user.type(screen.getByPlaceholderText('Channel ID'), '23{Enter}')
  await waitFor(() =>
    expect(api.get).toHaveBeenLastCalledWith('/api/image-studio/logs', {
      params: expect.objectContaining({ scope: 'all', channel_id: 23 }),
    })
  )
  expect(router.state.location.search).toMatchObject({
    drawingSource: 'studio',
    channel: '23',
    page: 1,
  })
})
