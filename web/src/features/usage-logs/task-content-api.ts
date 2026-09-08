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

import { TASK_PLATFORMS } from './constants'

export interface TaskRequestBodyResponse {
  success: boolean
  message?: string
  data?: unknown
}

export async function getTaskRequestBody(
  taskId: string
): Promise<TaskRequestBodyResponse> {
  const res = await api.get<TaskRequestBodyResponse>(
    `/api/task/${encodeURIComponent(taskId)}/request-body`,
    { disableDuplicate: true }
  )
  return res.data
}

export async function getLogRequestBody(
  taskId: string | undefined,
  requestId: string | undefined
): Promise<TaskRequestBodyResponse> {
  const res = await api.get<TaskRequestBodyResponse>('/api/log/request-body', {
    params: { task_id: taskId, request_id: requestId },
    disableDuplicate: true,
  })
  return res.data
}

const TASK_INFO_PATHS: Record<string, (taskId: string) => string> = {
  [TASK_PLATFORMS.DOUBAO_VIDEO]: (taskId) => `/v1/video/generations/${taskId}`,
  [TASK_PLATFORMS.VOLC_NATIVE]: (taskId) =>
    `/api/v3/contents/generations/tasks/${taskId}`,
}

export function canGetTaskInformation(platform: string): boolean {
  return platform in TASK_INFO_PATHS
}

export async function getTaskInformation(
  taskId: string,
  platform: string
): Promise<unknown> {
  const pathBuilder = TASK_INFO_PATHS[platform]
  if (!pathBuilder) throw new Error('task information unsupported')
  const res = await api.get<unknown>(pathBuilder(encodeURIComponent(taskId)), {
    disableDuplicate: true,
    skipErrorHandler: true,
  })
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
