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
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import { upstreamCostCoverage } from '../lib/cost-coverage'
import type { UpstreamDayRow, UpstreamModelRow, UpstreamReport } from '../types'
import { CostAmount } from './cost-amount'

export function CostRows(props: {
  rows: (UpstreamDayRow | UpstreamModelRow)[]
  total?: UpstreamDayRow
  symbol: string
  mode: 'day' | 'model'
  scope?: Pick<UpstreamReport, 'group' | 'coverage'>
  onSelect?: (row: UpstreamDayRow | UpstreamModelRow) => void
}) {
  const { t } = useTranslation()
  return (
    <Table
      aria-label={
        props.mode === 'day'
          ? t('Daily upstream costs')
          : t('Upstream model usage')
      }
    >
      <TableHeader>
        <TableRow>
          <TableHead>{props.mode === 'day' ? t('Date') : t('Model')}</TableHead>
          <TableHead className='text-right'>{t('Cost amount')}</TableHead>
          <TableHead className='text-right'>{t('Refund')}</TableHead>
          <TableHead className='text-right'>{t('Net cost')}</TableHead>
          <TableHead className='text-right'>{t('Requests')}</TableHead>
          <TableHead className='text-right'>{t('Input tokens')}</TableHead>
          <TableHead className='text-right'>{t('Output tokens')}</TableHead>
          <TableHead className='text-right'>{t('Cache read tokens')}</TableHead>
          <TableHead className='text-right'>
            {t('Cache write tokens')}
          </TableHead>
          <TableHead className='text-right'>{t('Video tokens')}</TableHead>
          <TableHead className='text-right'>{t('Video seconds')}</TableHead>
          <TableHead>{t('Usage quality')}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {props.rows.map((row) => {
          const model = 'model_name' in row ? row : null
          const label = model?.model_name || row.label
          const key = model
            ? `${model.model_name}:${model.upstream_model_name || ''}:${model.billing_group || ''}`
            : row.label
          return (
            <TableRow key={key}>
              <TableCell className='max-w-72 min-w-36 break-words whitespace-normal'>
                {props.onSelect && row.count > 0 ? (
                  <Button
                    variant='link'
                    className='h-auto max-w-full p-0 text-left break-all whitespace-normal'
                    onClick={() => props.onSelect?.(row)}
                  >
                    {label}
                  </Button>
                ) : (
                  label
                )}
                {model?.upstream_model_name &&
                  model.upstream_model_name !== model.model_name && (
                    <p className='text-muted-foreground text-xs'>
                      {model.upstream_model_name}
                    </p>
                  )}
                {model?.billing_group && (
                  <p className='text-muted-foreground text-xs'>
                    {t('Billing group')}: {model.billing_group}
                  </p>
                )}
              </TableCell>
              <UsageCells row={row} symbol={props.symbol} scope={props.scope} />
            </TableRow>
          )
        })}
        {props.rows.length === 0 && (
          <TableRow>
            <TableCell colSpan={12}>{t('No usage records')}</TableCell>
          </TableRow>
        )}
      </TableBody>
      {props.total && (
        <TableFooter>
          <TableRow>
            <TableCell>{t('Total')}</TableCell>
            <UsageCells
              row={props.total}
              symbol={props.symbol}
              scope={props.scope}
              total
            />
          </TableRow>
        </TableFooter>
      )}
    </Table>
  )
}

function UsageCells(props: {
  row: UpstreamDayRow
  symbol: string
  scope?: Pick<UpstreamReport, 'group' | 'coverage'>
  total?: boolean
}) {
  const { t } = useTranslation()
  const coverage = upstreamCostCoverage(
    props.row,
    props.scope,
    props.total === true
  )
  return (
    <>
      <TableCell className='text-right tabular-nums'>
        <CostAmount
          amount={props.row.charge}
          symbol={props.symbol}
          coverage={coverage}
        />
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        <CostAmount
          amount={props.row.refund}
          symbol={props.symbol}
          coverage={coverage}
        />
      </TableCell>
      <TableCell className='text-right font-medium tabular-nums'>
        <CostAmount
          amount={props.row.amount}
          symbol={props.symbol}
          coverage={coverage}
        />
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {props.row.requests}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {props.row.input_tokens}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {props.row.output_tokens}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {props.row.cache_read_tokens}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {props.row.cache_write_tokens}
        {(props.row.cache_write_5m_tokens > 0 ||
          props.row.cache_write_1h_tokens > 0) && (
          <p className='text-muted-foreground text-xs'>
            {t('5 minutes')}: {props.row.cache_write_5m_tokens} · {t('1 hour')}:{' '}
            {props.row.cache_write_1h_tokens}
          </p>
        )}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {props.row.missing_usage_records > 0 && props.row.video_tokens === 0
          ? t('Unknown')
          : props.row.video_tokens}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {props.row.missing_usage_records > 0 && props.row.video_seconds === 0
          ? t('Unknown')
          : props.row.video_seconds}
      </TableCell>
      <TableCell>
        <div className='flex flex-wrap gap-1'>
          {props.row.missing_usage_records > 0 && (
            <Badge variant='outline'>
              {t('Incomplete usage')} ({props.row.missing_usage_records})
            </Badge>
          )}
          {props.row.estimated_usage_records > 0 && (
            <Badge variant='secondary'>
              {t('Estimated usage')} ({props.row.estimated_usage_records})
            </Badge>
          )}
          {props.row.missing_cost_records > 0 && (
            <Badge variant='destructive'>
              {t('Missing cost')} ({props.row.missing_cost_records})
            </Badge>
          )}
          {props.row.pending_tasks > 0 && (
            <Badge variant='outline'>
              {t('Pending tasks')} ({props.row.pending_tasks})
            </Badge>
          )}
          {props.row.missing_usage_records === 0 &&
            props.row.estimated_usage_records === 0 &&
            props.row.missing_cost_records === 0 &&
            props.row.pending_tasks === 0 &&
            '—'}
        </div>
      </TableCell>
    </>
  )
}
