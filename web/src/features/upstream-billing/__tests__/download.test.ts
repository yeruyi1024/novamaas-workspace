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

import { api } from '@/lib/api'

import {
  downloadUpstreamAttachment,
  downloadUpstreamStatement,
  uploadUpstreamAttachment,
} from '../api'

afterEach(() => vi.restoreAllMocks())

test.each(['pdf', 'xlsx', 'details', 'snapshot', 'manifest', 'receipt'])(
  'a %s download uses the authenticated endpoint and preserves the server filename',
  async (kind) => {
    const filename = '供应商_2026-09.xlsx'
    const get = vi.spyOn(api, 'get').mockResolvedValue({
      data: new Blob(['archive']),
      headers: {
        'content-disposition': `attachment; filename*=UTF-8''${encodeURIComponent(filename)}`,
      },
    })
    vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:local-archive')
    const revoke = vi
      .spyOn(URL, 'revokeObjectURL')
      .mockImplementation(() => undefined)
    const click = vi
      .spyOn(HTMLAnchorElement.prototype, 'click')
      .mockImplementation(function (this: HTMLAnchorElement) {
        expect(this.download).toBe(filename)
      })
    await downloadUpstreamStatement('frozen-statement', kind, 0)
    expect(get).toHaveBeenCalledWith(
      `/api/upstream-billing/statements/frozen-statement/files/${kind}/0`,
      { responseType: 'blob' }
    )
    expect(click).toHaveBeenCalledOnce()
    expect(revoke).toHaveBeenCalledWith('blob:local-archive')
  }
)

test('supplier evidence uploads use the exact multipart file contract and preserve the UUID response', async () => {
  const file = new File(['invoice'], 'supplier.csv', { type: 'text/csv' })
  const post = vi.spyOn(api, 'post').mockResolvedValue({
    data: { success: true, data: { id: 'attachment-uuid' } },
  })
  expect(await uploadUpstreamAttachment('frozen-statement', file)).toEqual({
    id: 'attachment-uuid',
  })
  const [url, body] = post.mock.calls[0]
  expect(url).toBe(
    '/api/upstream-billing/statements/frozen-statement/attachments'
  )
  expect(body).toBeInstanceOf(FormData)
  expect([...(body as FormData).keys()]).toEqual(['file'])
  expect((body as FormData).get('file')).toBe(file)
})

test('supplier downloads remain authenticated and sanitize unsafe filename characters', async () => {
  const get = vi
    .spyOn(api, 'get')
    .mockResolvedValue({ data: new Blob(['invoice']), headers: {} })
  vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:evidence')
  const revoke = vi
    .spyOn(URL, 'revokeObjectURL')
    .mockImplementation(() => undefined)
  const click = vi
    .spyOn(HTMLAnchorElement.prototype, 'click')
    .mockImplementation(function (this: HTMLAnchorElement) {
      expect(this.download).toBe('.._supplier.pdf')
    })
  await downloadUpstreamAttachment(
    'frozen-statement',
    'uuid',
    '../supplier.pdf'
  )
  expect(get).toHaveBeenCalledWith(
    '/api/upstream-billing/statements/frozen-statement/attachments/uuid',
    { responseType: 'blob' }
  )
  expect(click).toHaveBeenCalledOnce()
  expect(revoke).toHaveBeenCalledWith('blob:evidence')
})
