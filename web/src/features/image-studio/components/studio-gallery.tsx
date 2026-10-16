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
  Download01Icon,
  FolderAddIcon,
  Image01Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { cn } from '@/lib/utils'

import { getStudioRun } from '../api'
import {
  downloadStudioImage,
  studioErrorMessage,
  studioStatusLabelKey,
  studioTaskName,
} from '../studio-utils'
import type { StudioResult, StudioRun } from '../types'

const canvasClassName =
  'mx-auto aspect-[4/3] min-h-48 w-full max-w-[48rem] flex-none'

export function StudioGallery(props: {
  run?: StudioRun
  resultId?: string
  onReusePrompt?: (prompt: string) => void
  reuseDisabled?: boolean
  ownerName?: string
  onSave?: (result: StudioResult) => void
  saving?: boolean
  compact?: boolean
}) {
  const { t } = useTranslation()
  const [preview, setPreview] = useState<StudioResult>()
  const [unavailable, setUnavailable] = useState<string[]>([])
  const layout = props.compact
    ? 'contents'
    : 'flex min-h-0 min-w-0 flex-col gap-3'
  const run = props.run
  const runId = run?.id || ''
  const original = useQuery({
    queryKey: ['image-studio', 'review-original', runId],
    queryFn: () => getStudioRun(runId, 'all'),
    enabled: Boolean(props.compact && preview && !preview.preview_url),
    gcTime: 0,
  })
  const previewURL =
    preview?.preview_url ||
    original.data?.results.find((result) => result.id === preview?.id)
      ?.preview_url
  const results =
    run?.results.filter(
      (result) => !props.resultId || result.id === props.resultId
    ) ?? []

  if (!run) {
    return (
      <Empty
        role='region'
        aria-label={t('Generation results')}
        className={cn(
          canvasClassName,
          'bg-muted/30 relative overflow-hidden rounded-2xl border'
        )}
      >
        <p className='text-muted-foreground absolute top-5 left-5 text-xs font-medium'>
          {t('Generation results')}
        </p>
        <EmptyHeader>
          <EmptyMedia
            variant='icon'
            aria-hidden='true'
            className='size-12 rounded-xl'
          >
            <HugeiconsIcon
              icon={Image01Icon}
              className='size-6'
              aria-hidden='true'
            />
          </EmptyMedia>
          <EmptyTitle className='text-base'>
            {t('Your image will appear here')}
          </EmptyTitle>
          <EmptyDescription>
            {t('Choose a prompt or try an example to start creating.')}
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  return (
    <section className={layout} aria-label={t('Generation results')}>
      {!props.compact && (
        <div className='flex shrink-0 flex-wrap items-center justify-between gap-2'>
          <p className='text-muted-foreground mr-auto text-xs font-medium'>
            {t('Generation results')}
          </p>
          {props.onReusePrompt && run.prompt ? (
            <Button
              type='button'
              size='sm'
              variant='outline'
              disabled={props.reuseDisabled}
              onClick={() => props.onReusePrompt?.(run.prompt || '')}
            >
              {t('Use this prompt')}
            </Button>
          ) : null}
          <Badge variant='secondary'>
            {t(studioStatusLabelKey(run.status))}
          </Badge>
        </div>
      )}
      {!props.compact && run.status === 'running' && (
        <p role='status' className='text-muted-foreground text-sm'>
          {t('This generation is still in progress')}
        </p>
      )}
      {!props.compact && run.error_message && (
        <p role='alert' className='text-destructive text-sm'>
          {run.error_message}
        </p>
      )}
      {run.results.length === 0 && (
        <Empty
          className={cn('border', props.compact ? 'min-h-24' : canvasClassName)}
        >
          <EmptyHeader>
            <EmptyTitle>{t(studioStatusLabelKey(run.status))}</EmptyTitle>
            {props.compact && (
              <StudioRunSummary run={run} ownerName={props.ownerName} compact />
            )}
            <EmptyDescription>
              {run.error_message || run.prompt}
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
      <div className={layout}>
        {results.map((result) => {
          const imageURL = props.compact
            ? result.thumbnail_url || result.preview_url
            : result.preview_url
          return (
            <article
              key={result.id}
              className={cn(
                'relative flex min-w-0 flex-col gap-3',
                props.compact ? 'bg-card rounded-xl border p-2' : 'min-h-0'
              )}
            >
              {imageURL && !unavailable.includes(imageURL) ? (
                <button
                  type='button'
                  aria-label={t('Preview')}
                  onClick={() => setPreview(result)}
                  className={cn(
                    'bg-muted/30 focus-visible:outline-ring rounded-xl focus-visible:outline-2',
                    props.compact
                      ? ''
                      : cn(
                          canvasClassName,
                          'relative flex items-center justify-center overflow-hidden border'
                        )
                  )}
                >
                  <img
                    src={imageURL}
                    alt={
                      result.description || run.prompt || t('Generated image')
                    }
                    loading='lazy'
                    className={cn(
                      'w-full rounded-lg object-contain',
                      props.compact ? 'h-40' : 'absolute inset-0 h-full'
                    )}
                    onError={(event) => {
                      const url = event.currentTarget.src
                      setUnavailable((current) => [...current, url])
                    }}
                  />
                </button>
              ) : (
                <Empty className={props.compact ? 'h-40' : canvasClassName}>
                  <EmptyHeader>
                    <EmptyTitle>
                      {result.expired
                        ? t('Image expired')
                        : t('Preview unavailable')}
                    </EmptyTitle>
                  </EmptyHeader>
                </Empty>
              )}
              <div
                className={
                  props.compact
                    ? 'absolute top-3 right-3'
                    : 'flex shrink-0 flex-wrap items-center justify-end gap-2'
                }
              >
                {!props.compact && (
                  <p className='text-muted-foreground mr-auto text-xs'>
                    {result.asset_id
                      ? t('Saved to asset library')
                      : t('Saved in generation records')}
                  </p>
                )}
                <Button
                  type='button'
                  size={props.compact ? 'icon-sm' : 'sm'}
                  variant='outline'
                  aria-label={t('Download')}
                  disabled={!imageURL}
                  onClick={async () => {
                    try {
                      let image = result
                      if (!result.download_url) {
                        const detail = await getStudioRun(run.id, 'all')
                        const found = detail.results.find(
                          (value) => value.id === result.id
                        )
                        if (!found) throw new Error(t('Preview unavailable'))
                        image = found
                      }
                      downloadStudioImage(image)
                    } catch (error) {
                      toast.error(studioErrorMessage(error))
                    }
                  }}
                >
                  <HugeiconsIcon icon={Download01Icon} aria-hidden='true' />
                  {!props.compact && t('Download')}
                </Button>
                {props.onSave && !result.asset_id && (
                  <Button
                    type='button'
                    size='sm'
                    variant='outline'
                    aria-describedby={
                      props.compact
                        ? undefined
                        : `${result.id}-asset-library-help`
                    }
                    disabled={props.saving || result.expired}
                    onClick={() => props.onSave?.(result)}
                  >
                    <HugeiconsIcon icon={FolderAddIcon} aria-hidden='true' />
                    {t('Save to asset library (for video generation)')}
                  </Button>
                )}
              </div>
              {!props.compact && props.onSave && !result.asset_id && (
                <p
                  id={`${result.id}-asset-library-help`}
                  className='text-muted-foreground shrink-0 text-xs leading-relaxed'
                >
                  {t(
                    'Images are saved in the generation records on the right for 30 days. Download them before expiry. Save to the asset library only for direct use in video generation; library copies are kept separately.'
                  )}
                </p>
              )}
              {props.compact && (
                <StudioRunSummary
                  run={run}
                  ownerName={props.ownerName}
                  compact
                />
              )}
              {(result.description || run.prompt) && (
                <p
                  className={cn(
                    'text-muted-foreground break-words',
                    props.compact
                      ? 'line-clamp-2 text-xs'
                      : 'max-h-20 shrink-0 overflow-y-auto whitespace-pre-wrap text-xs'
                  )}
                  title={result.description || run.prompt}
                >
                  {result.description || run.prompt}
                </p>
              )}
            </article>
          )
        })}
      </div>
      <Dialog
        open={Boolean(preview)}
        onOpenChange={(open) => {
          if (!open) setPreview(undefined)
        }}
      >
        <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-4xl'>
          <DialogHeader>
            <DialogTitle>{t('Generated image')}</DialogTitle>
          </DialogHeader>
          {original.isFetching && !previewURL && (
            <p role='status'>{t('Loading...')}</p>
          )}
          {original.isError && <p role='alert'>{t('Failed to load data')}</p>}
          {previewURL && (
            <img
              src={previewURL}
              alt={preview?.description || run.prompt || t('Generated image')}
              className='max-h-[75vh] w-full object-contain'
            />
          )}
        </DialogContent>
      </Dialog>
    </section>
  )
}

export function StudioRunSummary(props: {
  run: StudioRun
  ownerName?: string
  compact?: boolean
}) {
  const { t } = useTranslation()
  const run = props.run
  return (
    <div className='min-w-0 flex-1 space-y-1'>
      <p
        className={cn('truncate font-semibold', props.compact && 'text-sm')}
        title={studioTaskName(run)}
      >
        {studioTaskName(run)}
      </p>
      <p className='text-muted-foreground truncate text-xs' title={run.id}>
        {t('Task ID')}: {run.id}
      </p>
      <div className='flex flex-wrap items-center justify-between gap-1'>
        <p className='text-muted-foreground truncate text-xs'>
          {props.ownerName || run.username} · {t('User ID')}:{' '}
          {run.owner_user_id}
          {!props.compact && ` · ${run.model}`}
        </p>
        {props.compact && (
          <Badge variant='secondary'>
            {t(studioStatusLabelKey(run.status))}
          </Badge>
        )}
      </div>
    </div>
  )
}
