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
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { SupplyConvergenceVisual } from '@/features/home/components/control-plane-visuals'

const authCapabilities = [
  'Multi-provider aggregation',
  'Commercial distribution',
  'Metered settlement',
] as const

export function SignInShowcase() {
  const { t } = useTranslation()

  return (
    <div className='mx-auto w-full max-w-[34rem]'>
      <Badge
        variant='outline'
        className='maas-hero-badge h-7 gap-2 rounded-full px-3'
      >
        <span aria-hidden className='maas-live-dot size-1.5 rounded-full' />
        {t('AI supply infrastructure')}
      </Badge>

      <h2 className='mt-5 text-[2.5rem] leading-[1.08] font-semibold tracking-[-0.04em] text-balance 2xl:text-[2.75rem]'>
        {t('Turn fragmented AI supply into')}{' '}
        <span className='maas-gradient-text'>
          {t('one programmable market')}
        </span>
      </h2>

      <p className='text-muted-foreground mt-4 text-base leading-7 text-pretty'>
        {t(
          'Aggregate tokens across providers, distribute access with commercial control, and prepare your platform for the next layer of compute supply.'
        )}
      </p>

      <div className='mt-7'>
        <SupplyConvergenceVisual />
      </div>

      <ul className='text-muted-foreground mt-5 flex flex-wrap gap-x-5 gap-y-2 text-sm'>
        {authCapabilities.map((capability) => (
          <li key={capability} className='flex items-center gap-2'>
            <span aria-hidden className='maas-live-dot size-1.5 rounded-full' />
            {t(capability)}
          </li>
        ))}
      </ul>
    </div>
  )
}
