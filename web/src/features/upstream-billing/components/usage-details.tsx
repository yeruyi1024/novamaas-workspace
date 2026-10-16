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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import { getUpstreamUsageDetails } from '../api'
import type { UpstreamModelRow, UpstreamRange } from '../types'
import { CostRows } from './cost-rows'
import { UsageRecordRow } from './usage-record-row'

export function UsageDetails(props: {
  groupId: number
  range: UpstreamRange
  filters: { day?: string; model?: string }
  symbol: string
  models?: UpstreamModelRow[]
  onClose: () => void
}) {
  const { t } = useTranslation()
  const [cursors, setCursors] = useState<number[]>([])
  const cursor = cursors.at(-1)
  const query = useQuery({
    queryKey: [
      'upstream-billing',
      'usage-details',
      props.groupId,
      props.range,
      props.filters,
      cursor,
    ],
    queryFn: () =>
      getUpstreamUsageDetails(props.groupId, props.range, {
        ...props.filters,
        cursor,
      }),
  })
  return (
    <Dialog open onOpenChange={(open) => !open && props.onClose()}>
      <DialogContent className='flex max-h-[90dvh] flex-col overflow-hidden sm:max-w-[calc(100%-2rem)]'>
        <DialogHeader className='shrink-0 pr-8'>
          <DialogTitle>{t('Upstream usage details')}</DialogTitle>
          <DialogDescription>
            {props.filters.day ||
              ('month' in props.range
                ? props.range.month
                : `${props.range.start_date} – ${props.range.end_date}`)}{' '}
            · {props.filters.model || t('All models')} · Asia/Shanghai
          </DialogDescription>
        </DialogHeader>
        <div className='min-h-0 overflow-y-auto'>
          {props.models && (
            <CostRows mode='model' rows={props.models} symbol={props.symbol} />
          )}
          {query.isPending && <p role='status'>{t('Loading...')}</p>}
          {query.isError && <p role='alert'>{t('Failed to load data')}</p>}
          {query.data && !query.isError && (
            <Table
              aria-label={t('Upstream request details')}
              className='min-w-[100rem]'
            >
              <TableHeader>
                <TableRow>
                  <TableHead className='bg-popover sticky left-0 z-10 min-w-44'>
                    {t('Time')}
                  </TableHead>
                  <TableHead>{t('Channel')}</TableHead>
                  <TableHead className='min-w-48'>{t('Model')}</TableHead>
                  <TableHead className='text-right'>{t('Net cost')}</TableHead>
                  <TableHead className='text-right'>
                    {t('Input tokens')}
                  </TableHead>
                  <TableHead className='text-right'>
                    {t('Output tokens')}
                  </TableHead>
                  <TableHead className='text-right'>
                    {t('Cache read tokens')}
                  </TableHead>
                  <TableHead className='text-right'>
                    {t('Cache write tokens')}
                  </TableHead>
                  <TableHead className='text-right'>
                    {t('Video tokens')}
                  </TableHead>
                  <TableHead className='text-right'>
                    {t('Video seconds')}
                  </TableHead>
                  <TableHead>{t('Usage quality')}</TableHead>
                  <TableHead className='min-w-64'>{t('Request ID')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {query.data.records.map((record) => (
                  <UsageRecordRow
                    key={`${record.source_log_id}:${record.sequence}`}
                    record={record}
                    symbol={props.symbol}
                  />
                ))}
                {query.data.records.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={12}>{t('No usage records')}</TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </div>
        <div className='flex shrink-0 flex-wrap justify-end gap-2'>
          <Button
            variant='outline'
            disabled={query.isFetching}
            onClick={() => void query.refetch()}
          >
            {t('Refresh')}
          </Button>
          <Button
            variant='outline'
            disabled={cursors.length === 0 || query.isFetching}
            onClick={() => setCursors((values) => values.slice(0, -1))}
          >
            {t('Previous')}
          </Button>
          <Button
            variant='outline'
            disabled={
              !query.data?.next_cursor || query.isFetching || query.isError
            }
            onClick={() => {
              const next = query.data?.next_cursor
              if (next) {
                setCursors((values) => [...values, next])
              }
            }}
          >
            {t('Next')}
          </Button>
          <Button onClick={props.onClose}>{t('Close')}</Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
