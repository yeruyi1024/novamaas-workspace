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
  UpstreamAttachment,
  UpstreamChannel,
  UpstreamGroup,
  UpstreamRange,
  UpstreamReport,
  UpstreamStatement,
  UpstreamStatementDetail,
  UpstreamStorageProfile,
  UpstreamUsageDetails,
} from './types'

interface Response<T> {
  success: boolean
  data: T
  message?: string
}
const BASE = '/api/upstream-billing'

function unwrap<T>(response: Response<T>): T {
  if (!response.success) {
    throw new Error(response.message || 'Something went wrong!')
  }
  return response.data
}

export async function getUpstreamGroups(): Promise<UpstreamGroup[]> {
  return unwrap(
    (await api.get<Response<UpstreamGroup[]>>(`${BASE}/groups`)).data
  )
}

export async function saveUpstreamGroup(input: {
  id?: number
  name: string
  version: number
  channel_ids: number[]
}): Promise<UpstreamGroup> {
  const { id, ...body } = input
  const response = id
    ? await api.put<Response<UpstreamGroup>>(`${BASE}/groups/${id}`, body)
    : await api.post<Response<UpstreamGroup>>(`${BASE}/groups`, body)
  return unwrap(response.data)
}

export async function getUpstreamChannels(q = ''): Promise<UpstreamChannel[]> {
  return unwrap(
    (
      await api.get<Response<UpstreamChannel[]>>(`${BASE}/channels`, {
        params: { q },
      })
    ).data
  )
}

export async function getUpstreamReport(
  groupId: number,
  range: UpstreamRange
): Promise<UpstreamReport> {
  return unwrap(
    (
      await api.get<Response<UpstreamReport>>(`${BASE}/report`, {
        params: { group_id: groupId, ...range },
      })
    ).data
  )
}

export async function getUpstreamUsageDetails(
  groupId: number,
  range: UpstreamRange,
  filters: {
    day?: string
    model?: string
    cursor?: number
  }
): Promise<UpstreamUsageDetails> {
  return unwrap(
    (
      await api.get<Response<UpstreamUsageDetails>>(`${BASE}/usage-details`, {
        params: { group_id: groupId, ...range, ...filters },
      })
    ).data
  )
}

export async function getUpstreamStorageProfiles(): Promise<
  UpstreamStorageProfile[]
> {
  return unwrap(
    (
      await api.get<Response<UpstreamStorageProfile[]>>(
        `${BASE}/storage-profiles`
      )
    ).data
  )
}

export async function getUpstreamStatements(
  groupId: number,
  status: string,
  cursor?: { at: number; id: string }
): Promise<UpstreamStatement[]> {
  return unwrap(
    (
      await api.get<Response<UpstreamStatement[]>>(`${BASE}/statements`, {
        params: {
          group_id: groupId,
          status: status === 'all' ? undefined : status,
          before: cursor?.at,
          before_id: cursor?.id,
        },
      })
    ).data
  )
}

export async function prepareUpstreamStatement(
  groupId: number,
  month: string,
  storageProfileId: number
): Promise<UpstreamStatement> {
  return unwrap(
    (
      await api.post<Response<UpstreamStatement>>(`${BASE}/statements`, {
        group_id: groupId,
        month,
        storage_profile_id: storageProfileId,
      })
    ).data
  )
}

export async function getUpstreamStatement(
  id: string
): Promise<UpstreamStatementDetail> {
  return unwrap(
    (
      await api.get<Response<UpstreamStatementDetail>>(
        `${BASE}/statements/${encodeURIComponent(id)}`
      )
    ).data
  )
}

export async function upstreamStatementAction(
  id: string,
  input: {
    action: 'confirm' | 'void' | 'retry'
    manifest_sha256: string
    evidence_sha256?: string
    reason: string
    acknowledged: boolean
  }
): Promise<UpstreamStatement> {
  return unwrap(
    (
      await api.post<Response<UpstreamStatement>>(
        `${BASE}/statements/${encodeURIComponent(id)}/actions`,
        input
      )
    ).data
  )
}

export async function uploadUpstreamAttachment(
  id: string,
  file: File
): Promise<UpstreamAttachment> {
  const body = new FormData()
  body.append('file', file)
  return unwrap(
    (
      await api.post<Response<UpstreamAttachment>>(
        `${BASE}/statements/${encodeURIComponent(id)}/attachments`,
        body
      )
    ).data
  )
}

export async function downloadUpstreamAttachment(
  statementId: string,
  attachmentId: string,
  filename: string
): Promise<void> {
  const response = await api.get<Blob>(
    `${BASE}/statements/${encodeURIComponent(statementId)}/attachments/${encodeURIComponent(attachmentId)}`,
    { responseType: 'blob' }
  )
  saveUpstreamDownload(
    response.data,
    response.headers['content-disposition'],
    filename
  )
}

export async function downloadUpstreamStatement(
  id: string,
  kind: string,
  ordinal = 0
): Promise<void> {
  const response = await api.get<Blob>(
    `${BASE}/statements/${encodeURIComponent(id)}/files/${encodeURIComponent(kind)}/${ordinal}`,
    { responseType: 'blob' }
  )
  let extension = 'json'
  if (kind === 'details') {
    extension = 'jsonl.gz'
  } else if (kind === 'xlsx') {
    extension = 'xlsx'
  } else if (kind === 'pdf' || kind === 'receipt') {
    extension = 'pdf'
  }
  saveUpstreamDownload(
    response.data,
    response.headers['content-disposition'],
    `upstream-${id}-${kind}.${extension}`
  )
}

function saveUpstreamDownload(
  blob: Blob,
  header: unknown,
  fallback: string
): void {
  const disposition = String(header || '')
  const encoded = disposition.match(/(?:^|;)\s*filename\*=UTF-8''([^;]+)/i)
  const plain = disposition.match(/(?:^|;)\s*filename=(?:"([^"]+)"|([^;]+))/i)
  let filename = plain?.[1] || plain?.[2]?.trim() || ''
  if (encoded) {
    try {
      filename = decodeURIComponent(encoded[1].trim())
    } catch {
      /* Use the plain name when encoding is invalid. */
    }
  }
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = (filename || fallback).replaceAll(/[\\/\r\n]/g, '_')
  try {
    anchor.click()
  } finally {
    URL.revokeObjectURL(url)
  }
}
