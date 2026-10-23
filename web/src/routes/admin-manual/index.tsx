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
import { createFileRoute, notFound } from "@tanstack/react-router";
import { z } from "zod";

import { AdminManual } from "@/features/admin-manual";
import { ROLE } from "@/lib/roles";
import { useAuthStore } from "@/stores/auth-store";

export const Route = createFileRoute("/admin-manual/")({
  validateSearch: z.object({
    article: z.string().max(80).optional().catch(undefined),
  }),
  beforeLoad: () => {
    const { auth } = useAuthStore.getState();
    if (!auth.accessToken || !auth.user || auth.user.role < ROLE.ADMIN) {
      throw notFound();
    }
  },
  component: AdminManual,
});
