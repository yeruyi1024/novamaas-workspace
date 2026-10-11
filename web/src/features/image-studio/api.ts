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

import type {
  CreateStudioRunInput,
  StudioResponse,
  StudioRun,
  StudioRunList,
  StudioSettings,
} from './types'

export async function getStudioSettings() {
  return (
    await api.get<StudioResponse<StudioSettings>>('/api/image-studio/settings')
  ).data
}

export async function listStudioModels(tokenId: number, kind?: 'text') {
  return (
    await api.get<StudioResponse<string[]>>('/api/image-studio/models', {
      params: { token_id: tokenId, kind },
    })
  ).data
}

export async function assistStudioPrompt(input: {
  token_id: number
  model: string
  prompt: string
}) {
  return (
    await api.post<StudioResponse<{ suggested_prompt: string }>>(
      '/api/image-studio/prompt-assist',
      input,
      {
        timeout: 150_000,
        skipBusinessError: true,
        skipErrorHandler: true,
      }
    )
  ).data
}

export async function createStudioRun(input: CreateStudioRunInput) {
  return (
    await api.post<StudioResponse<StudioRun>>('/api/image-studio/runs', input, {
      timeout: 300_000,
      headers: { 'Idempotency-Key': input.idempotency_key },
      skipBusinessError: true,
      skipErrorHandler: true,
    })
  ).data
}

export async function listStudioRuns(
  page = 1,
  pageSize = 12,
  scope?: 'all',
  userId = 0
) {
  return (
    await api.get<StudioResponse<StudioRunList>>('/api/image-studio/runs', {
      params: {
        p: page,
        page_size: pageSize,
        scope,
        user_id: userId || undefined,
      },
    })
  ).data
}

export async function getStudioRun(id: string, scope?: 'all') {
  return (
    await api.get<StudioResponse<StudioRun>>(
      `/api/image-studio/runs/${encodeURIComponent(id)}`,
      { params: { scope } }
    )
  ).data
}

export async function saveStudioResult(resultId: string, groupId: string) {
  return (
    await api.post<StudioResponse<undefined>>(
      `/api/image-studio/results/${encodeURIComponent(resultId)}/save`,
      {
        group_id: groupId,
      },
      { skipBusinessError: true, skipErrorHandler: true }
    )
  ).data
}

export async function deleteStudioResult(resultId: string) {
  return (
    await api.delete<StudioResponse<undefined>>(
      `/api/image-studio/results/${encodeURIComponent(resultId)}`,
      { skipBusinessError: true, skipErrorHandler: true }
    )
  ).data
}
