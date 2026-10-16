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

import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import { upstreamCostCoverage } from '../lib/cost-coverage'
import type { UpstreamChannelUserRow } from '../types'
import { CostAmount } from './cost-amount'

export function ChannelUserCosts(props: {
  rows?: UpstreamChannelUserRow[]
  schemaVersion: number
  symbol: string
  frozen?: boolean
}) {
  const { t } = useTranslation()
  if (props.frozen && props.schemaVersion < 2 && !props.rows) {
    return (
      <p className='text-muted-foreground text-sm'>
        {t(
          'This older snapshot did not freeze user cost allocation by channel.'
        )}
      </p>
    )
  }
  // The API returns the complete user distribution. Combine evidence counts only.
  const channels = new Map<
    number,
    { count: number; missing_cost_records: number }
  >()
  for (const row of props.rows || []) {
    const channel = channels.get(row.channel_id) || {
      count: 0,
      missing_cost_records: 0,
    }
    channel.count += row.count
    channel.missing_cost_records += row.missing_cost_records
    channels.set(row.channel_id, channel)
  }
  return (
    <Table aria-label={t('User cost allocation by channel')}>
      <TableCaption>
        {t(
          'Shares use each channel’s net upstream cost. A share is unavailable when costs are missing or the channel net cost is not positive.'
        )}
      </TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead>{t('Channel')}</TableHead>
          <TableHead>{t('User')}</TableHead>
          <TableHead className='text-right'>{t('Cost charges')}</TableHead>
          <TableHead className='text-right'>{t('Cost refunds')}</TableHead>
          <TableHead className='text-right'>{t('Net upstream cost')}</TableHead>
          <TableHead className='text-right'>{t('Channel net cost')}</TableHead>
          <TableHead className='text-right'>
            {t('Share of channel net cost')}
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {props.rows?.map((row) => {
          const coverage = upstreamCostCoverage(row)
          const channelCoverage = upstreamCostCoverage(
            channels.get(row.channel_id) || row
          )
          return (
            <TableRow key={`${row.channel_id}:${row.user_id}`}>
              <TableCell>
                {row.channel_name || t('Unknown')} (#{row.channel_id})
              </TableCell>
              <TableCell>
                <span>
                  {row.display_name || row.username || t('Unknown')} (#
                  {row.user_id})
                </span>
                {row.username &&
                  row.display_name &&
                  row.username !== row.display_name && (
                    <span className='text-muted-foreground ml-2'>
                      {row.username}
                    </span>
                  )}
              </TableCell>
              <TableCell className='text-right'>
                <CostAmount
                  amount={row.charge}
                  symbol={props.symbol}
                  coverage={coverage}
                />
              </TableCell>
              <TableCell className='text-right'>
                <CostAmount
                  amount={row.refund}
                  symbol={props.symbol}
                  coverage={coverage}
                />
              </TableCell>
              <TableCell className='text-right'>
                <CostAmount
                  amount={row.amount}
                  symbol={props.symbol}
                  coverage={coverage}
                />
              </TableCell>
              <TableCell className='text-right'>
                <CostAmount
                  amount={row.channel_amount}
                  symbol={props.symbol}
                  coverage={channelCoverage}
                />
              </TableCell>
              <TableCell className='text-right'>
                {row.share_percent === null ? '—' : `${row.share_percent}%`}
              </TableCell>
            </TableRow>
          )
        })}
        {!props.rows?.length && (
          <TableRow>
            <TableCell
              colSpan={7}
              className='text-muted-foreground text-center'
            >
              {t('No records')}
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  )
}
