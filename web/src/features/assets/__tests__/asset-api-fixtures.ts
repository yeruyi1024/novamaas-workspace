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
import { vi } from 'vitest'

import { previewMediaAssets, summarizeMediaAssets } from '../api'

export const THUMBNAIL_URL = 'https://example.com/thumbnail.jpg'

/**
 * Gives the page-level summary and batch thumbnail endpoints neutral answers.
 * Test files must already mock `../api`.
 */
export function mockAssetPageApis() {
  vi.mocked(summarizeMediaAssets).mockResolvedValue({
    success: true,
    data: {
      total: 1,
      by_type: { image: 1, video: 0, audio: 0 },
      by_status: { ready: 1, processing: 0, unavailable: 0 },
    },
  })
  vi.mocked(previewMediaAssets).mockImplementation(async (items) => ({
    success: true,
    data: Object.fromEntries(
      items.map((item) => [
        `${item.owner_user_id}:${item.id}`,
        { url: THUMBNAIL_URL, expires_at: 4102444800 },
      ])
    ),
  }))
}
