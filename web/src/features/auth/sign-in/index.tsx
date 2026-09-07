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
import { Link, useSearch } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { useStatus } from '@/hooks/use-status'

import { AuthLayout } from '../auth-layout'
import { TermsFooter } from '../components/terms-footer'
import { SignInShowcase } from './components/sign-in-showcase'
import { UserAuthForm } from './components/user-auth-form'

export function SignIn() {
  const { t } = useTranslation()
  const { redirect } = useSearch({ from: '/(auth)/sign-in' })
  const { status } = useStatus()
  const hasLegalFooter = Boolean(
    status?.user_agreement_enabled || status?.privacy_policy_enabled
  )

  return (
    <AuthLayout showcase={<SignInShowcase />}>
      <div className='flex flex-col'>
        <header className='flex flex-col gap-2.5'>
          <h1 className='text-2xl leading-tight font-semibold tracking-[-0.025em] sm:text-3xl'>
            {t('Welcome back!')}
          </h1>
          {!status?.self_use_mode_enabled &&
          status?.register_enabled !== false ? (
            <p className='text-muted-foreground text-sm leading-6'>
              {t("Don't have an account?")}{' '}
              <Link
                to='/sign-up'
                className='text-foreground decoration-border hover:text-primary font-medium underline underline-offset-4 transition-colors'
              >
                {t('Sign up')}
              </Link>
              .
            </p>
          ) : null}
        </header>

        <UserAuthForm redirectTo={redirect} className='mt-7' />

        {hasLegalFooter ? (
          <div className='border-border/70 mt-7 border-t pt-5'>
            <TermsFooter
              variant='sign-in'
              status={status}
              className='text-center leading-5'
            />
          </div>
        ) : null}
      </div>
    </AuthLayout>
  )
}
