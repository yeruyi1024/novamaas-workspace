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
import { createFileRoute } from '@tanstack/react-router'
import { useCallback } from 'react'
import { z } from 'zod'

import { AssetLibrary } from '@/features/assets'
import type { AssetLibraryFilters } from '@/features/assets/types'

const assetLibrarySearchSchema = z.object({
  group: z.string().trim().min(1).max(64).optional().catch(undefined),
  owner: z.coerce.number().int().positive().optional().catch(undefined),
  groupOwner: z.coerce.number().int().positive().optional().catch(undefined),
  type: z.enum(['image', 'video', 'audio']).optional().catch(undefined),
  status: z
    .enum(['ready', 'processing', 'unavailable'])
    .optional()
    .catch(undefined),
  q: z.string().trim().min(1).max(200).optional().catch(undefined),
  sort: z.enum(['newest', 'oldest']).optional().catch(undefined),
  page: z.coerce.number().int().min(2).optional().catch(undefined),
})

export const Route = createFileRoute('/_authenticated/assets/')({
  validateSearch: assetLibrarySearchSchema,
  component: AssetLibraryRoute,
})

function AssetLibraryRoute() {
  const filters = Route.useSearch()
  const navigate = Route.useNavigate()
  const changeFilters = useCallback(
    (patch: Partial<AssetLibraryFilters>) =>
      navigate({
        replace: true,
        search: (previous) => {
          const next: AssetLibraryFilters = { ...previous, ...patch }
          if (next.sort === 'newest') next.sort = undefined
          if (next.page !== undefined && next.page < 2) next.page = undefined
          return next
        },
      }),
    [navigate]
  )
  return <AssetLibrary filters={filters} onFiltersChange={changeFilters} />
}
