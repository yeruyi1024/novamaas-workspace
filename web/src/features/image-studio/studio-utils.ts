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

import type {
  StudioImageSize,
  StudioResponse,
  StudioResult,
  StudioRun,
} from './types'

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
    failed: 'Failed',
    running: 'Running',
  }
  return labels[status] ?? status
}

export function studioStatusBadgeVariant(
  status: string
): 'secondary' | 'destructive' | 'warning' {
  if (status === 'failed') return 'destructive'
  if (status === 'running') return 'warning'
  return 'secondary'
}

// The signed download URL carries Content-Disposition with the stored file
// name and extension, so the browser saves it without fetching through CORS.
export function downloadStudioImage(result: StudioResult): void {
  if (!result.download_url) throw new Error('Preview unavailable')
  const link = document.createElement('a')
  link.href = result.download_url
  link.rel = 'noopener noreferrer'
  link.click()
}

export const studioAspectRatios: Record<string, StudioImageSize> = {
  '1:1': '1024x1024',
  '3:2': '1536x1024',
  '2:3': '1024x1536',
  '16:9': '1792x1024',
  '9:16': '1024x1792',
}

// Only map ratios to the server's bounded size presets. Explicit UI choices win.
export function studioImageSize(
  prompt: string,
  selection: StudioImageSize
): StudioImageSize | undefined {
  if (selection !== 'auto') return selection
  const matches = [
    ...prompt.matchAll(
      /(?<![\d.])(16|9|3|2|1)\s*[:：/]\s*(16|9|3|2|1)(?![\d.])/g
    ),
  ]
  for (const match of matches.reverse()) {
    const size = studioAspectRatios[`${match[1]}:${match[2]}`]
    if (size) return size
  }
  return undefined
}

export function studioPollInterval(error: unknown, batches = 1): number {
  const interval = Math.max(3000, batches * 1500)
  if (!axios.isAxiosError(error) || error.response?.status !== 429) {
    return interval
  }
  const retryAfter = error.response.headers?.['retry-after']
  const seconds = Number(retryAfter)
  if (Number.isFinite(seconds) && seconds > 0) {
    return Math.max(interval, seconds * 1000)
  }
  const date =
    typeof retryAfter === 'string' ? Date.parse(retryAfter) : Number.NaN
  if (Number.isFinite(date)) return Math.max(interval, date - Date.now())
  return Math.max(interval, 30_000)
}
