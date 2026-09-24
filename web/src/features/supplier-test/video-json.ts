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
import type { VideoForm } from './types'

export function buildVideoRequestPayload(
  model: string,
  video: VideoForm
): Record<string, unknown> | null {
  const content: Array<Record<string, unknown>> = []
  const prompt = video.prompt.trim()
  if (prompt) {
    content.push({
      type: 'text',
      text: prompt,
    })
  }

  if (video.hasImage) {
    const role = video.role || 'reference_image'
    if (video.uploadMode === 'url' && video.imageUrl.trim()) {
      content.push({
        type: 'image_url',
        image_url: { url: video.imageUrl.trim() },
        role,
      })
    } else if (video.uploadMode === 'base64' && video.base64Data.trim()) {
      content.push({
        type: 'image_url',
        image_url: { url: video.base64Data.trim() },
        role,
      })
    }
  }

  if (video.hasLastFrame) {
    if (video.lastFrameMode === 'url' && video.lastFrameUrl.trim()) {
      content.push({
        type: 'image_url',
        image_url: { url: video.lastFrameUrl.trim() },
        role: 'last_frame',
      })
    } else if (
      video.lastFrameMode === 'base64' &&
      video.lastFrameBase64.trim()
    ) {
      content.push({
        type: 'image_url',
        image_url: { url: video.lastFrameBase64.trim() },
        role: 'last_frame',
      })
    }
  }

  const payload: Record<string, unknown> = {}
  if (content.length > 0) payload.content = content

  if (video.hasResolution && video.resolution) {
    payload.resolution = video.resolution
  }
  if (video.hasRatio && video.ratio) {
    payload.ratio = video.ratio
  }
  if (video.hasDuration && video.duration > 0) {
    payload.duration = video.duration
  }
  if (video.hasWatermark) {
    payload.watermark = video.watermark
  }
  if (video.hasSeed && video.seed.trim() !== '') {
    const parsed = Number.parseInt(video.seed.trim(), 10)
    if (!Number.isNaN(parsed)) {
      payload.seed = parsed
    }
  }
  if (video.hasGenerateAudio) {
    payload.generate_audio = video.generateAudio
  }
  if (video.hasReturnLastFrame) {
    payload.return_last_frame = video.returnLastFrame
  }

  if (video.hasCustomJson && video.customJson.trim()) {
    try {
      const extra = JSON.parse(video.customJson.trim())
      if (extra && typeof extra === 'object' && !Array.isArray(extra)) {
        Object.assign(payload, extra)
      }
    } catch {
      // Ignore invalid extra JSON in build preview
    }
  }

  if (Object.keys(payload).length === 0) return null
  const trimmedModel = model.trim()
  if (trimmedModel) payload.model = trimmedModel
  return payload
}

export type ParseVideoPayloadResult = {
  success: boolean
  error?: string
  videoPatch?: Partial<VideoForm>
  model?: string
}

export function parseVideoPayloadToForm(raw: string): ParseVideoPayloadResult {
  const trimmed = raw.trim()
  if (!trimmed) {
    return { success: false, error: 'Empty JSON string' }
  }

  let obj: Record<string, unknown>
  try {
    const parsed = JSON.parse(trimmed)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      return { success: false, error: 'Root payload must be a JSON object' }
    }
    obj = parsed as Record<string, unknown>
  } catch (err) {
    return {
      success: false,
      error: err instanceof Error ? err.message : 'Invalid JSON syntax',
    }
  }

  const patch: Partial<VideoForm> = {}
  let extractedModel: string | undefined

  if (typeof obj.model === 'string' && obj.model.trim()) {
    extractedModel = obj.model.trim()
  }

  const handledKeys = new Set([
    'model',
    'content',
    'resolution',
    'ratio',
    'duration',
    'watermark',
    'seed',
    'generate_audio',
    'return_last_frame',
  ])

  if (Array.isArray(obj.content)) {
    for (const item of obj.content) {
      if (!item || typeof item !== 'object') continue
      const itemType = (item as Record<string, unknown>).type
      if (
        itemType === 'text' &&
        typeof (item as Record<string, unknown>).text === 'string'
      ) {
        patch.prompt = String((item as Record<string, unknown>).text)
      } else if (itemType === 'image_url') {
        const imgObj = (item as Record<string, unknown>).image_url as
          | Record<string, unknown>
          | undefined
        const urlStr = typeof imgObj?.url === 'string' ? imgObj.url : ''
        const roleStr =
          typeof (item as Record<string, unknown>).role === 'string'
            ? String((item as Record<string, unknown>).role)
            : ''

        if (roleStr === 'last_frame') {
          patch.hasLastFrame = true
          if (urlStr.startsWith('data:image/')) {
            patch.lastFrameMode = 'base64'
            patch.lastFrameBase64 = urlStr
          } else {
            patch.lastFrameMode = 'url'
            patch.lastFrameUrl = urlStr
          }
        } else {
          patch.hasImage = true
          patch.role = roleStr || 'reference_image'
          if (urlStr.startsWith('data:image/')) {
            patch.uploadMode = 'base64'
            patch.base64Data = urlStr
          } else {
            patch.uploadMode = 'url'
            patch.imageUrl = urlStr
          }
        }
      }
    }
  }

  if (typeof obj.resolution === 'string' && obj.resolution.trim()) {
    patch.hasResolution = true
    patch.resolution = obj.resolution.trim()
  }

  if (typeof obj.ratio === 'string' && obj.ratio.trim()) {
    patch.hasRatio = true
    patch.ratio = obj.ratio.trim()
  }

  if (typeof obj.duration === 'number' && obj.duration > 0) {
    patch.hasDuration = true
    patch.duration = obj.duration
  }

  if (typeof obj.watermark === 'boolean') {
    patch.hasWatermark = true
    patch.watermark = obj.watermark
  }

  if (obj.seed !== undefined && obj.seed !== null) {
    patch.hasSeed = true
    patch.seed = String(obj.seed)
  }

  if (typeof obj.generate_audio === 'boolean') {
    patch.hasGenerateAudio = true
    patch.generateAudio = obj.generate_audio
  }

  if (typeof obj.return_last_frame === 'boolean') {
    patch.hasReturnLastFrame = true
    patch.returnLastFrame = obj.return_last_frame
  }

  const extra: Record<string, unknown> = {}
  for (const [k, v] of Object.entries(obj)) {
    if (!handledKeys.has(k)) {
      extra[k] = v
    }
  }

  if (Object.keys(extra).length > 0) {
    patch.hasCustomJson = true
    patch.customJson = JSON.stringify(extra, null, 2)
  }

  return {
    success: true,
    videoPatch: patch,
    model: extractedModel,
  }
}
