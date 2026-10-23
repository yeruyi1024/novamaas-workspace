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
  focusManager,
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query'
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} from '@tanstack/react-router'
import { act, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AxiosError, AxiosHeaders } from 'axios'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { getApiKeys } from '@/features/keys/api'
import type { ApiKey } from '@/features/keys/types'
import { useAuthStore } from '@/stores/auth-store'

import * as api from '../api'
import { ImageStudio } from '../image-studio'
import type { StudioRun } from '../types'

vi.mock('../api')
vi.mock('@/features/keys/api')

const completed: StudioRun = {
  id: 'run-1',
  owner_user_id: 7,
  status: 'completed',
  prompt: 'A mountain',
  created_at: 1,
  model: 'dall-e-3',
  quota: 0,
  results: [
    {
      id: 'result-1',
      index: 0,
      preview_url: 'https://example.test/original.png',
      thumbnail_url: 'https://example.test/thumbnail.png',
    },
  ],
}
const key = { id: 1, name: 'Image key', status: 1 } as ApiKey

function promptInput() {
  return screen.getByRole('textbox', { name: 'Describe your image' })
}
function createImages() {
  return screen.getByRole('button', { name: 'Create images' })
}

function showRuns(...items: StudioRun[]) {
  vi.mocked(api.listStudioRuns).mockResolvedValue({
    items,
    total: items.length,
  })
}

let client: QueryClient
beforeEach(() => {
  client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  client.setQueryData(['status'], { ImageStudioEnabled: true })
  useAuthStore.getState().auth.setBundle({
    access_token: 'studio-test-session',
    token_type: 'Bearer',
    access_expires_at: 2000000000,
    user: { id: 7, username: 'artist', role: 1, group: 'default' },
    session: {
      sid: 'studio-session',
      current: true,
      login_method: 'password',
      ip: '127.0.0.1',
      user_agent: 'vitest',
      created_at: 1,
      last_active_at: 1,
      expires_at: 2000000000,
    },
  })
  vi.mocked(api.getStudioSettings).mockResolvedValue({
    enabled: true,
    max_images: 4,
    max_prompt_characters: 1000,
    max_reference_images: 8,
    max_reference_bytes: 20 * 1024 * 1024,
    max_running_runs: 5,
  })
  vi.mocked(getApiKeys).mockResolvedValue({
    success: true,
    data: { items: [key], total: 1, page: 1, page_size: 100 },
  })
  vi.mocked(api.listStudioModels).mockImplementation(async (_id, kind) =>
    kind === 'text' ? ['text-model'] : ['dall-e-3']
  )
  vi.mocked(api.createStudioRun).mockResolvedValue(completed)
  showRuns()
  vi.mocked(api.pollStudioRuns).mockResolvedValue({ items: [], total: 0 })
  vi.mocked(api.listStudioReferences).mockResolvedValue({ items: [], total: 0 })
})
afterEach(() => {
  client.clear()
  useAuthStore.getState().auth.reset()
  vi.resetAllMocks()
})

async function openStudio() {
  const root = createRootRoute({ component: Outlet })
  const studio = createRoute({
    getParentRoute: () => root,
    path: '/image-studio',
    component: ImageStudio,
  })
  const router = createRouter({
    routeTree: root.addChildren([studio]),
    history: createMemoryHistory({ initialEntries: ['/image-studio'] }),
  })
  await router.load()
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
  await screen.findByText('dall-e-3', { exact: true })
  return userEvent.setup()
}

test('guests can try examples without loading private records and signing out clears the draft', async () => {
  const user = await openStudio()
  await user.type(promptInput(), 'Private draft')
  await act(() => useAuthStore.getState().auth.reset())
  expect(promptInput()).toHaveValue('')
  expect(
    screen.getByRole('link', { name: 'Sign in to create images' })
  ).toBeVisible()
  vi.mocked(getApiKeys).mockClear()
  vi.mocked(api.listStudioRuns).mockClear()
  await user.click(
    screen.getByRole('button', { name: 'Cinematic photography' })
  )
  expect(promptInput()).not.toHaveValue('')
  expect(getApiKeys).not.toHaveBeenCalled()
  expect(api.listStudioRuns).not.toHaveBeenCalled()
})

test('accepted tasks leave a blank editor, allow another task, and open details only on selection', async () => {
  const first: StudioRun = { ...completed, status: 'running', results: [] }
  const second = { ...first, id: 'run-2', prompt: 'Second composition' }
  vi.mocked(api.createStudioRun)
    .mockResolvedValueOnce(first)
    .mockResolvedValueOnce(second)
  vi.mocked(api.pollStudioRuns).mockResolvedValue({ items: [first], total: 1 })
  const user = await openStudio()
  showRuns(first)
  await user.type(promptInput(), 'A mountain')
  await user.click(createImages())
  await waitFor(() => expect(promptInput()).toHaveValue(''))
  expect(promptInput()).toHaveFocus()
  expect(screen.getByRole('heading', { name: /Imagine it/ })).toBeVisible()
  showRuns(second, first)
  vi.mocked(api.pollStudioRuns).mockResolvedValue({
    items: [first, second],
    total: 2,
  })
  await user.type(promptInput(), second.prompt)
  await user.click(createImages())
  await waitFor(() => expect(api.createStudioRun).toHaveBeenCalledTimes(2))
  expect(
    vi.mocked(api.createStudioRun).mock.calls[0][0].idempotency_key
  ).not.toBe(vi.mocked(api.createStudioRun).mock.calls[1][0].idempotency_key)
  const history = screen.getByRole('region', { name: 'Generation records' })
  expect(
    await within(history).findByRole('button', { name: /Second composition/ })
  ).toHaveTextContent('Running')
  showRuns(completed, second)
  vi.mocked(api.pollStudioRuns).mockResolvedValue({
    items: [completed, second],
    total: 2,
  })
  await waitFor(
    () =>
      expect(
        within(history).getByRole('button', { name: /A mountain/ })
      ).toHaveTextContent('Completed'),
    { timeout: 4000 }
  )
  expect(
    screen.queryByRole('button', { name: 'Preview' })
  ).not.toBeInTheDocument()
  vi.mocked(api.getStudioRun).mockResolvedValue(completed)
  await user.click(within(history).getByRole('button', { name: /A mountain/ }))
  expect(await screen.findByRole('button', { name: 'Preview' })).toBeVisible()
  await user.click(screen.getByRole('button', { name: 'New creation' }))
  expect(screen.queryByRole('button', { name: 'Preview' })).toBeNull()
})

test('failed submissions preserve the draft and retry with the same idempotency key', async () => {
  vi.mocked(api.createStudioRun).mockRejectedValueOnce(new Error('network'))
  const user = await openStudio()
  await user.type(promptInput(), 'A mountain')
  await user.click(createImages())
  expect(await screen.findByRole('alert')).toHaveTextContent('network')
  expect(promptInput()).toHaveValue('A mountain')
  await user.click(createImages())
  await waitFor(() => expect(promptInput()).toHaveValue(''))
  const calls = vi.mocked(api.createStudioRun).mock.calls
  expect(calls).toHaveLength(2)
  expect(calls[1][0].idempotency_key).toBe(calls[0][0].idempotency_key)
  expect(api.saveStudioResult).not.toHaveBeenCalled()
})

test('stored reference IDs are reused and submission clears references and returns to generation records', async () => {
  vi.mocked(api.listStudioReferences).mockResolvedValue({
    items: [
      {
        id: 'stored',
        url: 'https://example.test/reference.png',
        expires_at: 2000000000,
      },
    ],
    total: 1,
  })
  const user = await openStudio()
  await user.click(screen.getByRole('tab', { name: 'Upload history' }))
  await user.click(
    await screen.findByRole('button', { name: 'Add reference image 1' })
  )
  await user.type(promptInput(), 'Keep this image')
  expect(screen.getByAltText('Reference image 1')).toBeVisible()
  await user.click(createImages())
  await waitFor(() => expect(promptInput()).toHaveValue(''))
  expect(
    vi.mocked(api.createStudioRun).mock.calls[0][0].request.reference_ids
  ).toEqual(['stored'])
  expect(api.uploadStudioReference).not.toHaveBeenCalled()
  expect(screen.queryByAltText('Reference image 1')).not.toBeInTheDocument()
  expect(
    screen.getByRole('tab', { name: 'Generation records' })
  ).toHaveAttribute('aria-selected', 'true')
})

test('reference uploads preserve IDs and reject invalid formats before upload', async () => {
  vi.mocked(api.uploadStudioReference).mockResolvedValue({
    id: 'uploaded',
    url: 'https://example.test/upload.png',
    expires_at: 2000000000,
  })
  const user = userEvent.setup({ applyAccept: false })
  await openStudio()
  const input = screen.getByLabelText('Upload reference images')
  await user.upload(
    input,
    new File(['invalid'], 'file.pdf', { type: 'application/pdf' })
  )
  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Reference images must be PNG, JPEG or WebP'
  )
  expect(api.uploadStudioReference).not.toHaveBeenCalled()
  await user.upload(
    input,
    new File(['image'], 'reference.png', { type: 'image/png' })
  )
  expect(await screen.findByAltText('Reference image 1')).toBeVisible()
  await user.type(promptInput(), 'Use the uploaded image')
  await user.click(createImages())
  await waitFor(() => expect(api.createStudioRun).toHaveBeenCalled())
  expect(
    vi.mocked(api.createStudioRun).mock.calls[0][0].request.reference_ids
  ).toEqual(['uploaded'])
})

test('the records sidebar starts collapsed and can open without losing the draft', async () => {
  const user = await openStudio()
  const toggle = screen.getByRole('button', { name: 'Show generation records' })
  expect(toggle).toHaveAttribute('aria-expanded', 'false')
  await user.type(promptInput(), 'Keep my draft')
  await user.click(
    screen.getByRole('button', { name: 'Collapse prompt editor' })
  )
  expect(
    screen.queryByRole('textbox', { name: 'Describe your image' })
  ).not.toBeInTheDocument()
  await user.click(screen.getByRole('button', { name: 'Show prompt editor' }))
  expect(promptInput()).toHaveValue('Keep my draft')
  await user.click(toggle)
  expect(toggle).toHaveAttribute('aria-expanded', 'true')
})

test('the create button is disabled with a notice while the running task limit is reached', async () => {
  const running = Array.from({ length: 5 }, (_, index) => ({
    ...completed,
    id: `run-${index}`,
    status: 'running' as const,
    results: [],
  }))
  showRuns(...running)
  vi.mocked(api.pollStudioRuns).mockResolvedValue({
    items: running,
    total: running.length,
  })
  const user = await openStudio()
  await user.type(promptInput(), 'Another mountain')
  expect(
    await screen.findByText(/Up to 5 images can be generating at once/)
  ).toBeVisible()
  expect(createImages()).toBeDisabled()
})

test('adding a task or regaining focus during a 429 cooldown does not restart polling early', async () => {
  const first = { ...completed, status: 'running' as const, results: [] }
  const second = { ...first, id: 'run-during-cooldown' }
  showRuns(first)
  let rejectPoll!: (error: Error) => void
  vi.mocked(api.pollStudioRuns).mockImplementationOnce(
    () =>
      new Promise((_resolve, reject) => {
        rejectPoll = reject
      })
  )
  await openStudio()
  await waitFor(() => expect(api.pollStudioRuns).toHaveBeenCalledTimes(1))
  vi.useFakeTimers()
  try {
    await act(async () => {
      rejectPoll(
        new AxiosError('rate limited', undefined, undefined, undefined, {
          status: 429,
          statusText: '',
          data: '',
          headers: new AxiosHeaders({ 'retry-after': '20' }),
          config: { headers: new AxiosHeaders() },
        })
      )
      await vi.advanceTimersByTimeAsync(1)
    })
    vi.mocked(api.pollStudioRuns).mockResolvedValue({
      items: [first, second],
      total: 2,
    })
    await act(async () => {
      client.setQueryData(['image-studio', 'runs', 7, 1], {
        items: [second, first],
        total: 2,
      })
      await vi.advanceTimersByTimeAsync(1)
      focusManager.setFocused(false)
      focusManager.setFocused(true)
      await vi.advanceTimersByTimeAsync(1)
    })
    expect(api.pollStudioRuns).toHaveBeenCalledTimes(1)
    await act(async () => {
      await vi.advanceTimersByTimeAsync(19000)
    })
    expect(api.pollStudioRuns).toHaveBeenCalledTimes(1)
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000)
    })
    expect(api.pollStudioRuns).toHaveBeenCalledTimes(2)
    expect(api.pollStudioRuns).toHaveBeenLastCalledWith(
      expect.arrayContaining([first.id, second.id])
    )
  } finally {
    focusManager.setFocused(undefined)
    vi.useRealTimers()
  }
})
