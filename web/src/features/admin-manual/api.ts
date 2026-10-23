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
import { queryOptions } from "@tanstack/react-query";

import { api, getFreshAuthHeaders } from "@/lib/api";

export type ManualDocument = { id: string; title: string; content: string };

export function adminManualAccessOptions(
  userID: number | null,
  sessionID: string | null,
  role: number,
) {
  return queryOptions({
    queryKey: ["admin-manual", "access", userID, sessionID, role],
    queryFn: async ({ signal }) => {
      try {
        const headers = await getFreshAuthHeaders();
        const response = await api.get<{
          success: boolean;
          data?: { available?: boolean };
        }>("/api/admin-manual/access", {
          signal,
          headers,
          // React Query owns cancellation/deduplication for this query key.
          // HTTP-level reuse can return a request aborted by a prior mount.
          disableDuplicate: true,
          skipErrorHandler: true,
        });
        return response.data.success && response.data.data?.available === true;
      } catch {
        return false;
      }
    },
    staleTime: 0,
    gcTime: 0,
    retry: false,
    refetchOnWindowFocus: "always",
  });
}

export function adminManualDocumentOptions(
  userID: number | null,
  sessionID: string | null,
  role: number,
) {
  return queryOptions({
    queryKey: ["admin-manual", "documents", userID, sessionID, role],
    queryFn: async ({ signal }) => {
      try {
        const headers = await getFreshAuthHeaders();
        const response = await api.get<{
          success: boolean;
          data: ManualDocument[];
        }>("/api/admin-manual", {
          signal,
          headers,
          disableDuplicate: true,
          skipErrorHandler: true,
        });
        if (response.data.success && Array.isArray(response.data.data)) {
          return response.data.data;
        }
      } catch {
        // Replace previously delivered text on any failed revalidation.
      }
      return [];
    },
    staleTime: 0,
    gcTime: 0,
    retry: false,
    refetchOnWindowFocus: "always",
  });
}
