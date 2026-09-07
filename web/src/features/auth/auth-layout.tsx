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
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { LanguageSwitcher } from '@/components/language-switcher'
import { ThemeSwitch } from '@/components/theme-switch'
import { useSystemConfig } from '@/hooks/use-system-config'

type AuthLayoutProps = {
  children: React.ReactNode
  showcase?: React.ReactNode
}

export function AuthLayout(props: AuthLayoutProps) {
  const { t } = useTranslation()
  const { systemName, logo } = useSystemConfig()
  const hasShowcase = Boolean(props.showcase)

  return (
    <div className='maas-public relative min-h-svh overflow-x-hidden'>
      <div aria-hidden className='maas-hero-grid absolute inset-0 opacity-60' />
      <div
        aria-hidden
        className='maas-hero-orb maas-hero-orb-primary absolute'
      />
      <div
        aria-hidden
        className='maas-hero-orb maas-hero-orb-secondary absolute'
      />

      <header className='absolute inset-x-0 top-0 z-10 flex h-20 items-center justify-between gap-4 px-5 sm:px-8 lg:px-10'>
        <Link
          to='/'
          className='focus-visible:ring-ring/50 flex min-w-0 items-center gap-2.5 rounded-xl transition-opacity outline-none hover:opacity-80 focus-visible:ring-3'
        >
          <span className='relative size-9 shrink-0'>
            <img
              src={logo}
              alt={t('Logo')}
              className='ring-foreground/10 size-9 rounded-xl object-cover shadow-sm ring-1'
            />
          </span>
          <span className='truncate text-lg font-semibold tracking-tight'>
            {systemName}
          </span>
        </Link>

        <div className='border-border/70 bg-background/65 flex shrink-0 items-center gap-0.5 rounded-xl border p-1 shadow-sm backdrop-blur-xl'>
          <LanguageSwitcher />
          <ThemeSwitch />
        </div>
      </header>

      <main
        data-testid='auth-layout'
        className='relative mx-auto flex min-h-svh max-w-[90rem] items-center justify-center px-4 py-20 sm:px-6 sm:py-24 lg:px-8 lg:py-[clamp(6rem,10vh,9rem)]'
      >
        {hasShowcase ? (
          <div
            data-testid='auth-shell'
            className='grid w-full max-w-[30rem] gap-10 lg:min-h-[clamp(30rem,65svh,42rem)] lg:max-w-[76rem] lg:grid-cols-[minmax(0,1.08fr)_minmax(25rem,0.92fr)] lg:gap-[clamp(3rem,6vw,6rem)]'
          >
            <aside
              className='hidden min-w-0 flex-col justify-center lg:flex'
              aria-label={t('AI supply infrastructure')}
            >
              {props.showcase}
            </aside>

            <section className='relative flex min-w-0 items-center justify-center px-2 py-4 sm:px-4 lg:p-0'>
              <div
                data-testid='auth-content'
                className='mx-auto flex w-full max-w-[26rem] flex-col justify-center'
              >
                {props.children}
              </div>
            </section>
          </div>
        ) : (
          <section className='relative flex w-full min-w-0 items-center justify-center py-8'>
            <div
              data-testid='auth-content'
              className='mx-auto flex w-full max-w-[30rem] flex-col justify-center'
            >
              {props.children}
            </div>
          </section>
        )}
      </main>
    </div>
  )
}
