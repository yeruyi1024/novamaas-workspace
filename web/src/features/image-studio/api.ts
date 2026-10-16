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

import { assertStudioSuccess } from './studio-utils'
import type {
  CreateStudioRunInput,
  StudioLinkedAssets,
  StudioResponse,
  StudioRun,
  StudioRunList,
  StudioSettings,
} from './types'

async function studioData<T>(
  request: Promise<{ data: StudioResponse<T> }>
): Promise<T> {
  return assertStudioSuccess((await request).data)
}

const quietErrors = { skipBusinessError: true, skipErrorHandler: true }

export function getStudioSettings() {
  return studioData(
    api.get<StudioResponse<StudioSettings>>('/api/image-studio/settings')
  )
}

export function listStudioModels(tokenId: number, kind?: 'text') {
  return studioData(
    api.get<StudioResponse<string[]>>('/api/image-studio/models', {
      params: { token_id: tokenId, kind },
    })
  )
}

export function assistStudioPrompt(input: {
  token_id: number
  model: string
  prompt: string
}) {
  return studioData(
    api.post<StudioResponse<{ suggested_prompt: string }>>(
      '/api/image-studio/prompt-assist',
      input,
      { timeout: 150_000, ...quietErrors }
    )
  )
}

export async function createStudioRun(
  input: CreateStudioRunInput
): Promise<{ run: StudioRun; replay: boolean }> {
  const response = await api.post<StudioResponse<StudioRun>>(
    '/api/image-studio/runs',
    input,
    {
      timeout: 300_000,
      headers: { 'Idempotency-Key': input.idempotency_key },
      ...quietErrors,
    }
  )
  return {
    run: assertStudioSuccess(response.data),
    replay: response.data.idempotent_replay === true,
  }
}

export function listStudioRuns(
  page = 1,
  pageSize = 12,
  scope?: 'all',
  userId = 0
) {
  return studioData(
    api.get<StudioResponse<StudioRunList>>('/api/image-studio/runs', {
      params: {
        p: page,
        page_size: pageSize,
        scope,
        user_id: userId || undefined,
      },
    })
  )
}

export function getStudioRun(id: string, scope?: 'all') {
  return studioData(
    api.get<StudioResponse<StudioRun>>(
      `/api/image-studio/runs/${encodeURIComponent(id)}`,
      { params: { scope } }
    )
  )
}

export function deleteStudioRun(id: string, linkedAssets?: StudioLinkedAssets) {
  return studioData(
    api.delete<StudioResponse<undefined>>(
      `/api/image-studio/runs/${encodeURIComponent(id)}`,
      { ...quietErrors, params: { linked_assets: linkedAssets } }
    )
  )
}

export function saveStudioResult(resultId: string, groupId: string) {
  return studioData(
    api.post<StudioResponse<{ result_id: string; asset_id: string }>>(
      `/api/image-studio/results/${encodeURIComponent(resultId)}/save`,
      { group_id: groupId },
      quietErrors
    )
  )
}

export function deleteStudioResult(
  resultId: string,
  linkedAssets?: StudioLinkedAssets
) {
  return studioData(
    api.delete<StudioResponse<undefined>>(
      `/api/image-studio/results/${encodeURIComponent(resultId)}`,
      { ...quietErrors, params: { linked_assets: linkedAssets } }
    )
  )
}
