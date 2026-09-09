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
import { beforeEach, describe, expect, test } from 'vitest'

import { initializeFrontendCache } from './frontend-cache'

describe('initializeFrontendCache', () => {
  beforeEach(() => {
    window.localStorage.clear()
  })

  test('clears stale UI cache while preserving login state', () => {
    window.localStorage.setItem('newapi:default:cache-version', 'default-v1')
    window.localStorage.setItem('status', '{"system_name":"stale"}')
    window.localStorage.setItem(
      'usage-logs:common:user:column-visibility',
      '{}'
    )
    window.localStorage.setItem('user', '{"id":1}')
    window.localStorage.setItem('uid', '1')

    initializeFrontendCache()

    expect(window.localStorage.getItem('newapi:default:cache-version')).toBe(
      'default-v2'
    )
    expect(window.localStorage.getItem('status')).toBeNull()
    expect(
      window.localStorage.getItem('usage-logs:common:user:column-visibility')
    ).toBeNull()
    expect(window.localStorage.getItem('user')).toBe('{"id":1}')
    expect(window.localStorage.getItem('uid')).toBe('1')
  })
})
