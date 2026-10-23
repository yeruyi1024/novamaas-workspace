import { renderHook } from '@testing-library/react'
import i18next from 'i18next'
import { expect, test, vi } from 'vitest'

import { useSidebarData } from '../use-sidebar-data'

vi.mock('@/features/admin-manual/use-admin-manual-access', () => ({
  useAdminManualAccess: () => ({ available: false }),
}))

test('personal navigation lists profile, statements and wallet in order', () => {
  const { result } = renderHook(() => useSidebarData())
  const items = result.current.navGroups.find(
    (group) => group.id === 'personal'
  )?.items
  expect(items?.map((item) => item.title)).toEqual([
    i18next.t('Profile'),
    i18next.t('Billing statements'),
    i18next.t('Wallet'),
  ])
  expect(items?.map((item) => ('url' in item ? item.url : null))).toEqual([
    '/profile',
    '/billing',
    '/wallet',
  ])
})
