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
  Building02Icon,
  CloudServerIcon,
  ComputerProgramming01Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'

export function MarketNetwork() {
  const { t } = useTranslation()

  const participants = [
    {
      icon: CloudServerIcon,
      label: t('For suppliers'),
      title: t('Turn fragmented inventory into reachable demand'),
      description: t(
        'Connect upstream token resources today and prepare schedulable compute supply for tomorrow.'
      ),
      outcome: t('Expand utilization'),
    },
    {
      icon: Building02Icon,
      label: t('For operators'),
      title: t('Build a distribution business with control'),
      description: t(
        'Define products, routes, permissions and commercial policies without rebuilding the infrastructure layer.'
      ),
      outcome: t('Control every margin'),
    },
    {
      icon: ComputerProgramming01Icon,
      label: t('For AI builders'),
      title: t('Access the right supply through one endpoint'),
      description: t(
        'Give applications and teams reliable model access while the platform handles upstream complexity.'
      ),
      outcome: t('Ship with confidence'),
    },
  ]

  return (
    <section
      className='maas-deferred-section relative px-5 py-24 sm:px-6 md:py-32'
      aria-labelledby='market-network-title'
    >
      <div className='mx-auto max-w-7xl'>
        <AnimateInView className='max-w-3xl'>
          <p className='maas-section-kicker'>{t('The market network')}</p>
          <h2
            id='market-network-title'
            className='mt-4 text-3xl leading-tight font-semibold tracking-[-0.035em] text-balance md:text-5xl'
          >
            {t('One platform, multiple business models')}
          </h2>
          <p className='text-muted-foreground mt-5 max-w-2xl text-base leading-7 text-pretty'>
            {t(
              'Create value for every participant without fragmenting the operating experience.'
            )}
          </p>
        </AnimateInView>

        <div className='mt-14 grid gap-px overflow-hidden rounded-3xl border md:grid-cols-3'>
          {participants.map((participant, index) => (
            <AnimateInView
              key={participant.label}
              delay={index * 100}
              className='maas-market-card flex min-h-80 flex-col p-6 md:p-8'
            >
              <div className='flex items-center justify-between gap-4'>
                <span className='maas-market-icon flex size-11 items-center justify-center rounded-2xl'>
                  <HugeiconsIcon
                    icon={participant.icon}
                    className='size-5'
                    strokeWidth={1.7}
                    aria-hidden='true'
                  />
                </span>
                <span className='text-muted-foreground font-mono text-[10px] tracking-[0.15em] uppercase'>
                  {participant.label}
                </span>
              </div>
              <h3 className='mt-10 text-xl leading-snug font-semibold tracking-tight text-balance'>
                {participant.title}
              </h3>
              <p className='text-muted-foreground mt-4 text-sm leading-6 text-pretty'>
                {participant.description}
              </p>
              <div className='maas-market-outcome mt-auto flex items-center gap-3 pt-8 text-sm font-medium'>
                <span aria-hidden className='h-px flex-1' />
                {participant.outcome}
              </div>
            </AnimateInView>
          ))}
        </div>
      </div>
    </section>
  )
}
