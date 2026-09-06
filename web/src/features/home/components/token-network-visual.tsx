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
import {
  AiNetworkIcon,
  Building02Icon,
  CloudServerIcon,
  CpuIcon,
  DatabaseSync01Icon,
  Key01Icon,
  Share01Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon, type IconSvgElement } from '@hugeicons/react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'

type NetworkNodeProps = {
  icon: IconSvgElement
  label: string
  meta: string
  roadmap?: boolean
}

function NetworkNode(props: NetworkNodeProps) {
  const { t } = useTranslation()

  return (
    <div className='maas-network-node flex items-center gap-3 rounded-xl px-3 py-3'>
      <span className='maas-node-icon flex size-9 shrink-0 items-center justify-center rounded-lg'>
        <HugeiconsIcon
          icon={props.icon}
          className='size-4.5'
          strokeWidth={1.8}
          aria-hidden='true'
        />
      </span>
      <span className='min-w-0 flex-1'>
        <span className='block text-xs leading-tight font-medium'>
          {props.label}
        </span>
        <span className='text-muted-foreground mt-0.5 block truncate text-[10px]'>
          {props.meta}
        </span>
        {props.roadmap ? (
          <Badge
            variant='outline'
            className='maas-roadmap-badge mt-1.5 h-4 px-1.5 text-[8px] leading-none'
          >
            {t('Roadmap')}
          </Badge>
        ) : null}
      </span>
    </div>
  )
}

export function TokenNetworkVisual() {
  const { t } = useTranslation()

  const supplyNodes = [
    {
      icon: CloudServerIcon,
      label: t('Model providers'),
      meta: t('Public API supply'),
    },
    {
      icon: DatabaseSync01Icon,
      label: t('Private token pools'),
      meta: t('Managed inventory'),
    },
    {
      icon: CpuIcon,
      label: t('Compute capacity'),
      meta: t('Schedulable supply'),
      roadmap: true,
    },
  ]

  const demandNodes = [
    {
      icon: Key01Icon,
      label: t('API consumers'),
      meta: t('Metered access'),
    },
    {
      icon: Share01Icon,
      label: t('Channel partners'),
      meta: t('Commercial distribution'),
    },
    {
      icon: Building02Icon,
      label: t('Enterprise teams'),
      meta: t('Governed usage'),
    },
  ]

  return (
    <div
      className='maas-network-panel relative mx-auto max-w-3xl overflow-hidden rounded-[1.75rem] p-3 sm:p-4'
      role='group'
      aria-label={t('AI supply network')}
    >
      <div aria-hidden className='maas-network-scan absolute inset-x-0 h-px' />

      <div className='maas-network-surface relative overflow-hidden rounded-[1.35rem] p-4 sm:p-5'>
        <div className='flex items-center justify-between gap-4'>
          <div className='flex min-w-0 items-center gap-2.5'>
            <span aria-hidden className='maas-live-dot size-2 rounded-full' />
            <span className='truncate font-mono text-[10px] tracking-[0.16em] uppercase sm:text-xs'>
              {t('AI supply network')}
            </span>
          </div>
          <Badge variant='outline' className='maas-online-badge shrink-0'>
            {t('Control plane online')}
          </Badge>
        </div>

        <div className='mt-5 grid items-center gap-3 lg:grid-cols-[minmax(0,1fr)_2.5rem_minmax(9rem,0.8fr)_2.5rem_minmax(0,1fr)]'>
          <div className='flex min-w-0 flex-col gap-2.5'>
            <p className='text-muted-foreground mb-0.5 font-mono text-[10px] tracking-[0.14em] uppercase'>
              {t('Supply')}
            </p>
            {supplyNodes.map((node) => (
              <NetworkNode key={node.label} {...node} />
            ))}
          </div>

          <div aria-hidden className='maas-network-connector hidden lg:block' />

          <div className='maas-network-core relative mx-auto flex aspect-square w-full max-w-40 items-center justify-center rounded-full p-5'>
            <div
              aria-hidden
              className='maas-core-ring absolute inset-2 rounded-full'
            />
            <div className='relative flex flex-col items-center text-center'>
              <span className='maas-core-icon flex size-12 items-center justify-center rounded-2xl'>
                <HugeiconsIcon
                  icon={AiNetworkIcon}
                  className='size-6'
                  strokeWidth={1.7}
                  aria-hidden='true'
                />
              </span>
              <span className='mt-3 text-xs font-semibold'>
                {t('Unified token pool')}
              </span>
              <span className='text-muted-foreground mt-1 font-mono text-[9px] tracking-wider uppercase'>
                {t('Policy · Routing · Billing')}
              </span>
            </div>
          </div>

          <div aria-hidden className='maas-network-connector hidden lg:block' />

          <div className='flex min-w-0 flex-col gap-2.5'>
            <p className='text-muted-foreground mb-0.5 font-mono text-[10px] tracking-[0.14em] uppercase'>
              {t('Demand')}
            </p>
            {demandNodes.map((node) => (
              <NetworkNode key={node.label} {...node} />
            ))}
          </div>
        </div>

        <div className='maas-network-footer mt-5 grid grid-cols-3 gap-px overflow-hidden rounded-xl'>
          {[
            { label: t('Routing policy'), value: t('Adaptive') },
            { label: t('Metered settlement'), value: t('Traceable') },
            { label: t('Access control'), value: t('Governed') },
          ].map((item) => (
            <div key={item.label} className='px-2 py-3 text-center'>
              <div className='maas-network-value text-xs font-medium'>
                {item.value}
              </div>
              <div className='text-muted-foreground mt-1 text-[9px] leading-tight sm:text-[10px]'>
                {item.label}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
