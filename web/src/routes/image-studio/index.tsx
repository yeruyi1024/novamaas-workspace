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
import { createFileRoute, redirect } from '@tanstack/react-router'

import { Main } from '@/components/layout'
import { SkipToMain } from '@/components/skip-to-main'
import { ImageStudio } from '@/features/image-studio'
import { getStudioSettings } from '@/features/image-studio/api'
import { LoginNoticeDialog } from '@/features/login-notice/login-notice-dialog'

export const Route = createFileRoute('/image-studio/')({
  beforeLoad: async () => {
    const settings = await getStudioSettings()
    if (!settings.enabled) {
      throw redirect({ to: '/' })
    }
  },
  component: ImageStudioPage,
})

export function ImageStudioPage() {
  return (
    <div className='bg-background flex h-svh min-h-0 w-full flex-col overflow-hidden'>
      <SkipToMain />
      <Main id='content' tabIndex={-1} className='p-0'>
        <ImageStudio />
      </Main>
      <LoginNoticeDialog />
    </div>
  )
}
