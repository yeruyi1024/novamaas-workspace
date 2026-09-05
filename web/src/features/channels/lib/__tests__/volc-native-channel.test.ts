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

import { CHANNEL_TYPE_OPTIONS, CHANNEL_TYPE_VOLC_NATIVE } from '../../constants'
import { getChannelTypeConfig, getDefaultBaseUrl } from '../channel-type-config'
import { getChannelTypeIcon, getKeyPromptForType } from '../channel-utils'

describe('Volc Native channel', () => {
  test('registers the native channel with the Fire Ark connection defaults', () => {
    expect(
      CHANNEL_TYPE_OPTIONS.find(
        (item) => item.value === CHANNEL_TYPE_VOLC_NATIVE
      )
    ).toEqual({ value: CHANNEL_TYPE_VOLC_NATIVE, label: 'Volc Native' })
    expect(getChannelTypeIcon(CHANNEL_TYPE_VOLC_NATIVE)).toBe('Volcengine')
    expect(getKeyPromptForType(CHANNEL_TYPE_VOLC_NATIVE)).toBe(
      'Volcengine Ark API Key'
    )
    expect(getChannelTypeConfig(CHANNEL_TYPE_VOLC_NATIVE).icon).toBe(
      'volcengine'
    )
    expect(getDefaultBaseUrl(CHANNEL_TYPE_VOLC_NATIVE)).toBe(
      'https://ark.cn-beijing.volces.com'
    )
  })
})
