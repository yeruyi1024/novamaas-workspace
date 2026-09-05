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
import { Layers, Cable, ChartNoAxesCombined, KeyRound } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

export function Stats(props: { className?: string }) {
  const { t } = useTranslation()
  const capabilities = [
    {
      icon: Layers,
      title: t('Multi-model access'),
      description: t('One model service catalog'),
    },
    {
      icon: Cable,
      title: t('Multi-protocol Compatible'),
      description: t('Connect your existing AI tools'),
    },
    {
      icon: ChartNoAxesCombined,
      title: t('Transparent Billing'),
      description: t('Understand usage and costs'),
    },
    {
      icon: KeyRound,
      title: t('Access control'),
      description: t('Manage keys and permissions'),
    },
  ]

  return (
    <div
      className={cn(
        'maas-capabilities relative z-10 border-y border-border/50',
        props.className
      )}
    >
      <div className='mx-auto grid max-w-6xl grid-cols-2 gap-8 px-6 py-10 md:grid-cols-4 md:py-12'>
        {capabilities.map((item) => (
          <div
            key={item.title}
            className='flex min-w-0 flex-col items-start gap-2'
          >
            <item.icon
              aria-hidden
              className='maas-accent mb-2 size-5'
              strokeWidth={1.5}
            />
            <span className='text-sm font-semibold md:text-base'>
              {item.title}
            </span>
            <span className='text-muted-foreground text-xs leading-relaxed'>
              {item.description}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
