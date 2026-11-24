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

import type { AssetLibraryResponse, AssetRef, MediaAsset } from './types'

export function assertAssetSuccess<T>(response: AssetLibraryResponse<T>): T {
  if (!response.success) throw new Error(response.message || 'Request failed')
  return response.data
}

export function assetErrorMessage(error: unknown): string {
  if (axios.isAxiosError<AssetLibraryResponse>(error)) {
    const serverMessage = error.response?.data?.message
    if (serverMessage) return serverMessage
    if (error.response?.status === 429) {
      return i18next.t('Too many requests')
    }
    return error.message
  }
  return error instanceof Error ? error.message : 'Request failed'
}

export function formatAssetBytes(value: number): string {
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KiB`
  return `${(value / 1024 / 1024).toFixed(1)} MiB`
}

const IMAGE_EXTENSIONS = new Set(['jpg', 'jpeg', 'png', 'webp'])
const VIDEO_EXTENSIONS = new Set(['mp4', 'webm', 'mov'])
const AUDIO_EXTENSIONS = new Set(['mp3', 'wav', 'm4a', 'aac', 'ogg'])

export function inferAssetType(file: File): MediaAsset['type'] | null {
  if (file.type.startsWith('image/')) return 'image'
  if (file.type.startsWith('video/')) return 'video'
  if (file.type.startsWith('audio/')) return 'audio'
  const ext = file.name.split('.').pop()?.toLowerCase() ?? ''
  if (IMAGE_EXTENSIONS.has(ext)) return 'image'
  if (VIDEO_EXTENSIONS.has(ext)) return 'video'
  if (AUDIO_EXTENSIONS.has(ext)) return 'audio'
  return null
}

export function assetKey(asset: AssetRef): string {
  return `${asset.owner_user_id}:${asset.id}`
}

export function formatMediaDuration(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds <= 0) return ''
  const total = Math.round(seconds)
  const minutes = Math.floor(total / 60)
  const rest = String(total % 60).padStart(2, '0')
  if (minutes < 60) return `${minutes}:${rest}`
  const hours = Math.floor(minutes / 60)
  return `${hours}:${String(minutes % 60).padStart(2, '0')}:${rest}`
}

/**
 * Decorative waveform heights (0.2–1) derived from the file checksum, so each
 * audio card keeps the same distinct silhouette without decoding the audio.
 */
export function assetWaveform(
  seed: string,
  bars: number
): { id: string; height: number }[] {
  const source = seed || 'asset'
  const heights: { id: string; height: number }[] = []
  for (let index = 0; index < bars; index += 1) {
    const code = source.charCodeAt(index % source.length)
    const neighbour = source.charCodeAt((index * 7 + 3) % source.length)
    const wave = Math.sin((index / bars) * Math.PI)
    const jitter = ((code * 31 + neighbour) % 97) / 97
    heights.push({
      id: `bar-${index}`,
      height: 0.2 + 0.8 * (jitter * 0.6 + wave * 0.4),
    })
  }
  return heights
}
