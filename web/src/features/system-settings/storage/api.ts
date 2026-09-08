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
  StoragePolicy,
  StoragePolicyInput,
  StorageProfile,
  StorageProfileInput,
} from './types'

export async function listStorageProfiles() {
  const response = await api.get<StorageAPIResponse<StorageProfile[]>>(
    '/api/storage/profiles'
  )
  return response.data
}

export async function createStorageProfile(input: StorageProfileInput) {
  const response = await api.post<StorageAPIResponse<StorageProfile>>(
    '/api/storage/profiles',
    input
  )
  return response.data
}

export async function updateStorageProfile(
  id: number,
  input: StorageProfileInput
) {
  const response = await api.put<StorageAPIResponse<StorageProfile>>(
    `/api/storage/profiles/${id}`,
    input
  )
  return response.data
}

export async function archiveStorageProfile(id: number) {
  const response = await api.delete<StorageAPIResponse>(
    `/api/storage/profiles/${id}`
  )
  return response.data
}

export async function testStorageProfile(input: StorageProfileInput) {
  const response = await api.post<StorageAPIResponse>(
    '/api/storage/profiles/test',
    input
  )
  return response.data
}

export async function testSavedStorageProfile(id: number) {
  const response = await api.post<StorageAPIResponse>(
    `/api/storage/profiles/${id}/test`
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
    input
  )
  return response.data
}
