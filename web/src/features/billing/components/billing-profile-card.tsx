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
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useId } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { z } from 'zod'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Field,
  FieldGroup,
  FieldLabel,
  FieldDescription,
  FieldError,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { handleServerError } from '@/lib/handle-server-error'

import { billingTimestamp, getBillingAccount, saveBillingAccount } from '../api'
import type { BillingAccount } from '../types'

const schema = z.object({
  company_title: z.string().trim().max(200),
  tax_id: z.string().trim().max(64),
  start: z.boolean(),
  start_time: z.string(),
})
type Values = z.infer<typeof schema>
export function BillingProfileCard(props: { userId: number; admin?: boolean }) {
  const { t } = useTranslation()
  const account = useQuery({
    queryKey: ['billing', 'account', props.userId],
    queryFn: () => getBillingAccount(props.userId),
    enabled: props.userId > 0,
  })
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Billing identity')}</CardTitle>
      </CardHeader>
      <CardContent>
        {account.isPending && <p>{t('Loading...')}</p>}
        {account.isError && <p role='alert'>{t('Failed to load data')}</p>}
        {account.data && (
          <BillingProfileForm
            key={`${props.userId}:${account.data.profile_version}`}
            account={account.data}
            admin={props.admin}
          />
        )}
      </CardContent>
    </Card>
  )
}
function BillingProfileForm(props: {
  account: BillingAccount
  admin?: boolean
}) {
  const { t } = useTranslation()
  const id = useId()
  const queryClient = useQueryClient()
  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      company_title: props.account.company_title,
      tax_id: props.account.tax_id,
      start: false,
      start_time: '',
    },
  })
  const mutation = useMutation({
    mutationFn: (values: Values) => {
      const data: Partial<BillingAccount> = {
        company_title: values.company_title,
        tax_id: values.tax_id,
        profile_version: props.account.profile_version,
      }
      if (props.admin && values.start && !props.account.accounting_start_at) {
        data.accounting_start_at = values.start_time
          ? Math.floor(
              new Date(`${values.start_time}:00+08:00`).getTime() / 1000
            )
          : 0
      }
      return saveBillingAccount(props.account.user_id, data)
    },
    onSuccess: () => {
      toast.success(t('Saved successfully'))
      void queryClient.invalidateQueries({ queryKey: ['billing'] })
    },
    onError: handleServerError,
  })
  return (
    <form onSubmit={form.handleSubmit((values) => mutation.mutate(values))}>
      <FieldGroup>
        <Field>
          <FieldLabel htmlFor={`${id}-title`}>{t('Company title')}</FieldLabel>
          <Input
            id={`${id}-title`}
            {...form.register('company_title')}
            maxLength={200}
            aria-invalid={Boolean(form.formState.errors.company_title)}
            aria-describedby={
              form.formState.errors.company_title
                ? `${id}-title-error`
                : undefined
            }
          />
          <FieldError id={`${id}-title-error`}>
            {form.formState.errors.company_title &&
              t('Maximum {{count}} characters', { count: 200 })}
          </FieldError>
        </Field>
        <Field>
          <FieldLabel htmlFor={`${id}-tax`}>{t('Tax ID')}</FieldLabel>
          <Input
            id={`${id}-tax`}
            {...form.register('tax_id')}
            maxLength={64}
            aria-invalid={Boolean(form.formState.errors.tax_id)}
            aria-describedby={
              form.formState.errors.tax_id ? `${id}-tax-error` : undefined
            }
          />
          <FieldError id={`${id}-tax-error`}>
            {form.formState.errors.tax_id &&
              t('Maximum {{count}} characters', { count: 64 })}
          </FieldError>
        </Field>
        <p className='text-muted-foreground text-sm'>
          {t('Issued statements retain their original billing identity.')}
        </p>
        <FieldDescription>
          {t('Accounting start')} (Asia/Shanghai):{' '}
          {billingTimestamp(props.account.accounting_start_at)}
        </FieldDescription>
        {props.admin && !props.account.accounting_start_at && (
          <>
            <Field orientation='horizontal'>
              <Checkbox
                id={`${id}-start`}
                checked={form.watch('start')}
                onCheckedChange={(checked) =>
                  form.setValue('start', checked === true)
                }
              />
              <FieldLabel htmlFor={`${id}-start`}>
                {t('Enable formal accounting')}
              </FieldLabel>
            </Field>
            <Field>
              <FieldLabel htmlFor={`${id}-time`}>
                {t('Accounting start')}
              </FieldLabel>
              <Input
                id={`${id}-time`}
                type='datetime-local'
                {...form.register('start_time')}
                disabled={!form.watch('start')}
              />
              <FieldDescription>
                {t(
                  'Leave empty to start now. Earlier usage logs remain historical records.'
                )}
              </FieldDescription>
            </Field>
          </>
        )}
        <Button type='submit' disabled={mutation.isPending}>
          {mutation.isPending ? t('Saving...') : t('Save')}
        </Button>
      </FieldGroup>
    </form>
  )
}
