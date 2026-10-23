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
import { getRouteApi, Link } from '@tanstack/react-router'
import type { ColumnDef } from '@tanstack/react-table'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { DataTablePage, useDataTable } from '@/components/data-table'
import { StatusBadge } from '@/components/status-badge'
import { Badge } from '@/components/ui/badge'
import {
  studioStatusBadgeVariant,
  studioStatusLabelKey,
} from '@/features/image-studio/studio-utils'
import { useTableUrlState } from '@/hooks/use-table-url-state'
import { formatTimestampToDate } from '@/lib/format'

import { getStudioTaskLogs, type StudioTaskLog } from '../api'
import { getDefaultTimeRange } from '../lib/utils'
import { createChannelColumn } from './columns/column-helpers'
import { TaskLogsFilterBar } from './task-logs-filter-bar'
import { useLogsViewScope } from './usage-logs-provider'

const route = getRouteApi('/_authenticated/usage-logs/$section')

export function StudioTaskLogs() {
  const { t } = useTranslation()
  const { isAdminView } = useLogsViewScope()
  const search = route.useSearch()
  const state = useTableUrlState({
    search,
    navigate: route.useNavigate(),
    pagination: { defaultPage: 1, defaultPageSize: 20 },
    globalFilter: { enabled: false },
  })
  const [{ start, end }] = useState(getDefaultTimeRange)
  const params = {
    p: state.pagination.pageIndex + 1,
    page_size: state.pagination.pageSize,
    scope: isAdminView ? ('all' as const) : ('self' as const),
    filter: search.filter,
    channel_id:
      isAdminView && search.channel ? Number(search.channel) : undefined,
    start_timestamp: Math.floor((search.startTime ?? start.getTime()) / 1000),
    end_timestamp: Math.floor((search.endTime ?? end.getTime()) / 1000),
  }
  const logs = useQuery({
    queryKey: ['logs', 'image-studio', params],
    queryFn: () => getStudioTaskLogs(params),
    retry: false,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
  })
  const columns: ColumnDef<StudioTaskLog>[] = [
    {
      accessorKey: 'created_at',
      header: t('Submit Time'),
      size: 180,
      cell: ({ row }) => (
        <span className='text-xs whitespace-nowrap tabular-nums'>
          {formatTimestampToDate(row.original.created_at, 'seconds')}
        </span>
      ),
    },
    ...(isAdminView
      ? [
          createChannelColumn<StudioTaskLog>({ headerLabel: t('Channel') }),
          { accessorKey: 'owner_user_id', header: t('User ID') },
        ]
      : []),
    { accessorKey: 'model', header: t('Model') },
    {
      accessorKey: 'status',
      header: t('Status'),
      cell: ({ row }) => (
        <Badge variant={studioStatusBadgeVariant(row.original.status)}>
          {row.original.status === 'deleted'
            ? t('Record deleted')
            : t(studioStatusLabelKey(row.original.status))}
        </Badge>
      ),
    },
    {
      accessorKey: 'id',
      header: t('Task ID'),
      size: 220,
      cell: ({ row }) => (
        <StatusBadge
          label={row.original.id}
          copyText={row.original.id}
          variant='neutral'
          size='sm'
          className='max-w-[220px] truncate font-mono'
        />
      ),
    },
    {
      accessorKey: 'request_id',
      header: t('Request ID'),
      size: 220,
      cell: ({ row }) =>
        row.original.request_id ? (
          <Link
            className='text-primary block max-w-[220px] truncate font-mono text-xs underline underline-offset-4'
            to='/usage-logs/$section'
            params={{ section: 'common' }}
            search={{
              requestId: row.original.request_id,
              startTime: (row.original.created_at - 60) * 1000,
              endTime:
                (Math.max(row.original.created_at, row.original.updated_at) +
                  60) *
                1000,
            }}
          >
            {row.original.request_id}
          </Link>
        ) : (
          '—'
        ),
    },
    {
      accessorKey: 'duration_ms',
      header: t('Duration'),
      cell: ({ row }) =>
        row.original.duration_ms > 0
          ? `${(row.original.duration_ms / 1000).toFixed(2)} ${t('s')}`
          : '—',
    },
    {
      accessorKey: 'error_message',
      header: t('Fail Reason'),
      size: 260,
      cell: ({ row }) => (
        <span className='block max-w-[320px] text-xs break-words whitespace-pre-wrap'>
          {row.original.error_message || '—'}
        </span>
      ),
    },
  ]
  const { table } = useDataTable({
    data: logs.data?.items ?? [],
    columns,
    pagination: state.pagination,
    onPaginationChange: state.onPaginationChange,
    manualPagination: true,
    manualFiltering: true,
    enableRowSelection: false,
    totalCount: logs.data?.total ?? 0,
    ensurePageInRange: state.ensurePageInRange,
  })

  return (
    <div className='flex h-full min-h-0 flex-col gap-3'>
      {logs.isError && (
        <p role='alert' className='text-destructive text-sm'>
          {t('Failed to load logs')}
        </p>
      )}
      <DataTablePage
        table={table}
        columns={columns}
        isLoading={logs.isPending}
        isFetching={logs.isFetching}
        emptyTitle={t('No Logs Found')}
        emptyDescription={t(
          'Image task logs show progress and request IDs. Charges are recorded in common logs.'
        )}
        toolbar={<TaskLogsFilterBar table={table} logCategory='drawing' />}
      />
    </div>
  )
}
