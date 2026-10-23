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
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";

import { ROLE } from "@/lib/roles";
import { useAuthStore } from "@/stores/auth-store";

import { adminManualAccessOptions, adminManualDocumentOptions } from "./api";

export function useAdminManualAccess() {
  const userID = useAuthStore((state) => state.auth.user?.id ?? null);
  const role = useAuthStore((state) => state.auth.user?.role ?? 0);
  const sessionID = useAuthStore((state) => state.auth.session?.sid ?? null);
  const hasToken = useAuthStore((state) => !!state.auth.accessToken);
  const eligible = hasToken && role >= ROLE.ADMIN;
  const client = useQueryClient();
  const access = useQuery({
    ...adminManualAccessOptions(userID, sessionID, role),
    enabled: eligible,
  });

  useEffect(() => {
    if (eligible && access.data !== false && !access.isError) return;
    const queryKey = adminManualDocumentOptions(
      userID,
      sessionID,
      role,
    ).queryKey;
    void client.cancelQueries({ queryKey }).then(() => {
      client.removeQueries({ queryKey });
    });
  }, [eligible, access.data, access.isError, client, userID, sessionID, role]);

  return {
    available: eligible && access.data === true && !access.isError,
    pending: eligible && access.isPending,
    userID,
    sessionID,
    role,
  };
}
