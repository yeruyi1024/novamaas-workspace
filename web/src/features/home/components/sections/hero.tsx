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
  ArrowRight01Icon,
  BookOpen01Icon,
  Tick02Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useStatus } from '@/hooks/use-status'
import { cn } from '@/lib/utils'

import { TokenNetworkVisual } from '../token-network-visual'

interface HeroProps {
  className?: string
  isAuthenticated?: boolean
}

export function Hero(props: HeroProps) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const docsUrl =
    (status?.docs_link as string | undefined) || 'https://docs.newapi.pro'
  const isExternalDocs = docsUrl.startsWith('http')

  return (
    <section
      className={cn(
        'maas-hero relative overflow-hidden px-5 pt-28 pb-20 sm:px-6 md:pt-36 md:pb-28 lg:pt-40 lg:pb-32',
        props.className
      )}
      aria-labelledby='home-hero-title'
    >
      <div aria-hidden className='maas-hero-grid absolute inset-0 -z-20' />
      <div
        aria-hidden
        className='maas-hero-orb maas-hero-orb-primary absolute -z-10'
      />
      <div
        aria-hidden
        className='maas-hero-orb maas-hero-orb-secondary absolute -z-10'
      />

      <div className='mx-auto grid max-w-7xl items-center gap-16 xl:grid-cols-[minmax(0,0.9fr)_minmax(32rem,1.1fr)]'>
        <div className='flex min-w-0 flex-col items-start'>
          <Badge
            variant='outline'
            className='maas-hero-badge landing-animate-fade-up h-7 gap-2 rounded-full px-3 opacity-0'
          >
            <span aria-hidden className='maas-live-dot size-1.5 rounded-full' />
            {t('AI supply infrastructure')}
          </Badge>

          <h1
            id='home-hero-title'
            className='landing-animate-fade-up mt-7 max-w-4xl text-[clamp(2.75rem,6vw,5.75rem)] leading-[0.98] font-semibold tracking-[-0.055em] text-balance opacity-0'
            style={{ animationDelay: '70ms' }}
          >
            {t('Turn fragmented AI supply into')}{' '}
            <span className='maas-gradient-text'>
              {t('one programmable market.')}
            </span>
          </h1>

          <p
            className='landing-animate-fade-up text-muted-foreground mt-7 max-w-2xl text-base leading-7 text-pretty opacity-0 md:text-lg md:leading-8'
            style={{ animationDelay: '140ms' }}
          >
            {t(
              'Aggregate tokens across providers, distribute access with commercial control, and prepare your platform for the next layer of compute supply.'
            )}
          </p>

          <div
            className='landing-animate-fade-up mt-9 flex flex-wrap items-center gap-3 opacity-0'
            style={{ animationDelay: '210ms' }}
          >
            <Button
              size='lg'
              className='maas-primary-action h-11 rounded-full px-5'
              render={
                <Link to={props.isAuthenticated ? '/dashboard' : '/sign-up'} />
              }
            >
              {props.isAuthenticated
                ? t('Go to Dashboard')
                : t('Start building')}
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

            {isExternalDocs ? (
              <Button
                variant='ghost'
                size='lg'
                className='h-11 rounded-full px-4'
                render={
                  <a href={docsUrl} target='_blank' rel='noopener noreferrer' />
                }
              >
                <HugeiconsIcon icon={BookOpen01Icon} data-icon='inline-start' />
                {t('Docs')}
              </Button>
            ) : (
              <Button
                variant='ghost'
                size='lg'
                className='h-11 rounded-full px-4'
                render={<Link to={docsUrl} />}
              >
                <HugeiconsIcon icon={BookOpen01Icon} data-icon='inline-start' />
                {t('Docs')}
              </Button>
            )}
          </div>

          <div
            className='landing-animate-fade-up text-muted-foreground mt-9 flex flex-wrap gap-x-5 gap-y-2 text-xs opacity-0 sm:text-sm'
            style={{ animationDelay: '280ms' }}
            aria-label={t('Platform capabilities')}
          >
            {[
              t('Multi-provider aggregation'),
              t('Commercial distribution'),
              t('Metered settlement'),
            ].map((capability) => (
              <span key={capability} className='flex items-center gap-2'>
                <span className='maas-check-icon flex size-4 items-center justify-center rounded-full'>
                  <HugeiconsIcon
                    icon={Tick02Icon}
                    className='size-2.5'
                    strokeWidth={2.5}
                    aria-hidden='true'
                  />
                </span>
                {capability}
              </span>
            ))}
          </div>
        </div>

        <div
          className='landing-animate-fade-up min-w-0 opacity-0'
          style={{ animationDelay: '220ms' }}
        >
          <TokenNetworkVisual />
        </div>
      </div>
    </section>
  )
}
