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
import { useQuery } from '@tanstack/react-query'
import { useRouteContext, useSearch } from '@tanstack/react-router'

import { api } from '@/lib/api'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import { DocumentationReader, type APIDocument } from './reader'

export function APIDocs() {
  const { section, article } = useSearch({ from: '/docs/' })
  const { access } = useRouteContext({ from: '/docs/' })
  const userID = useAuthStore((state) => state.auth.user?.id)
  const sessionID = useAuthStore((state) => state.auth.session?.sid)
  const role = useAuthStore((state) => state.auth.user?.role ?? 0)
  const hasToken = useAuthStore((state) => !!state.auth.accessToken)
  const isAdmin = hasToken && role >= ROLE.ADMIN
  const allowed =
    access.enabled &&
    (!access.requireAuth || !!userID) &&
    (section !== 'admin' || isAdmin)
  const docs = useQuery({
    queryKey: [
      'api-docs',
      section ?? 'api',
      userID ?? null,
      sessionID ?? null,
      role,
    ],
    queryFn: async ({ signal }) => {
      const response = await api.get<{ success: boolean; data: APIDocument[] }>(
        section === 'admin' ? '/api/docs/admin' : '/api/docs',
        { signal, disableDuplicate: true, skipErrorHandler: true }
      )
      return response.data.data
    },
    enabled: allowed,
    gcTime: 0,
    staleTime: 0,
    refetchOnWindowFocus: 'always',
    retry: false,
  })
  return (
    <DocumentationReader
      key={section ?? 'api'}
      query={docs}
      article={article ?? (section === 'admin' ? 'overview' : 'quickstart')}
      section={section}
      allowed={allowed}
      manualAvailable={isAdmin}
    />
  )
}
