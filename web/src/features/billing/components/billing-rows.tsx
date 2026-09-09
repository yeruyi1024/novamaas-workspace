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
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableFooter,
} from '@/components/ui/table'

import type { BillingRow } from '../types'

export function BillingRows(props: {
  rows: BillingRow[]
  total: BillingRow
  symbol: string
  roundingDifference?: string
  onSelectRow?: (row: BillingRow) => void
}) {
  const { t } = useTranslation()
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{t('Date / hour')}</TableHead>
          <TableHead className='text-right'>{t('Consumption')}</TableHead>
          <TableHead className='text-right'>{t('Refund')}</TableHead>
          <TableHead className='text-right'>{t('Net amount')}</TableHead>
          <TableHead className='text-right'>{t('Records')}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {props.rows.map((row) => (
          <TableRow key={row.label}>
            <TableCell>
              {props.onSelectRow &&
              row.count > 0 &&
              row.state !== 'future' &&
              row.state !== 'outside_period' ? (
                <Button
                  variant='link'
                  className='h-auto p-0'
                  onClick={() => props.onSelectRow?.(row)}
                >
                  {row.label}
                </Button>
              ) : (
                row.label
              )}
              {row.state === 'in_progress' && (
                <Badge variant='outline' className='ml-2'>
                  {t('In progress')}
                </Badge>
              )}
            </TableCell>
            {row.state === 'future' || row.state === 'outside_period' ? (
              <TableCell
                colSpan={4}
                className='text-muted-foreground text-right'
              >
                —{' '}
                {row.state === 'future'
                  ? t('Not yet occurred')
                  : t('Outside accounting period')}
              </TableCell>
            ) : (
              <>
                <TableCell className='text-right tabular-nums'>
                  {props.symbol} {row.charge}
                </TableCell>
                <TableCell className='text-right tabular-nums'>
                  {props.symbol} {row.refund}
                </TableCell>
                <TableCell className='text-right font-medium tabular-nums'>
                  {props.symbol} {row.amount}
                </TableCell>
                <TableCell className='text-right tabular-nums'>
                  {row.count}
                </TableCell>
              </>
            )}
          </TableRow>
        ))}
      </TableBody>
      <TableFooter>
        <TableRow>
          <TableCell>{t('Total')}</TableCell>
          <TableCell className='text-right tabular-nums'>
            {props.symbol} {props.total.charge}
          </TableCell>
          <TableCell className='text-right tabular-nums'>
            {props.symbol} {props.total.refund}
          </TableCell>
          <TableCell className='text-right tabular-nums'>
            {props.symbol} {props.total.amount}
          </TableCell>
          <TableCell className='text-right tabular-nums'>
            {props.total.count}
          </TableCell>
        </TableRow>
        {props.roundingDifference &&
          props.roundingDifference !== '0.000000' && (
            <TableRow>
              <TableCell colSpan={3}>
                {t('Display rounding difference')}
              </TableCell>
              <TableCell className='text-right tabular-nums'>
                {props.symbol} {props.roundingDifference}
              </TableCell>
              <TableCell />
            </TableRow>
          )}
      </TableFooter>
    </Table>
  )
}
