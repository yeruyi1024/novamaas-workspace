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

export interface TaskRequestBodyResponse {
  success: boolean
  message?: string
  data?: unknown
}

export async function getTaskRequestBody(
  taskId: string,
  isAdmin: boolean
): Promise<TaskRequestBodyResponse> {
  const prefix = isAdmin ? '/api/task' : '/api/task/self'
  const res = await api.get<TaskRequestBodyResponse>(
    `${prefix}/${encodeURIComponent(taskId)}/request-body`,
    { disableDuplicate: true }
  )
  return res.data
}

export interface TaskVideoDownloadResponse {
  data: Blob
  headers: { 'content-type': string }
}

export interface TaskVideoContentInfoResponse {
  success: boolean
  message?: string
  data?: {
    delivery_mode: 'proxy' | 'redirect'
    url?: string
  }
}

export async function getTaskVideoContentInfo(
  taskId: string
): Promise<TaskVideoContentInfoResponse> {
  const res = await api.get<TaskVideoContentInfoResponse>(
    `/v1/videos/${encodeURIComponent(taskId)}/content-info`,
    {
      disableDuplicate: true,
      skipErrorHandler: true,
    }
  )
  return res.data
}

export async function downloadTaskVideo(
  taskId: string
): Promise<TaskVideoDownloadResponse> {
  const res = await api.get<Blob>(
    `/v1/videos/${encodeURIComponent(taskId)}/content`,
    {
      disableDuplicate: true,
      responseType: 'blob',
      skipErrorHandler: true,
    }
  )
  return {
    data: res.data,
    headers: {
      'content-type': String(res.headers['content-type'] || ''),
    },
  }
}
