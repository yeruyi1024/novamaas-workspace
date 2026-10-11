import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, expect, test } from 'vitest'

import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import { useSidebarData } from '../use-sidebar-data'

afterEach(() => useAuthStore.getState().auth.reset())

test('hidden studio removes the creation entry while keeping administrator review', () => {
  useAuthStore
    .getState()
    .auth.setUser({ id: 1, username: 'root', role: ROLE.SUPER_ADMIN })
  const client = new QueryClient({
    defaultOptions: { queries: { staleTime: Infinity } },
  })
  client.setQueryData(['status'], { ImageStudioVisible: false })
  const hook = renderHook(() => useSidebarData(), {
    wrapper: (props: { children: ReactNode }) => (
      <QueryClientProvider client={client}>
        {props.children}
      </QueryClientProvider>
    ),
  })
  const urls = hook.result.current.navGroups
    .flatMap((group) => group.items)
    .filter((item) => 'url' in item)
    .map((item) => item.url)
  expect(urls).not.toContain('/image-studio')
  expect(urls).toContain('/image-review')
  hook.unmount()
  client.clear()
})

test.each([
  { role: ROLE.SUPER_ADMIN, read: false, contentRead: false, visible: true },
  { role: ROLE.ADMIN, read: true, contentRead: true, visible: true },
  { role: ROLE.ADMIN, read: false, contentRead: true, visible: false },
  { role: ROLE.ADMIN, read: true, contentRead: false, visible: false },
  { role: ROLE.USER, read: true, contentRead: true, visible: false },
])(
  'review entry visibility respects role $role and read=$read/content=$contentRead permissions',
  (input) => {
    useAuthStore.getState().auth.setUser({
      id: 8,
      username: 'viewer',
      role: input.role,
      permissions: {
        admin_permissions: {
          image_studio: { read: input.read, content_read: input.contentRead },
        },
      },
    })
    const client = new QueryClient({
      defaultOptions: { queries: { staleTime: Infinity } },
    })
    client.setQueryData(['status'], { ImageStudioVisible: true })
    const hook = renderHook(() => useSidebarData(), {
      wrapper: (props: { children: ReactNode }) => (
        <QueryClientProvider client={client}>
          {props.children}
        </QueryClientProvider>
      ),
    })
    const visible = hook.result.current.navGroups
      .flatMap((group) => group.items)
      .some((item) => 'url' in item && item.url === '/image-review')
    expect(visible).toBe(input.visible)
    hook.unmount()
    client.clear()
  }
)
