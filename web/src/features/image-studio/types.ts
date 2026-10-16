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
  max_images: number
  max_prompt_characters: number
}

export type StudioRunStatus = 'running' | 'completed' | 'failed'

export type StudioResult = {
  id: string
  index: number
  preview_url?: string
  thumbnail_url?: string
  // Signed attachment URL named after the stored format (png, jpg or webp).
  download_url?: string
  description?: string
  asset_id?: string
  expired?: boolean
}

export type StudioRun = {
  id: string
  owner_user_id: number
  username?: string
  model: string
  status: StudioRunStatus
  prompt?: string
  error_message?: string
  // Only returned to authorized reviewers.
  request_id?: string
  quota: number
  duration_ms?: number
  created_at: number
  results: StudioResult[]
}

// How library assets saved from deleted images are handled.
export type StudioLinkedAssets = 'keep' | 'delete'

export type StudioResponse<T> = {
  success: boolean
  message?: string
  code?: string
  data: T
  idempotent_replay?: boolean
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
    size?: string
  }
}
