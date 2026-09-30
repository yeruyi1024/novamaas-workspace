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
  ChartRelationshipIcon,
  CheckmarkCircle02Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useTranslation } from 'react-i18next'

import { IconWeChat } from '@/assets/brand-icons'
import { AnimateInView } from '@/components/animate-in-view'
import { Badge } from '@/components/ui/badge'

export function MiniProgram() {
  const { t } = useTranslation()

  const benefits = [
    {
      icon: CheckmarkCircle02Icon,
      title: t('See generation results'),
      description: t(
        'Check completed work without opening the desktop console.'
      ),
    },
    {
      icon: ChartRelationshipIcon,
      title: t('Follow token usage'),
      description: t('Keep consumption visible as work happens.'),
    },
    {
      icon: Building02Icon,
      title: t('Stay in the team loop'),
      description: t('Spot usage changes and act sooner.'),
    },
  ]

  return (
    <section
      className='maas-mini-section maas-deferred-section relative overflow-hidden px-5 py-24 sm:px-6 md:py-32'
      aria-labelledby='mini-program-title'
    >
      <div className='mx-auto grid max-w-7xl items-center gap-12 lg:grid-cols-[minmax(0,1fr)_minmax(0,0.9fr)] lg:gap-20'>
        <AnimateInView animation='fade-right' className='max-w-2xl'>
          <div className='flex flex-wrap items-center gap-3'>
            <p className='maas-section-kicker'>{t('The next touchpoint')}</p>
            <Badge variant='outline' className='maas-roadmap-badge'>
              {t('Coming soon')}
            </Badge>
          </div>
          <h2
            id='mini-program-title'
            className='mt-5 text-3xl leading-tight font-semibold tracking-[-0.035em] text-balance md:text-5xl'
          >
            {t('AI results and usage, always close')}
          </h2>
          <p className='text-muted-foreground mt-6 text-base leading-7 text-pretty'>
            {t(
              'The planned mini program brings generation results, token usage and team visibility to your phone. Check routine work on the move, then return to the console for deeper governance.'
            )}
          </p>
          <div className='mt-9 grid gap-5 sm:grid-cols-3 lg:grid-cols-1'>
            {benefits.map((benefit) => (
              <div key={benefit.title} className='flex items-start gap-3'>
                <span className='maas-market-icon flex size-10 shrink-0 items-center justify-center rounded-xl'>
                  <HugeiconsIcon
                    icon={benefit.icon}
                    className='size-5'
                    strokeWidth={1.7}
                    aria-hidden='true'
                  />
                </span>
                <div>
                  <h3 className='text-sm font-semibold'>{benefit.title}</h3>
                  <p className='text-muted-foreground mt-1 text-sm leading-6'>
                    {benefit.description}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </AnimateInView>

        <AnimateInView animation='scale-in' delay={100}>
          <div className='maas-mini-stage flex min-h-[27rem] items-center justify-center rounded-[2rem] p-8 sm:min-h-[32rem]'>
            <div
              role='img'
              aria-label={t('WeChat mini program')}
              className='maas-mini-phone relative w-full max-w-[17rem] rounded-[2.4rem] p-3'
            >
              <div className='maas-mini-screen overflow-hidden rounded-[1.8rem] px-5 pt-7 pb-6'>
                <div
                  className='bg-foreground/15 mx-auto mb-8 h-1.5 w-16 rounded-full'
                  aria-hidden='true'
                />
                <div className='flex items-center justify-center'>
                  <span
                    className='flex size-11 items-center justify-center rounded-2xl bg-[#07c160]/10 text-[#07c160]'
                    aria-hidden='true'
                  >
                    <IconWeChat className='size-7' aria-hidden='true' />
                  </span>
                </div>
                <div className='mt-8 space-y-3'>
                  {[
                    {
                      label: t('Generation results'),
                      icon: CheckmarkCircle02Icon,
                    },
                    { label: t('Token usage'), icon: ChartRelationshipIcon },
                    { label: t('Team view'), icon: Building02Icon },
                  ].map((item) => (
                    <div
                      key={item.label}
                      className='maas-mini-row flex items-center gap-3 rounded-xl px-3 py-3'
                    >
                      <HugeiconsIcon
                        icon={item.icon}
                        className='size-4 shrink-0'
                        aria-hidden='true'
                      />
                      <span className='text-xs font-medium'>{item.label}</span>
                      <span
                        className='maas-mini-row-dot ml-auto size-1.5 rounded-full'
                        aria-hidden='true'
                      />
                    </div>
                  ))}
                </div>
                <div className='maas-mini-caption mt-8 rounded-xl px-3 py-3 text-center text-[11px] leading-5'>
                  {t('Quick checks on mobile. Deeper control on desktop.')}
                </div>
              </div>
            </div>
          </div>
        </AnimateInView>
      </div>
    </section>
  )
}
