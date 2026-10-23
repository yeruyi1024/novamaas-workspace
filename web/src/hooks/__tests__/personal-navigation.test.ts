import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook } from '@testing-library/react'
import i18next from 'i18next'
import { createElement, type ReactNode } from 'react'
import { afterEach, expect, test } from 'vitest'

import { useSidebarData } from '../use-sidebar-data'

let client: QueryClient
afterEach(() => client?.clear())

test('personal navigation lists profile, statements and wallet in order', () => {
  client = new QueryClient({
    defaultOptions: { queries: { staleTime: Infinity } },
  })
  client.setQueryData(['status'], {})
  const { result } = renderHook(() => useSidebarData(), {
    wrapper: (props: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client }, props.children),
  })
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
