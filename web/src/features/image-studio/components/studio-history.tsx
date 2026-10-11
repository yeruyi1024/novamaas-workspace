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
import { Delete02Icon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useMutation, type UseQueryResult } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { Button } from '@/components/ui/button'

import { deleteStudioResult } from '../api'
import { studioErrorMessage } from '../studio-utils'
import type { StudioResult, StudioRun, StudioRunList } from '../types'
import { StudioRunSummary } from './studio-gallery'

export function StudioHistory(props: {
  runs: UseQueryResult<StudioRunList>
  page: number
  runId: string
  resultId?: string
  ownerName: string
  disabled: boolean
  onSelect: (run: StudioRun, resultId?: string) => void
  onPageChange: (page: number) => void
  onRemoved: () => Promise<void>
}) {
  const { t } = useTranslation()
  const runs = props.runs
  const page = props.page
  const runId = props.runId
  const resultId = props.resultId
  const ownerName = props.ownerName
  const [deleteTarget, setDeleteTarget] = useState<StudioResult>()
  const remove = useMutation({
    mutationFn: (id: string) => deleteStudioResult(id),
    onSuccess: async () => {
      await props.onRemoved()
      setDeleteTarget(undefined)
    },
  })
  return (
    <>
      <section
        className='bg-card/90 flex min-h-0 min-w-0 flex-col gap-3 rounded-2xl border p-3 shadow-sm xl:col-start-3'
        aria-label={t('Generation records')}
      >
        <div className='flex shrink-0 items-center justify-between gap-3'>
          <h2 className='font-semibold'>{t('Generation records')}</h2>
          <Button
            variant='ghost'
            disabled={runs.isFetching}
            onClick={() => void runs.refetch()}
          >
            {t('Refresh')}
          </Button>
        </div>
        <p className='text-muted-foreground text-xs'>
          {t('Images are saved here automatically.')}
        </p>
        {runs.isPending && <p role='status'>{t('Loading...')}</p>}
        {runs.isError && <p role='alert'>{t('Failed to load data')}</p>}
        {runs.data?.items.length === 0 && (
          <p className='text-muted-foreground text-sm'>
            {t('No generation records')}
          </p>
        )}
        <div className='grid max-h-[28rem] min-h-0 grid-cols-2 content-start gap-2 overflow-y-auto xl:max-h-none xl:flex-1 xl:grid-cols-1'>
          {runs.data?.items.flatMap((run) =>
            (run.results.length ? run.results : [undefined]).map((result) => (
              <article key={result?.id || run.id} className='relative min-w-0'>
                <button
                  type='button'
                  onClick={() => props.onSelect(run, result?.id)}
                  aria-pressed={
                    runId === run.id && (!result || resultId === result.id)
                  }
                  className='hover:bg-muted/60 aria-pressed:border-primary/60 aria-pressed:bg-primary/5 focus-visible:outline-ring flex w-full min-w-0 items-start gap-2 rounded-xl border p-2 pr-8 text-left transition-colors focus-visible:outline-2'
                >
                  {(result?.thumbnail_url || result?.preview_url) && (
                    <img
                      src={result.thumbnail_url || result.preview_url}
                      alt={t('Generated image')}
                      loading='lazy'
                      className='bg-muted size-14 shrink-0 rounded object-contain'
                    />
                  )}
                  <div className='flex min-w-0 flex-1 flex-col gap-1'>
                    <StudioRunSummary run={run} ownerName={ownerName} compact />
                    <span
                      className='text-muted-foreground line-clamp-2 text-xs break-words'
                      title={result?.description || run.prompt}
                    >
                      {result?.description || run.prompt}
                    </span>
                    <span className='text-muted-foreground text-[11px]'>
                      {new Date(run.created_at * 1000).toLocaleDateString()}
                    </span>
                  </div>
                </button>
                {result && (
                  <Button
                    type='button'
                    size='icon-xs'
                    variant='ghost'
                    aria-label={t('Delete image')}
                    title={t('Delete image')}
                    className='absolute top-2 right-1'
                    disabled={props.disabled || remove.isPending}
                    onClick={() => {
                      remove.reset()
                      setDeleteTarget(result)
                    }}
                  >
                    <HugeiconsIcon icon={Delete02Icon} aria-hidden='true' />
                  </Button>
                )}
              </article>
            ))
          )}
        </div>
        <StudioRunPagination
          page={page}
          runs={runs}
          onPageChange={props.onPageChange}
        />
      </section>
      <ConfirmDialog
        open={Boolean(deleteTarget)}
        onOpenChange={(open) => {
          if (!open && !remove.isPending) setDeleteTarget(undefined)
        }}
        title={t('Delete image')}
        desc={t(
          'Delete this generated image? Images already saved to the asset library are kept.'
        )}
        destructive
        isLoading={remove.isPending}
        confirmText={remove.isPending ? t('Deleting...') : t('Delete')}
        handleConfirm={() => {
          if (deleteTarget) remove.mutate(deleteTarget.id)
        }}
      >
        {remove.isError && (
          <p role='alert' className='text-destructive text-sm'>
            {studioErrorMessage(remove.error)}
          </p>
        )}
      </ConfirmDialog>
    </>
  )
}

export function StudioRunPagination(props: {
  page: number
  runs: UseQueryResult<StudioRunList>
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
