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
import { t } from 'i18next'

import { api } from '@/lib/api'

import type {
  BillingAccount,
  BillingDay,
  BillingMonthPreview,
  BillingUsageDetails,
  BillingStatement,
  StatementDetail,
  BillingHistoryReview,
  BillingHistoryImport,
} from './types'

interface Response<T> {
  success: boolean
  data: T
  message?: string
}
function unwrap<T>(response: Response<T>): T {
  if (!response.success) {
    throw new Error(response.message || 'Something went wrong!')
  }
  return response.data
}
export async function getBillingDay(userId: number, date: string) {
  return unwrap(
    (
      await api.get<Response<BillingDay>>('/api/billing/day', {
        params: { user_id: userId, date },
      })
    ).data
  )
}
export async function getBillingAccount(userId: number) {
  return unwrap(
    (
      await api.get<Response<BillingAccount>>('/api/billing/account', {
        params: { user_id: userId },
      })
    ).data
  )
}
export async function saveBillingAccount(
  userId: number,
  input: Partial<BillingAccount>
) {
  return unwrap(
    (
      await api.put<Response<BillingAccount>>('/api/billing/account', input, {
        params: { user_id: userId },
      })
    ).data
  )
}
export async function getStatements(
  userId: number,
  cursor?: { at: number; id: string },
  status?: string
) {
  return unwrap(
    (
      await api.get<Response<BillingStatement[]>>('/api/billing/statements', {
        params: {
          user_id: userId,
          before: cursor?.at,
          before_id: cursor?.id,
          status,
        },
      })
    ).data
  )
}
export async function getStatement(id: string) {
  return unwrap(
    (await api.get<Response<StatementDetail>>(`/api/billing/statements/${id}`))
      .data
  )
}
export async function previewStatement(
  userId: number,
  month: string,
  storageProfileId = 0
) {
  return unwrap(
    (
      await api.get<Response<BillingMonthPreview>>(
        '/api/billing/admin/preview',
        {
          params: {
            user_id: userId,
            month,
            storage_profile_id: storageProfileId,
          },
        }
      )
    ).data
  )
}
export async function getBillingUsageDetails(
  userId: number,
  date: string,
  hour: number,
  cursor = ''
) {
  return unwrap(
    (
      await api.get<Response<BillingUsageDetails>>(
        '/api/billing/usage-details',
        {
          params: { user_id: userId, date, hour, cursor },
        }
      )
    ).data
  )
}
export async function prepareStatement(
  userId: number,
  month: string,
  storageProfileId: number
) {
  return unwrap(
    (
      await api.post<Response<BillingStatement>>(
        '/api/billing/admin/statements',
        { month, storage_profile_id: storageProfileId },
        { params: { user_id: userId } }
      )
    ).data
  )
}
export async function getBillingStorageProfiles() {
  return unwrap(
    (
      await api.get<Response<{ id: number; name: string }[]>>(
        '/api/billing/admin/storage-profiles'
      )
    ).data
  )
}
export async function statementAction(
  id: string,
  action: string,
  manifest: string,
  note: string
) {
  return unwrap(
    (
      await api.post<Response<BillingStatement>>(
        `/api/billing/statements/${id}/actions`,
        { action, manifest_sha256: manifest, note }
      )
    ).data
  )
}
export async function downloadStatement(id: string, kind: string, ordinal = 0) {
  const response = await api.get<Blob>(
    `/api/billing/statements/${id}/files/${kind}/${ordinal}`,
    { responseType: 'blob' }
  )
  const url = URL.createObjectURL(response.data)
  const anchor = document.createElement('a')
  anchor.href = url
  let extension = 'pdf'
  if (kind === 'details') extension = 'jsonl.gz'
  if (kind === 'manifest' || kind === 'snapshot') extension = 'json'
  let filename = ''
  const disposition = response.headers['content-disposition']
  if (typeof disposition === 'string') {
    const encoded = disposition.match(/(?:^|;)\s*filename\*=UTF-8''([^;]+)/i)
    if (encoded) {
      try {
        filename = decodeURIComponent(encoded[1].trim())
      } catch {
        // A malformed or older response still gets a safe non-UUID filename.
      }
    }
    if (!filename) {
      const plain = disposition.match(
        /(?:^|;)\s*filename=(?:"([^"]+)"|([^;]+))/i
      )
      filename = plain?.[1] || plain?.[2]?.trim() || ''
    }
  }
  anchor.download =
    filename.replaceAll(/[\\/\r\n]/g, '_') ||
    `${t('Download')}_${Date.now()}.${extension}`
  try {
    anchor.click()
  } finally {
    URL.revokeObjectURL(url)
  }
}
export async function reviewBillingHistory(userId: number, month: string) {
  return unwrap(
    (
      await api.get<Response<BillingHistoryReview>>(
        '/api/billing/admin/history-review',
        { params: { user_id: userId, month } }
      )
    ).data
  )
}
export async function getBillingHistoryImports(userId: number, month: string) {
  return unwrap(
    (
      await api.get<Response<BillingHistoryImport[]>>(
        '/api/billing/admin/history-imports',
        { params: { user_id: userId, month } }
      )
    ).data
  )
}
export async function confirmBillingHistoryImport(
  userId: number,
  input: {
    id: string
    month: string
    storage_profile_id: number
    source_sha256: string
    acknowledged: boolean
    note: string
  }
) {
  return unwrap(
    (
      await api.post<Response<BillingHistoryImport>>(
        '/api/billing/admin/history-imports',
        input,
        { params: { user_id: userId } }
      )
    ).data
  )
}
export async function downloadBillingHistorySource(userId: number, id: string) {
  const response = await api.get<Blob>(
    `/api/billing/admin/history-imports/${id}/source`,
    { params: { user_id: userId }, responseType: 'blob' }
  )
  const url = URL.createObjectURL(response.data)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `billing-${id}.json.gz`
  anchor.click()
  URL.revokeObjectURL(url)
}
export function billingToday() {
  return new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(new Date())
}
export function billingTimestamp(value: number) {
  if (!value) return '-'
  return new Intl.DateTimeFormat('sv-SE', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date(value * 1000))
}
export function billingPreviousMonth() {
  const [year, month] = billingToday().split('-').map(Number)
  const previous = new Date(Date.UTC(year, month - 2, 1))
  return previous.toISOString().slice(0, 7)
}
