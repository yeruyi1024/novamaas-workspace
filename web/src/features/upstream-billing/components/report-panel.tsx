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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import {
  billingPreviousMonth,
  billingTimestamp,
  billingToday,
} from '@/features/billing/api'

import { getUpstreamReport } from '../api'
import type { UpstreamRange } from '../types'
import { ChannelUserCosts } from './channel-user-costs'
import { CostRows } from './cost-rows'
import { ReportSummary } from './report-summary'
import { UsageDetails } from './usage-details'

export function ReportPanel(props: { groupId: number }) {
  const { t } = useTranslation()
  const [mode, setMode] = useState('month')
  const [month, setMonth] = useState(billingPreviousMonth)
  const [startDate, setStartDate] = useState(
    () => `${billingPreviousMonth()}-01`
  )
  const [endDate, setEndDate] = useState(billingToday)
  const [selection, setSelection] = useState<{
    day?: string
    model?: string
  } | null>(null)
  const range: UpstreamRange =
    mode === 'month' ? { month } : { start_date: startDate, end_date: endDate }
  const valid =
    mode === 'month'
      ? /^\d{4}-\d{2}$/.test(month)
      : Boolean(startDate && endDate && startDate <= endDate)
  const report = useQuery({
    queryKey: ['upstream-billing', 'report', props.groupId, range],
    queryFn: () => getUpstreamReport(props.groupId, range),
    enabled: props.groupId > 0 && valid,
  })
  return (
    <Card>
      <CardHeader className='gap-4'>
        <CardTitle>{t('Upstream reconciliation')}</CardTitle>
        <CardDescription>
          {t(
            'Review frozen upstream costs and usage across the channels in this accounting group.'
          )}
        </CardDescription>
        <ToggleGroup
          value={[mode]}
          onValueChange={(values) => values[0] && setMode(values[0])}
          spacing={1}
          aria-label={t('Period')}
        >
          <ToggleGroupItem value='month'>{t('By month')}</ToggleGroupItem>
          <ToggleGroupItem value='range'>{t('Date range')}</ToggleGroupItem>
        </ToggleGroup>
        <FieldGroup className='grid gap-4 sm:grid-cols-3'>
          {mode === 'month' ? (
            <Field>
              <FieldLabel htmlFor='upstream-month'>
                {t('Statement month')} (Asia/Shanghai)
              </FieldLabel>
              <Input
                id='upstream-month'
                type='month'
                min='1970-02'
                max={billingToday().slice(0, 7)}
                value={month}
                onChange={(event) => setMonth(event.target.value)}
              />
            </Field>
          ) : (
            <>
              <Field data-invalid={!valid}>
                <FieldLabel htmlFor='upstream-start'>
                  {t('Start date')} (Asia/Shanghai)
                </FieldLabel>
                <Input
                  id='upstream-start'
                  type='date'
                  max={endDate || billingToday()}
                  value={startDate}
                  aria-invalid={!valid}
                  onChange={(event) => setStartDate(event.target.value)}
                />
              </Field>
              <Field data-invalid={!valid}>
                <FieldLabel htmlFor='upstream-end'>
                  {t('End date')} (Asia/Shanghai)
                </FieldLabel>
                <Input
                  id='upstream-end'
                  type='date'
                  min={startDate}
                  max={billingToday()}
                  value={endDate}
                  aria-invalid={!valid}
                  onChange={(event) => setEndDate(event.target.value)}
                />
              </Field>
            </>
          )}
          <div className='flex items-end'>
            <Button
              variant='outline'
              disabled={!valid || report.isFetching}
              onClick={() => void report.refetch()}
            >
              {t('Refresh')}
            </Button>
          </div>
        </FieldGroup>
        {!valid && (
          <p role='alert' className='text-destructive text-sm'>
            {t('Select a valid date range.')}
          </p>
        )}
      </CardHeader>
      <CardContent className='flex flex-col gap-5'>
        {report.isFetching && <p role='status'>{t('Loading...')}</p>}
        {report.isError && <p role='alert'>{t('Failed to load data')}</p>}
        {valid && report.data && !report.isError && (
          <>
            <p className='text-muted-foreground text-sm'>
              {report.data.start_date} – {report.data.end_date} ·{' '}
              {report.data.currency.code} · {t('Data as of')}:{' '}
              {billingTimestamp(report.data.updated_at)}
            </p>
            <ReportSummary report={report.data} />
            <Tabs defaultValue='days'>
              <TabsList>
                <TabsTrigger value='days'>
                  {t('Daily upstream costs')}
                </TabsTrigger>
                <TabsTrigger value='models'>
                  {t('Upstream model usage')}
                </TabsTrigger>
                <TabsTrigger value='channel-users'>
                  {t('User cost allocation by channel')}
                </TabsTrigger>
              </TabsList>
              <TabsContent value='days'>
                <CostRows
                  mode='day'
                  scope={report.data}
                  rows={report.data.days}
                  total={report.data.total}
                  symbol={report.data.currency.symbol}
                  onSelect={(row) => setSelection({ day: row.label })}
                />
              </TabsContent>
              <TabsContent value='models'>
                <CostRows
                  mode='model'
                  scope={report.data}
                  rows={report.data.models}
                  total={report.data.total}
                  symbol={report.data.currency.symbol}
                  onSelect={(row) =>
                    'model_name' in row &&
                    setSelection({ model: row.model_name })
                  }
                />
              </TabsContent>
              <TabsContent value='channel-users'>
                <ChannelUserCosts
                  rows={report.data.channel_users}
                  schemaVersion={report.data.schema_version}
                  symbol={report.data.currency.symbol}
                />
              </TabsContent>
            </Tabs>
            <p className='text-muted-foreground text-sm'>
              {t(
                'Usage totals include known values only. Incomplete and estimated usage is marked explicitly.'
              )}
            </p>
          </>
        )}
        {selection && report.data && (
          <UsageDetails
            key={`${props.groupId}:${JSON.stringify(range)}:${JSON.stringify(selection)}`}
            groupId={props.groupId}
            range={range}
            filters={selection}
            symbol={report.data.currency.symbol}
            models={
              selection.day
                ? report.data.day_models.filter(
                    (row) => row.date === selection.day
                  )
                : undefined
            }
            onClose={() => setSelection(null)}
          />
        )}
      </CardContent>
    </Card>
  )
}
