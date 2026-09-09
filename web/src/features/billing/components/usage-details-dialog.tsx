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

import { billingTimestamp, getBillingUsageDetails } from '../api'

export function UsageDetailsDialog(props: {
  userId: number
  date: string
  hour: number
  onClose: () => void
}) {
  const { t } = useTranslation()
  const [cursors, setCursors] = useState<string[]>([])
  const cursor = cursors.at(-1) ?? ''
  const query = useQuery({
    queryKey: [
      'billing',
      'usage-details',
      props.userId,
      props.date,
      props.hour,
      cursor,
    ],
    queryFn: () =>
      getBillingUsageDetails(props.userId, props.date, props.hour, cursor),
  })
  return (
    <Dialog open onOpenChange={(open) => !open && props.onClose()}>
      <DialogContent className='max-h-[90dvh] overflow-y-auto sm:max-w-4xl'>
        <DialogHeader>
          <DialogTitle>{t('Hourly usage details')}</DialogTitle>
          <DialogDescription>
            {props.date} · {String(props.hour).padStart(2, '0')}:00–
            {String(props.hour + 1).padStart(2, '0')}:00 (Asia/Shanghai)
          </DialogDescription>
        </DialogHeader>
        <p className='text-muted-foreground text-sm'>
          {t('Usage logs are reference data, not archived statement evidence.')}
        </p>
        {query.isPending && <p>{t('Loading...')}</p>}
        {query.isError && <p role='alert'>{t('Failed to load data')}</p>}
        {query.data && !query.isError && (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Time')}</TableHead>
                <TableHead>{t('Model')}</TableHead>
                <TableHead>{t('Type')}</TableHead>
                <TableHead className='text-right'>{t('Net amount')}</TableHead>
                <TableHead>{t('Request ID')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {query.data.items.map((item) => (
                <TableRow key={`${item.id}:${item.request_id}`}>
                  <TableCell className='whitespace-nowrap'>
                    {billingTimestamp(item.created_at)}
                  </TableCell>
                  <TableCell className='max-w-80 break-all whitespace-normal'>
                    {item.model_name || '—'}
                  </TableCell>
                  <TableCell>
                    {item.kind === 'refund' ? t('Refund') : t('Consumption')}
                  </TableCell>
                  <TableCell className='text-right tabular-nums'>
                    {query.data.currency.symbol} {item.amount}
                  </TableCell>
                  <TableCell className='max-w-80 break-all whitespace-normal'>
                    {item.request_id || '—'}
                  </TableCell>
                </TableRow>
              ))}
              {query.data.items.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5}>
                    {t('No usage records for this hour.')}
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        )}
        <div className='flex flex-wrap justify-end gap-2'>
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
            onClick={() => setCursors(cursors.slice(0, -1))}
          >
            {t('Previous')}
          </Button>
          <Button
            variant='outline'
            disabled={
              !query.data?.next_cursor || query.isFetching || query.isError
            }
            onClick={() =>
              query.data?.next_cursor &&
              setCursors([...cursors, query.data.next_cursor])
            }
          >
            {t('Next')}
          </Button>
          <Button onClick={props.onClose}>{t('Close')}</Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
