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
import { api } from '@/lib/api'

import type {
  AssetAccessKey,
  AssetCopyInput,
  AssetCopyResult,
  AssetFilterDeleteInput,
  AssetFilterDeleteResult,
  AssetGroup,
  AssetGroupList,
  AssetGroupPurgeResult,
  AssetLibraryResponse,
  AssetOwnerList,
  AssetPreviewMap,
  AssetRef,
  AssetRequestLogList,
  AssetRequestLogWithDetail,
  AssetSortOrder,
  AssetStatusFilter,
  AssetSummary,
  AssetSyncJobList,
  CreatedAssetAccessKey,
  MediaAsset,
  MediaAssetList,
  MediaAssetType,
  UserAssetPurgeResult,
} from './types'

export async function listAssetAccessKeys() {
  const response = await api.get<AssetLibraryResponse<AssetAccessKey[]>>(
    '/api/asset-library/access-keys'
  )
  return response.data
}

export async function createAssetAccessKey(name: string) {
  const response = await api.post<AssetLibraryResponse<CreatedAssetAccessKey>>(
    '/api/asset-library/access-keys',
    { name },
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function deleteAssetAccessKey(id: number) {
  const response = await api.delete<AssetLibraryResponse>(
    `/api/asset-library/access-keys/${id}`,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function listAssetGroups(includeAllOwners = false) {
  const response = await api.get<AssetLibraryResponse<AssetGroup[]>>(
    '/api/asset-library/groups',
    { params: includeAllOwners ? { scope: 'all' } : undefined }
  )
  return response.data
}

export async function listAssetGroupsPage(params: {
  includeAllOwners: boolean
  ownerUserId?: number
  groupId?: string
  page: number
  pageSize: number
  search: string
}) {
  const response = await api.get<AssetLibraryResponse<AssetGroupList>>(
    '/api/asset-library/groups',
    {
      params: {
        p: params.page,
        page_size: params.pageSize,
        search: params.search || undefined,
        scope: params.includeAllOwners ? 'all' : undefined,
        owner_user_id: params.ownerUserId,
        group_id: params.groupId || undefined,
      },
    }
  )
  return response.data
}

export async function copyAssetGroups(input: AssetCopyInput) {
  const response = await api.post<AssetLibraryResponse<AssetCopyResult>>(
    '/api/asset-library/admin/copies',
    input,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function getAssetCopy(id: string) {
  const response = await api.get<AssetLibraryResponse<AssetCopyResult>>(
    `/api/asset-library/admin/copies/${encodeURIComponent(id)}`,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function deleteAssetCopySources(id: string) {
  const response = await api.post<AssetLibraryResponse<AssetCopyResult>>(
    `/api/asset-library/admin/copies/${encodeURIComponent(id)}/delete-sources`,
    undefined,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function purgeUserAssets(userId: number) {
  const response = await api.post<AssetLibraryResponse<UserAssetPurgeResult>>(
    `/api/asset-library/admin/users/${userId}/purge`,
    undefined,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function createAssetGroup(input: {
  name: string
  description: string
}) {
  const response = await api.post<AssetLibraryResponse<AssetGroup>>(
    '/api/asset-library/groups',
    input,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function deleteAssetGroup(
  id: string,
  ownerUserId?: number,
  options?: { withAssets?: boolean }
) {
  const response = await api.delete<
    AssetLibraryResponse<AssetGroupPurgeResult | undefined>
  >(`/api/asset-library/groups/${encodeURIComponent(id)}`, {
    params: {
      owner_user_id: ownerUserId || undefined,
      with_assets: options?.withAssets ? 'true' : undefined,
    },
    skipBusinessError: true,
    skipErrorHandler: true,
  })
  return response.data
}

export async function updateAssetGroup(
  id: string,
  input: { name: string; description: string },
  ownerUserId?: number
) {
  const response = await api.put<AssetLibraryResponse<AssetGroup>>(
    `/api/asset-library/groups/${encodeURIComponent(id)}`,
    input,
    {
      params: ownerUserId ? { owner_user_id: ownerUserId } : undefined,
      skipBusinessError: true,
      skipErrorHandler: true,
    }
  )
  return response.data
}

type AssetScopeParams = {
  groupId?: string
  ownerUserId?: number
  search?: string
  includeAllOwners?: boolean
}

function assetScopeQuery(params: AssetScopeParams) {
  return {
    group_id: params.groupId || undefined,
    owner_user_id: params.ownerUserId || undefined,
    search: params.search || undefined,
    scope: params.includeAllOwners ? 'all' : undefined,
  }
}

export async function listMediaAssets(
  params: AssetScopeParams & {
    type?: MediaAssetType
    status?: AssetStatusFilter
    sort?: AssetSortOrder
    page: number
    pageSize: number
  }
) {
  const response = await api.get<AssetLibraryResponse<MediaAssetList>>(
    '/api/asset-library/assets',
    {
      params: {
        ...assetScopeQuery(params),
        type: params.type || undefined,
        status: params.status || undefined,
        sort: params.sort === 'oldest' ? 'oldest' : undefined,
        p: params.page,
        page_size: params.pageSize,
      },
    }
  )
  return response.data
}

export async function summarizeMediaAssets(params: AssetScopeParams) {
  const response = await api.get<AssetLibraryResponse<AssetSummary>>(
    '/api/asset-library/assets/summary',
    { params: assetScopeQuery(params) }
  )
  return response.data
}

export async function listAssetOwners(params: {
  search: string
  page: number
  pageSize: number
}) {
  const response = await api.get<AssetLibraryResponse<AssetOwnerList>>(
    '/api/asset-library/admin/owners',
    {
      params: {
        search: params.search || undefined,
        p: params.page,
        page_size: params.pageSize,
      },
    }
  )
  return response.data
}

export async function previewMediaAssets(items: AssetRef[]) {
  const response = await api.post<AssetLibraryResponse<AssetPreviewMap>>(
    '/api/asset-library/assets/previews',
    { items },
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function deleteMediaAssets(items: AssetRef[]) {
  const response = await api.post<
    AssetLibraryResponse<{ deleted_count: number }>
  >(
    '/api/asset-library/assets/batch-delete',
    { items },
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function deleteMediaAssetsByFilter(input: AssetFilterDeleteInput) {
  const response = await api.post<
    AssetLibraryResponse<AssetFilterDeleteResult>
  >('/api/asset-library/assets/batch-delete-by-filter', input, {
    skipBusinessError: true,
    skipErrorHandler: true,
  })
  return response.data
}

export async function uploadMediaAsset(
  formData: FormData,
  onProgress?: (value: number) => void
) {
  const response = await api.post<AssetLibraryResponse<MediaAsset>>(
    '/api/asset-library/assets',
    formData,
    {
      skipBusinessError: true,
      skipErrorHandler: true,
      onUploadProgress: (event) => {
        if (!event.total) return
        onProgress?.(
          Math.min(100, Math.round((event.loaded / event.total) * 100))
        )
      },
    }
  )
  return response.data
}

export async function getMediaAssetPreview(
  id: string,
  variant: 'thumbnail' | 'original' | 'download' = 'original',
  ownerUserId?: number
) {
  const response = await api.get<
    AssetLibraryResponse<{ url: string; expires_at: number }>
  >(`/api/asset-library/assets/${encodeURIComponent(id)}/preview`, {
    params: {
      variant,
      owner_user_id: ownerUserId || undefined,
    },
    skipBusinessError: true,
    skipErrorHandler: true,
  })
  return response.data
}

export async function deleteMediaAsset(id: string, ownerUserId?: number) {
  const response = await api.delete<AssetLibraryResponse>(
    `/api/asset-library/assets/${encodeURIComponent(id)}`,
    {
      params: ownerUserId ? { owner_user_id: ownerUserId } : undefined,
      skipBusinessError: true,
      skipErrorHandler: true,
    }
  )
  return response.data
}

export async function listAssetSyncJobs(params: {
  page: number
  pageSize: number
  channelId?: number
  libraryId?: number
  status?: string
}) {
  const response = await api.get<AssetLibraryResponse<AssetSyncJobList>>(
    '/api/asset-library/admin/sync-jobs',
    {
      params: {
        p: params.page,
        page_size: params.pageSize,
        channel_id: params.channelId || undefined,
        library_id: params.libraryId || undefined,
        status: params.status || undefined,
      },
    }
  )
  return response.data
}

export async function retryAssetSyncJob(id: number) {
  const response = await api.post<AssetLibraryResponse>(
    `/api/asset-library/admin/sync-jobs/${id}/retry`,
    undefined,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function listAssetRequestLogs(params: {
  channelId?: number
  libraryId?: number
  replicaId?: number
  source?: string
  result?: string
  requestId?: string
  cursor?: string
  pageSize?: number
  startMS?: number
  endMS?: number
}) {
  const response = await api.get<AssetLibraryResponse<AssetRequestLogList>>(
    '/api/asset-library/admin/request-logs',
    {
      params: {
        channel_id: params.channelId || undefined,
        library_id: params.libraryId || undefined,
        replica_id: params.replicaId || undefined,
        source: params.source || undefined,
        result: params.result || undefined,
        request_id: params.requestId || undefined,
        cursor: params.cursor || undefined,
        page_size: params.pageSize || 50,
        start_ms: params.startMS,
        end_ms: params.endMS,
      },
    }
  )
  return response.data
}

export async function getAssetRequestLogDetail(id: number) {
  const response = await api.get<
    AssetLibraryResponse<AssetRequestLogWithDetail>
  >(`/api/asset-library/admin/request-logs/${id}`)
  return response.data
}
