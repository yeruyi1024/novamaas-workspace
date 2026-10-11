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
import axios from 'axios'

import { getMediaAssetPreview } from '@/features/assets/api'

import type { StudioResponse, StudioResult, StudioRun } from './types'

export function studioTaskName(run: StudioRun): string {
  return run.prompt?.trim().replaceAll(/\s+/g, ' ').slice(0, 40) || run.model
}

export function assertStudioSuccess<T>(response: StudioResponse<T>): T {
  if (!response.success) throw new Error(response.message || 'Request failed')
  return response.data
}

export function studioErrorMessage(error: unknown): string {
  if (axios.isAxiosError<StudioResponse<unknown>>(error)) {
    return error.response?.data?.message || error.message
  }
  return error instanceof Error ? error.message : 'Request failed'
}

export function studioStatusLabelKey(status: string): string {
  const labels: Record<string, string> = {
    completed: 'Completed',
    partial: 'Partial',
    failed: 'Failed',
    running: 'Running',
    pending: 'Pending',
    canceled: 'Canceled',
  }
  return labels[status] ?? status
}

export async function downloadStudioImage(
  result: StudioResult,
  usePreviewURL = false
): Promise<void> {
  let url = result.preview_url
  if (result.asset_id && !usePreviewURL) {
    try {
      url = assertStudioSuccess(
        await getMediaAssetPreview(result.asset_id, 'download')
      ).url
    } catch (error) {
      if (!url) throw error
    }
  }
  if (!url) throw new Error('Preview unavailable')
  let objectUrl: string | undefined
  try {
    const response = await fetch(url)
    if (response.ok) {
      objectUrl = URL.createObjectURL(await response.blob())
    }
  } catch {
    // An upstream image may disallow CORS; its direct link remains usable.
  }
  const link = document.createElement('a')
  link.href = objectUrl || url
  link.download = `${result.id}.png`
  link.target = '_blank'
  link.rel = 'noopener noreferrer'
  link.click()
  if (objectUrl) setTimeout(() => URL.revokeObjectURL(objectUrl), 1000)
}
