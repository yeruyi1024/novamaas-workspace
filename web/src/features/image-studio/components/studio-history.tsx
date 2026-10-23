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
  Alert02Icon,
  Clock01Icon,
  Delete02Icon,
  Image01Icon,
  RefreshIcon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import {
  useMutation,
  useQuery,
  type UseQueryResult,
} from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

import {
  deleteStudioResult,
  deleteStudioRun,
  listStudioReferences,
} from '../api'
import {
  studioErrorMessage,
  studioStatusBadgeVariant,
  studioStatusLabelKey,
} from '../studio-utils'
import type {
  StudioLinkedAssets,
  StudioResult,
  StudioRun,
  StudioRunList,
  StudioReference,
} from '../types'
import { StudioUploadHistory } from './studio-upload-history'

export function StudioHistory(props: {
  tab: string
  onTabChange: (tab: string) => void
  authenticated: boolean
  runs: UseQueryResult<StudioRunList>
  page: number
  runId: string
  resultId?: string
  disabled: boolean
  onSelect: (run: StudioRun, resultId?: string) => void
  onPageChange: (page: number) => void
  onRemoved: () => Promise<void>
  referenceIds: string[]
  canAddReference: boolean
  uploading: boolean
  onUpload: (files: File[]) => void
  onUseReference: (reference: StudioReference) => void
  uploadId?: string
  onSelectUpload: (reference: StudioReference) => void
}) {
  const { t } = useTranslation()
  const runs = props.runs
  const tab = props.tab
  const [uploadPage, setUploadPage] = useState(1)
  const userId = useAuthStore((state) =>
    props.authenticated ? (state.auth.user?.id ?? 0) : 0
  )
  const uploads = useQuery({
    queryKey: ['image-studio', 'references', userId, uploadPage],
    queryFn: () => listStudioReferences(uploadPage),
    enabled: Boolean(userId) && tab === 'uploads',
  })
  // A run without images (for example a failed one) is removed as a whole.
  const [deleteTarget, setDeleteTarget] = useState<StudioRun | StudioResult>()
  const deletingImage = Boolean(deleteTarget && !('results' in deleteTarget))
  // Images still saved in the asset library need an explicit choice.
  const hasLinkedAsset =
    deleteTarget &&
    ('results' in deleteTarget
      ? deleteTarget.results.some((result) => Boolean(result.asset_id))
      : Boolean(deleteTarget.asset_id))
  let deleteDescription = t(
    'Delete this generation record? This cannot be undone.'
  )
  if (deletingImage) deleteDescription = t('Delete this generated image?')
  if (hasLinkedAsset) {
    deleteDescription = t(
      'This image is also saved in the asset library. Delete only the generated image, or delete the asset library copy as well?'
    )
  }
  const remove = useMutation({
    mutationFn: (linkedAssets?: StudioLinkedAssets) => {
      if (!deleteTarget) return Promise.resolve(undefined)
      return 'results' in deleteTarget
        ? deleteStudioRun(deleteTarget.id, linkedAssets)
        : deleteStudioResult(deleteTarget.id, linkedAssets)
    },
    onSuccess: async () => {
      await props.onRemoved()
      setDeleteTarget(undefined)
    },
  })
  return (
    <>
      <Tabs
        value={tab}
        onValueChange={(value) => props.onTabChange(String(value))}
        className='min-h-0 w-full flex-1 gap-3'
      >
        <div className='mx-4 flex min-h-14 shrink-0 items-center gap-2 border-b border-(--studio-subtle-border)'>
          <TabsList className='min-w-0 flex-1' variant='line'>
            <TabsTrigger value='generated'>
              {t('Generation records')}
            </TabsTrigger>
            <TabsTrigger value='uploads'>{t('Upload history')}</TabsTrigger>
          </TabsList>
          {props.authenticated && (
            <Button
              type='button'
              variant='ghost'
              size='icon-sm'
              aria-label={t('Refresh')}
              disabled={
                tab === 'uploads' ? uploads.isFetching : runs.isFetching
              }
              onClick={() => {
                if (tab === 'uploads') void uploads.refetch()
                else void runs.refetch()
              }}
            >
              <HugeiconsIcon icon={RefreshIcon} aria-hidden='true' />
            </Button>
          )}
        </div>
        <TabsContent value='generated' className='flex min-h-0 flex-1 flex-col'>
          <section
            className='flex min-h-0 w-full min-w-0 flex-1 flex-col gap-4 px-4 pb-4'
            aria-label={t('Generation records')}
          >
            {(!props.authenticated || runs.data?.items.length === 0) && (
              <Empty className='studio-history-empty flex-1 px-2.5 py-7 min-[75rem]:py-6'>
                <EmptyHeader>
                  <EmptyMedia variant='icon'>
                    <HugeiconsIcon icon={Clock01Icon} aria-hidden='true' />
                  </EmptyMedia>
                  <EmptyTitle>
                    {props.authenticated
                      ? t('No generation records')
                      : t('Your creative journey starts here')}
                  </EmptyTitle>
                  <EmptyDescription>
                    {t(
                      'Generated images will be collected here, ready to revisit and download.'
                    )}
                  </EmptyDescription>
                </EmptyHeader>
                {!props.authenticated && (
                  <Link
                    to='/sign-in'
                    search={{ redirect: '/image-studio' }}
                    className='text-primary text-sm hover:underline'
                  >
                    {t('Sign in to view your generation records')}
                  </Link>
                )}
              </Empty>
            )}
            {props.authenticated && runs.isPending && (
              <p role='status'>{t('Loading...')}</p>
            )}
            {runs.isError && <p role='alert'>{t('Failed to load data')}</p>}
            <div className='grid max-h-[480px] min-h-0 grid-cols-1 content-start gap-1.5 overflow-y-auto overscroll-contain has-[article]:flex-1 md:grid-cols-2 min-[75rem]:max-h-none min-[75rem]:grid-cols-1'>
              {runs.data?.items.flatMap((run) =>
                (run.results.length ? run.results : [undefined]).map(
                  (result) => (
                    <article
                      key={result?.id || run.id}
                      className='relative min-w-0'
                    >
                      <button
                        type='button'
                        onClick={() => props.onSelect(run, result?.id)}
                        disabled={props.disabled}
                        aria-pressed={
                          props.runId === run.id &&
                          (!result || props.resultId === result.id)
                        }
                        className='focus-visible:outline-ring flex w-full min-w-0 items-start gap-3 rounded-xl border border-transparent bg-(--studio-history-item) p-2 pr-7 text-left transition-colors hover:border-(--studio-hover-border) hover:bg-(--studio-panel) focus-visible:outline-2 focus-visible:-outline-offset-2 disabled:cursor-not-allowed disabled:opacity-55 aria-pressed:border-(--studio-accent)/60 aria-pressed:bg-(--studio-selection)'
                      >
                        {result?.thumbnail_url || result?.preview_url ? (
                          <img
                            src={result.thumbnail_url || result.preview_url}
                            alt={t('Generated image')}
                            loading='lazy'
                            onDragStart={(event) => {
                              event.dataTransfer.setData(
                                'text/uri-list',
                                result.preview_url || result.thumbnail_url || ''
                              )
                            }}
                            className='bg-background size-14 shrink-0 rounded-lg object-cover'
                          />
                        ) : (
                          <span
                            aria-hidden='true'
                            className={cn(
                              'text-muted-foreground grid size-14 shrink-0 place-items-center rounded-lg border border-dashed border-(--studio-panel-border)',
                              run.status === 'failed' && 'text-destructive'
                            )}
                          >
                            <HugeiconsIcon
                              icon={
                                run.status === 'failed'
                                  ? Alert02Icon
                                  : Image01Icon
                              }
                              className='size-5'
                            />
                          </span>
                        )}
                        <div className='flex min-w-0 flex-1 flex-col gap-1'>
                          <span
                            className='line-clamp-2 text-xs leading-relaxed break-words text-(--studio-soft-text)'
                            title={result?.description || run.prompt}
                          >
                            {result?.description ||
                              run.prompt ||
                              t(studioStatusLabelKey(run.status))}
                          </span>
                          <span
                            className='text-muted-foreground truncate text-[10px]'
                            title={run.model}
                          >
                            {run.model}
                          </span>
                          <div className='flex flex-wrap items-center justify-between gap-1'>
                            <time
                              className='text-muted-foreground text-[10px]'
                              dateTime={new Date(
                                run.created_at * 1000
                              ).toISOString()}
                            >
                              {new Date(
                                run.created_at * 1000
                              ).toLocaleDateString()}
                            </time>
                            <Badge
                              variant={studioStatusBadgeVariant(run.status)}
                            >
                              {t(studioStatusLabelKey(run.status))}
                            </Badge>
                          </div>
                        </div>
                      </button>
                      <Button
                        type='button'
                        size='icon-xs'
                        variant='ghost'
                        aria-label={
                          result ? t('Delete image') : t('Delete record')
                        }
                        title={result ? t('Delete image') : t('Delete record')}
                        className='absolute top-2 right-1'
                        disabled={
                          props.disabled ||
                          remove.isPending ||
                          run.status === 'running'
                        }
                        onClick={() => {
                          remove.reset()
                          setDeleteTarget(result ?? run)
                        }}
                      >
                        <HugeiconsIcon icon={Delete02Icon} aria-hidden='true' />
                      </Button>
                    </article>
                  )
                )
              )}
            </div>
            {props.authenticated && (
              <StudioRunPagination
                page={props.page}
                runs={runs}
                onPageChange={props.onPageChange}
              />
            )}
          </section>
        </TabsContent>
        <TabsContent value='uploads' className='flex min-h-0 flex-1 flex-col'>
          <StudioUploadHistory
            authenticated={props.authenticated}
            uploads={uploads}
            selectedId={props.uploadId}
            disabled={props.disabled}
            onSelect={props.onSelectUpload}
            referenceIds={props.referenceIds}
            canAddReference={props.canAddReference}
            uploading={props.uploading}
            onUpload={props.onUpload}
            onUseReference={props.onUseReference}
          />
          {props.authenticated && (
            <div className='px-4 pb-4'>
              <StudioRunPagination
                page={uploadPage}
                runs={uploads}
                onPageChange={setUploadPage}
              />
            </div>
          )}
        </TabsContent>
      </Tabs>
      <StudioDeleteDialog
        open={Boolean(deleteTarget)}
        busy={remove.isPending}
        title={deletingImage ? t('Delete image') : t('Delete record')}
        description={
          <>
            {deleteDescription}
            {deletingImage && (
              <span className='mt-2 block'>
                {t(
                  'Deleted images cannot be restored. Files are removed in the background.'
                )}
              </span>
            )}
          </>
        }
        error={remove.error}
        onClose={() => setDeleteTarget(undefined)}
      >
        {hasLinkedAsset ? (
          <>
            <Button
              variant='outline'
              disabled={remove.isPending}
              onClick={() => remove.mutate('keep')}
            >
              {t('Delete generated image only')}
            </Button>
            <Button
              variant='destructive'
              disabled={remove.isPending}
              onClick={() => remove.mutate('delete')}
            >
              {t('Also delete asset library image')}
            </Button>
          </>
        ) : (
          <Button
            variant='destructive'
            disabled={remove.isPending}
            onClick={() => remove.mutate(undefined)}
          >
            {remove.isPending ? t('Deleting...') : t('Delete')}
          </Button>
        )}
      </StudioDeleteDialog>
    </>
  )
}

export function StudioRunPagination(props: {
  page: number
  runs: { data?: { total: number }; isFetching: boolean }
  onPageChange: (page: number) => void
  showPage?: boolean
}) {
  const { t } = useTranslation()
  return (
    <div className='flex shrink-0 items-center justify-end gap-2'>
      <Button
        variant='outline'
        disabled={props.page === 1 || props.runs.isFetching}
        onClick={() => props.onPageChange(props.page - 1)}
      >
        {t('Previous')}
      </Button>
      {props.showPage && <span>{props.page}</span>}
      <Button
        variant='outline'
        disabled={
          !props.runs.data ||
          props.page * 12 >= props.runs.data.total ||
          props.runs.isFetching
        }
        onClick={() => props.onPageChange(props.page + 1)}
      >
        {t('Next')}
      </Button>
    </div>
  )
}

export function StudioDeleteDialog(props: {
  open: boolean
  busy: boolean
  title: string
  description: ReactNode
  error: unknown
  onClose: () => void
  children: ReactNode
}) {
  const { t } = useTranslation()
  return (
    <AlertDialog
      open={props.open}
      onOpenChange={(open) => {
        if (!open && !props.busy) props.onClose()
      }}
    >
      <AlertDialogContent className='image-studio-theme'>
        <AlertDialogHeader className='text-start'>
          <AlertDialogTitle>{props.title}</AlertDialogTitle>
          <AlertDialogDescription>{props.description}</AlertDialogDescription>
        </AlertDialogHeader>
        {Boolean(props.error) && (
          <p role='alert' className='text-destructive text-sm'>
            {studioErrorMessage(props.error)}
          </p>
        )}
        <AlertDialogFooter>
          <AlertDialogCancel disabled={props.busy}>
            {t('Cancel')}
          </AlertDialogCancel>
          {props.children}
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
