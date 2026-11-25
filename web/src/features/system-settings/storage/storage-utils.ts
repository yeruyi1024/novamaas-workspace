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
import i18next from 'i18next'

import type { AssetLibraryConnection, StorageAPIResponse } from './types'

export function assertStorageSuccess<T>(response: StorageAPIResponse<T>): T {
  if (!response.success) {
    throw new Error(response.message || i18next.t('Storage operation failed'))
  }
  return response.data
}

export function storageErrorMessage(error: unknown): string {
  if (axios.isAxiosError<StorageAPIResponse>(error)) {
    return (
      error.response?.data?.message ||
      error.message ||
      i18next.t('Storage operation failed')
    )
  }
  return error instanceof Error
    ? error.message
    : i18next.t('Storage operation failed')
}

/** An account uploads only while enabled and linked to an enabled channel. */
export function canUpload(library: AssetLibraryConnection): boolean {
  return library.enabled && library.channels.some((channel) => channel.enabled)
}

/**
 * Uploads an account that cannot upload keeps paused; "Clear failures" can
 * release them like failures.
 */
export function pausedUploadCount(library: AssetLibraryConnection): number {
  if (canUpload(library)) return 0
  const summary = library.summary
  return summary.pending + summary.processing - (summary.deleting ?? 0)
}
