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
  assertStudioSuccess,
  downloadStudioImage,
  studioErrorMessage,
  studioStatusLabelKey,
  studioTaskName,
} from '../studio-utils'
import type { StudioResult, StudioRun } from '../types'

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
  const run = props.run
  const runId = run?.id || ''
  const original = useQuery({
    queryKey: ['image-studio', 'review-original', runId],
    queryFn: async () => assertStudioSuccess(await getStudioRun(runId, 'all')),
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
        className='from-primary/5 via-background to-muted/40 relative isolate min-h-80 overflow-hidden rounded-2xl border border-solid bg-gradient-to-br'
      >
        <div
          aria-hidden='true'
          className='pointer-events-none absolute inset-0 -z-10 [background-image:radial-gradient(var(--border)_1px,transparent_1px)] [background-size:20px_20px] opacity-40'
        />
        <div
          aria-hidden='true'
          className='pointer-events-none absolute inset-4 rounded-xl border border-dashed'
        />
        <p className='text-muted-foreground absolute top-7 left-7 text-xs font-medium'>
          {t('Generation results')}
        </p>
        <EmptyHeader className='relative gap-3 py-6'>
          <EmptyMedia aria-hidden='true' className='relative h-32 w-44'>
            <div className='bg-primary/5 absolute left-3 h-24 w-20 -rotate-12 rounded-xl border shadow-sm' />
            <div className='bg-muted absolute right-3 h-24 w-20 rotate-12 rounded-xl border shadow-sm' />
            <div className='from-background to-primary/10 text-primary relative flex size-24 items-center justify-center rounded-2xl border bg-gradient-to-br shadow-sm'>
              <HugeiconsIcon
                icon={Image01Icon}
                className='size-9'
                aria-hidden='true'
              />
            </div>
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
    <section
      className={
        props.compact
          ? 'contents'
          : 'flex min-h-0 min-w-0 flex-1 flex-col gap-3'
      }
      aria-label={t('Generation results')}
    >
      {!props.compact && (
        <div className='flex shrink-0 flex-wrap items-center justify-between gap-2'>
          <div className='min-w-0 flex-1'>
            <h2 className='truncate font-semibold' title={studioTaskName(run)}>
              {studioTaskName(run)}
            </h2>
            <p className='text-muted-foreground mt-1 text-xs break-all'>
              {t('Task ID')}: {run.id}
            </p>
            <p className='text-muted-foreground mt-1 text-xs'>
              {props.ownerName} · {t('User ID')}: {run.owner_user_id} ·{' '}
              {run.model}
            </p>
          </div>
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
      {!props.compact && run.error_message && (
        <p role='alert' className='text-destructive text-sm'>
          {run.error_message}
        </p>
      )}
      {run.results.length === 0 && (
        <Empty
          className={cn('border', props.compact ? 'min-h-24' : 'min-h-80')}
        >
          <EmptyHeader>
            <EmptyTitle>{t(studioStatusLabelKey(run.status))}</EmptyTitle>
            {props.compact && (
              <>
                <p className='text-sm'>{studioTaskName(run)}</p>
                <p className='text-muted-foreground text-xs break-all'>
                  {t('Task ID')}: {run.id}
                </p>
                <p className='text-muted-foreground text-xs'>
                  {props.ownerName} · {t('User ID')}: {run.owner_user_id}
                </p>
              </>
            )}
          </EmptyHeader>
          {props.compact && (
            <p className='text-muted-foreground line-clamp-3 text-xs break-words'>
              {run.error_message || run.prompt}
            </p>
          )}
        </Empty>
      )}
      <div
        className={
          props.compact
            ? 'contents'
            : 'flex min-h-0 min-w-0 flex-1 flex-col gap-3'
        }
      >
        {results.map((result) => {
          const imageURL = props.compact
            ? result.thumbnail_url || result.preview_url
            : result.preview_url
          return (
            <article
              key={result.id}
              className={cn(
                'relative flex min-w-0 flex-col gap-3',
                props.compact
                  ? 'bg-card rounded-xl border p-2'
                  : 'min-h-0 flex-1'
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
                      : 'relative flex min-h-48 flex-1 items-center justify-center overflow-hidden border'
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
                      props.compact
                        ? 'h-40'
                        : 'max-h-[50vh] xl:absolute xl:inset-0 xl:h-full xl:max-h-none'
                    )}
                    onError={(event) => {
                      const url = event.currentTarget.src
                      setUnavailable((current) => [...current, url])
                    }}
                  />
                </button>
              ) : (
                <Empty className={props.compact ? 'h-40' : 'min-h-64'}>
                  <EmptyHeader>
                    <EmptyTitle>{t('Preview unavailable')}</EmptyTitle>
                  </EmptyHeader>
                </Empty>
              )}
              {!props.compact && result.save_error && (
                <p role='alert' className='text-destructive text-xs'>
                  {result.save_error}
                </p>
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
                      if (!result.preview_url) {
                        const detail = assertStudioSuccess(
                          await getStudioRun(run.id, 'all')
                        )
                        const found = detail.results.find(
                          (value) => value.id === result.id
                        )
                        if (!found) throw new Error(t('Preview unavailable'))
                        image = found
                      }
                      await downloadStudioImage(image, props.compact)
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
                    disabled={props.saving}
                    onClick={() => props.onSave?.(result)}
                  >
                    <HugeiconsIcon icon={FolderAddIcon} aria-hidden='true' />
                    {t('Save to asset library')}
                  </Button>
                )}
              </div>
              {!props.compact && props.onSave && !result.asset_id && (
                <p className='text-muted-foreground shrink-0 text-[11px] leading-relaxed'>
                  {t(
                    'Only add images to an asset group when you need them in the asset library for video generation or other workflows.'
                  )}
                </p>
              )}
              {props.compact && (
                <>
                  <p
                    className='truncate text-sm font-medium'
                    title={studioTaskName(run)}
                  >
                    {studioTaskName(run)}
                  </p>
                  <p
                    className='text-muted-foreground truncate text-xs'
                    title={run.id}
                  >
                    {t('Task ID')}: {run.id}
                  </p>
                  <div className='flex flex-wrap items-center justify-between gap-1'>
                    <p className='text-muted-foreground truncate text-xs'>
                      {props.ownerName} · {t('User ID')}: {run.owner_user_id}
                    </p>
                    <Badge variant='secondary'>
                      {t(studioStatusLabelKey(run.status))}
                    </Badge>
                  </div>
                </>
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
