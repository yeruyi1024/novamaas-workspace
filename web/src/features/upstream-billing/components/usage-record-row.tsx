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
import { TableCell, TableRow } from '@/components/ui/table'
import { billingTimestamp } from '@/features/billing/api'
import { LOG_TYPE_ENUM } from '@/features/usage-logs/constants'

import { upstreamUsageSourceKeys } from '../lib/usage-source'
import type { UpstreamUsageRecord } from '../types'

export function UsageRecordRow(props: {
  record: UpstreamUsageRecord
  symbol: string
}) {
  const { t } = useTranslation()
  const notApplicable =
    props.record.usage_source === 'task_failed_refunded' ||
    props.record.usage_source === 'task_historical_failed_refunded'
  const usageUnavailable = props.record.usage_missing || notApplicable
  const videoFallback = notApplicable ? '—' : t('Unknown')
  const sourceKey = Object.hasOwn(
    upstreamUsageSourceKeys,
    props.record.usage_source
  )
    ? upstreamUsageSourceKeys[props.record.usage_source]
    : 'Unknown'
  return (
    <TableRow>
      <TableCell className='bg-popover sticky left-0 z-10 min-w-44 whitespace-nowrap'>
        {billingTimestamp(props.record.occurred_at)}
      </TableCell>
      <TableCell>#{props.record.channel_id}</TableCell>
      <TableCell className='max-w-64 min-w-48 break-all whitespace-normal'>
        {props.record.model_name}
        {props.record.upstream_model_name &&
          props.record.upstream_model_name !== props.record.model_name && (
            <p className='text-muted-foreground text-xs'>
              {props.record.upstream_model_name}
            </p>
          )}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {props.record.cost_missing
          ? '—'
          : `${props.symbol} ${props.record.amount}`}{' '}
        {props.record.log_type === LOG_TYPE_ENUM.REFUND && (
          <Badge variant='outline'>{t('Refund')}</Badge>
        )}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {usageUnavailable ? '—' : props.record.input_tokens}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {usageUnavailable ? '—' : props.record.output_tokens}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {usageUnavailable ? '—' : props.record.cache_read_tokens}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {usageUnavailable ? '—' : props.record.cache_write_tokens}
        {!usageUnavailable &&
          (props.record.cache_write_5m_tokens > 0 ||
            props.record.cache_write_1h_tokens > 0) && (
            <p className='text-muted-foreground text-xs'>
              {t('5 minutes')}: {props.record.cache_write_5m_tokens} ·{' '}
              {t('1 hour')}: {props.record.cache_write_1h_tokens}
            </p>
          )}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {usageUnavailable ? videoFallback : props.record.video_tokens}
      </TableCell>
      <TableCell className='text-right tabular-nums'>
        {usageUnavailable ? videoFallback : props.record.video_seconds}
      </TableCell>
      <TableCell className='max-w-64 min-w-56 whitespace-normal'>
        <div className='flex flex-wrap gap-1'>
          {props.record.cost_missing && (
            <Badge variant='destructive'>{t('Missing cost')}</Badge>
          )}
          {props.record.usage_missing && (
            <Badge variant='outline'>{t('Incomplete usage')}</Badge>
          )}
          {props.record.usage_estimated && (
            <Badge variant='secondary'>{t('Estimated usage')}</Badge>
          )}
          {props.record.task_pending && (
            <Badge variant='outline'>{t('Pending tasks')}</Badge>
          )}
        </div>
        <p className='text-muted-foreground mt-1 text-xs'>
          {t('Source')}: {t(sourceKey)}
        </p>
      </TableCell>
      <TableCell className='max-w-80 min-w-64 break-all whitespace-normal'>
        <p>{props.record.request_id || '—'}</p>
        <p className='text-muted-foreground mt-1 text-xs'>
          {t('Source log ID')}: {props.record.source_log_id || '—'}
        </p>
        {props.record.task_id && (
          <p className='text-muted-foreground mt-1 text-xs'>
            {t('Task ID')}: {props.record.task_id}
          </p>
        )}
        {props.record.upstream_request_id && (
          <p className='text-muted-foreground mt-1 text-xs'>
            {t('Upstream Request ID')}: {props.record.upstream_request_id}
          </p>
        )}
      </TableCell>
    </TableRow>
  )
}
