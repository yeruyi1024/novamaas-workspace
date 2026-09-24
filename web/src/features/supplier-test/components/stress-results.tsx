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

import { Card } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import {
  displayMeasured,
  displayThreshold,
  type Assessment,
} from '../baselines'
import type { StressMetrics } from '../types'

function duration(ms: number): string {
  if (ms <= 0) return '—'
  if (ms < 1) return `${ms.toFixed(2)} ms`
  if (ms >= 1000) return `${(ms / 1000).toFixed(1)} s`
  return `${ms.toFixed(1)} ms`
}

export function StressResults(props: {
  metrics: StressMetrics
  assessment: Assessment | null
  stream: boolean
}) {
  const { t } = useTranslation()
  const metrics = props.metrics
  const slowRows =
    props.assessment?.rows.filter(
      (row) => row.verdict === 'slow' || row.verdict === 'abnormal'
    ) ?? []
  const missingUsage = Math.max(
    0,
    metrics.succeeded - (metrics.usage_n ?? metrics.succeeded)
  )
  const completeUsage = missingUsage === 0 && metrics.succeeded > 0
  const issues = metrics.issues ?? []
  const otherIssues = metrics.other_issue_count ?? 0
  const missingTTFT = props.stream
    ? Math.max(0, metrics.succeeded - metrics.ttft_n)
    : 0
  const missingTPOT = props.stream
    ? Math.max(0, metrics.ttft_n - metrics.tpot_n)
    : 0
  const notRun =
    metrics.not_run ??
    Math.max(0, metrics.total - (metrics.attempted ?? metrics.total))
  const hasProblems =
    metrics.failed > 0 ||
    notRun > 0 ||
    missingUsage > 0 ||
    missingTTFT > 0 ||
    missingTPOT > 0 ||
    slowRows.length > 0

  return (
    <div className='space-y-4'>
      <Card className='space-y-3 p-4'>
        <h3 className='text-sm font-semibold'>{t('Timing and throughput')}</h3>
        <p className='text-muted-foreground text-xs'>
          {t(
            'Batch duration spans all workers. Request duration includes every attempted request; first output uses successful streamed requests.'
          )}
        </p>
        <div className='grid gap-3 text-sm sm:grid-cols-2 lg:grid-cols-3'>
          <div>
            <div className='text-muted-foreground'>{t('Batch wall time')}</div>
            <div className='font-medium'>{duration(metrics.elapsed_ms)}</div>
          </div>
          <div>
            <div className='text-muted-foreground'>
              {t('Request duration avg / P50 / P90')}
            </div>
            <div className='font-medium'>
              {duration(metrics.request_avg_ms ?? 0)} /{' '}
              {duration(metrics.request_p50_ms ?? 0)} /{' '}
              {duration(metrics.request_p90_ms ?? 0)}
            </div>
          </div>
          <div>
            <div className='text-muted-foreground'>
              {t('TTFT avg / P50 / P90')}
            </div>
            <div className='font-medium'>
              {metrics.ttft_n > 0
                ? `${duration(metrics.ttft_avg_ms)} / ${duration(metrics.ttft_p50_ms)} / ${duration(metrics.ttft_p90_ms)} (n=${metrics.ttft_n})`
                : t('No sample')}
            </div>
          </div>
          <div>
            <div className='text-muted-foreground'>
              {t('TPOT avg / P50 / P90')}
            </div>
            <div className='font-medium'>
              {metrics.tpot_n > 0
                ? `${duration(metrics.tpot_avg_ms)} / ${duration(metrics.tpot_p50_ms)} / ${duration(metrics.tpot_p90_ms)} (n=${metrics.tpot_n})`
                : t('No sample')}
            </div>
          </div>
          <div>
            <div className='text-muted-foreground'>
              {t('Batch output throughput')}
            </div>
            <div className='font-medium'>
              {completeUsage
                ? `${metrics.tokens_per_sec.toFixed(1)} tok/s`
                : t('No sample')}
            </div>
          </div>
          <div>
            <div className='text-muted-foreground'>
              {t('Per-request output rate')}
            </div>
            <div className='font-medium'>
              {completeUsage && metrics.request_tokens_per_sec
                ? `${metrics.request_tokens_per_sec.toFixed(1)} tok/s`
                : t('No sample')}
            </div>
          </div>
          <div>
            <div className='text-muted-foreground'>
              {t('Requests completed / not run')}
            </div>
            <div className='font-medium'>
              {metrics.attempted ?? metrics.total} / {metrics.not_run ?? 0}
            </div>
          </div>
        </div>
      </Card>

      <Card className='space-y-3 p-4'>
        <h3 className='text-sm font-semibold'>{t('Problems found')}</h3>
        {!hasProblems ? (
          <p className='text-muted-foreground text-sm'>
            {t(
              'No request errors or threshold violations were found in this run.'
            )}
          </p>
        ) : null}
        {missingUsage > 0 ? (
          <p className='text-warning text-sm'>
            {t(
              '{{count}} successful requests had no complete token usage; token rates and TPM are unavailable.',
              { count: missingUsage }
            )}
          </p>
        ) : null}
        {notRun > 0 ? (
          <p className='text-warning text-sm'>
            {t(
              '{{count}} planned requests were not started because the run stopped or was cancelled.',
              { count: notRun }
            )}
          </p>
        ) : null}
        {missingTTFT > 0 ? (
          <p className='text-warning text-sm'>
            {t(
              '{{count}} successful stream requests had no observable first output; TTFT is unavailable for them.',
              { count: missingTTFT }
            )}
          </p>
        ) : null}
        {missingTPOT > 0 ? (
          <p className='text-warning text-sm'>
            {t(
              '{{count}} streamed responses lacked enough timed output and token usage to calculate TPOT.',
              { count: missingTPOT }
            )}
          </p>
        ) : null}
        {metrics.failed > 0 && issues.length === 0 && otherIssues === 0 ? (
          <p className='text-warning text-sm'>
            {t('Failure details are unavailable for this run.')}
          </p>
        ) : null}
        {issues.length > 0 ? (
          <div className='overflow-x-auto'>
            <Table className='min-w-[680px]'>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('Count')}</TableHead>
                  <TableHead>{t('HTTP status')}</TableHead>
                  <TableHead>{t('Error or problem')}</TableHead>
                  <TableHead>{t('Example request')}</TableHead>
                  <TableHead>{t('Example duration')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {issues.map((issue) => (
                  <TableRow key={`${issue.status_code}-${issue.message}`}>
                    <TableCell>{issue.count}</TableCell>
                    <TableCell>
                      {issue.status_code || t('No HTTP response')}
                    </TableCell>
                    <TableCell className='max-w-[420px] break-all'>
                      {issue.message}
                    </TableCell>
                    <TableCell>
                      W{issue.worker} / R{issue.round}
                    </TableCell>
                    <TableCell>{duration(issue.elapsed_ms)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ) : null}
        {otherIssues > 0 ? (
          <p className='text-muted-foreground text-sm'>
            {t(
              '{{count}} additional failures are included in the failed total but omitted from the detailed list.',
              { count: otherIssues }
            )}
          </p>
        ) : null}
        {slowRows.map((row) => (
          <p key={row.id} className='text-warning text-sm'>
            {t(row.label)}: {displayMeasured(row, t)} ·{' '}
            {displayThreshold(row, t)}
          </p>
        ))}
      </Card>
    </div>
  )
}
