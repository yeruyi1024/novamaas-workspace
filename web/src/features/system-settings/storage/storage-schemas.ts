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
import * as z from 'zod'

import {
  STORAGE_AUTH_ENVIRONMENT,
  STORAGE_AUTH_STATIC,
  STORAGE_PROVIDER_ALIYUN_OSS,
  type StoragePolicy,
  type StoragePolicyInput,
} from './types'

type Translate = (key: string) => string

export const RELAY_MEDIA_ALLOWED_MIME_TYPES =
  'image/jpeg,image/png,image/webp,video/mp4,video/webm,video/quicktime'

export const createStorageProfileSchema = (
  t: Translate,
  requireStaticCredentials = false
) =>
  z
    .object({
      name: z.string().trim().min(1, t('Storage profile name is required')),
      provider_type: z.literal(STORAGE_PROVIDER_ALIYUN_OSS),
      status: z.number().int().min(0).max(1),
      endpoint: z
        .string()
        .trim()
        .url(t('Enter a valid HTTPS endpoint'))
        .refine(
          (value) => value.startsWith('https://'),
          t('Enter a valid HTTPS endpoint')
        ),
      region: z.string().trim().min(1, t('Region is required')),
      bucket: z.string().trim().min(1, t('Bucket is required')),
      auth_type: z.enum([STORAGE_AUTH_STATIC, STORAGE_AUTH_ENVIRONMENT]),
      access_key_id: z.string(),
      access_key_secret: z.string(),
      security_token: z.string(),
    })
    .superRefine((value, context) => {
      if (value.auth_type !== STORAGE_AUTH_STATIC) return
      const hasAccessKey = value.access_key_id.trim() !== ''
      const hasSecret = value.access_key_secret.trim() !== ''
      if (requireStaticCredentials && !hasAccessKey && !hasSecret) {
        context.addIssue({
          code: 'custom',
          path: ['access_key_id'],
          message: t('Access Key ID and secret are required'),
        })
        return
      }
      if (hasAccessKey === hasSecret) return
      context.addIssue({
        code: 'custom',
        path: [hasAccessKey ? 'access_key_secret' : 'access_key_id'],
        message: t('Access Key ID and secret must be provided together'),
      })
    })

export type StorageProfileFormValues = z.infer<
  ReturnType<typeof createStorageProfileSchema>
>

export const createStoragePolicySchema = (t: Translate) =>
  z
    .object({
      enabled: z.boolean(),
      storage_profile_id: z.number().int().min(0),
      object_prefix: z
        .string()
        .trim()
        .min(1, t('Object prefix is required'))
        .refine(
          (value) => !value.includes('..') && !value.includes('\\'),
          t('Object prefix contains invalid characters')
        ),
      signed_url_ttl_hours: z.number().min(1).max(168),
      retention_hours: z.number().min(1).max(720),
      max_file_mib: z.number().min(1).max(512),
      max_total_mib: z.number().min(1).max(512),
      max_files: z.number().int().min(1).max(100),
    })
    .superRefine((value, context) => {
      if (value.enabled && value.storage_profile_id === 0) {
        context.addIssue({
          code: 'custom',
          path: ['storage_profile_id'],
          message: t('Select an enabled storage profile'),
        })
      }
      if (value.retention_hours < value.signed_url_ttl_hours) {
        context.addIssue({
          code: 'custom',
          path: ['retention_hours'],
          message: t('Retention must be at least the signed URL lifetime'),
        })
      }
      if (value.max_total_mib < value.max_file_mib) {
        context.addIssue({
          code: 'custom',
          path: ['max_total_mib'],
          message: t('Total size must be at least the per-file size'),
        })
      }
    })

export type StoragePolicyFormValues = z.infer<
  ReturnType<typeof createStoragePolicySchema>
>

export function storagePolicyToForm(
  policy: StoragePolicy
): StoragePolicyFormValues {
  return {
    enabled: policy.enabled,
    storage_profile_id: policy.storage_profile_id,
    object_prefix: policy.object_prefix,
    signed_url_ttl_hours: policy.signed_url_ttl_seconds / 3600,
    retention_hours: policy.retention_seconds / 3600,
    max_file_mib: policy.max_file_bytes / 1024 / 1024,
    max_total_mib: policy.max_total_bytes / 1024 / 1024,
    max_files: policy.max_files,
  }
}

export function storagePolicyToInput(
  values: StoragePolicyFormValues
): StoragePolicyInput {
  return {
    enabled: values.enabled,
    storage_profile_id: values.storage_profile_id,
    object_prefix: values.object_prefix.trim().replaceAll(/^\/+|\/+$/g, ''),
    signed_url_ttl_seconds: Math.round(values.signed_url_ttl_hours * 3600),
    retention_seconds: Math.round(values.retention_hours * 3600),
    max_file_bytes: Math.round(values.max_file_mib * 1024 * 1024),
    max_total_bytes: Math.round(values.max_total_mib * 1024 * 1024),
    max_files: values.max_files,
    allowed_mime_types: RELAY_MEDIA_ALLOWED_MIME_TYPES,
  }
}
