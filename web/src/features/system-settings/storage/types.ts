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
export const STORAGE_PROVIDER_ALIYUN_OSS = 'aliyun_oss'
export const STORAGE_AUTH_STATIC = 'static_access_key'
export const STORAGE_AUTH_ENVIRONMENT = 'environment'

export type StorageProfile = {
  id: number
  name: string
  provider_type: string
  status: number
  endpoint: string
  region: string
  bucket: string
  auth_type: string
  credential_configured: boolean
  access_key_hint: string
  created_at: number
  updated_at: number
}

export type StorageProfileInput = {
  name: string
  provider_type: string
  status: number
  endpoint: string
  region: string
  bucket: string
  auth_type: string
  access_key_id: string
  access_key_secret: string
  security_token: string
}

export type StoragePolicy = {
  id: number
  key: string
  name: string
  purpose: string
  storage_profile_id: number
  object_prefix: string
  signed_url_ttl_seconds: number
  retention_seconds: number
  max_file_bytes: number
  max_total_bytes: number
  max_files: number
  allowed_mime_types: string
  enabled: boolean
  created_at: number
  updated_at: number
}

export type StoragePolicyInput = Omit<
  StoragePolicy,
  'id' | 'key' | 'name' | 'purpose' | 'created_at' | 'updated_at'
>

export type StorageAPIResponse<T = undefined> = {
  success: boolean
  message?: string
  data: T
}
