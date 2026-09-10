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

import {
  createStoragePolicySchema,
  createStorageProfileSchema,
  storagePolicyToInput,
} from '../storage-schemas'
import { STORAGE_AUTH_STATIC, STORAGE_PROVIDER_ALIYUN_OSS } from '../types'

const t = (key: string) => key

describe('storage settings validation', () => {
  test('requires both static credential fields for a new profile', () => {
    const result = createStorageProfileSchema(t, true).safeParse({
      name: 'primary',
      provider_type: STORAGE_PROVIDER_ALIYUN_OSS,
      status: 1,
      endpoint: 'https://oss-cn-hangzhou.aliyuncs.com',
      region: 'oss-cn-hangzhou',
      bucket: 'private-media',
      auth_type: STORAGE_AUTH_STATIC,
      access_key_id: '',
      access_key_secret: '',
      security_token: '',
    })

    expect(result.success).toBe(false)
  })

  test('rejects a policy whose retention is shorter than its signed URL', () => {
    const result = createStoragePolicySchema(t).safeParse({
      enabled: true,
      storage_profile_id: 1,
      object_prefix: 'temporary/relay-media',
      signed_url_ttl_hours: 72,
      retention_hours: 24,
      max_file_mib: 10,
      max_total_mib: 20,
      max_files: 10,
    })

    expect(result.success).toBe(false)
  })

  test('converts display units to the exact API units', () => {
    expect(
      storagePolicyToInput({
        enabled: true,
        storage_profile_id: 7,
        object_prefix: '/temporary/relay-media/',
        signed_url_ttl_hours: 72,
        retention_hours: 96,
        max_file_mib: 10,
        max_total_mib: 20,
        max_files: 8,
      })
    ).toEqual({
      enabled: true,
      storage_profile_id: 7,
      object_prefix: 'temporary/relay-media',
      signed_url_ttl_seconds: 259200,
      retention_seconds: 345600,
      max_file_bytes: 10485760,
      max_total_bytes: 20971520,
      max_files: 8,
      allowed_mime_types:
        'image/jpeg,image/png,image/webp,video/mp4,video/webm,video/quicktime',
    })
  })
})
