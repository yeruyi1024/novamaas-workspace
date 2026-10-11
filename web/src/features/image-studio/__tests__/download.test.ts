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
import { afterEach, expect, test, vi } from 'vitest'

import { getMediaAssetPreview } from '@/features/assets/api'

import { downloadStudioImage } from '../studio-utils'

vi.mock('@/features/assets/api')
afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
  vi.resetAllMocks()
})

test('saved images download the signed original as a blob with the result filename', async () => {
  vi.useFakeTimers()
  const original = 'https://oss.example.test/original?signature=download'
  vi.mocked(getMediaAssetPreview).mockResolvedValue({
    success: true,
    message: '',
    data: { url: original, expires_at: 1900000000 },
  })
  const blob = new Blob(['image bytes'], { type: 'image/png' })
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({ ok: true, blob: async () => blob })
  )
  const createObjectURL = vi.fn().mockReturnValue('blob:studio-download')
  const revokeObjectURL = vi.fn()
  vi.stubGlobal(
    'URL',
    class extends URL {
      static createObjectURL = createObjectURL
      static revokeObjectURL = revokeObjectURL
    }
  )
  const click = vi
    .spyOn(HTMLAnchorElement.prototype, 'click')
    .mockImplementation(() => {})

  await downloadStudioImage({
    id: 'result-original',
    index: 0,
    asset_id: 'asset-owned',
    preview_url: 'https://oss.example.test/thumbnail',
  })

  expect(getMediaAssetPreview).toHaveBeenCalledWith('asset-owned', 'download')
  expect(fetch).toHaveBeenCalledWith(original)
  expect(createObjectURL).toHaveBeenCalledWith(blob)
  expect(click.mock.instances[0]).toHaveAttribute(
    'href',
    'blob:studio-download'
  )
  expect(click.mock.instances[0]).toHaveAttribute(
    'download',
    'result-original.png'
  )
  vi.runAllTimers()
  expect(revokeObjectURL).toHaveBeenCalledWith('blob:studio-download')
})

test('review downloads use the authorized preview and fall back to its direct URL when CORS blocks fetch', async () => {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockRejectedValue(new TypeError('Failed to fetch'))
  )
  const click = vi
    .spyOn(HTMLAnchorElement.prototype, 'click')
    .mockImplementation(() => {})
  const preview = 'https://oss.example.test/review?signature=content-read'

  await downloadStudioImage(
    {
      id: 'review-image',
      index: 0,
      asset_id: 'other-owner-asset',
      preview_url: preview,
    },
    true
  )

  expect(getMediaAssetPreview).not.toHaveBeenCalled()
  expect(fetch).toHaveBeenCalledWith(preview)
  expect(click.mock.instances[0]).toHaveAttribute('href', preview)
  expect(click.mock.instances[0]).toHaveAttribute(
    'download',
    'review-image.png'
  )
  expect(click.mock.instances[0]).toHaveAttribute('rel', 'noopener noreferrer')
})

test('deleting a library asset still allows downloading the retained generated image', async () => {
  vi.mocked(getMediaAssetPreview).mockRejectedValue(
    new Error('Asset not found')
  )
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false }))
  const click = vi
    .spyOn(HTMLAnchorElement.prototype, 'click')
    .mockImplementation(() => {})
  const preview = 'data:image/png;base64,aW1hZ2U='

  await downloadStudioImage({
    id: 'retained-image',
    index: 0,
    asset_id: 'deleted-asset',
    preview_url: preview,
  })

  expect(fetch).toHaveBeenCalledWith(preview)
  expect(click).toHaveBeenCalledOnce()
  expect(click.mock.instances[0]).toHaveAttribute('href', preview)
})
