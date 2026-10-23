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
  StorageAPIResponse,
  AssetChannelConfig,
  AssetChannelConfigInput,
  StoragePolicy,
  StoragePolicyInput,
  StorageProfile,
  StorageProfileInput,
  RequestSnapshotJob,
  RequestSnapshotStatus,
} from './types'

export async function getImageStudioStoragePolicies() {
  return (
    await api.get<StorageAPIResponse<StoragePolicy[]>>(
      '/api/storage/policies/image-studio'
    )
  ).data
}

export async function updateImageStudioStoragePolicy(
  purpose: string,
  input: StoragePolicyInput
) {
  return (
    await api.put<StorageAPIResponse<StoragePolicy>>(
      `/api/storage/policies/image-studio/${encodeURIComponent(purpose)}`,
      input,
      { skipBusinessError: true, skipErrorHandler: true }
    )
  ).data
}

export async function listStorageProfiles() {
  const response = await api.get<StorageAPIResponse<StorageProfile[]>>(
    '/api/storage/profiles'
  )
  return response.data
}

export async function createStorageProfile(input: StorageProfileInput) {
  const response = await api.post<StorageAPIResponse<StorageProfile>>(
    '/api/storage/profiles',
    input,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function updateStorageProfile(
  id: number,
  input: StorageProfileInput
) {
  const response = await api.put<StorageAPIResponse<StorageProfile>>(
    `/api/storage/profiles/${id}`,
    input,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function archiveStorageProfile(id: number) {
  const response = await api.delete<StorageAPIResponse>(
    `/api/storage/profiles/${id}`,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function testStorageProfile(input: StorageProfileInput) {
  const response = await api.post<StorageAPIResponse>(
    '/api/storage/profiles/test',
    input,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function testSavedStorageProfile(id: number) {
  const response = await api.post<StorageAPIResponse>(
    `/api/storage/profiles/${id}/test`,
    undefined,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function getRelayMediaStoragePolicy() {
  const response = await api.get<StorageAPIResponse<StoragePolicy>>(
    '/api/storage/policies/relay-media-temp'
  )
  return response.data
}

export async function updateRelayMediaStoragePolicy(input: StoragePolicyInput) {
  const response = await api.put<StorageAPIResponse<StoragePolicy>>(
    '/api/storage/policies/relay-media-temp',
    input,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function getRequestSnapshotStoragePolicy() {
  const response = await api.get<StorageAPIResponse<StoragePolicy>>(
    '/api/storage/policies/request-snapshots'
  )
  return response.data
}

export async function updateRequestSnapshotStoragePolicy(
  input: StoragePolicyInput
) {
  const response = await api.put<StorageAPIResponse<StoragePolicy>>(
    '/api/storage/policies/request-snapshots',
    input,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function getRequestSnapshotStatus() {
  const response = await api.get<StorageAPIResponse<RequestSnapshotStatus>>(
    '/api/storage/request-snapshots/status',
    { disableDuplicate: true, skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function startRequestSnapshotMigration() {
  const response = await api.post<StorageAPIResponse<RequestSnapshotJob>>(
    '/api/storage/request-snapshots/migrate',
    { batch_size: 50, clear_database_body: false },
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function startRequestSnapshotCleanup() {
  const response = await api.post<StorageAPIResponse<RequestSnapshotJob>>(
    '/api/storage/request-snapshots/cleanup',
    { batch_size: 50 },
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function getAssetLibraryStoragePolicy() {
  const response = await api.get<StorageAPIResponse<StoragePolicy>>(
    '/api/storage/policies/asset-library'
  )
  return response.data
}

export async function updateAssetLibraryStoragePolicy(
  input: StoragePolicyInput
) {
  const response = await api.put<StorageAPIResponse<StoragePolicy>>(
    '/api/storage/policies/asset-library',
    input,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function listAssetChannelConfigs() {
  const response = await api.get<StorageAPIResponse<AssetChannelConfig[]>>(
    '/api/asset-library/admin/channels'
  )
  return response.data
}

export async function updateAssetChannelConfig(
  channelId: number,
  input: AssetChannelConfigInput
) {
  const response = await api.put<StorageAPIResponse<AssetChannelConfig>>(
    `/api/asset-library/admin/channels/${channelId}`,
    input,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}

export async function testAssetChannelConfig(
  channelId: number,
  input?: AssetChannelConfigInput
) {
  const response = await api.post<StorageAPIResponse>(
    `/api/asset-library/admin/channels/${channelId}/test`,
    input,
    { skipBusinessError: true, skipErrorHandler: true }
  )
  return response.data
}
