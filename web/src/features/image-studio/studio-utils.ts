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
    failed: 'Failed',
    running: 'Running',
  }
  return labels[status] ?? status
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
