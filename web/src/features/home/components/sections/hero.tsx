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
import { ArrowRight01Icon, BookOpen01Icon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useStatus } from '@/hooks/use-status'
import { cn } from '@/lib/utils'

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
        'maas-hero relative isolate overflow-hidden px-5 pt-36 pb-[21rem] sm:px-6 md:flex md:min-h-[720px] md:items-center md:py-32 lg:min-h-[740px]',
        props.className
      )}
      aria-labelledby='home-hero-title'
    >
      <div
        aria-hidden='true'
        data-testid='home-hero-artwork'
        className='maas-hero-image absolute inset-0 -z-10'
      />

      <div
        className='mx-auto w-full max-w-7xl 2xl:max-w-[clamp(80rem,80vw,112rem)]'
        data-testid='home-hero-layout'
      >
        <div className='flex max-w-xl min-w-0 flex-col items-start text-left xl:max-w-[700px] 2xl:max-w-[780px]'>
          <Badge
            variant='outline'
            className='maas-hero-badge landing-animate-fade-up h-7 gap-2 rounded-full px-3 opacity-0'
          >
            <span aria-hidden className='maas-live-dot size-1.5 rounded-full' />
            {t('AI compute gateway')}
          </Badge>

          <h1
            id='home-hero-title'
            className='maas-hero-title landing-animate-fade-up mt-6 font-semibold opacity-0'
            style={{ animationDelay: '70ms' }}
          >
            <span className='block'>
              {t('One gateway to diverse AI supply')}
            </span>{' '}
            <span className='maas-hero-accent block'>
              {t('Choose every token with purpose')}
            </span>
          </h1>

          <p
            className='landing-animate-fade-up text-muted-foreground mt-6 max-w-lg text-base leading-7 text-pretty opacity-0 2xl:max-w-[640px]'
            style={{ animationDelay: '140ms' }}
          >
            {t(
              'Connect models and token channels through one gateway. Next, discover evaluated supply and manage enterprise usage in a clearer procurement journey.'
            )}
          </p>

          <div
            className='landing-animate-fade-up mt-8 flex flex-wrap items-center gap-3 opacity-0'
            style={{ animationDelay: '210ms' }}
          >
            <Button
              size='lg'
              className='maas-primary-action h-11 rounded-full px-5'
              render={
                <Link to={props.isAuthenticated ? '/dashboard' : '/sign-up'} />
              }
            >
              {props.isAuthenticated ? t('Go to Dashboard') : t('Get Started')}
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
        </div>
      </div>
    </section>
  )
}
