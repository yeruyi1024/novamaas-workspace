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
import {
  AlertCircleIcon,
  ArrowLeft01Icon,
  ArrowRight01Icon,
  CheckmarkCircle02Icon,
  Clock01Icon,
  Refresh01Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { TFunction } from 'i18next'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from '@/components/ui/empty'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { listAssetSyncJobs, retryAssetSyncJob } from '@/features/assets/api'
import {
  assertAssetSuccess,
  assetErrorMessage,
} from '@/features/assets/asset-utils'
import type { AssetSyncJob } from '@/features/assets/types'

import { listAssetLibraryConnections } from './api'
import { AssetRequestLogDialog } from './asset-request-log-dialog'
import { AssetSyncLibraryOverview } from './asset-sync-library-overview'

const PAGE_SIZE = 20
const JOBS_QUERY_KEY = ['asset-library', 'admin', 'sync-jobs'] as const

function statusVariant(status: string) {
  if (status === 'active') {
    return 'default' as const
  }
  if (status === 'failed' || status === 'rejected') {
    return 'destructive' as const
  }
  return 'secondary' as const
}

function formatTime(value: number) {
  if (!value) return '—'
  return new Date(value * 1000).toLocaleString()
}

function syncStatusLabel(t: TFunction, status: string) {
  switch (status) {
    case 'pending':
      return t('Pending synchronization')
    case 'syncing':
      return t('Synchronizing')
    case 'processing':
      return t('Upstream processing')
    case 'active':
      return t('Synchronized')
    case 'failed':
      return t('Synchronization failed')
    case 'rejected':
      return t('Asset rejected by upstream')
    case 'deleting':
      return t('Deleting upstream copy')
    default:
      return status
  }
}

function assetTypeLabel(t: TFunction, type: string) {
  if (type === 'video') return t('Video')
  if (type === 'audio') return t('Audio')
  return t('Image')
}

function SyncSummaryCard(props: {
  title: string
  value: number
  icon: typeof Clock01Icon
}) {
  return (
    <Card size='sm'>
      <CardHeader className='flex-row items-center justify-between gap-2'>
        <CardTitle className='text-sm font-medium'>{props.title}</CardTitle>
        <HugeiconsIcon
          icon={props.icon}
          className='text-muted-foreground size-4'
        />
      </CardHeader>
      <CardContent>
        <p className='text-2xl font-semibold tabular-nums'>{props.value}</p>
      </CardContent>
    </Card>
  )
}

function SyncStatusCell(props: {
  job: AssetSyncJob
  retrying: boolean
  onRetry: (id: number) => void
  onViewLogs: (channelId: number, replicaId: number, libraryId?: number) => void
}) {
  const { t } = useTranslation()
  const inProgress = ['pending', 'syncing', 'processing', 'deleting'].includes(
    props.job.status
  )
  let rejectionMessage = t('The upstream provider rejected this asset.')
  if (props.job.last_error === 'real_person') {
    rejectionMessage = t(
      'Real-person content was rejected by the upstream provider.'
    )
  } else if (props.job.last_error === 'sensitive_content') {
    rejectionMessage = t(
      'Sensitive content was rejected by the upstream provider.'
    )
  }
  return (
    <div className='flex min-w-0 flex-col gap-2 whitespace-normal'>
      <div className='flex flex-wrap items-center gap-2'>
        <Badge variant={statusVariant(props.job.status)}>
          {syncStatusLabel(t, props.job.status)}
        </Badge>
        <span className='text-muted-foreground text-xs tabular-nums'>
          {props.job.progress}%
        </span>
        <Button
          size='xs'
          variant='outline'
          className='ml-auto'
          onClick={() =>
            props.onViewLogs(
              props.job.channel_id,
              props.job.id,
              props.job.library_id
            )
          }
        >
          {t('View logs')}
        </Button>
        {props.job.status === 'failed' ? (
          <Button
            size='xs'
            variant='outline'
            disabled={props.retrying}
            onClick={() => props.onRetry(props.job.id)}
          >
            <HugeiconsIcon icon={Refresh01Icon} data-icon='inline-start' />
            {t('Retry')}
          </Button>
        ) : null}
      </div>
      {inProgress ? <Progress value={props.job.progress} /> : null}
      {props.job.status === 'rejected' && (
        <span className='text-destructive text-xs leading-relaxed'>
          {rejectionMessage}{' '}
          {t('Upload revised material and use its new asset ID.')}
        </span>
      )}
      {props.job.status !== 'rejected' && props.job.last_error && (
        <span
          className='text-destructive line-clamp-2 text-xs leading-relaxed'
          title={props.job.last_error}
        >
          {props.job.last_error}
        </span>
      )}
    </div>
  )
}

function AssetIdentityCell(props: { job: AssetSyncJob }) {
  const { t } = useTranslation()
  return (
    <div className='min-w-0 whitespace-normal'>
      <p className='truncate font-medium' title={props.job.asset_name}>
        {props.job.asset_name}
      </p>
      <div className='mt-1 flex min-w-0 items-center gap-2'>
        <Badge variant='outline' className='shrink-0 font-normal'>
          {assetTypeLabel(t, props.job.asset_type)}
        </Badge>
        <span
          className='text-muted-foreground truncate text-xs'
          title={props.job.group_name}
        >
          {props.job.group_name}
        </span>
      </div>
      <code
        className='text-muted-foreground mt-1.5 block truncate text-[11px]'
        title={props.job.asset_id}
      >
        {props.job.asset_id}
      </code>
    </div>
  )
}

export function AssetSyncManagement(props: { initialLibraryId?: number } = {}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [libraryId, setLibraryId] = useState(props.initialLibraryId ?? 0)
  const [showDetails, setShowDetails] = useState(
    (props.initialLibraryId ?? 0) > 0
  )
  const [status, setStatus] = useState('')
  const [logTarget, setLogTarget] = useState<{
    channelId: number
    libraryId?: number
    replicaId: number
  } | null>(null)
  const librariesQuery = useQuery({
    queryKey: ['asset-library', 'admin', 'libraries'],
    queryFn: async () =>
      assertAssetSuccess(await listAssetLibraryConnections()),
    staleTime: 60_000,
    refetchInterval: (query) =>
      query.state.data?.some(
        (library) =>
          library.channels.length > 0 &&
          library.summary.pending + library.summary.processing > 0
      )
        ? 5000
        : false,
  })
  const jobsQuery = useQuery({
    queryKey: [...JOBS_QUERY_KEY, page, libraryId, status],
    enabled: showDetails,
    queryFn: async () =>
      assertAssetSuccess(
        await listAssetSyncJobs({
          page,
          pageSize: PAGE_SIZE,
          libraryId: libraryId || undefined,
          status: status || undefined,
        })
      ),
    refetchInterval: (query) => {
      const summary = query.state.data?.summary
      return summary && summary.pending + summary.processing > 0 ? 5000 : false
    },
  })
  const retryMutation = useMutation({
    mutationFn: async (id: number) =>
      assertAssetSuccess(await retryAssetSyncJob(id)),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: JOBS_QUERY_KEY })
      toast.success(t('Synchronization retry queued'))
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })

  const libraries = librariesQuery.data ?? []
  const jobs = jobsQuery.data?.items ?? []
  const summary = jobsQuery.data?.summary ?? {
    total: 0,
    pending: 0,
    processing: 0,
    active: 0,
    failed: 0,
  }
  const pageCount = Math.max(
    1,
    Math.ceil((jobsQuery.data?.total ?? 0) / PAGE_SIZE)
  )

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        {t('Asset synchronization')}
      </SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='flex flex-col gap-4 pb-4'>
          <Alert>
            <AlertTitle>{t('Shared library synchronization')}</AlertTitle>
            <AlertDescription>
              {t(
                'Uploads synchronize once per linked resource account. Select an unlinked library to inspect its historical synchronization records.'
              )}
            </AlertDescription>
          </Alert>

          {!showDetails && librariesQuery.isPending && (
            <Skeleton className='h-48 w-full' />
          )}
          {!showDetails && librariesQuery.isError && (
            <Alert variant='destructive'>
              <AlertTitle>
                {t('Unable to load asset library settings')}
              </AlertTitle>
              <AlertDescription>
                {assetErrorMessage(librariesQuery.error)}
              </AlertDescription>
            </Alert>
          )}
          {!showDetails && librariesQuery.data && (
            <AssetSyncLibraryOverview
              libraries={libraries}
              onSelect={(id) => {
                setLibraryId(id)
                setPage(1)
                setShowDetails(true)
              }}
            />
          )}
          {showDetails && (
            <>
              <div>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={() => setShowDetails(false)}
                >
                  {t('Library overview')}
                </Button>
              </div>
              <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-4'>
                <SyncSummaryCard
                  title={t('Pending synchronization')}
                  value={summary.pending}
                  icon={Clock01Icon}
                />
                <SyncSummaryCard
                  title={t('Synchronizing')}
                  value={summary.processing}
                  icon={Refresh01Icon}
                />
                <SyncSummaryCard
                  title={t('Synchronized')}
                  value={summary.active}
                  icon={CheckmarkCircle02Icon}
                />
                <SyncSummaryCard
                  title={t('Synchronization failed')}
                  value={summary.failed}
                  icon={AlertCircleIcon}
                />
              </div>

              <Card>
                <CardContent className='flex flex-col gap-3 py-4 md:flex-row md:items-center'>
                  <NativeSelect
                    value={String(libraryId)}
                    aria-label={t('Channel')}
                    className='md:w-60'
                    onChange={(event) => {
                      setPage(1)
                      setLibraryId(Number(event.target.value))
                    }}
                  >
                    <NativeSelectOption value='0'>
                      {t('Linked asset libraries')}
                    </NativeSelectOption>
                    {libraries.map((library) => (
                      <NativeSelectOption
                        key={library.id}
                        value={String(library.id)}
                      >
                        {library.provider_name} · {library.name} · #{library.id}
                      </NativeSelectOption>
                    ))}
                  </NativeSelect>
                  <NativeSelect
                    value={status}
                    aria-label={t('Status')}
                    className='md:w-52'
                    onChange={(event) => {
                      setPage(1)
                      setStatus(event.target.value)
                    }}
                  >
                    <NativeSelectOption value=''>
                      {t('All statuses')}
                    </NativeSelectOption>
                    {[
                      'pending',
                      'syncing',
                      'processing',
                      'active',
                      'failed',
                      'rejected',
                      'deleting',
                    ].map((value) => (
                      <NativeSelectOption key={value} value={value}>
                        {syncStatusLabel(t, value)}
                      </NativeSelectOption>
                    ))}
                  </NativeSelect>
                  {(summary.pending > 0 || summary.processing > 0) && (
                    <span className='text-muted-foreground flex items-center gap-2 text-xs md:ml-auto'>
                      <Spinner />
                      {t('Updates automatically')}
                    </span>
                  )}
                </CardContent>
              </Card>

              {jobsQuery.isLoading && <Skeleton className='h-80 w-full' />}
              {jobsQuery.isError && (
                <Alert variant='destructive'>
                  <AlertTitle>
                    {t('Unable to load synchronization jobs')}
                  </AlertTitle>
                  <AlertDescription>
                    {assetErrorMessage(jobsQuery.error)}
                  </AlertDescription>
                </Alert>
              )}
              {!jobsQuery.isLoading &&
                !jobsQuery.isError &&
                jobs.length === 0 && (
                  <Empty className='min-h-56 border'>
                    <EmptyHeader>
                      <EmptyTitle>{t('No synchronization jobs')}</EmptyTitle>
                      <EmptyDescription>
                        {t(
                          'Jobs appear here automatically when assets are uploaded to an enabled channel.'
                        )}
                      </EmptyDescription>
                    </EmptyHeader>
                  </Empty>
                )}
              {jobs.length > 0 && (
                <Card className='overflow-hidden'>
                  <Table className='min-w-[1080px] table-fixed'>
                    <colgroup>
                      <col className='w-[300px]' />
                      <col className='w-[170px]' />
                      <col className='w-[230px]' />
                      <col className='w-[240px]' />
                      <col className='w-[170px]' />
                    </colgroup>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t('Asset')}</TableHead>
                        <TableHead>{t('Owner')}</TableHead>
                        <TableHead>{t('Upstream asset library')}</TableHead>
                        <TableHead>{t('Status')}</TableHead>
                        <TableHead>{t('Last updated')}</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {jobs.map((job) => (
                        <TableRow
                          key={job.id}
                          className='[contain-intrinsic-size:76px] [content-visibility:auto]'
                        >
                          <TableCell className='py-3 align-middle'>
                            <AssetIdentityCell job={job} />
                          </TableCell>
                          <TableCell className='py-3 align-middle whitespace-normal'>
                            <div className='min-w-0'>
                              <p
                                className='truncate font-medium'
                                title={
                                  job.owner_name || `#${job.owner_user_id}`
                                }
                              >
                                {job.owner_name || `#${job.owner_user_id}`}
                              </p>
                              <p className='text-muted-foreground text-xs'>
                                {t('User ID')} · {job.owner_user_id}
                              </p>
                            </div>
                          </TableCell>
                          <TableCell className='py-3 align-middle whitespace-normal'>
                            <div className='min-w-0'>
                              <p
                                className='truncate font-medium'
                                title={
                                  job.library_name ||
                                  job.channel_name ||
                                  `#${job.channel_id}`
                                }
                              >
                                {job.library_name ||
                                  job.channel_name ||
                                  `#${job.channel_id}`}
                              </p>
                              <p className='text-muted-foreground text-xs'>
                                {job.library_id
                                  ? t('Upstream asset library ID')
                                  : t('Channel ID')}{' '}
                                · {job.library_id || job.channel_id}
                                {job.channels && job.channels.length > 0 && (
                                  <span className='mt-1 block'>
                                    {job.channels
                                      .map((channel) => channel.name)
                                      .join(', ')}
                                  </span>
                                )}
                              </p>
                            </div>
                          </TableCell>
                          <TableCell className='py-3 align-middle'>
                            <SyncStatusCell
                              job={job}
                              retrying={retryMutation.isPending}
                              onRetry={(id) => retryMutation.mutate(id)}
                              onViewLogs={(
                                selectedChannelId,
                                replicaId,
                                selectedLibraryId
                              ) =>
                                setLogTarget({
                                  channelId: selectedChannelId,
                                  libraryId: selectedLibraryId,
                                  replicaId,
                                })
                              }
                            />
                          </TableCell>
                          <TableCell className='py-3 align-middle whitespace-normal'>
                            <div className='min-w-0'>
                              <p className='font-medium'>
                                {formatTime(job.updated_at)}
                              </p>
                              <div className='text-muted-foreground mt-1 flex flex-wrap gap-x-3 gap-y-1 text-xs'>
                                <span>
                                  {t('Task ID')} · {job.id}
                                </span>
                                <span>
                                  {t('Attempts')} · {job.attempts}
                                </span>
                              </div>
                            </div>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </Card>
              )}

              {jobsQuery.data && jobsQuery.data.total > PAGE_SIZE && (
                <div className='flex items-center justify-between gap-3'>
                  <span className='text-muted-foreground text-xs'>
                    {t('Page {{page}} of {{pages}}', {
                      page,
                      pages: pageCount,
                    })}
                  </span>
                  <div className='flex gap-2'>
                    <Button
                      size='sm'
                      variant='outline'
                      disabled={page <= 1}
                      onClick={() => setPage((value) => Math.max(1, value - 1))}
                    >
                      <HugeiconsIcon
                        icon={ArrowLeft01Icon}
                        data-icon='inline-start'
                      />
                      {t('Previous')}
                    </Button>
                    <Button
                      size='sm'
                      variant='outline'
                      disabled={page >= pageCount}
                      onClick={() => setPage((value) => value + 1)}
                    >
                      {t('Next')}
                      <HugeiconsIcon
                        icon={ArrowRight01Icon}
                        data-icon='inline-end'
                      />
                    </Button>
                  </div>
                </div>
              )}
              {logTarget !== null && (
                <AssetRequestLogDialog
                  open
                  channelId={
                    logTarget.libraryId ? undefined : logTarget.channelId
                  }
                  libraryId={logTarget.libraryId}
                  replicaId={logTarget.replicaId}
                  onOpenChange={(open) => !open && setLogTarget(null)}
                />
              )}
            </>
          )}
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
