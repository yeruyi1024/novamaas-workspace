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

export interface LoginNoticePeriodStats {
  generated: number
  violations: number
}

export interface LoginNoticeAnnouncement {
  id?: number
  content: string
  publishDate?: string
  type?: 'default' | 'ongoing' | 'success' | 'warning' | 'error'
  extra?: string
}

export interface LoginNoticeData {
  announcements: LoginNoticeAnnouncement[]
  statistics: {
    today: LoginNoticePeriodStats
    seven_days: LoginNoticePeriodStats
    thirty_days: LoginNoticePeriodStats
  }
  requires_acknowledgement: boolean
  acknowledged: boolean
}

interface LoginNoticeResponse {
  success: boolean
  message?: string
  data?: LoginNoticeData
}

export async function getLoginNotice(): Promise<LoginNoticeData> {
  const response = await api.get<LoginNoticeResponse>(
    '/api/user/login-notice',
    {
      disableDuplicate: true,
      skipErrorHandler: true,
    }
  )
  if (!response.data.success || !response.data.data) {
    throw new Error(response.data.message || 'login notice unavailable')
  }
  return response.data.data
}

export async function acknowledgeLoginNotice(
  deviceFingerprint: string
): Promise<void> {
  const response = await api.post<LoginNoticeResponse>(
    '/api/user/login-notice/acknowledge',
    { device_fingerprint: deviceFingerprint },
    { skipErrorHandler: true }
  )
  if (!response.data.success) {
    throw new Error(
      response.data.message || 'login notice acknowledgement failed'
    )
  }
}
