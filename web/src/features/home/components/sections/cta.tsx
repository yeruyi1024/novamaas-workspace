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
import { ArrowRight01Icon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface CTAProps {
  className?: string
  isAuthenticated?: boolean
}

export function CTA(props: CTAProps) {
  const { t } = useTranslation()

  if (props.isAuthenticated) {
    return null
  }

  return (
    <section
      className={cn(
        'maas-deferred-section relative overflow-hidden px-5 py-24 sm:px-6 md:py-32',
        props.className
      )}
      aria-labelledby='home-cta-title'
    >
      <div aria-hidden className='maas-cta-orb absolute -z-10 rounded-full' />
      <AnimateInView animation='scale-in' className='mx-auto max-w-7xl'>
        <div className='maas-cta-panel relative overflow-hidden rounded-[2rem] px-6 py-16 text-center sm:px-10 md:py-24'>
          <div aria-hidden className='maas-cta-grid absolute inset-0' />
          <div className='relative mx-auto max-w-3xl'>
            <p className='maas-section-kicker'>{t('Build the market')}</p>
            <h2
              id='home-cta-title'
              className='mt-4 text-3xl leading-tight font-semibold tracking-[-0.04em] text-balance md:text-6xl'
            >
              {t('Build the operating system for AI supply')}
            </h2>
            <p className='text-muted-foreground mx-auto mt-6 max-w-2xl text-base leading-7 text-pretty md:text-lg'>
              {t(
                'Start with token aggregation and distribution today, then expand into the compute market as your business grows.'
              )}
            </p>
            <div className='mt-9 flex flex-wrap items-center justify-center gap-3'>
              <Button
                size='lg'
                className='maas-primary-action h-11 rounded-full px-5'
                render={<Link to='/sign-up' />}
              >
                {t('Get Started')}
                <HugeiconsIcon icon={ArrowRight01Icon} data-icon='inline-end' />
              </Button>
              <Button
                variant='outline'
                size='lg'
                className='maas-secondary-action h-11 rounded-full px-5'
                render={<Link to='/pricing' />}
              >
                {t('Explore model supply')}
              </Button>
            </div>
          </div>
        </div>
      </AnimateInView>
    </section>
  )
}
