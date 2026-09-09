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
import { useQuery } from '@tanstack/react-query'
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

import { billingTimestamp, billingToday, getBillingDay } from '../api'
import { BillingRows } from './billing-rows'
import { UsageDetailsDialog } from './usage-details-dialog'

export function ConsumptionPanel(props: {
  userId: number
  date: string
  onDateChange: (date: string) => void
}) {
  const { t } = useTranslation()
  const [hour, setHour] = useState<number | null>(null)
  const day = useQuery({
    queryKey: ['billing', 'day', props.userId, props.date],
    queryFn: () => getBillingDay(props.userId, props.date),
    enabled: props.userId > 0 && Boolean(props.date),
    refetchInterval: props.date === billingToday() ? 30000 : false,
  })
  return (
    <Card>
      <CardHeader className='gap-4'>
        <CardTitle>{t('Consumption lookup')}</CardTitle>
        <CardDescription>
          {t(
            'View historical consumption by day and hour. This reference view is separate from formal monthly statements.'
          )}
        </CardDescription>
        <FieldGroup className='flex flex-row items-end gap-3'>
          <Field className='max-w-60'>
            <FieldLabel htmlFor='billing-date'>
              {t('Date')} (Asia/Shanghai)
            </FieldLabel>
            <Input
              id='billing-date'
              type='date'
              min='1970-01-02'
              max={billingToday()}
              value={props.date}
              onChange={(event) => props.onDateChange(event.target.value)}
            />
          </Field>
          <Button
            variant='outline'
            onClick={() => void day.refetch()}
            disabled={!props.date || day.isFetching}
          >
            {t('Refresh')}
          </Button>
        </FieldGroup>
      </CardHeader>
      <CardContent className='flex flex-col gap-4'>
        {day.isPending && props.date && <p>{t('Loading...')}</p>}
        {day.isError && <p role='alert'>{t('Failed to load data')}</p>}
        {day.data && !day.isError && (
          <>
            <p className='text-muted-foreground text-sm'>
              {t(
                'Click a populated hour to view request details. Funding and temporary reservations are excluded.'
              )}
            </p>
            <p className='text-muted-foreground text-sm'>
              {day.data.currency.code} · {t('Data as of')}:{' '}
              {billingTimestamp(day.data.updated_at)} ·{' '}
              {t('Latest usage record')}:{' '}
              {billingTimestamp(day.data.last_log_at)}
            </p>
            {day.data.total.count === 0 && (
              <p role='status'>
                {t('No usage records for this date. Try another date.')}
              </p>
            )}
            <BillingRows
              rows={day.data.hours}
              total={day.data.total}
              symbol={day.data.currency.symbol}
              roundingDifference={day.data.rounding_difference}
              onSelectRow={(row) => setHour(Number(row.label.slice(0, 2)))}
            />
          </>
        )}
        {hour !== null && (
          <UsageDetailsDialog
            key={`${props.userId}:${props.date}:${hour}`}
            userId={props.userId}
            date={props.date}
            hour={hour}
            onClose={() => setHour(null)}
          />
        )}
      </CardContent>
    </Card>
  )
}
