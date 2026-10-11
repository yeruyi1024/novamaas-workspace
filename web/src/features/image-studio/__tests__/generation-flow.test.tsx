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

import { listAssetGroups, listMediaAssets } from '@/features/assets/api'
import { getApiKeys } from '@/features/keys/api'
import { useAuthStore } from '@/stores/auth-store'

import * as api from '../api'
import { ImageStudio } from '../image-studio'
import type { StudioResponse, StudioRun } from '../types'

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
  quota: 20000,
  results: [
    {
      id: 'result-1',
      asset_id: '',
      index: 0,
      preview_url: 'https://example.test/image.png',
      save_status: 'not_requested',
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
  group: 'default',
  auto_groups: null,
  cross_group_retry: false,
  model_limits_enabled: false,
  model_limits: '',
  allow_ips: '',
}
function success<T>(data: T): StudioResponse<T> {
  return { success: true, data }
}

let client: QueryClient
beforeEach(() => {
  client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  vi.mocked(api.getStudioSettings).mockResolvedValue(
    success({
      enabled: true,
      asset_library_enabled: true,
      max_images: 4,
      max_reference_images: 2,
      max_prompt_characters: 1000,
    })
  )
  vi.mocked(getApiKeys).mockResolvedValue(
    success({ items: [key], total: 1, page: 1, page_size: 100 })
  )
  vi.mocked(api.listStudioModels).mockImplementation(async (_tokenId, kind) =>
    success(kind === 'text' ? ['text-model'] : ['dall-e-3'])
  )
  vi.mocked(api.listStudioRuns).mockResolvedValue(
    success({ items: [], total: 0 })
  )
  vi.mocked(api.createStudioRun).mockResolvedValue(success(run))
  vi.mocked(api.getStudioRun).mockResolvedValue(success(run))
  vi.mocked(listMediaAssets).mockResolvedValue(
    success({ items: [], total: 0, page: 1, page_size: 30 })
  )
  vi.mocked(listAssetGroups).mockResolvedValue(
    success([
      {
        id: 'group-1',
        name: 'Campaign',
        owner_user_id: 7,
        description: '',
        status: 'ready',
        created_at: 1,
        updated_at: 1,
      },
    ])
  )
})
afterEach(() => {
  client.clear()
  useAuthStore.getState().auth.setUser(null)
  vi.unstubAllGlobals()
  vi.resetAllMocks()
})

async function renderStudio(model = 'dall-e-3') {
  render(
    <QueryClientProvider client={client}>
      <ImageStudio />
    </QueryClientProvider>
  )
  await screen.findByRole('textbox', { name: 'Describe your image' })
  await screen.findByText(model, { exact: true })
  return userEvent.setup()
}

async function submitImage(user: ReturnType<typeof userEvent.setup>) {
  await user.type(
    screen.getByRole('textbox', { name: 'Describe your image' }),
    'A mountain'
  )
  await user.click(screen.getByRole('button', { name: 'Create images' }))
}

async function generateImage() {
  const user = await renderStudio()
  await submitImage(user)
  await screen.findByRole('img', { name: 'A mountain' })
  return user
}

test('LAN HTTP generates with an existing key and leaves the asset library untouched by default', async () => {
  vi.stubGlobal('crypto', {
    getRandomValues: crypto.getRandomValues.bind(crypto),
  })
  await generateImage()
  expect(api.listStudioModels).toHaveBeenCalledWith(1)
  expect(api.createStudioRun).toHaveBeenCalledWith(
    expect.objectContaining({
      token_id: 1,
      idempotency_key: expect.any(String),
      request: { model: 'dall-e-3', prompt: 'A mountain' },
    })
  )
  expect(vi.mocked(api.createStudioRun).mock.calls[0][0]).not.toHaveProperty(
    'output_group_id'
  )
  expect(api.saveStudioResult).not.toHaveBeenCalled()
  expect(
    screen.getByRole('button', { name: 'Save to asset library' })
  ).toBeEnabled()
})

test('key selection shows its inherited or bound group and loads that key’s models', async () => {
  useAuthStore
    .getState()
    .auth.setUser({ id: 7, username: 'artist', role: 1, group: 'creative' })
  vi.mocked(getApiKeys)
    .mockResolvedValueOnce(
      success({
        items: Array.from({ length: 100 }, (_, index) => ({
          ...key,
          id: index + 1,
          status: index === 0 ? 1 : 2,
          group: '',
        })),
        total: 101,
        page: 1,
        page_size: 100,
      })
    )
    .mockResolvedValueOnce(
      success({
        items: [{ ...key, id: 101, name: 'Campaign key', group: 'campaign' }],
        total: 101,
        page: 2,
        page_size: 100,
      })
    )
  vi.mocked(api.listStudioModels).mockImplementation(async (tokenId, kind) => {
    if (kind === 'text') return success(['text-model'])
    return success(tokenId === 101 ? ['seedream'] : ['dall-e-3'])
  })
  const user = await renderStudio()
  await user.click(screen.getByRole('button', { name: 'Model settings' }))
  const dialog = await screen.findByRole('dialog')
  const imageModels = within(dialog).getByRole('group', { name: 'Image model' })
  const textModels = within(dialog).getByRole('group', { name: 'Text model' })
  expect(
    within(imageModels).getByRole('option', { name: 'Image key · creative' })
  ).toBeInTheDocument()
  await user.selectOptions(
    within(imageModels).getByRole('combobox', { name: 'API key' }),
    '101'
  )
  await waitFor(() =>
    expect(
      within(imageModels).getByRole('combobox', { name: 'Model' })
    ).toHaveValue('seedream')
  )
  expect(
    within(textModels).getByRole('combobox', { name: 'API key' })
  ).toHaveValue('1')
  expect(api.listStudioModels).toHaveBeenCalledWith(101)
  expect(getApiKeys).toHaveBeenCalledWith({ p: 2, size: 100 })
  expect(
    within(imageModels).getByRole('option', { name: 'Campaign key · campaign' })
  ).toBeInTheDocument()
  await user.click(within(dialog).getByRole('button', { name: 'Done' }))
  await submitImage(user)
  await waitFor(() =>
    expect(api.createStudioRun).toHaveBeenCalledWith(
      expect.objectContaining({
        token_id: 101,
        request: expect.objectContaining({ model: 'seedream' }),
      })
    )
  )
})

test('a pending HTTP request disables repeat submission until its result is displayed', async () => {
  let complete!: (value: StudioResponse<StudioRun>) => void
  vi.mocked(api.createStudioRun).mockImplementation(
    () =>
      new Promise((resolve) => {
        complete = resolve
      })
  )
  const user = await renderStudio()
  await submitImage(user)
  expect(screen.getByRole('button', { name: 'Generating...' })).toBeDisabled()
  complete(success(run))
  await screen.findByRole('img', { name: 'A mountain' })
  expect(api.createStudioRun).toHaveBeenCalledOnce()
})

test('disabled generation leaves history readable and blocks new requests', async () => {
  vi.mocked(api.getStudioSettings).mockResolvedValue(
    success({
      enabled: false,
      max_images: 4,
      max_reference_images: 2,
      max_prompt_characters: 1000,
    })
  )
  vi.mocked(api.listStudioRuns).mockResolvedValue(
    success({ items: [run], total: 1 })
  )
  const user = await renderStudio()
  await user.click(
    within(
      screen.getByRole('region', { name: 'Generation records' })
    ).getByRole('img')
  )
  await screen.findByRole('img', { name: 'A mountain' })
  expect(screen.getByRole('button', { name: 'Create images' })).toBeDisabled()
  expect(api.createStudioRun).not.toHaveBeenCalled()
})

test('disabled asset integration omits private asset queries and save controls', async () => {
  vi.mocked(api.getStudioSettings).mockResolvedValue(
    success({
      enabled: true,
      asset_library_enabled: false,
      max_images: 4,
      max_reference_images: 2,
      max_prompt_characters: 1000,
    })
  )
  await generateImage()
  expect(listMediaAssets).not.toHaveBeenCalled()
  expect(listAssetGroups).not.toHaveBeenCalled()
  expect(api.createStudioRun).toHaveBeenCalledWith(
    expect.objectContaining({ input_asset_ids: [] })
  )
  expect(
    screen.queryByRole('button', { name: 'Save to asset library' })
  ).not.toBeInTheDocument()
})

test('a failed save can be retried against the same result without regenerating', async () => {
  vi.mocked(api.saveStudioResult)
    .mockRejectedValueOnce(new Error('Storage unavailable'))
    .mockResolvedValueOnce(success(undefined))
  vi.mocked(api.getStudioRun)
    .mockResolvedValueOnce(
      success({
        ...run,
        results: [
          {
            ...run.results[0],
            save_status: 'failed',
            save_error: 'Storage unavailable',
          },
        ],
      })
    )
    .mockResolvedValueOnce(
      success({
        ...run,
        results: [
          { ...run.results[0], asset_id: 'asset-1', save_status: 'saved' },
        ],
      })
    )
  const user = await generateImage()
  await user.click(
    screen.getByRole('button', { name: 'Save to asset library' })
  )
  const dialog = await screen.findByRole('dialog')
  expect(
    within(dialog).getByText(
      'Only add images to an asset group when you need them in the asset library for video generation or other workflows.'
    )
  ).toBeInTheDocument()
  expect(within(dialog).getByRole('button', { name: 'Save' })).toBeDisabled()
  await user.selectOptions(
    within(dialog).getByRole('combobox', { name: 'Asset group' }),
    'group-1'
  )
  await user.click(within(dialog).getByRole('button', { name: 'Save' }))
  await within(dialog).findByRole('alert')
  await user.click(within(dialog).getByRole('button', { name: 'Save' }))
  const gallery = screen.getByRole('region', { name: 'Generation results' })
  await within(gallery).findByText('Saved to asset library')
  expect(api.saveStudioResult).toHaveBeenCalledTimes(2)
  expect(api.createStudioRun).toHaveBeenCalledOnce()
})

test('history loads thumbnails and only fetches originals when an image is selected', async () => {
  useAuthStore.getState().auth.setUser({ id: 7, username: 'artist', role: 1 })
  const images = {
    ...run,
    results: [
      {
        ...run.results[0],
        thumbnail_url: 'https://example.test/thumbnail.png',
        description: 'A mountain',
      },
      {
        ...run.results[0],
        id: 'result-2',
        index: 1,
        preview_url: 'https://example.test/forest.png',
        thumbnail_url: 'https://example.test/forest-thumbnail.png',
        description: 'A forest',
      },
    ],
  }
  vi.mocked(api.listStudioRuns).mockResolvedValue(
    success({
      items: [
        {
          ...images,
          results: images.results.map((image) => ({
            ...image,
            preview_url: undefined,
          })),
        },
      ],
      total: 1,
    })
  )
  vi.mocked(api.getStudioRun).mockResolvedValue(success(images))
  const user = await renderStudio()
  const history = screen.getByRole('region', { name: 'Generation records' })
  expect(history).toHaveClass('xl:col-start-3')
  expect(await within(history).findAllByRole('img')).toHaveLength(2)
  expect(within(history).getAllByRole('img')[0]).toHaveAttribute(
    'src',
    'https://example.test/thumbnail.png'
  )
  expect(screen.getByText('Your image will appear here')).toBeInTheDocument()
  expect(api.getStudioRun).not.toHaveBeenCalled()
  expect(
    screen.getByRole('textbox', { name: 'Describe your image' })
  ).toHaveAttribute('rows', '4')
  await user.click(within(history).getAllByRole('img')[0])
  const gallery = screen.getByRole('region', { name: 'Generation results' })
  const mainImage = await within(gallery).findByRole('img', {
    name: 'A mountain',
  })
  expect(mainImage).toHaveAttribute('src', 'https://example.test/image.png')
  expect(api.getStudioRun).toHaveBeenCalledWith(run.id)
  expect(mainImage).toHaveClass('object-contain')
  expect(mainImage.closest('button')).toHaveClass('min-h-48')
  expect(gallery.parentElement).toHaveClass('overflow-y-auto')
  const collection = within(history)
    .getAllByRole('img')[0]
    .closest('article')?.parentElement
  expect(collection).toHaveClass(
    'max-h-[28rem]',
    'overflow-y-auto',
    'xl:grid-cols-1'
  )
  expect(within(gallery).getAllByRole('img')).toHaveLength(1)
  expect(within(gallery).getByText(/Task ID.*run-1/)).toBeInTheDocument()
  expect(within(gallery).getByText(/artist.*User ID.*7/)).toBeInTheDocument()
  const second = within(history).getByRole('button', { name: /A forest/ })
  await user.click(second)
  expect(
    await within(gallery).findByRole('img', { name: 'A forest' })
  ).toHaveAttribute('src', 'https://example.test/forest.png')
  expect(second).toHaveAttribute('aria-pressed', 'true')
})

test('failed generation only displays the reason and leaves history empty for retry', async () => {
  vi.mocked(api.createStudioRun).mockRejectedValueOnce(
    new Error('Upstream unavailable')
  )
  const user = await renderStudio()
  await submitImage(user)
  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Upstream unavailable'
  )
  expect(
    within(
      screen.getByRole('region', { name: 'Generation records' })
    ).queryByRole('img')
  ).not.toBeInTheDocument()
  expect(screen.getByText('Your image will appear here')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Create images' })).toBeEnabled()
  await user.click(screen.getByRole('button', { name: 'Create images' }))
  await screen.findByRole('img', { name: 'A mountain' })
  expect(api.createStudioRun).toHaveBeenCalledTimes(2)
})

test('history deletes individual images with cancel and retry, and clears the last image without touching assets', async () => {
  let results = [
    { ...run.results[0], asset_id: 'asset-1', description: 'A mountain' },
    {
      ...run.results[0],
      id: 'result-2',
      index: 1,
      description: 'A forest',
      preview_url: 'https://example.test/forest.png',
    },
  ]
  vi.mocked(api.getStudioRun).mockResolvedValue(success({ ...run, results }))
  vi.mocked(api.listStudioRuns).mockImplementation(async () =>
    success({
      items: results.length ? [{ ...run, results }] : [],
      total: results.length ? 1 : 0,
    })
  )
  vi.mocked(api.deleteStudioResult)
    .mockRejectedValueOnce(new Error('Delete unavailable'))
    .mockImplementation(async (id) => {
      results = results.filter((result) => result.id !== id)
      return success(undefined)
    })
  const user = await renderStudio()
  const history = screen.getByRole('region', { name: 'Generation records' })
  await within(history).findAllByRole('img')
  await user.click(within(history).getAllByRole('img')[0])
  await user.click(
    within(history).getAllByRole('button', { name: 'Delete image' })[0]
  )
  let dialog = await screen.findByRole('dialog', { name: 'Delete image' })
  expect(
    within(dialog).getByText(
      'Delete this generated image? Images already saved to the asset library are kept.'
    )
  ).toBeInTheDocument()
  await user.click(within(dialog).getByRole('button', { name: 'Cancel' }))
  expect(api.deleteStudioResult).not.toHaveBeenCalled()
  await user.click(
    within(history).getAllByRole('button', { name: 'Delete image' })[0]
  )
  dialog = await screen.findByRole('dialog', { name: 'Delete image' })
  await user.click(within(dialog).getByRole('button', { name: 'Delete' }))
  expect(await within(dialog).findByRole('alert')).toHaveTextContent(
    'Delete unavailable'
  )
  expect(within(history).getAllByRole('img', { hidden: true })).toHaveLength(2)
  await user.click(within(dialog).getByRole('button', { name: 'Delete' }))
  await waitFor(() =>
    expect(within(history).getAllByRole('img')).toHaveLength(1)
  )
  const gallery = screen.getByRole('region', { name: 'Generation results' })
  await within(gallery).findByRole('img', { name: 'A forest' })
  await user.click(
    within(history).getByRole('button', { name: 'Delete image' })
  )
  dialog = await screen.findByRole('dialog', { name: 'Delete image' })
  await user.click(within(dialog).getByRole('button', { name: 'Delete' }))
  await within(history).findByText('No generation records')
  expect(screen.getByText('Your image will appear here')).toBeInTheDocument()
  expect(api.deleteStudioResult).toHaveBeenNthCalledWith(2, 'result-1')
  expect(api.deleteStudioResult).toHaveBeenNthCalledWith(3, 'result-2')
  expect(api.saveStudioResult).not.toHaveBeenCalled()
  expect(api.createStudioRun).not.toHaveBeenCalled()
})

test('direct prompt optimization applies the result without a dialog, prevents conflicting actions and supports undo', async () => {
  let complete!: (value: StudioResponse<{ suggested_prompt: string }>) => void
  vi.mocked(api.assistStudioPrompt).mockImplementation(
    () =>
      new Promise((resolve) => {
        complete = resolve
      })
  )
  const user = await renderStudio()
  const prompt = screen.getByRole('textbox', { name: 'Describe your image' })
  expect(screen.getByRole('button', { name: 'Optimize prompt' })).toBeDisabled()
  await user.type(prompt, 'A mountain')
  await user.click(screen.getByRole('button', { name: 'Optimize prompt' }))
  expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  expect(prompt).toBeDisabled()
  expect(screen.getByRole('button', { name: 'Create images' })).toBeDisabled()
  expect(screen.getByRole('button', { name: 'Model settings' })).toBeDisabled()
  expect(prompt).toHaveValue('A mountain')
  expect(api.assistStudioPrompt).toHaveBeenCalledWith({
    token_id: 1,
    model: 'text-model',
    prompt: 'A mountain',
  })
  expect(api.createStudioRun).not.toHaveBeenCalled()
  complete(success({ suggested_prompt: 'A sunlit mountain, wide composition' }))
  await waitFor(() =>
    expect(prompt).toHaveValue('A sunlit mountain, wide composition')
  )
  expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  expect(api.createStudioRun).not.toHaveBeenCalled()
  await user.click(screen.getByRole('button', { name: 'Undo' }))
  expect(prompt).toHaveValue('A mountain')
  expect(screen.queryByRole('button', { name: 'Undo' })).not.toBeInTheDocument()
})

test('optimization with another owned key handles failure without replacing the original prompt', async () => {
  vi.mocked(getApiKeys).mockResolvedValue(
    success({
      items: [key, { ...key, id: 2, name: 'Text key' }],
      total: 2,
      page: 1,
      page_size: 100,
    })
  )
  vi.mocked(api.listStudioModels).mockImplementation(async (tokenId, kind) => {
    if (kind !== 'text') return success(['dall-e-3'])
    return success(tokenId === 2 ? ['text-model'] : [])
  })
  vi.mocked(api.assistStudioPrompt).mockRejectedValue(
    new Error('Text provider unavailable')
  )
  const user = await renderStudio()
  const prompt = screen.getByRole('textbox', { name: 'Describe your image' })
  await user.type(prompt, 'A mountain')
  expect(screen.getByRole('button', { name: 'Optimize prompt' })).toBeDisabled()
  await user.click(screen.getByRole('button', { name: 'Model settings' }))
  const dialog = await screen.findByRole('dialog')
  const textModels = within(dialog).getByRole('group', { name: 'Text model' })
  await within(textModels).findByText('No models available')
  await user.selectOptions(
    within(textModels).getByRole('combobox', { name: 'API key' }),
    '2'
  )
  await waitFor(() =>
    expect(
      within(textModels).getByRole('combobox', { name: 'Model' })
    ).toHaveValue('text-model')
  )
  await user.click(within(dialog).getByRole('button', { name: 'Done' }))
  await user.click(screen.getByRole('button', { name: 'Optimize prompt' }))
  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Text provider unavailable'
  )
  expect(api.assistStudioPrompt).toHaveBeenCalledWith({
    token_id: 2,
    model: 'text-model',
    prompt: 'A mountain',
  })
  expect(prompt).toHaveValue('A mountain')
  expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  expect(api.createStudioRun).not.toHaveBeenCalled()
})

test('recommended examples fill the prompt and style suggestions append without generating automatically', async () => {
  const user = await renderStudio()
  const prompt = screen.getByRole('textbox', { name: 'Describe your image' })
  const examples = screen.getByRole('region', { name: 'Try an example' })
  await user.click(
    within(examples).getByRole('button', { name: 'Cinematic photography' })
  )
  expect(prompt).toHaveValue(
    'A ginger cat on a sunny windowsill, warm afternoon light, cinematic photography, soft shadows and a calm composition.'
  )
  const styles = screen.getByRole('group', { name: 'Add a style' })
  await user.click(within(styles).getByRole('button', { name: 'Ink painting' }))
  expect(prompt).toHaveValue(
    'A ginger cat on a sunny windowsill, warm afternoon light, cinematic photography, soft shadows and a calm composition., ink painting, flowing brushwork, poetic negative space'
  )
  expect(api.createStudioRun).not.toHaveBeenCalled()
})

test('generation omits quality and DALL-E 3 explains its single-image limit', async () => {
  const user = await renderStudio()
  expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  expect(
    screen.queryByRole('group', { name: 'Image model' })
  ).not.toBeInTheDocument()
  expect(
    screen.queryByRole('textbox', { name: 'Image size' })
  ).not.toBeInTheDocument()
  const count = screen.getByRole('combobox', { name: 'Image count' })
  expect(within(count).getAllByRole('option')).toHaveLength(1)
  expect(
    screen.getByText('This model supports one image per request.')
  ).toBeInTheDocument()
  expect(
    screen.queryByRole('combobox', { name: 'Image quality' })
  ).not.toBeInTheDocument()
  await user.click(screen.getByRole('button', { name: 'Portrait' }))
  expect(screen.getByRole('button', { name: 'Portrait' })).toHaveAttribute(
    'aria-pressed',
    'true'
  )
  await submitImage(user)
  await waitFor(() =>
    expect(api.createStudioRun).toHaveBeenCalledWith(
      expect.objectContaining({
        request: expect.objectContaining({
          size: '1024x1792',
        }),
      })
    )
  )
  expect(
    vi.mocked(api.createStudioRun).mock.calls[0][0].request
  ).not.toHaveProperty('quality')
})

test('switching image models preserves default quality and clamps the count before submission', async () => {
  vi.mocked(api.listStudioModels).mockImplementation(async (_id, kind) =>
    success(kind === 'text' ? ['text-model'] : ['dall-e-3', 'gpt-image-1'])
  )
  const user = await renderStudio()
  expect(
    screen.queryByRole('combobox', { name: 'Image quality' })
  ).not.toBeInTheDocument()
  await user.click(screen.getByRole('button', { name: 'Model settings' }))
  let dialog = await screen.findByRole('dialog')
  await user.selectOptions(
    within(
      within(dialog).getByRole('group', { name: 'Image model' })
    ).getByRole('combobox', { name: 'Model' }),
    'gpt-image-1'
  )
  await user.click(within(dialog).getByRole('button', { name: 'Done' }))
  expect(
    screen.queryByRole('combobox', { name: 'Image quality' })
  ).not.toBeInTheDocument()
  expect(
    within(screen.getByRole('combobox', { name: 'Image count' }))
      .getAllByRole('option')
      .map((option) => option.getAttribute('value'))
  ).toEqual(['1', '2', '3', '4'])
  await user.selectOptions(
    screen.getByRole('combobox', { name: 'Image count' }),
    '3'
  )
  await user.click(screen.getByRole('button', { name: 'Landscape' }))
  await submitImage(user)
  await waitFor(() =>
    expect(api.createStudioRun).toHaveBeenCalledWith(
      expect.objectContaining({
        request: expect.objectContaining({
          model: 'gpt-image-1',
          n: 3,
          size: '1536x1024',
        }),
      })
    )
  )
  expect(
    vi.mocked(api.createStudioRun).mock.calls[0][0].request
  ).not.toHaveProperty('quality')
  await waitFor(() =>
    expect(screen.getByRole('button', { name: 'Model settings' })).toBeEnabled()
  )
  await user.click(screen.getByRole('button', { name: 'Model settings' }))
  dialog = await screen.findByRole('dialog')
  await user.selectOptions(
    within(
      within(dialog).getByRole('group', { name: 'Image model' })
    ).getByRole('combobox', { name: 'Model' }),
    'dall-e-3'
  )
  await user.click(within(dialog).getByRole('button', { name: 'Done' }))
  expect(screen.getByRole('combobox', { name: 'Image count' })).toHaveValue('1')
  expect(
    screen.queryByRole('combobox', { name: 'Image quality' })
  ).not.toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Model default' })).toHaveAttribute(
    'aria-pressed',
    'true'
  )
})

test('reusing a generated prompt focuses the composer without submitting or writing assets', async () => {
  vi.mocked(api.listStudioRuns).mockResolvedValue(
    success({ items: [run], total: 1 })
  )
  const user = await renderStudio()
  await user.click(
    within(
      screen.getByRole('region', { name: 'Generation records' })
    ).getByRole('img')
  )
  await user.click(
    await screen.findByRole('button', { name: 'Use this prompt' })
  )
  expect(
    screen.getByRole('textbox', { name: 'Describe your image' })
  ).toHaveValue('A mountain')
  expect(
    screen.getByRole('textbox', { name: 'Describe your image' })
  ).toHaveFocus()
  expect(api.createStudioRun).not.toHaveBeenCalled()
  expect(api.saveStudioResult).not.toHaveBeenCalled()
})

test('square-only models default to omitted size and submit explicit square size and count', async () => {
  vi.mocked(api.listStudioModels).mockImplementation(async (_id, kind) =>
    success(kind === 'text' ? ['text-model'] : ['dall-e-2'])
  )
  const user = await renderStudio('dall-e-2')
  const proportions = screen.getByRole('group', { name: 'Image proportions' })
  expect(within(proportions).getAllByRole('button')).toHaveLength(2)
  expect(
    within(proportions).getByRole('button', { name: 'Model default' })
  ).toHaveAttribute('aria-pressed', 'true')
  expect(
    screen.queryByRole('textbox', { name: 'Image size' })
  ).not.toBeInTheDocument()
  await user.click(screen.getByRole('button', { name: 'Square image' }))
  await user.selectOptions(
    screen.getByRole('combobox', { name: 'Image count' }),
    '2'
  )
  await submitImage(user)
  await waitFor(() =>
    expect(api.createStudioRun).toHaveBeenCalledWith(
      expect.objectContaining({
        request: expect.objectContaining({
          model: 'dall-e-2',
          n: 2,
          size: '1024x1024',
        }),
      })
    )
  )
})

test('selecting the model default after a proportion omits size and response format', async () => {
  const user = await renderStudio()
  await user.click(screen.getByRole('button', { name: 'Portrait' }))
  await user.click(screen.getByRole('button', { name: 'Model default' }))
  await submitImage(user)
  await waitFor(() => expect(api.createStudioRun).toHaveBeenCalled())
  expect(vi.mocked(api.createStudioRun).mock.calls[0][0].request).toEqual({
    model: 'dall-e-3',
    prompt: 'A mountain',
  })
})

test('Seedream 4.5 sends only required image parameters by default', async () => {
  vi.mocked(api.listStudioModels).mockImplementation(async (_id, kind) =>
    success(kind === 'text' ? ['text-model'] : ['doubao-seedream-4-5-251128'])
  )
  const user = await renderStudio('doubao-seedream-4-5-251128')
  await submitImage(user)
  await waitFor(() => expect(api.createStudioRun).toHaveBeenCalled())
  expect(vi.mocked(api.createStudioRun).mock.calls[0][0].request).toEqual({
    model: 'doubao-seedream-4-5-251128',
    prompt: 'A mountain',
  })
})

test('Seedream 4.5 explicitly selected proportions use a supported pixel size', async () => {
  vi.mocked(api.listStudioModels).mockImplementation(async (_id, kind) =>
    success(kind === 'text' ? ['text-model'] : ['doubao-seedream-4-5-251128'])
  )
  const user = await renderStudio('doubao-seedream-4-5-251128')
  await user.click(screen.getByRole('button', { name: 'Landscape' }))
  await submitImage(user)
  await waitFor(() => expect(api.createStudioRun).toHaveBeenCalled())
  expect(vi.mocked(api.createStudioRun).mock.calls[0][0].request).toEqual({
    model: 'doubao-seedream-4-5-251128',
    prompt: 'A mountain',
    size: '2496x1664',
  })
})

test('proportion choices explain model support and how the default determines image size', async () => {
  await renderStudio()
  expect(
    screen.getByRole('group', { name: 'Image proportions' })
  ).toHaveAccessibleDescription(
    'Aspect ratio support depends on the selected model. Model default lets the model choose the size and aspect ratio.'
  )
})
