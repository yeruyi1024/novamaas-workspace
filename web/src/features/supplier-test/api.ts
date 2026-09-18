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
import { api } from '@/lib/api'

import { API_ENDPOINTS } from './constants'

export async function fetchSupplierModels(payload: {
  base_url: string
  api_key: string
}): Promise<string[]> {
  const res = await api.post(API_ENDPOINTS.MODELS, payload, {
    skipBusinessError: true,
    skipErrorHandler: true,
  })
  if (!res.data?.success) {
    throw new Error(String(res.data?.message || 'Failed to fetch models'))
  }
  return (res.data?.data ?? []) as string[]
}
