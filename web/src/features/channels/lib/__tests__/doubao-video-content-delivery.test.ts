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

import { CHANNEL_TYPE_DOUBAO_VIDEO } from '../../constants'
import type { Channel } from '../../types'
import {
  CHANNEL_FORM_DEFAULT_VALUES,
  channelFormSchema,
  transformChannelToFormDefaults,
  transformFormDataToUpdatePayload,
} from '../channel-form'

function doubaoChannel(settings: string): Channel {
  return {
    id: 54,
    type: CHANNEL_TYPE_DOUBAO_VIDEO,
    key: '',
    status: 1,
    name: 'doubao-video',
    created_time: 0,
    test_time: 0,
    response_time: 0,
    balance: 0,
    balance_updated_time: 0,
    models: 'doubao-seedance-1-0-pro-250528',
    group: 'default',
    used_quota: 0,
    other: '',
    other_info: '',
    remark: '',
    max_input_tokens: 0,
    channel_info: {
      is_multi_key: false,
      multi_key_size: 0,
      multi_key_polling_index: 0,
      multi_key_mode: 'random',
    },
    settings,
  }
}

describe('DoubaoVideo content delivery settings', () => {
  test('defaults existing channels to server proxy', () => {
    const defaults = transformChannelToFormDefaults(doubaoChannel('{}'))

    expect(defaults.video_content_delivery_mode).toBe('proxy')
  })

  test('loads and persists redirect mode for DoubaoVideo', () => {
    const defaults = transformChannelToFormDefaults(
      doubaoChannel('{"video_content_delivery_mode":"redirect"}')
    )
    const payload = transformFormDataToUpdatePayload(defaults, 54)

    expect(defaults.video_content_delivery_mode).toBe('redirect')
    expect(JSON.parse(payload.settings || '{}')).toMatchObject({
      video_content_delivery_mode: 'redirect',
    })
  })

  test('removes DoubaoVideo delivery mode after changing the channel type', () => {
    const defaults = transformChannelToFormDefaults(
      doubaoChannel('{"video_content_delivery_mode":"redirect"}')
    )
    const payload = transformFormDataToUpdatePayload(
      { ...defaults, type: 1 },
      54
    )

    expect(JSON.parse(payload.settings || '{}')).not.toHaveProperty(
      'video_content_delivery_mode'
    )
  })

  test('rejects unsupported delivery modes', () => {
    const result = channelFormSchema.safeParse({
      ...CHANNEL_FORM_DEFAULT_VALUES,
      name: 'doubao-video',
      type: CHANNEL_TYPE_DOUBAO_VIDEO,
      key: 'test-key',
      models: 'doubao-seedance-1-0-pro-250528',
      video_content_delivery_mode: 'download',
    })

    expect(result.success).toBe(false)
  })
})
