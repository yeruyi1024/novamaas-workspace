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

import { billingTimestamp } from '../api'
import type { BillingMonthPreview } from '../types'

export function BillingReadiness(props: {
  readiness: BillingMonthPreview['readiness']
  onConfigureIdentity?: () => void
  onViewStatement: (id: string) => void
}) {
  const { t } = useTranslation()
  const readiness = props.readiness
  const labels: Record<string, string> = {
    accounting_configured: t('Formal accounting enabled'),
    period_covered: t('Month overlaps the accounting period'),
    identity_complete: t('Company title and tax ID complete'),
    month_closed: t('Month closed plus 24 hours'),
    operations_settled: t('All reservations settled'),
    summary_valid: t('Summary data available'),
    history_reconciled: t(
      'Historical usage does not conflict with an empty ledger'
    ),
    no_active_statement: t('No active statement for this month'),
    storage_missing: t('Select archive storage'),
    storage_ready: t('Archive credentials readable'),
    storage_unavailable: t('Archive credentials unavailable'),
  }
  const descriptions: Record<
    BillingMonthPreview['readiness']['status'],
    string
  > = {
    not_configured: t(
      'Formal accounting is not enabled. Historical usage is shown below for reference.'
    ),
    outside_period: t(
      'This month is outside the formal accounting period. Historical usage is reference only.'
    ),
    data_error: t(
      'Formal summary data is invalid. Draft creation is blocked; contact the administrator.'
    ),
    no_consumption: t(
      'No formal entries in this period. This does not mean historical usage is zero.'
    ),
    historical_data_unreconciled: t(
      'Historical usage exists without formal entries. Review and import verified historical records before creating a statement. Changing the accounting start alone does not import history.'
    ),
    blocked: t('Complete the checks below before creating a draft.'),
    ready: t(
      'Ready to prepare. Evidence integrity and storage access are verified during archiving.'
    ),
  }
  return (
    <section
      aria-label={t('Statement readiness')}
      className='flex flex-col gap-3 rounded-lg border p-4 text-sm'
    >
      <h3 className='font-medium'>{t('Statement readiness')}</h3>
      <p role='status'>{descriptions[readiness.status]}</p>
      <p>
        {t('Accounting start')}:{' '}
        {billingTimestamp(readiness.accounting_start_at)} ·{' '}
        {t('Earliest draft creation')}: {billingTimestamp(readiness.prepare_at)}{' '}
        (Asia/Shanghai)
      </p>
      <ul className='grid gap-2 sm:grid-cols-2'>
        {readiness.checks.map((check) => (
          <li key={check.code} className='flex items-start gap-2'>
            <Badge variant={check.passed ? 'secondary' : 'outline'}>
              {check.passed ? t('Passed') : t('Not ready')}
            </Badge>
            <span>{labels[check.code] ?? check.code}</span>
          </li>
        ))}
      </ul>
      {readiness.pending > 0 && (
        <p>
          {t('Unsettled reservations')}: {readiness.pending}
        </p>
      )}
      <div className='flex flex-wrap gap-2'>
        {readiness.checks.some(
          (check) =>
            !check.passed &&
            (check.code === 'accounting_configured' ||
              check.code === 'identity_complete')
        ) && (
          <Button variant='outline' onClick={props.onConfigureIdentity}>
            {t('Configure billing identity')}
          </Button>
        )}
        {readiness.existing_statement && (
          <Button
            variant='outline'
            onClick={() => props.onViewStatement(readiness.existing_statement)}
          >
            {t('View existing statement')}
          </Button>
        )}
      </div>
    </section>
  )
}
