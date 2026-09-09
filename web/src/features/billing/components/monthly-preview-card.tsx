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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { handleServerError } from '@/lib/handle-server-error'

import {
  billingPreviousMonth,
  billingTimestamp,
  billingToday,
  getBillingStorageProfiles,
  prepareStatement,
  previewStatement,
} from '../api'
import { BillingReadiness } from './billing-readiness'
import { BillingRows } from './billing-rows'
import { HistoryImportDialog } from './history-import-dialog'

export function MonthlyPreviewCard(props: {
  userId: number
  onViewStatement: (id: string) => void
  onConfigureIdentity?: () => void
  onSelectDay?: (date: string) => void
}) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const [month, setMonth] = useState(billingPreviousMonth)
  const [profileId, setProfileId] = useState<number | null>(null)
  const profiles = useQuery({
    queryKey: ['billing', 'storage'],
    queryFn: getBillingStorageProfiles,
  })
  const preview = useQuery({
    queryKey: ['billing', 'preview', props.userId, month, profileId],
    queryFn: () => previewStatement(props.userId, month, profileId ?? 0),
    enabled: props.userId > 0 && Boolean(month),
  })
  const create = useMutation({
    mutationFn: () => prepareStatement(props.userId, month, profileId ?? 0),
    onSuccess: (statement) => {
      props.onViewStatement(statement.id)
      void client.invalidateQueries({ queryKey: ['billing'] })
    },
    onError: handleServerError,
  })
  const profileOptions = [
    { value: null, label: t('Select storage profile') },
    ...(profiles.data || []).map((item) => ({
      value: item.id,
      label: item.name,
    })),
  ]
  const data = preview.isError ? undefined : preview.data
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Monthly reconciliation')}</CardTitle>
        <CardDescription>
          {t(
            'Set billing identity → preview a closed month → create and review a draft → issue → customer confirms → download PDF.'
          )}
        </CardDescription>
      </CardHeader>
      <CardContent className='flex flex-col gap-4'>
        <FieldGroup className='grid gap-4 sm:grid-cols-3'>
          <Field>
            <FieldLabel htmlFor='statement-month'>{t('Month')}</FieldLabel>
            <Input
              id='statement-month'
              type='month'
              min='1970-02'
              max={billingToday().slice(0, 7)}
              value={month}
              onChange={(event) => setMonth(event.target.value)}
            />
          </Field>
          <Field>
            <FieldLabel htmlFor='statement-storage'>
              {t('Archive storage')}
            </FieldLabel>
            <Select
              items={profileOptions}
              value={profileId}
              onValueChange={setProfileId}
            >
              <SelectTrigger id='statement-storage' className='w-full'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {profileOptions.map((item) => (
                    <SelectItem key={item.value ?? 'empty'} value={item.value}>
                      {item.label}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
          <div className='flex items-end'>
            <Button
              variant='outline'
              disabled={!month}
              onClick={() => {
                void preview.refetch()
                void profiles.refetch()
              }}
            >
              {t('Preview')}
            </Button>
          </div>
        </FieldGroup>
        {profiles.isError && (
          <p role='alert'>
            {t('Failed to load archive storage. Refresh to retry.')}
          </p>
        )}
        {profiles.data?.length === 0 && (
          <p role='alert'>
            {t('Configure an enabled Aliyun OSS storage profile first.')}
          </p>
        )}
        {preview.isFetching && <p>{t('Loading...')}</p>}
        {preview.isError && <p role='alert'>{t('Failed to load data')}</p>}
        {data && (
          <>
            <BillingReadiness
              readiness={data.readiness}
              onConfigureIdentity={props.onConfigureIdentity}
              onViewStatement={props.onViewStatement}
            />
            <HistoryImportDialog
              key={`${props.userId}:${month}`}
              userId={props.userId}
              month={month}
              profileId={profileId}
              onViewStatement={props.onViewStatement}
            />
            {data.formal && (
              <section
                aria-label={t('Formal statement preview')}
                className='flex flex-col gap-3'
              >
                <h3 className='font-medium'>{t('Formal statement preview')}</h3>
                <p className='text-sm'>
                  {data.formal.company_title || '—'} ·{' '}
                  {data.formal.tax_id || '—'} · {data.formal.currency.code}
                </p>
                <p className='text-muted-foreground text-sm'>
                  {t('Covered period')}:{' '}
                  {billingTimestamp(data.readiness.period_start_at)} –{' '}
                  {billingTimestamp(data.readiness.period_end_at - 1)}
                </p>
                <BillingRows
                  rows={data.formal.days}
                  total={data.formal.total}
                  symbol={data.formal.currency.symbol}
                  roundingDifference={data.formal.rounding_difference}
                />
              </section>
            )}
          </>
        )}
        <Button
          className='self-start'
          disabled={
            !profileId ||
            !data?.readiness.ready ||
            preview.isFetching ||
            create.isPending
          }
          onClick={() => create.mutate()}
        >
          {create.isPending ? t('Preparing') : t('Create archived draft')}
        </Button>
        {data && (
          <section
            aria-label={t('Historical consumption (reference only)')}
            className='flex flex-col gap-3 border-t pt-4'
          >
            <h3 className='font-medium'>
              {t('Historical consumption (reference only)')}
            </h3>
            <p className='text-muted-foreground text-sm'>
              {t(
                'Usage logs are reference data, not archived statement evidence.'
              )}{' '}
              {t('Select a populated date to view its hourly breakdown.')}
            </p>
            <p className='text-muted-foreground text-sm'>
              {data.reference.currency.code} · {t('Data as of')}:{' '}
              {billingTimestamp(data.reference.updated_at)} ·{' '}
              {t('Latest usage record')}:{' '}
              {billingTimestamp(data.reference.last_log_at)}
            </p>
            <BillingRows
              rows={data.reference.days}
              total={data.reference.total}
              symbol={data.reference.currency.symbol}
              roundingDifference={data.reference.rounding_difference}
              onSelectRow={
                props.onSelectDay
                  ? (row) => props.onSelectDay?.(row.label)
                  : undefined
              }
            />
          </section>
        )}
      </CardContent>
    </Card>
  )
}
