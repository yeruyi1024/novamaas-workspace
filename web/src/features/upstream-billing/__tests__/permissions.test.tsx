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
import { renderHook, screen, waitFor } from '@testing-library/react'
import { afterEach, expect, test, vi } from 'vitest'

import { useSidebarData } from '@/hooks/use-sidebar-data'
import { useAuthStore } from '@/stores/auth-store'

import { getUpstreamGroups, getUpstreamReport } from '../api'
import { UpstreamBilling } from '../index'
import { group, renderUpstream, report } from './fixtures'

vi.mock('../api', async (original) => ({
  ...(await original<typeof import('../api')>()),
  getUpstreamGroups: vi.fn(),
  getUpstreamReport: vi.fn(),
}))
afterEach(() => {
  useAuthStore.getState().auth.reset()
  vi.resetAllMocks()
})

test.each([
  { role: 1, finance: false, allowed: false },
  { role: 1, finance: true, allowed: false },
  { role: 10, finance: false, allowed: false },
  { role: 10, finance: true, allowed: true },
  { role: 100, finance: false, allowed: true },
])(
  'role $role and finance $finance control upstream page queries and navigation',
  async (scenario) => {
    useAuthStore.getState().auth.setUser({
      id: 1,
      username: 'viewer',
      role: scenario.role,
      permissions: {
        admin_permissions: {
          financial_accounting: { view: scenario.finance },
        },
      },
    })
    vi.mocked(getUpstreamGroups).mockResolvedValue([group])
    vi.mocked(getUpstreamReport).mockResolvedValue(report)
    renderUpstream(<UpstreamBilling />)
    const navigation = renderHook(() => useSidebarData())
    const item = navigation.result.current.navGroups
      .flatMap((nav) => nav.items)
      .find((nav) => nav.url === '/upstream-billing')
    expect(item).toBeUndefined()
    if (scenario.allowed) {
      await waitFor(() => expect(getUpstreamReport).toHaveBeenCalled())
      expect(
        screen.getByRole('combobox', { name: 'Upstream accounting group' })
      ).toBeVisible()
    } else {
      expect(screen.getByRole('alert')).toHaveTextContent('Access denied')
      expect(getUpstreamGroups).not.toHaveBeenCalled()
      expect(getUpstreamReport).not.toHaveBeenCalled()
    }
  }
)
