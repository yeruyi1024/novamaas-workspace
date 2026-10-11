import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { getStudioRun, listStudioRuns } from '../api'
import { StudioReview } from '../components/studio-review'
import type { StudioRun } from '../types'

vi.mock('../api', () => ({ listStudioRuns: vi.fn(), getStudioRun: vi.fn() }))
const run: StudioRun = {
  id: 'run-review',
  owner_user_id: 7,
  status: 'completed',
  prompt: 'Review this mountain',
  created_at: 1,
  model: 'seedream',
  quota: 0,
  results: [
    {
      id: 'result-review',
      asset_id: '',
      thumbnail_url: 'https://example.test/thumbnail.png',
      index: 0,
    },
    {
      id: 'result-review-2',
      asset_id: '',
      thumbnail_url: 'https://example.test/thumbnail-2.png',
      index: 1,
    },
  ],
}
const users = [
  { id: 7, name: 'artist' },
  { id: 8, name: 'designer' },
]
let client: QueryClient
beforeEach(() => {
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  vi.mocked(listStudioRuns).mockResolvedValue({
    success: true,
    data: { items: [run], total: 13, users },
  })
  vi.mocked(getStudioRun).mockResolvedValue({
    success: true,
    data: {
      ...run,
      results: run.results.map((result) => ({
        ...result,
        preview_url: 'https://example.test/image.png',
      })),
    },
  })
})
afterEach(() => {
  client.clear()
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
  vi.resetAllMocks()
})

function renderReview() {
  render(
    <QueryClientProvider client={client}>
      <StudioReview />
    </QueryClientProvider>
  )
  return userEvent.setup()
}

test('review shows every image in compact scrollable cards and opens preview by keyboard', async () => {
  const user = renderReview()
  const images = await screen.findAllByRole('img', { name: run.prompt })
  expect(images).toHaveLength(2)
  for (const image of images) {
    expect(image).toHaveClass('h-40', 'object-contain')
  }
  expect(images[0]).toHaveAttribute('src', 'https://example.test/thumbnail.png')
  expect(getStudioRun).not.toHaveBeenCalled()
  const list = screen.getByRole('region', { name: 'Image review' })
  expect(list).toHaveClass('grid', 'grid-cols-2', 'min-h-0', 'overflow-y-auto')
  expect(list).toHaveAttribute('tabindex', '0')
  expect(within(list).getAllByText(/artist.*User ID.*7/)).toHaveLength(2)
  expect(within(list).getAllByText(/Task ID.*run-review/)).toHaveLength(2)
  expect(listStudioRuns).toHaveBeenCalledWith(1, 12, 'all', 0)
  expect(
    screen.queryByRole('button', { name: 'Save to asset library' })
  ).not.toBeInTheDocument()
  const preview = within(list).getAllByRole('button', { name: 'Preview' })[0]
  preview.focus()
  await user.keyboard('{Enter}')
  const dialog = await screen.findByRole('dialog')
  expect(dialog).toHaveClass('overflow-y-auto')
  await waitFor(() =>
    expect(within(dialog).getByRole('img')).toHaveAttribute(
      'src',
      'https://example.test/image.png'
    )
  )
  expect(getStudioRun).toHaveBeenCalledWith(run.id, 'all')
  await user.keyboard('{Escape}')
  await waitFor(() =>
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  )
})

test('selecting a user resets pagination and returning to all users clears the filter', async () => {
  const user = renderReview()
  await screen.findAllByRole('img')
  const filter = screen.getByRole('combobox', { name: 'User' })
  expect(filter).toHaveValue('0')
  await user.click(screen.getByRole('button', { name: 'Next' }))
  await waitFor(() =>
    expect(listStudioRuns).toHaveBeenCalledWith(2, 12, 'all', 0)
  )
  await user.selectOptions(filter, '7')
  await waitFor(() =>
    expect(listStudioRuns).toHaveBeenCalledWith(1, 12, 'all', 7)
  )
  expect(
    screen.getByRole('option', { name: /designer.*User ID.*8/ })
  ).toBeInTheDocument()
  await user.selectOptions(filter, '0')
  expect(filter).toHaveValue('0')
  expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
})

test('download is available on the review card without opening the record or preview', async () => {
  vi.stubGlobal(
    'fetch',
    vi
      .fn()
      .mockResolvedValue({ ok: true, blob: async () => new Blob(['image']) })
  )
  vi.stubGlobal(
    'URL',
    class extends URL {
      static createObjectURL = vi.fn().mockReturnValue('blob:review-download')
      static revokeObjectURL = vi.fn()
    }
  )
  const click = vi
    .spyOn(HTMLAnchorElement.prototype, 'click')
    .mockImplementation(() => {})
  const user = renderReview()
  const buttons = await screen.findAllByRole('button', { name: 'Download' })
  await user.click(buttons[0])
  await waitFor(() => expect(click).toHaveBeenCalledOnce())
  expect(fetch).toHaveBeenCalledWith('https://example.test/image.png')
  expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
})
