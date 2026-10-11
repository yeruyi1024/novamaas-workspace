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
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { listAssetGroupsPage, listMediaAssets } from '@/features/assets/api'
import { getApiKeys } from '@/features/keys/api'
import { useAuthStore } from '@/stores/auth-store'

import * as api from '../api'
import { ImageStudio } from '../image-studio'
import type { StudioRun, StudioSettings } from '../types'

vi.mock('../api')
vi.mock('@/features/assets/api')
vi.mock('@/features/keys/api')

const run: StudioRun = {
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
      asset_id: '',
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
  remain_quota: 100,
  used_quota: 0,
  unlimited_quota: true,
  expired_time: -1,
  created_time: 1,
  accessed_time: 1,
  group: '',
  auto_groups: null,
  cross_group_retry: false,
  model_limits_enabled: false,
  model_limits: '',
  allow_ips: '',
}
const settings: StudioSettings = {
  enabled: true,
  asset_library_enabled: true,
  max_images: 4,
  max_reference_images: 2,
  max_prompt_characters: 1000,
}
const success = <T,>(data: T) => ({ success: true, data })
function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (error: Error) => void
  const promise = new Promise<T>((yes, no) => {
    resolve = yes
    reject = no
  })
  return { promise, resolve, reject }
}

let client: QueryClient
beforeEach(() => {
  client = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
  useAuthStore
    .getState()
    .auth.setUser({ id: 7, username: 'artist', role: 1, group: 'creative' })
  vi.mocked(api.getStudioSettings).mockResolvedValue(settings)
  vi.mocked(getApiKeys).mockResolvedValue(
    success({ items: [key], total: 1, page: 1, page_size: 100 })
  )
  vi.mocked(api.listStudioModels).mockImplementation(async (_id, kind) =>
    kind === 'text' ? ['text-model'] : ['dall-e-3']
  )
  vi.mocked(api.listStudioRuns).mockResolvedValue({ items: [], total: 0 })
  vi.mocked(api.createStudioRun).mockResolvedValue(run)
  vi.mocked(api.getStudioRun).mockResolvedValue(run)
  vi.mocked(listMediaAssets).mockResolvedValue(
    success({ items: [], total: 0, page: 1, page_size: 30 })
  )
  vi.mocked(listAssetGroupsPage).mockResolvedValue(
    success({
      items: [
        {
          id: 'group-1',
          name: 'Campaign',
          owner_user_id: 7,
          description: '',
          status: 'ready',
          created_at: 1,
          updated_at: 1,
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
  )
})
afterEach(() => {
  client.clear()
  useAuthStore.getState().auth.reset()
  vi.unstubAllGlobals()
  vi.resetAllMocks()
})

async function studio(model = 'dall-e-3') {
  render(
    <QueryClientProvider client={client}>
      <ImageStudio />
    </QueryClientProvider>
  )
  await screen.findByText(model, { exact: true })
  return userEvent.setup()
}
const prompt = () =>
  screen.getByRole('textbox', { name: 'Describe your image' })
const button = (name: string) => screen.getByRole('button', { name })
async function generate(user: ReturnType<typeof userEvent.setup>) {
  await user.type(prompt(), 'A mountain')
  await user.click(button('Create images'))
}

test('default generation on LAN HTTP omits optional fields and leaves assets untouched', async () => {
  vi.stubGlobal('crypto', {
    getRandomValues: crypto.getRandomValues.bind(crypto),
  })
  await generate(await studio())
  await screen.findByRole('img', { name: 'A mountain' })
  expect(api.createStudioRun).toHaveBeenCalledWith({
    token_id: 1,
    idempotency_key: expect.any(String),
    request: { model: 'dall-e-3', prompt: 'A mountain' },
    input_asset_ids: [],
  })
  expect(api.saveStudioResult).not.toHaveBeenCalled()
  expect(button('Save to asset library')).toBeEnabled()
})

test.each([
  ['dall-e-2', 'Square image', '1024x1024', '2'],
  ['dall-e-3', 'Portrait', '1024x1792', '1'],
  ['gpt-image-1', 'Landscape', '1536x1024', '3'],
  ['doubao-seedream-4-5-251128', 'Landscape', '2496x1664', '2'],
])(
  '%s sends chosen proportions and count, then resets them on model change',
  async (model, ratio, size, count) => {
    vi.mocked(api.listStudioModels).mockImplementation(async (_id, kind) =>
      kind === 'text' ? ['text-model'] : [model, 'other-model']
    )
    const user = await studio(model)
    await user.click(button(ratio))
    if (count !== '1') {
      await user.selectOptions(
        screen.getByRole('combobox', { name: 'Image count' }),
        count
      )
    }
    await generate(user)
    await waitFor(() => expect(api.createStudioRun).toHaveBeenCalled())
    expect(vi.mocked(api.createStudioRun).mock.calls[0][0].request).toEqual({
      model,
      prompt: 'A mountain',
      size,
      ...(count !== '1' ? { n: Number(count) } : {}),
    })
    await user.click(button('Model settings'))
    const group = within(screen.getByRole('group', { name: 'Image model' }))
    await user.selectOptions(
      group.getByRole('combobox', { name: 'Model' }),
      'other-model'
    )
    await user.click(button('Done'))
    expect(button('Model default')).toHaveAttribute('aria-pressed', 'true')
    expect(
      screen.queryByRole('combobox', { name: 'Image quality' })
    ).not.toBeInTheDocument()
  }
)

test('independent key selection applies inherited groups, bound models and paginated keys', async () => {
  vi.mocked(getApiKeys)
    .mockResolvedValueOnce(
      success({ items: [key], total: 2, page: 1, page_size: 100 })
    )
    .mockResolvedValueOnce(
      success({
        items: [{ ...key, id: 2, name: 'Campaign', group: 'campaign' }],
        total: 2,
        page: 2,
        page_size: 100,
      })
    )
  vi.mocked(api.listStudioModels).mockImplementation(async (id, kind) =>
    kind === 'text' ? ['text-model'] : [id === 2 ? 'seedream' : 'dall-e-3']
  )
  const user = await studio()
  await user.click(button('Model settings'))
  const image = within(screen.getByRole('group', { name: 'Image model' }))
  expect(
    image.getByRole('option', { name: 'Image key · creative' })
  ).toBeInTheDocument()
  await user.selectOptions(
    image.getByRole('combobox', { name: 'API key' }),
    '2'
  )
  await waitFor(() =>
    expect(image.getByRole('combobox', { name: 'Model' })).toHaveValue(
      'seedream'
    )
  )
  expect(
    within(screen.getByRole('group', { name: 'Text model' })).getByRole(
      'combobox',
      { name: 'API key' }
    )
  ).toHaveValue('1')
  await user.click(button('Done'))
  await generate(user)
  await waitFor(() => expect(api.createStudioRun).toHaveBeenCalled())
  expect(vi.mocked(api.createStudioRun).mock.calls[0][0].token_id).toBe(2)
  expect(getApiKeys).toHaveBeenCalledWith({ p: 2, size: 100 })
})

test('pending generation blocks repeat submission; failure preserves the prompt for retry', async () => {
  const pending = deferred<StudioRun>()
  vi.mocked(api.createStudioRun).mockReturnValueOnce(pending.promise)
  const user = await studio()
  await generate(user)
  expect(button('Generating...')).toBeDisabled()
  pending.reject(new Error('Provider unavailable'))
  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Provider unavailable'
  )
  expect(prompt()).toHaveValue('A mountain')
  await user.click(button('Create images'))
  await screen.findByRole('img', { name: 'A mountain' })
  expect(api.createStudioRun).toHaveBeenCalledTimes(2)
})

test.each([
  ['enabled', 'Create images'],
  ['asset_library_enabled', 'Save to asset library'],
] as const)(
  'disabling %s retains readable history and blocks that capability',
  async (field, action) => {
    vi.mocked(api.getStudioSettings).mockResolvedValue({
      ...settings,
      [field]: false,
    })
    vi.mocked(api.listStudioRuns).mockResolvedValue({ items: [run], total: 1 })
    const user = await studio()
    await user.click(
      await within(
        screen.getByRole('region', { name: 'Generation records' })
      ).findByRole('img')
    )
    await screen.findByRole('img', { name: 'A mountain' })
    if (field === 'enabled') expect(button(action)).toBeDisabled()
    else {
      expect(
        screen.queryByRole('button', { name: action })
      ).not.toBeInTheDocument()
      expect(listMediaAssets).not.toHaveBeenCalled()
      expect(listAssetGroupsPage).not.toHaveBeenCalled()
    }
    expect(api.createStudioRun).not.toHaveBeenCalled()
  }
)

test('history uses thumbnails, selects originals and reuses a prompt without generating', async () => {
  vi.mocked(api.listStudioRuns).mockResolvedValue({ items: [run], total: 1 })
  const user = await studio()
  const history = within(
    screen.getByRole('region', { name: 'Generation records' })
  )
  const thumbnail = await history.findByRole('img')
  expect(thumbnail).toHaveAttribute('src', run.results[0].thumbnail_url)
  expect(api.getStudioRun).not.toHaveBeenCalled()
  await user.click(thumbnail)
  const gallery = within(
    screen.getByRole('region', { name: 'Generation results' })
  )
  expect(await gallery.findByRole('img')).toHaveAttribute(
    'src',
    run.results[0].preview_url
  )
  await user.click(button('Use this prompt'))
  expect(prompt()).toHaveValue(run.prompt)
  expect(prompt()).toHaveFocus()
  expect(api.createStudioRun).not.toHaveBeenCalled()
})

test('explicit save retries the same result after storage failure', async () => {
  vi.mocked(api.saveStudioResult).mockRejectedValueOnce(
    new Error('Storage unavailable')
  )
  const user = await studio()
  await generate(user)
  await user.click(
    await screen.findByRole('button', { name: 'Save to asset library' })
  )
  const dialog = within(screen.getByRole('dialog'))
  expect(dialog.getByRole('button', { name: 'Save' })).toBeDisabled()
  await user.click(dialog.getByRole('combobox', { name: 'Asset group' }))
  await user.click(await screen.findByRole('option', { name: /Campaign/ }))
  await user.click(dialog.getByRole('button', { name: 'Save' }))
  expect(await dialog.findByRole('alert')).toHaveTextContent(
    'Storage unavailable'
  )
  const pending = deferred<undefined>()
  vi.mocked(api.saveStudioResult).mockReturnValueOnce(pending.promise)
  await user.click(dialog.getByRole('button', { name: 'Save' }))
  expect(dialog.getByRole('combobox', { name: 'Asset group' })).toBeDisabled()
  pending.resolve(undefined)
  await waitFor(() =>
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  )
  expect(api.saveStudioResult).toHaveBeenNthCalledWith(2, 'result-1', 'group-1')
  expect(api.createStudioRun).toHaveBeenCalledOnce()
})

test('deletion supports cancel and retry, then clears the last selected image', async () => {
  let remaining = [run]
  vi.mocked(api.listStudioRuns).mockImplementation(async () => ({
    items: remaining,
    total: remaining.length,
  }))
  vi.mocked(api.deleteStudioResult)
    .mockRejectedValueOnce(new Error('Delete unavailable'))
    .mockImplementation(async () => {
      remaining = []
    })
  const user = await studio()
  const history = within(
    screen.getByRole('region', { name: 'Generation records' })
  )
  await user.click(await history.findByRole('img'))
  await user.click(history.getByRole('button', { name: 'Delete image' }))
  await user.click(button('Cancel'))
  expect(api.deleteStudioResult).not.toHaveBeenCalled()
  await user.click(history.getByRole('button', { name: 'Delete image' }))
  await user.click(button('Delete'))
  expect(
    await within(screen.getByRole('alertdialog')).findByRole('alert')
  ).toHaveTextContent('Delete unavailable')
  await user.click(button('Delete'))
  await history.findByText('No generation records')
  expect(screen.getByText('Your image will appear here')).toBeInTheDocument()
  expect(api.saveStudioResult).not.toHaveBeenCalled()
})

test('optimization blocks conflicting actions, applies text output and supports undo', async () => {
  const pending = deferred<{ suggested_prompt: string }>()
  vi.mocked(api.assistStudioPrompt).mockReturnValueOnce(pending.promise)
  const user = await studio()
  await user.type(prompt(), 'A mountain')
  await user.click(button('Optimize prompt'))
  expect(prompt()).toBeDisabled()
  expect(button('Create images')).toBeDisabled()
  expect(button('Model settings')).toBeDisabled()
  pending.resolve({ suggested_prompt: 'A sunlit mountain' })
  await waitFor(() => expect(prompt()).toHaveValue('A sunlit mountain'))
  expect(api.assistStudioPrompt).toHaveBeenCalledWith({
    token_id: 1,
    model: 'text-model',
    prompt: 'A mountain',
  })
  await user.click(button('Undo'))
  expect(prompt()).toHaveValue('A mountain')
  expect(api.createStudioRun).not.toHaveBeenCalled()
})

test('examples replace and styles append without automatic submission', async () => {
  const user = await studio()
  await user.click(
    within(screen.getByRole('region', { name: 'Try an example' })).getByRole(
      'button',
      { name: 'Cinematic photography' }
    )
  )
  expect((prompt() as HTMLTextAreaElement).value).toContain('A ginger cat')
  await user.click(
    within(screen.getByRole('group', { name: 'Add a style' })).getByRole(
      'button',
      { name: 'Ink painting' }
    )
  )
  expect((prompt() as HTMLTextAreaElement).value).toContain('ink painting')
  expect(api.createStudioRun).not.toHaveBeenCalled()
})
