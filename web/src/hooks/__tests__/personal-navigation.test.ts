import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook } from '@testing-library/react'
import i18next from 'i18next'
import { createElement, type ReactNode } from 'react'
import { afterEach, expect, test } from 'vitest'

import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import { useSidebarData } from '../use-sidebar-data'

function renderSidebar(status: Record<string, unknown> = {}) {
  const client = new QueryClient({
    defaultOptions: { queries: { staleTime: Infinity } },
  })
  client.setQueryData(['status'], status)
  return renderHook(() => useSidebarData(), {
    wrapper: (props: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client }, props.children),
  })
}

function urls(
  groupId: string,
  navGroups: ReturnType<typeof useSidebarData>['navGroups']
) {
  return (
    navGroups
      .find((group) => group.id === groupId)
      ?.items.map((item) => ('url' in item ? item.url : null)) ?? []
  )
}

afterEach(() => {
  useAuthStore.getState().auth.reset()
})

test('personal navigation lists profile, statements and wallet in order', () => {
  const { result } = renderSidebar()
  const items = result.current.navGroups.find(
    (group) => group.id === 'personal'
  )?.items
  expect(items?.map((item) => item.title)).toEqual([
    i18next.t('Profile'),
    i18next.t('Billing statements'),
    i18next.t('Wallet'),
  ])
  expect(urls('personal', result.current.navGroups)).toEqual([
    '/profile',
    '/billing',
    '/wallet',
  ])
})

test.each([
  ['root', ROLE.SUPER_ADMIN, false, true],
  ['admin without image review permission', ROLE.ADMIN, false, false],
  ['admin with image review permission', ROLE.ADMIN, true, true],
  ['user with a stray permission flag', ROLE.USER, true, false],
])('%s sees image review: %s', (_name, role, granted, visible) => {
  useAuthStore.getState().auth.setUser({
    id: 7,
    username: 'reviewer',
    role,
    permissions: {
      admin_permissions: granted ? { image_studio: { read: true } } : {},
    },
  })
  const { result } = renderSidebar()
  const adminUrls = result.current.navGroups.flatMap((group) =>
    urls(group.id ?? '', result.current.navGroups)
  )
  expect(adminUrls.includes('/image-review')).toBe(visible)
  const disabled = renderSidebar({ ImageStudioEnabled: false }).result.current
  expect(urls('admin', disabled.navGroups)).not.toContain('/image-review')
})

test('image creation is absent from the console sidebar even when enabled', () => {
  expect(urls('chat', renderSidebar().result.current.navGroups)).not.toContain(
    '/image-studio'
  )
  expect(
    urls(
      'chat',
      renderSidebar({ ImageStudioEnabled: false }).result.current.navGroups
    )
  ).not.toContain('/image-studio')
})
