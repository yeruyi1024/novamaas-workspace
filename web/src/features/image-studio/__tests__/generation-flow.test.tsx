import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
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
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { getApiKeys } from '@/features/keys/api'
import type { ApiKey } from '@/features/keys/types'
import { ImageStudioPage } from '@/routes/image-studio'
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
const key = {
  id: 1,
  name: 'Image key',
  key: '***',
  status: 1,
  group: '',
} as ApiKey

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
  })
  vi.mocked(getApiKeys).mockResolvedValue({
    success: true,
    data: { items: [key], total: 1, page: 1, page_size: 100 },
  })
  vi.mocked(api.listStudioModels).mockImplementation(async (_id, kind) =>
    kind === 'text' ? ['text-model'] : ['dall-e-3']
  )
  vi.mocked(api.listStudioRuns).mockResolvedValue({ items: [], total: 0 })
})
afterEach(() => {
  client.clear()
  useAuthStore.getState().auth.reset()
  vi.resetAllMocks()
})

async function openStudio(guest = false) {
  const root = createRootRoute({ component: Outlet })
  const studio = createRoute({
    getParentRoute: () => root,
    path: '/image-studio',
    component: guest ? ImageStudioPage : ImageStudio,
  })
  const dashboard = createRoute({
    getParentRoute: () => root,
    path: '/dashboard',
    component: () => <main>Dashboard page</main>,
  })
  const signIn = createRoute({
    getParentRoute: () => root,
    path: '/sign-in',
    validateSearch: (search: Record<string, unknown>) => ({
      redirect: String(search.redirect || ''),
    }),
    component: () => <main>Sign in page</main>,
  })
  const router = createRouter({
    routeTree: root.addChildren([studio, dashboard, signIn]),
    history: createMemoryHistory({ initialEntries: ['/image-studio'] }),
  })
  await router.load()
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
  if (guest) await screen.findByRole('heading', { name: 'Image creation' })
  else await screen.findByText('dall-e-3', { exact: true })
  const user = userEvent.setup()
  return user
}

test('guests can browse the editor and examples without fetching keys, models or private records', async () => {
  useAuthStore.getState().auth.reset()
  const user = await openStudio(true)
  expect(screen.getByRole('main').closest('.h-svh')).toHaveClass(
    'w-full',
    'overflow-hidden'
  )
  const retention = screen.getByText(
    'Generated images are kept for 30 days, then automatically deleted. Please download them in time.'
  )
  expect(retention).toBeVisible()
  expect(retention.closest('header')).toHaveClass('sticky', 'top-0')
  await user.click(
    screen.getByRole('button', { name: 'Cinematic photography' })
  )
  expect(
    screen.getByRole('textbox', { name: 'Describe your image' })
  ).not.toHaveValue('')
  expect(screen.getByRole('button', { name: 'Optimize prompt' })).toBeDisabled()
  expect(getApiKeys).not.toHaveBeenCalled()
  expect(api.listStudioModels).not.toHaveBeenCalled()
  expect(api.listStudioRuns).not.toHaveBeenCalled()
  expect(api.createStudioRun).not.toHaveBeenCalled()
  const signIn = screen.getByRole('link', { name: 'Sign in to create images' })
  expect(signIn).toHaveAttribute('href', '/sign-in?redirect=%2Fimage-studio')
  await user.click(signIn)
  expect(await screen.findByText('Sign in page')).toBeVisible()
})

test('signing out clears the prompt, private generation records and model selection from the editor', async () => {
  vi.mocked(api.listStudioRuns).mockResolvedValue({
    items: [completed],
    total: 1,
  })
  const user = await openStudio()
  await user.type(
    screen.getByRole('textbox', { name: 'Describe your image' }),
    'Private draft'
  )
  expect(screen.getByRole('button', { name: 'Delete image' })).toBeVisible()
  await act(() => useAuthStore.getState().auth.reset())
  expect(
    await screen.findByRole('link', { name: 'Sign in to create images' })
  ).toBeVisible()
  expect(
    screen.queryByRole('button', { name: 'Delete image' })
  ).not.toBeInTheDocument()
  expect(
    screen.queryByText('dall-e-3', { exact: true })
  ).not.toBeInTheDocument()
  expect(
    screen.getByRole('textbox', { name: 'Describe your image' })
  ).toHaveValue('')
})

test('desktop windows give the landscape canvas more width without stretching it to screen height', async () => {
  await openStudio()
  const editor = screen
    .getByRole('textbox', { name: 'Describe your image' })
    .closest('form')
  expect(editor?.parentElement).toHaveClass(
    'lg:grid-cols-[minmax(18rem,0.9fr)_minmax(0,1.6fr)_minmax(16rem,0.85fr)]',
    'lg:min-h-0'
  )
  expect(
    screen.getByRole('region', { name: 'Generation results' })
  ).toHaveClass('aspect-[4/3]', 'w-full', 'max-w-[48rem]', 'flex-none')
  expect(
    screen.getByRole('region', { name: 'Generation records' })
  ).toHaveClass('lg:col-start-3')
})

test.each([
  ['keep', 'Delete generated image only'],
  ['delete', 'Also delete asset library image'],
] as const)(
  'deleting a saved image explains cleanup and applies the %s library choice',
  async (choice, label) => {
    const saved = {
      ...completed,
      results: [{ ...completed.results[0], asset_id: 'asset-1' }],
    }
    vi.mocked(api.listStudioRuns).mockResolvedValueOnce({
      items: [saved],
      total: 1,
    })
    vi.mocked(api.deleteStudioResult).mockResolvedValue(undefined)
    const user = await openStudio()
    await user.click(screen.getByRole('button', { name: 'Delete image' }))
    const dialog = screen.getByRole('alertdialog', { name: 'Delete image' })
    expect(
      within(dialog).getByText(
        'Deleted images cannot be restored. Files are removed in the background.'
      )
    ).toBeVisible()
    await user.click(within(dialog).getByRole('button', { name: label }))
    await waitFor(() =>
      expect(api.deleteStudioResult).toHaveBeenCalledWith('result-1', choice)
    )
    await waitFor(() =>
      expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument()
    )
    expect(api.deleteStudioRun).not.toHaveBeenCalled()
  }
)

test('deleting a failed record confirms irreversible removal and deletes the record instead of an image', async () => {
  vi.mocked(api.listStudioRuns).mockResolvedValueOnce({
    items: [{ ...completed, status: 'failed', results: [] }],
    total: 1,
  })
  vi.mocked(api.deleteStudioRun).mockResolvedValue(undefined)
  const user = await openStudio()
  await user.click(screen.getByRole('button', { name: 'Delete record' }))
  expect(
    screen.getByText('Delete this generation record? This cannot be undone.')
  ).toBeVisible()
  await user.click(
    within(screen.getByRole('alertdialog')).getByRole('button', {
      name: 'Delete',
    })
  )
  await waitFor(() =>
    expect(api.deleteStudioRun).toHaveBeenCalledWith('run-1', undefined)
  )
  expect(api.deleteStudioResult).not.toHaveBeenCalled()
})

async function generate() {
  const user = await openStudio()
  await user.type(
    screen.getByRole('textbox', { name: 'Describe your image' }),
    'A mountain'
  )
  await user.click(screen.getByRole('button', { name: 'Create images' }))
  return user
}

test('completed images fit a landscape canvas with the prompt only below it and still open the full preview', async () => {
  vi.mocked(api.createStudioRun).mockResolvedValue({
    run: completed,
    replay: false,
  })
  const user = await generate()
  const preview = await screen.findByRole('button', { name: 'Preview' })
  expect(preview).toHaveClass('aspect-[4/3]', 'max-w-[48rem]', 'flex-none')
  expect(within(preview).getByRole('img', { name: 'A mountain' })).toHaveClass(
    'object-contain'
  )
  const gallery = within(
    screen.getByRole('region', { name: 'Generation results' })
  )
  expect(gallery.getAllByText('A mountain', { exact: true })).toHaveLength(1)
  expect(
    preview.compareDocumentPosition(
      gallery.getByText('A mountain', { exact: true })
    ) & Node.DOCUMENT_POSITION_FOLLOWING
  ).toBeTruthy()
  await user.click(preview)
  expect(
    within(screen.getByRole('dialog', { name: 'Generated image' })).getByRole(
      'img',
      { name: 'A mountain' }
    )
  ).toHaveAttribute('src', completed.results[0].preview_url)
})

test('asset saving explains automatic history storage and is optional for using images in video generation', async () => {
  vi.mocked(api.createStudioRun).mockResolvedValue({
    run: completed,
    replay: false,
  })
  const user = await generate()
  const explanation =
    'Images are saved in the generation records on the right for 30 days. Download them before expiry. Save to the asset library only for direct use in video generation; library copies are kept separately.'
  const save = await screen.findByRole('button', {
    name: 'Save to asset library (for video generation)',
  })
  expect(save).toHaveAccessibleDescription(explanation)
  expect(screen.getByText(explanation)).toBeVisible()
  expect(api.saveStudioResult).not.toHaveBeenCalled()
  await user.click(save)
  const dialog = screen.getByRole('dialog', {
    name: 'Save to asset library (for video generation)',
  })
  expect(dialog).toHaveAccessibleDescription(explanation)
  await user.click(within(dialog).getByRole('button', { name: 'Cancel' }))
  expect(api.saveStudioResult).not.toHaveBeenCalled()
})

test('prompt optimization identifies and calls the chosen text model separately from image generation', async () => {
  vi.mocked(api.listStudioModels).mockImplementation(async (_id, kind) =>
    kind === 'text' ? ['text-model', 'chosen-text-model'] : ['dall-e-3']
  )
  vi.mocked(api.assistStudioPrompt).mockResolvedValue({
    suggested_prompt: 'A mountain in soft light',
  })
  const user = await openStudio()
  await user.click(screen.getByRole('button', { name: 'Model settings' }))
  const textModels = within(screen.getByRole('group', { name: 'Text model' }))
  await user.selectOptions(
    textModels.getByRole('combobox', { name: 'Model' }),
    'chosen-text-model'
  )
  await user.click(screen.getByRole('button', { name: 'Done' }))
  expect(
    await screen.findByText('Text model: chosen-text-model')
  ).toBeInTheDocument()
  const prompt = screen.getByRole('textbox', { name: 'Describe your image' })
  await user.type(prompt, 'A mountain')
  await user.click(screen.getByRole('button', { name: 'Optimize prompt' }))
  await waitFor(() => expect(prompt).toHaveValue('A mountain in soft light'))
  expect(api.assistStudioPrompt).toHaveBeenCalledWith({
    token_id: 1,
    model: 'chosen-text-model',
    prompt: 'A mountain',
  })
  expect(screen.getByText('dall-e-3', { exact: true })).toBeInTheDocument()
  expect(api.createStudioRun).not.toHaveBeenCalled()
})

test('image proportions show model defaults without selectable sizes', async () => {
  await openStudio()
  expect(screen.getByText('Model default')).toBeVisible()
  expect(
    screen.queryByRole('combobox', { name: 'Image count' })
  ).not.toBeInTheDocument()
  expect(
    screen.queryByText('This model supports one image per request.')
  ).not.toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: 'Model default' })
  ).not.toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: 'Landscape' })
  ).not.toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: 'Portrait' })
  ).not.toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: 'Square image' })
  ).not.toBeInTheDocument()
})

test('opening image creation exposes the prompt editor without a start introduction', async () => {
  const user = await openStudio()
  expect(screen.getByRole('heading', { name: 'Image creation' })).toBeVisible()
  expect(
    screen.getByRole('textbox', { name: 'Describe your image' })
  ).toBeVisible()
  expect(screen.queryByText('Start with an idea')).not.toBeInTheDocument()
  expect(
    screen.queryByText('Describe an idea. Create something new.')
  ).not.toBeInTheDocument()
  await user.click(screen.getByRole('link', { name: 'Back to Dashboard' }))
  expect(await screen.findByText('Dashboard page')).toBeVisible()
})

test('a failed submission retries with the same idempotency key and nothing is saved by default', async () => {
  vi.mocked(api.createStudioRun)
    .mockRejectedValueOnce(new Error('network'))
    .mockResolvedValueOnce({ run: completed, replay: false })
  const user = await generate()
  await screen.findByText('network')
  await user.click(screen.getByRole('button', { name: 'Create images' }))
  await screen.findByRole('img', { name: 'A mountain' })
  const calls = vi.mocked(api.createStudioRun).mock.calls
  expect(calls).toHaveLength(2)
  expect(calls[0][0]).toEqual({
    token_id: 1,
    idempotency_key: expect.any(String),
    request: { model: 'dall-e-3', prompt: 'A mountain' },
  })
  expect(calls[1][0].idempotency_key).toBe(calls[0][0].idempotency_key)
  expect(api.saveStudioResult).not.toHaveBeenCalled()
  expect(
    screen.getByRole('button', {
      name: 'Save to asset library (for video generation)',
    })
  ).toBeEnabled()
})

test('a replayed run that is still generating polls until it completes', async () => {
  const running: StudioRun = { ...completed, status: 'running', results: [] }
  vi.mocked(api.createStudioRun).mockResolvedValue({
    run: running,
    replay: true,
  })
  vi.mocked(api.getStudioRun).mockResolvedValue(running)
  await generate()
  await screen.findByText('This generation is still in progress')
  vi.mocked(api.getStudioRun).mockResolvedValue(completed)
  await screen.findByRole('img', { name: 'A mountain' }, { timeout: 3000 })
  await waitFor(() =>
    expect(
      screen.queryByText('This generation is still in progress')
    ).not.toBeInTheDocument()
  )
})
