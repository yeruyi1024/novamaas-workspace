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
import { getRouteApi } from '@tanstack/react-router'
import { Download, Loader2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Dialog } from '@/components/dialog'
import { Button } from '@/components/ui/button'
import { formatTimestampToDate } from '@/lib/format'

import { exportLogs } from '../api'
import { LOG_TYPE_ALL_VALUE, LOG_TYPE_FILTERS } from '../constants'
import { buildApiParams } from '../lib/utils'
import { useLogsViewScope } from './usage-logs-provider'

const route = getRouteApi('/_authenticated/usage-logs/$section')

/**
 * Downloads every log matching the current filters as an Excel file, after the
 * user confirms the conditions. Every filter of the table is listed in the
 * confirmation, with the ones left empty shown as unlimited.
 */
export function ExportLogsButton() {
  const { t } = useTranslation()
  const searchParams = route.useSearch()
  const { canManageScope, isAdminView: isAdmin } = useLogsViewScope()
  const [confirming, setConfirming] = useState(false)
  const [exporting, setExporting] = useState(false)

  const params = buildApiParams({
    page: 1,
    pageSize: 1,
    searchParams,
    isAdmin,
  })
  delete params.p
  delete params.page_size

  const logType = LOG_TYPE_FILTERS.find(
    (type) => type.value === String(params.type ?? LOG_TYPE_ALL_VALUE)
  )
  const conditions = [
    {
      label: t('Date Range'),
      value: `${formatTimestampToDate(params.start_timestamp)} ~ ${formatTimestampToDate(params.end_timestamp)}`,
    },
    ...(canManageScope
      ? [
          {
            label: t('View Scope'),
            value: isAdmin ? t('All') : t('Only Mine'),
          },
        ]
      : []),
    { label: t('Log Type'), value: logType ? t(logType.label) : '' },
    ...(isAdmin ? [{ label: t('Username'), value: params.username }] : []),
    { label: t('Token'), value: params.token_name },
    { label: t('Model Name'), value: params.model_name },
    { label: t('Group'), value: params.group },
    ...(isAdmin
      ? [
          {
            label: t('Channel'),
            value: params.channel ? `#${params.channel}` : '',
          },
        ]
      : []),
    { label: t('Request ID'), value: params.request_id },
    { label: t('Upstream Request ID'), value: params.upstream_request_id },
  ]

  const handleExport = async () => {
    setConfirming(false)
    setExporting(true)
    try {
      const { blob, filename, count } = await exportLogs(params, isAdmin)
      const url = URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = filename
      anchor.click()
      URL.revokeObjectURL(url)
      toast.success(t('Exported {{count}} logs', { count }))
    } catch (error) {
      toast.error(
        error instanceof Error && error.message
          ? error.message
          : t('Failed to export logs')
      )
    } finally {
      setExporting(false)
    }
  }

  return (
    <>
      <Button
        variant='outline'
        size='sm'
        onClick={() => setConfirming(true)}
        disabled={exporting}
      >
        {exporting ? <Loader2 className='animate-spin' /> : <Download />}
        {t('Export')}
      </Button>
      <Dialog
        open={confirming}
        onOpenChange={setConfirming}
        title={t('Export logs')}
        description={t(
          'Logs matching the current filters will be exported as an Excel file.'
        )}
        contentClassName='sm:max-w-md'
        footer={
          <>
            <Button variant='outline' onClick={() => setConfirming(false)}>
              {t('Cancel')}
            </Button>
            <Button onClick={handleExport}>{t('Confirm export')}</Button>
          </>
        }
      >
        <dl className='space-y-2 text-sm'>
          {conditions.map((condition) => (
            <div key={condition.label} className='flex gap-3'>
              <dt className='text-muted-foreground w-24 shrink-0'>
                {condition.label}
              </dt>
              <dd className='min-w-0 break-words'>
                {condition.value || (
                  <span className='text-muted-foreground'>
                    {t('Unlimited')}
                  </span>
                )}
              </dd>
            </div>
          ))}
        </dl>
        <p className='text-muted-foreground mt-4 text-xs'>
          {t(
            'Only consumption and refund records are exported; top-up, management, system, error and login logs are not included.'
          )}
        </p>
        <p className='text-muted-foreground mt-1 text-xs'>
          {t('At most 366 days and 200,000 logs per export.')}
        </p>
      </Dialog>
    </>
  )
}
