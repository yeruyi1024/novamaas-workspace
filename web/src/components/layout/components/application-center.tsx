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
import { DashboardSquare01Icon, Image01Icon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  NavigationMenuTrigger,
} from '@/components/ui/navigation-menu'
import { useStatus } from '@/hooks/use-status'

export function ApplicationCenter(props: { onNavigate?: () => void }) {
  const { t } = useTranslation()
  const { status } = useStatus()

  return (
    <NavigationMenu render={<div />} className='flex-none'>
      <NavigationMenuList>
        <NavigationMenuItem>
          <NavigationMenuTrigger className='text-muted-foreground hover:text-foreground max-lg:size-8 max-lg:p-0 max-lg:[&>svg:last-child]:hidden'>
            <HugeiconsIcon
              icon={DashboardSquare01Icon}
              aria-hidden='true'
              className='size-4 lg:hidden'
            />
            <span className='sr-only lg:not-sr-only'>
              {t('Application center')}
            </span>
          </NavigationMenuTrigger>
          <NavigationMenuContent className='w-48'>
            {status?.ImageStudioEnabled !== false && (
              <NavigationMenuLink
                render={<Link to='/image-studio' onClick={props.onNavigate} />}
              >
                <HugeiconsIcon icon={Image01Icon} aria-hidden='true' />
                {t('Image creation')}
              </NavigationMenuLink>
            )}
          </NavigationMenuContent>
        </NavigationMenuItem>
      </NavigationMenuList>
    </NavigationMenu>
  )
}
