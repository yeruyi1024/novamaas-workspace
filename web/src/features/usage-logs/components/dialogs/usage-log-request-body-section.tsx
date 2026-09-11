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
import { AlertCircleIcon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'

import { getLogRequestSnapshots } from '../../task-content-api'
import { RequestSnapshotsViewer } from './request-snapshots-viewer'

interface UsageLogRequestBodySectionProps {
  isAdmin: boolean
  open: boolean
  requestBody?: unknown
  taskId?: string
  requestId?: string
}

export function UsageLogRequestBodySection(
  props: UsageLogRequestBodySectionProps
) {
  const { t } = useTranslation()
  const hasEmbeddedRequestBody = props.requestBody !== undefined
  const requestBodyQuery = useQuery({
    queryKey: [
      'usage-log-task-request-snapshots',
      props.isAdmin ? 'admin' : 'self',
      props.taskId,
      props.requestId,
    ],
    queryFn: async () => {
      if (!props.taskId && !props.requestId) {
        throw new Error('request body reference unavailable')
      }
      const response = await getLogRequestSnapshots(
        props.taskId,
        props.requestId
      )
      if (!response.success || response.data === undefined) {
        throw new Error(response.message || 'request body unavailable')
      }
      return response.data
    },
    enabled:
      props.isAdmin &&
      props.open &&
      !hasEmbeddedRequestBody &&
      Boolean(props.taskId || props.requestId),
    retry: false,
    staleTime: Number.POSITIVE_INFINITY,
  })
  const requestSnapshots = hasEmbeddedRequestBody
    ? { original: props.requestBody }
    : requestBodyQuery.data

  if (!props.open) return null

  return (
    <section className='flex min-w-0 flex-col gap-1.5'>
      <Label className='flex items-center gap-1.5 text-xs font-semibold'>
        {t('Request Snapshots')}
      </Label>
      {!hasEmbeddedRequestBody && requestBodyQuery.isFetching ? (
        <div className='bg-muted/30 text-muted-foreground flex min-h-32 items-center justify-center gap-2 rounded-md border text-sm'>
          <Spinner />
          <span>{t('Loading request body...')}</span>
        </div>
      ) : null}
      {!hasEmbeddedRequestBody && requestBodyQuery.isError ? (
        <Alert variant='destructive'>
          <HugeiconsIcon icon={AlertCircleIcon} strokeWidth={2} />
          <AlertTitle>{t('Request body unavailable')}</AlertTitle>
          <AlertDescription>{t('Request failed')}</AlertDescription>
        </Alert>
      ) : null}
      {requestSnapshots ? (
        <RequestSnapshotsViewer
          original={requestSnapshots.original}
          upstream={requestSnapshots.upstream}
        />
      ) : null}
    </section>
  )
}
