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
import { describe, expect, test } from 'vitest'

import { DEFAULT_VIDEO_FORM } from './constants'
import type { VideoForm } from './types'
import { buildVideoRequestPayload, parseVideoPayloadToForm } from './video-json'

describe('video-json helpers', () => {
  test('buildVideoRequestPayload builds standard Volcano Ark payload with only enabled fields', () => {
    const video: VideoForm = {
      ...DEFAULT_VIDEO_FORM,
      prompt: 'A cyber dog playing football',
      hasResolution: true,
      resolution: '1080p',
      hasRatio: true,
      ratio: '16:9',
      hasDuration: true,
      duration: 5,
      hasWatermark: true,
      watermark: false,
      hasSeed: true,
      seed: '998877',
      hasGenerateAudio: true,
      generateAudio: true,
    }

    const payload = buildVideoRequestPayload('doubao-seedance-1-0-pro', video)
    expect(payload).not.toBeNull()
    if (!payload) return
    expect(payload.model).toBe('doubao-seedance-1-0-pro')
    expect(payload.resolution).toBe('1080p')
    expect(payload.ratio).toBe('16:9')
    expect(payload.duration).toBe(5)
    expect(payload.watermark).toBe(false)
    expect(payload.seed).toBe(998877)
    expect(payload.generate_audio).toBe(true)
    expect(payload.return_last_frame).toBeUndefined()

    const content = payload.content as Array<Record<string, unknown>>
    expect(content).toHaveLength(1)
    expect(content[0]?.text).toBe('A cyber dog playing football')
  })

  test('buildVideoRequestPayload includes image and last frame when enabled', () => {
    const video: VideoForm = {
      ...DEFAULT_VIDEO_FORM,
      prompt: 'Morphing portrait',
      hasImage: true,
      uploadMode: 'url',
      imageUrl: 'https://example.com/first.png',
      role: 'first_frame',
      hasLastFrame: true,
      lastFrameMode: 'url',
      lastFrameUrl: 'https://example.com/last.png',
    }

    const payload = buildVideoRequestPayload('doubao-seedance-pro', video)
    expect(payload).not.toBeNull()
    if (!payload) return
    const content = payload.content as Array<Record<string, unknown>>
    expect(content).toHaveLength(3)
    expect(content[0]?.type).toBe('text')
    expect(content[1]?.type).toBe('image_url')
    expect(content[1]?.role).toBe('first_frame')
    expect((content[1]?.image_url as Record<string, unknown>)?.url).toBe(
      'https://example.com/first.png'
    )
    expect(content[2]?.type).toBe('image_url')
    expect(content[2]?.role).toBe('last_frame')
    expect((content[2]?.image_url as Record<string, unknown>)?.url).toBe(
      'https://example.com/last.png'
    )
  })

  test('parseVideoPayloadToForm extracts prompt, images, parameters and custom extra JSON', () => {
    const raw = JSON.stringify({
      model: 'doubao-seedance-custom',
      content: [
        { type: 'text', text: 'An astronaut riding a dinosaur' },
        {
          type: 'image_url',
          image_url: { url: 'https://example.com/ref.jpg' },
          role: 'reference_image',
        },
        {
          type: 'image_url',
          image_url: { url: 'https://example.com/tail.jpg' },
          role: 'last_frame',
        },
      ],
      resolution: '720p',
      ratio: '4:3',
      duration: 10,
      watermark: true,
      seed: 42,
      generate_audio: true,
      return_last_frame: true,
      extra_experimental_param: 'value_123',
    })

    const res = parseVideoPayloadToForm(raw)
    expect(res.success).toBe(true)
    expect(res.model).toBe('doubao-seedance-custom')
    expect(res.videoPatch?.prompt).toBe('An astronaut riding a dinosaur')
    expect(res.videoPatch?.hasImage).toBe(true)
    expect(res.videoPatch?.uploadMode).toBe('url')
    expect(res.videoPatch?.imageUrl).toBe('https://example.com/ref.jpg')
    expect(res.videoPatch?.role).toBe('reference_image')
    expect(res.videoPatch?.hasLastFrame).toBe(true)
    expect(res.videoPatch?.lastFrameUrl).toBe('https://example.com/tail.jpg')
    expect(res.videoPatch?.hasResolution).toBe(true)
    expect(res.videoPatch?.resolution).toBe('720p')
    expect(res.videoPatch?.hasRatio).toBe(true)
    expect(res.videoPatch?.ratio).toBe('4:3')
    expect(res.videoPatch?.duration).toBe(10)
    expect(res.videoPatch?.watermark).toBe(true)
    expect(res.videoPatch?.seed).toBe('42')
    expect(res.videoPatch?.generateAudio).toBe(true)
    expect(res.videoPatch?.returnLastFrame).toBe(true)
    expect(res.videoPatch?.hasCustomJson).toBe(true)
    expect(res.videoPatch?.customJson).toContain('extra_experimental_param')
  })

  test('buildVideoRequestPayload does not invent a default request body', () => {
    expect(buildVideoRequestPayload('', DEFAULT_VIDEO_FORM)).toBeNull()
    expect(
      buildVideoRequestPayload('doubao-seedance-1-0-pro', DEFAULT_VIDEO_FORM)
    ).toBeNull()
  })

  test('parseVideoPayloadToForm handles invalid JSON gracefully', () => {
    const res = parseVideoPayloadToForm('not a json')
    expect(res.success).toBe(false)
    expect(res.error).toBeDefined()
  })
})
