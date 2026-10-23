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
  ASSET_PROTOCOL_VOLC_ACTION,
  ASSET_PROTOCOL_YOUFANG_REST,
  STORAGE_AUTH_ENVIRONMENT,
  STORAGE_AUTH_STATIC,
  STORAGE_PROVIDER_ALIYUN_OSS,
  type StoragePolicy,
  type StoragePolicyInput,
  type AssetChannelConfig,
  type AssetChannelConfigInput,
} from './types'

type Translate = (key: string) => string

export const RELAY_MEDIA_ALLOWED_MIME_TYPES =
  'image/jpeg,image/png,image/webp,video/mp4,video/webm,video/quicktime'

export const ASSET_LIBRARY_ALLOWED_MIME_TYPES =
  'image/jpeg,image/png,image/webp,video/mp4,video/webm,video/quicktime,audio/mpeg,audio/wav,audio/x-wav,audio/mp4,audio/aac,audio/ogg'

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

export const createStoragePolicySchema = (t: Translate, imageStudio = false) =>
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
      retention_hours: z
        .number()
        .min(1)
        .max(imageStudio ? 365 * 24 : 720),
      max_file_mib: z
        .number()
        .min(1)
        .max(imageStudio ? 64 : 512),
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
      if (!imageStudio && value.max_total_mib < value.max_file_mib) {
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

export const createAssetLibraryPolicySchema = (t: Translate) =>
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
      max_file_mib: z.number().min(1).max(512),
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
    })

export type AssetLibraryPolicyFormValues = z.infer<
  ReturnType<typeof createAssetLibraryPolicySchema>
>

export function assetLibraryPolicyToForm(
  policy: StoragePolicy
): AssetLibraryPolicyFormValues {
  return {
    enabled: policy.enabled,
    storage_profile_id: policy.storage_profile_id,
    object_prefix: policy.object_prefix,
    signed_url_ttl_hours: policy.signed_url_ttl_seconds / 3600,
    max_file_mib: policy.max_file_bytes / 1024 / 1024,
    max_files: policy.max_files,
  }
}

export function assetLibraryPolicyToInput(
  values: AssetLibraryPolicyFormValues
): StoragePolicyInput {
  return {
    enabled: values.enabled,
    storage_profile_id: values.storage_profile_id,
    object_prefix: values.object_prefix.trim().replaceAll(/^\/+|\/+$/g, ''),
    signed_url_ttl_seconds: Math.round(values.signed_url_ttl_hours * 3600),
    retention_seconds: 0,
    max_file_bytes: Math.round(values.max_file_mib * 1024 * 1024),
    max_total_bytes: Math.round(values.max_file_mib * 1024 * 1024),
    max_files: values.max_files,
    allowed_mime_types: ASSET_LIBRARY_ALLOWED_MIME_TYPES,
  }
}

export const createAssetChannelConfigSchema = (
  t: Translate,
  savedConfig?: Pick<
    AssetChannelConfig,
    'credential_configured' | 'protocol' | 'auth_type'
  >
) =>
  z
    .object({
      enabled: z.boolean(),
      protocol: z.enum([
        ASSET_PROTOCOL_VOLC_ACTION,
        ASSET_PROTOCOL_YOUFANG_REST,
      ]),
      auth_type: z.enum(['ak_sk', 'bearer']),
      base_url: z.string().trim().url(t('Enter a valid upstream URL')),
      region: z.string().trim(),
      service: z.string().trim(),
      api_version: z.string().trim(),
      project_name: z.string().trim(),
      qpm: z.number().int().min(1).max(1000),
      access_key_id: z.string().trim(),
      credential: z.string(),
    })
    .superRefine((value, context) => {
      const canKeepCredential =
        savedConfig?.credential_configured === true &&
        savedConfig.protocol === value.protocol &&
        savedConfig.auth_type === value.auth_type
      if (
        value.protocol === ASSET_PROTOCOL_YOUFANG_REST &&
        value.auth_type !== 'bearer'
      ) {
        context.addIssue({
          code: 'custom',
          path: ['auth_type'],
          message: t('YooFang REST requires a Bearer sk key'),
        })
      }
      if (
        value.protocol === ASSET_PROTOCOL_VOLC_ACTION &&
        (value.region === '' ||
          value.service === '' ||
          value.api_version === '')
      ) {
        context.addIssue({
          code: 'custom',
          path: ['region'],
          message: t('Region, service, and API version are required'),
        })
      }
      if (
        value.protocol === ASSET_PROTOCOL_VOLC_ACTION &&
        value.auth_type === 'ak_sk' &&
        value.access_key_id === '' &&
        !canKeepCredential
      ) {
        context.addIssue({
          code: 'custom',
          path: ['access_key_id'],
          message: t('Access Key ID is required'),
        })
      }
      if (value.credential.trim() === '' && !canKeepCredential) {
        context.addIssue({
          code: 'custom',
          path: ['credential'],
          message: t('Credential is required'),
        })
      }
    })

export type AssetChannelConfigFormValues = z.infer<
  ReturnType<typeof createAssetChannelConfigSchema>
>

export function assetChannelConfigToForm(
  config: AssetChannelConfig
): AssetChannelConfigFormValues {
  return {
    enabled: config.enabled,
    protocol: config.protocol || ASSET_PROTOCOL_VOLC_ACTION,
    auth_type: config.auth_type || 'ak_sk',
    base_url: config.base_url || 'https://ark.cn-beijing.volcengineapi.com',
    region: config.region || 'cn-beijing',
    service: config.service || 'ark',
    api_version: config.api_version || '2024-01-01',
    project_name: config.project_name || 'default',
    qpm: config.qpm || 60,
    access_key_id: '',
    credential: '',
  }
}

export function assetChannelConfigToInput(
  values: AssetChannelConfigFormValues
): AssetChannelConfigInput {
  return {
    ...values,
    base_url: values.base_url.trim(),
    region: values.region.trim(),
    service: values.service.trim(),
    api_version: values.api_version.trim(),
    project_name: values.project_name.trim(),
    access_key_id: values.access_key_id.trim(),
    credential: values.credential.trim(),
  }
}
