import { createFileRoute, redirect } from '@tanstack/react-router'

import { Main } from '@/components/layout'
import { StudioReview } from '@/features/image-studio/components/studio-review'
import { hasPermission } from '@/lib/admin-permissions'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

export const Route = createFileRoute('/_authenticated/image-review/')({
  beforeLoad: () => {
    const user = useAuthStore.getState().auth.user
    if (
      !user ||
      user.role < ROLE.ADMIN ||
      !hasPermission(user, 'image_studio', 'read') ||
      !hasPermission(user, 'image_studio', 'content_read')
    ) {
      throw redirect({ to: '/dashboard' })
    }
  },
  component: () => (
    <Main>
      <StudioReview />
    </Main>
  ),
})
