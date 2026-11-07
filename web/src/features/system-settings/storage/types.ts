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
export const ASSET_PROTOCOL_VOLC_ACTION = 'volc_action'
export const ASSET_PROTOCOL_YOUFANG_REST = 'youfang_rest'

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

export type RequestSnapshotLogCursor = {
  ID: number
  CreatedAt: number
  RequestID: string
  OtherSHA256: string
}

export type RequestSnapshotJob = {
  status: 'idle' | 'running' | 'completed' | 'failed'
  mode: 'migrate' | 'cleanup'
  cursor: number
  processed: number
  failed: number
  error?: string
  phase?: string
  initialized?: boolean
  task_cursor?: number
  task_upper_id?: number
  log_cursor?: RequestSnapshotLogCursor
  log_upper_cursor?: RequestSnapshotLogCursor
  archive_upper_id?: number
  scanned?: number
  updated_at?: number
  active_files?: number
  workers?: number
}

export type RequestSnapshotStatus = {
  node_name: string
  spool_directory: string
  workers: number
  pending_files: number
  pending_bytes: number
  last_error?: string
  migration: RequestSnapshotJob
}

export type AssetChannelConfig = {
  library_id?: number
  library_name?: string
  library_enabled?: boolean
  channel_id: number
  channel_name: string
  channel_type: number
  enabled: boolean
  protocol: 'volc_action' | 'youfang_rest'
  auth_type: 'ak_sk' | 'bearer'
  base_url: string
  region: string
  service: string
  api_version: string
  project_name: string
  qpm: number
  access_key_hint: string
  credential_configured: boolean
  updated_at: number
}

export type AssetChannelConfigInput = {
  library_id?: number
  enabled: boolean
  protocol: 'volc_action' | 'youfang_rest'
  auth_type: 'ak_sk' | 'bearer'
  base_url: string
  region: string
  service: string
  api_version: string
  project_name: string
  qpm: number
  access_key_id: string
  credential: string
}

export type AssetLibraryChannel = {
  id: number
  name: string
  enabled: boolean
}

export type AssetLibraryConnection = AssetChannelConfig & {
  id: number
  name: string
  provider_name: string
  legacy_channel_id?: number | null
  reused_existing?: boolean
  credential_status?: 'available' | 'unavailable'
  channels: AssetLibraryChannel[]
  summary: {
    total: number
    pending: number
    processing: number
    active: number
    failed: number
  }
  missing: number
}

export type AssetLibraryConnectionInput = AssetChannelConfigInput & {
  name: string
  provider_name: string
}

export type AssetLibraryMergeReport = {
  merged_groups: number
  deferred_groups: number
  unreadable_accounts: number
}
