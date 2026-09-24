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
import { Fragment, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  getAssetRequestLogDetail,
  listAssetRequestLogs,
} from '@/features/assets/api'
import {
  assertAssetSuccess,
  assetErrorMessage,
} from '@/features/assets/asset-utils'

const PAGE_SIZE = 20

function RequestLogDetail(props: { id: number }) {
  const { t } = useTranslation()
  const detailQuery = useQuery({
    queryKey: ['asset-library', 'admin', 'request-log-detail', props.id],
    queryFn: async () =>
      assertAssetSuccess(await getAssetRequestLogDetail(props.id)),
  })
  if (detailQuery.isPending) return <Skeleton className='h-44 w-full' />
  if (detailQuery.isError) {
    return (
      <Alert variant='destructive'>
        <AlertTitle>{t('Unable to load request details')}</AlertTitle>
        <AlertDescription>
          {assetErrorMessage(detailQuery.error)}
        </AlertDescription>
      </Alert>
    )
  }
  const { log, detail } = detailQuery.data
  if (!detail) {
    return (
      <p className='text-muted-foreground text-sm'>
        {t('Details unavailable for this log')}
      </p>
    )
  }
  const fields = [
    { label: t('Request URL'), value: detail.request_url || '—' },
    {
      label: t('Request body'),
      value: detail.request_body_omitted
        ? t('Body omitted (over 16 KiB)')
        : detail.request_body || t('No request body'),
    },
    { label: t('Response status'), value: String(log.http_status || '—') },
    {
      label: t('Response body'),
      value: detail.response_body_omitted
        ? t('Body omitted (over 16 KiB)')
        : detail.response_body || t('No response body'),
    },
  ]
  return (
    <div className='grid gap-3 text-left'>
      <p className='text-muted-foreground text-xs'>
        {t('Addresses and bodies are redacted; details kept for 7 days.')}
      </p>
      {fields.map((field) => (
        <div key={field.label} className='min-w-0'>
          <div className='mb-1 text-sm font-medium'>{field.label}</div>
          <pre className='bg-muted max-h-56 overflow-auto rounded-md p-3 font-mono text-xs break-all whitespace-pre-wrap'>
            {field.value}
          </pre>
        </div>
      ))}
    </div>
  )
}

export function AssetRequestLogDialog(props: {
  open: boolean
  channelId: number
  replicaId?: number
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  const [source, setSource] = useState('')
  const [result, setResult] = useState('')
  const [draftRequestId, setDraftRequestId] = useState('')
  const [requestId, setRequestId] = useState('')
  const [rangeDays, setRangeDays] = useState(1)
  const [rangeEnd, setRangeEnd] = useState(() => Date.now())
  const [cursors, setCursors] = useState([''])
  const [pageIndex, setPageIndex] = useState(0)
  const [expandedId, setExpandedId] = useState<number | null>(null)
  const cursor = cursors[pageIndex]
  const resetPage = () => {
    setExpandedId(null)
    setCursors([''])
    setPageIndex(0)
  }
  const logsQuery = useQuery({
    queryKey: [
      'asset-library',
      'admin',
      'request-logs',
      props.channelId,
      props.replicaId,
      source,
      result,
      requestId,
      cursor,
      rangeDays,
      rangeEnd,
    ],
    queryFn: async () =>
      assertAssetSuccess(
        await listAssetRequestLogs({
          channelId: props.channelId,
          replicaId: props.replicaId,
          source,
          result,
          requestId,
          cursor,
          pageSize: PAGE_SIZE,
          startMS: rangeEnd - rangeDays * 24 * 60 * 60 * 1000,
          endMS: rangeEnd,
        })
      ),
    enabled: props.open,
  })
  const logs = logsQuery.data?.items ?? []

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-5xl'>
        <DialogHeader>
          <DialogTitle>{t('Request logs')}</DialogTitle>
          <DialogDescription>
            {t('Channel ID')} · {props.channelId}
            {props.replicaId ? ` · ${t('Task ID')} ${props.replicaId}` : ''}
          </DialogDescription>
        </DialogHeader>

        <div className='flex flex-wrap gap-2'>
          <NativeSelect
            value={String(rangeDays)}
            aria-label={t('Time')}
            className='w-36'
            onChange={(event) => {
              setRangeDays(Number(event.target.value))
              setRangeEnd(Date.now())
              resetPage()
            }}
          >
            <NativeSelectOption value='1'>
              {t('Last 24 hours')}
            </NativeSelectOption>
            <NativeSelectOption value='7'>
              {t('Last 7 days')}
            </NativeSelectOption>
          </NativeSelect>
          <NativeSelect
            value={source}
            aria-label={t('Source')}
            className='w-36'
            onChange={(event) => {
              setSource(event.target.value)
              resetPage()
            }}
          >
            <NativeSelectOption value=''>{t('All')}</NativeSelectOption>
            <NativeSelectOption value='test'>{t('Test')}</NativeSelectOption>
            <NativeSelectOption value='sync'>{t('Sync')}</NativeSelectOption>
          </NativeSelect>
          <NativeSelect
            value={result}
            aria-label={t('Result')}
            className='w-36'
            onChange={(event) => {
              setResult(event.target.value)
              resetPage()
            }}
          >
            <NativeSelectOption value=''>{t('All')}</NativeSelectOption>
            <NativeSelectOption value='success'>
              {t('Success')}
            </NativeSelectOption>
            <NativeSelectOption value='failure'>
              {t('Failed')}
            </NativeSelectOption>
          </NativeSelect>
          <Input
            aria-label={t('Request ID')}
            placeholder={t('Request ID')}
            className='min-w-40 flex-1'
            value={draftRequestId}
            onChange={(event) => setDraftRequestId(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === 'Enter') {
                setRequestId(draftRequestId.trim())
                resetPage()
              }
            }}
          />
          <Button
            type='button'
            variant='outline'
            onClick={() => {
              setRequestId(draftRequestId.trim())
              resetPage()
            }}
          >
            {t('Search')}
          </Button>
          <Button
            type='button'
            variant='outline'
            onClick={() => {
              setRangeEnd(Date.now())
              resetPage()
            }}
          >
            {t('Refresh')}
          </Button>
        </div>

        {(logsQuery.data?.dropped_on_this_node ?? 0) > 0 && (
          <Alert variant='destructive'>
            <AlertTitle>{t('Request logs')}</AlertTitle>
            <AlertDescription>
              {t('Some request logs were dropped on this node: {{count}}', {
                count: logsQuery.data?.dropped_on_this_node,
              })}
            </AlertDescription>
          </Alert>
        )}
        {logsQuery.isLoading && <Skeleton className='h-64 w-full' />}
        {logsQuery.isError && (
          <Alert variant='destructive'>
            <AlertTitle>{t('Unable to load request logs')}</AlertTitle>
            <AlertDescription>
              {assetErrorMessage(logsQuery.error)}
            </AlertDescription>
          </Alert>
        )}
        {!logsQuery.isLoading && !logsQuery.isError && logs.length === 0 && (
          <p className='text-muted-foreground py-10 text-center text-sm'>
            {t('No logs')}
          </p>
        )}
        {logs.length > 0 && (
          <div className='overflow-x-auto rounded-md border'>
            <Table className='min-w-[1000px]'>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('Time')}</TableHead>
                  <TableHead>{t('Operation')}</TableHead>
                  <TableHead>{t('Source')}</TableHead>
                  <TableHead>{t('Result')}</TableHead>
                  <TableHead>{t('HTTP status')}</TableHead>
                  <TableHead>{t('Duration')}</TableHead>
                  <TableHead>{t('Task ID')}</TableHead>
                  <TableHead>{t('Request ID')}</TableHead>
                  <TableHead>{t('Details')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {logs.map((log) => (
                  <Fragment key={log.id}>
                    <TableRow>
                      <TableCell className='whitespace-nowrap'>
                        {new Date(log.created_at).toLocaleString()}
                      </TableCell>
                      <TableCell>
                        <span className='font-medium'>{log.operation}</span>
                        <code className='text-muted-foreground block text-xs'>
                          {log.protocol} · {log.method} {log.path_template}
                        </code>
                      </TableCell>
                      <TableCell>
                        {log.source === 'test' ? t('Test') : t('Sync')}
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={
                            log.result === 'success'
                              ? 'secondary'
                              : 'destructive'
                          }
                        >
                          {log.result === 'success'
                            ? t('Success')
                            : t('Failed')}
                        </Badge>
                        {log.error_kind && (
                          <code className='text-muted-foreground mt-1 block text-xs'>
                            {log.error_kind}
                          </code>
                        )}
                      </TableCell>
                      <TableCell>{log.http_status || '—'}</TableCell>
                      <TableCell>
                        {log.duration_ms} {t('ms')}
                      </TableCell>
                      <TableCell>{log.replica_id || '—'}</TableCell>
                      <TableCell>
                        <code
                          className='block max-w-44 truncate'
                          title={log.request_id}
                        >
                          {log.request_id}
                        </code>
                      </TableCell>
                      <TableCell>
                        <Button
                          type='button'
                          variant='outline'
                          size='sm'
                          aria-expanded={expandedId === log.id}
                          aria-controls={`request-log-detail-${log.id}`}
                          onClick={() =>
                            setExpandedId(expandedId === log.id ? null : log.id)
                          }
                        >
                          {expandedId === log.id
                            ? t('Hide details')
                            : t('Show details')}
                        </Button>
                      </TableCell>
                    </TableRow>
                    {expandedId === log.id && (
                      <TableRow id={`request-log-detail-${log.id}`}>
                        <TableCell colSpan={9}>
                          <RequestLogDetail id={log.id} />
                        </TableCell>
                      </TableRow>
                    )}
                  </Fragment>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
        {(pageIndex > 0 || logsQuery.data?.next_cursor) && (
          <div className='flex items-center justify-between gap-2'>
            <Button
              variant='outline'
              size='sm'
              disabled={pageIndex === 0 || logsQuery.isFetching}
              onClick={() => {
                setExpandedId(null)
                setPageIndex((value) => value - 1)
              }}
            >
              {t('Previous')}
            </Button>
            <span className='text-muted-foreground text-sm tabular-nums'>
              {t('Page')} {pageIndex + 1}
            </span>
            <Button
              variant='outline'
              size='sm'
              disabled={!logsQuery.data?.next_cursor || logsQuery.isFetching}
              onClick={() => {
                const nextCursor = logsQuery.data?.next_cursor
                if (!nextCursor) return
                setExpandedId(null)
                setCursors((values) => [
                  ...values.slice(0, pageIndex + 1),
                  nextCursor,
                ])
                setPageIndex((value) => value + 1)
              }}
            >
              {t('Next')}
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
