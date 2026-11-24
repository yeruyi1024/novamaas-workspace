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
export type AssetGroup = {
  id: string
  owner_user_id: number
  owner_name?: string
  asset_count?: number
  name: string
  description: string
  status: string
  created_at: number
  updated_at: number
}

export type AssetGroupList = {
  items: AssetGroup[]
  total: number
  page: number
  page_size: number
}

export type MediaAsset = {
  id: string
  group_id: string
  group_name?: string
  owner_user_id: number
  owner_name: string
  name: string
  type: 'image' | 'video' | 'audio'
  content_type: string
  size: number
  sha256: string
  status: string
  unavailable_reason?: 'real_person' | 'sensitive_content' | 'policy_rejected'
  thumbnail_url?: string
  poster_url?: string
  thumbnail_expires_at?: number
  created_at: number
  updated_at: number
}

export type MediaAssetList = {
  items: MediaAsset[]
  total: number
  page: number
  page_size: number
}

export type MediaAssetType = MediaAsset['type']

export type AssetStatusFilter = 'ready' | 'processing' | 'unavailable'

export type AssetSortOrder = 'newest' | 'oldest'

export type AssetLibraryFilters = {
  group?: string
  owner?: number
  /** Owner of the chosen group; group IDs are only unique per owner. */
  groupOwner?: number
  type?: MediaAssetType
  status?: AssetStatusFilter
  q?: string
  sort?: AssetSortOrder
  page?: number
}

export type AssetSummary = {
  total: number
  by_type: Record<MediaAssetType, number>
  by_status: Record<AssetStatusFilter, number>
}

export type AssetRef = { id: string; owner_user_id: number }

export type AssetOwner = {
  user_id: number
  username: string
  group_count: number
  asset_count: number
}

export type AssetOwnerList = {
  items: AssetOwner[]
  total: number
  page: number
  page_size: number
}

export type AssetPreviewMap = Record<
  string,
  { url: string; poster_url?: string; expires_at: number }
>

export type AssetFilterDeleteInput = {
  scope?: 'all'
  owner_user_id?: number
  group_id?: string
  type?: MediaAssetType
  status?: AssetStatusFilter
  search?: string
  exclude?: AssetRef[]
}

export type AssetFilterDeleteResult = {
  deleted_count: number
  owner_user_ids: number[]
}

export type AssetGroupPurgeResult = {
  owner_user_id: number
  deleted_asset_count: number
  deleted_group_count: number
}

export type AssetAccessKey = {
  id: number
  name: string
  access_key_id: string
  secret_hint: string
  status: string
  last_used_at: number
  created_at: number
}

export type CreatedAssetAccessKey = AssetAccessKey & {
  secret_access_key: string
}

export type AssetSyncJob = {
  library_id?: number
  library_name?: string
  channels?: { id: number; name: string; enabled: boolean }[]
  upstream_asset_id?: string
  id: number
  asset_id: string
  asset_name: string
  asset_type: string
  group_id: string
  group_name: string
  owner_user_id: number
  owner_name: string
  channel_id: number
  channel_name: string
  operation: string
  status: string
  progress: number
  attempts: number
  last_error?: string
  last_synced_at: number
  updated_at: number
}

export type AssetSyncSummary = {
  total: number
  pending: number
  processing: number
  active: number
  failed: number
}

export type AssetSyncJobList = {
  items: AssetSyncJob[]
  total: number
  page: number
  page_size: number
  summary: AssetSyncSummary
}

export type AssetRequestLog = {
  library_id?: number
  id: number
  created_at: number
  channel_id: number
  replica_id: number
  asset_id: number
  request_id: string
  source: 'test' | 'sync'
  protocol: string
  operation: string
  method: string
  path_template: string
  http_status: number
  duration_ms: number
  request_bytes: number
  response_bytes: number
  result: 'success' | 'failure'
  error_kind: string
}

export type AssetRequestLogList = {
  items: AssetRequestLog[]
  next_cursor: string
  dropped_on_this_node: number
}

export type AssetRequestLogDetail = {
  log_id: number
  request_url: string
  request_body: string
  response_body: string
  request_body_omitted: boolean
  response_body_omitted: boolean
}

export type AssetRequestLogWithDetail = {
  log: AssetRequestLog
  detail: AssetRequestLogDetail | null
}

export type AssetLibraryResponse<T = undefined> = {
  success: boolean
  message?: string
  data: T
}

export type AssetCopyInput = {
  request_id: string
  source_user_id: number
  target_user_id: number
  group_ids: string[]
}

export type AssetCopyResult = {
  id: string
  source_user_id: number
  target_user_id: number
  ready_count: number
  processing_count: number
  failed_count: number
  unavailable_count: number
  can_delete_sources: boolean
  groups: { id: string; name: string; asset_count: number }[]
  asset_count: number
  sources_deleted_at: number
}

export type UserAssetPurgeResult = {
  owner_user_id: number
  deleted_asset_count: number
  deleted_group_count: number
}
