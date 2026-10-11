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
export type StudioSettings = {
  enabled: boolean
  visible?: boolean
  asset_library_enabled?: boolean
  max_images: number
  max_reference_images: number
  max_prompt_characters: number
}

export type StudioResult = {
  id: string
  asset_id: string
  preview_url?: string
  thumbnail_url?: string
  description?: string
  index: number
  save_status?: string
  save_error?: string
}

export type StudioRun = {
  id: string
  owner_user_id: number
  status: string
  prompt?: string
  error_message?: string
  created_at: number
  model: string
  quota: number
  results: StudioResult[]
}

export type StudioResponse<T> = {
  success: boolean
  message?: string
  data: T
}

export type StudioRunList = {
  items: StudioRun[]
  total: number
  users?: { id: number; name: string }[]
}

export type CreateStudioRunInput = {
  token_id: number
  idempotency_key: string
  request: {
    model: string
    prompt: string
    n?: number
    size?: string
    response_format?: 'url' | 'b64_json'
  }
  input_asset_ids: string[]
}
