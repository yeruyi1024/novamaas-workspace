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

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'

import { upstreamCostCoverage } from '../lib/cost-coverage'
import type { UpstreamReport } from '../types'
import { CostAmount } from './cost-amount'

export function ReportSummary(props: { report: UpstreamReport }) {
  const { t } = useTranslation()
  const report = props.report
  return (
    <div className='flex flex-col gap-4'>
      <dl className='grid gap-4 rounded-lg border p-4 sm:grid-cols-3'>
        <div>
          <dt className='text-muted-foreground text-sm'>{t('Net cost')}</dt>
          <dd className='text-xl font-semibold tabular-nums'>
            <CostAmount
              amount={report.total.amount}
              symbol={report.currency.symbol}
              coverage={upstreamCostCoverage(report.total, report)}
              align='start'
            />
          </dd>
        </div>
        <div>
          <dt className='text-muted-foreground text-sm'>{t('Requests')}</dt>
          <dd className='text-xl font-semibold tabular-nums'>
            {report.total.requests}
          </dd>
        </div>
        <div>
          <dt className='text-muted-foreground text-sm'>{t('Channels')}</dt>
          <dd className='text-xl font-semibold tabular-nums'>
            {report.group.channels.filter((channel) => channel.eligible).length}
          </dd>
        </div>
      </dl>
      <Alert
        className={
          report.coverage.complete
            ? undefined
            : 'border-amber-500/40 bg-amber-500/5 text-amber-900 dark:text-amber-200'
        }
      >
        <AlertTitle>
          {report.coverage.complete
            ? t('Coverage complete')
            : t('Reconciliation warnings')}
        </AlertTitle>
        <AlertDescription className='flex flex-col gap-2'>
          <p>
            {t(
              'Channel tests and channels without a cost discount are excluded. A zero discount is included.'
            )}
          </p>
          {!report.coverage.complete && (
            <p>
              {t(
                'Warnings do not block creating, reviewing or archiving this statement. The archive retains the warnings and recorded data for reconciliation.'
              )}
            </p>
          )}
          <div className='flex flex-wrap gap-x-4 gap-y-1'>
            <span>
              {t('Missing channels')}: {report.coverage.missing_channels}
            </span>
            <span>
              {t('Missing cost')}: {report.coverage.missing_cost_records}
            </span>
            <span>
              {t('Incomplete usage')}: {report.coverage.missing_usage_records}
            </span>
            <span>
              {t('Estimated usage')}: {report.coverage.estimated_usage_records}
            </span>
            <span>
              {t('Pending tasks')}: {report.coverage.pending_tasks}
            </span>
          </div>
          {report.coverage.excluded_channels.length > 0 && (
            <p>
              {t('Excluded channels')}:{' '}
              {report.coverage.excluded_channels
                .map((channel) => `${channel.name} (#${channel.id})`)
                .join(', ')}
            </p>
          )}
        </AlertDescription>
      </Alert>
      <div className='flex flex-wrap gap-2' aria-label={t('Group channels')}>
        {report.group.channels.map((channel) => (
          <Badge
            key={channel.id}
            variant={channel.eligible ? 'secondary' : 'outline'}
          >
            {channel.name} (#{channel.id})
            {channel.missing && ` · ${t('Missing channel')}`}
            {!channel.missing && !channel.eligible && ` · ${t('Excluded')}`}
          </Badge>
        ))}
      </div>
    </div>
  )
}
