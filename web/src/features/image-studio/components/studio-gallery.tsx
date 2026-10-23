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
import { Spinner } from '@/components/ui/spinner'
import { cn } from '@/lib/utils'

import { getStudioRun } from '../api'
import {
  downloadStudioImage,
  studioErrorMessage,
  studioStatusBadgeVariant,
  studioStatusLabelKey,
  studioTaskName,
} from '../studio-utils'
import type { StudioResult, StudioRun } from '../types'
import { StudioWelcome } from './studio-welcome'

export function StudioGallery(props: {
  run?: StudioRun
  resultId?: string
  onReusePrompt?: (prompt: string) => void
  reuseDisabled?: boolean
  ownerName?: string
  onSave?: (result: StudioResult) => void
  saving?: boolean
  compact?: boolean
  maxPromptCharacters?: number
  onPreview?: () => void
}) {
  const { t } = useTranslation()
  const [preview, setPreview] = useState<StudioResult>()
  const [unavailable, setUnavailable] = useState<string[]>([])
  const layout = props.compact
    ? 'contents'
    : 'flex min-h-0 min-w-0 flex-1 flex-col'
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
      <StudioWelcome
        disabled={Boolean(props.reuseDisabled)}
        maxCharacters={props.maxPromptCharacters ?? 0}
        onPrompt={props.onReusePrompt}
      />
    )
  }

  return (
    <section className={layout} aria-label={t('Generation results')}>
      {!props.compact && (
        <div className='flex shrink-0 flex-wrap items-center gap-2 px-4 pt-4 md:px-5'>
          <Badge variant={studioStatusBadgeVariant(run.status)}>
            {t(studioStatusLabelKey(run.status))}
          </Badge>
          <p className='text-muted-foreground mr-auto truncate text-xs'>
            {t('Generation results')} · {run.model}
          </p>
          {props.onReusePrompt && run.prompt ? (
            <Button
              type='button'
              size='sm'
              variant='outline'
              className='rounded-full'
              disabled={props.reuseDisabled}
              onClick={() => props.onReusePrompt?.(run.prompt || '')}
            >
              {t('Use this prompt')}
            </Button>
          ) : null}
        </div>
      )}
      {!props.compact && run.status === 'running' && (
        <p
          role='status'
          className='text-muted-foreground flex items-center gap-2 px-5 pt-3 text-sm'
        >
          <Spinner />
          {t('This generation is still in progress')}
        </p>
      )}
      {!props.compact && run.error_message && run.results.length > 0 && (
        <p role='alert' className='text-destructive px-5 pt-3 text-sm'>
          {run.error_message}
        </p>
      )}
      {run.results.length === 0 && (
        <Empty
          className={cn(
            props.compact
              ? 'min-h-24 rounded-2xl border'
              : 'm-auto max-w-lg border-0 p-8'
          )}
        >
          <EmptyHeader>
            <EmptyMedia
              variant='icon'
              className={cn(
                run.status === 'failed' && 'bg-destructive/10 text-destructive'
              )}
            >
              <HugeiconsIcon
                icon={run.status === 'failed' ? Alert02Icon : Image01Icon}
                aria-hidden='true'
              />
            </EmptyMedia>
            <EmptyTitle>{t(studioStatusLabelKey(run.status))}</EmptyTitle>
            {props.compact && (
              <StudioRunSummary run={run} ownerName={props.ownerName} />
            )}
            <EmptyDescription
              role={run.error_message ? 'alert' : undefined}
              className='break-words'
            >
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
                props.compact
                  ? 'bg-card rounded-xl border p-2'
                  : 'grid min-h-0 flex-1 grid-rows-[minmax(0,1fr)_auto] gap-x-6 p-4 md:grid-cols-[minmax(0,1fr)_auto] md:p-5'
              )}
            >
              {imageURL && !unavailable.includes(imageURL) ? (
                <button
                  type='button'
                  aria-label={t('Preview')}
                  onClick={() => {
                    props.onPreview?.()
                    setPreview(result)
                  }}
                  className={cn(
                    'bg-muted/30 focus-visible:outline-ring rounded-xl focus-visible:outline-2',
                    props.compact
                      ? ''
                      : 'relative col-span-full min-h-64 cursor-zoom-in bg-transparent'
                  )}
                >
                  <img
                    src={imageURL}
                    alt={
                      result.description || run.prompt || t('Generated image')
                    }
                    loading='lazy'
                    className={cn(
                      'rounded-xl object-contain',
                      props.compact
                        ? 'h-40 w-full'
                        : 'absolute inset-0 m-auto h-auto max-h-full w-auto max-w-full shadow-[0_30px_80px_-30px_#000]'
                    )}
                    onError={(event) => {
                      const url = event.currentTarget.src
                      setUnavailable((current) => [...current, url])
                    }}
                  />
                </button>
              ) : (
                <Empty
                  className={props.compact ? 'h-40' : 'col-span-full min-h-64'}
                >
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
                    : 'order-2 flex flex-wrap items-center justify-end gap-2 self-center'
                }
              >
                {!props.compact && (
                  <p className='text-muted-foreground mr-1 text-[11px]'>
                    {result.asset_id
                      ? t('Saved to asset library')
                      : t('Saved in generation records')}
                  </p>
                )}
                <Button
                  type='button'
                  size={props.compact ? 'icon-sm' : 'sm'}
                  variant='outline'
                  className={cn(!props.compact && 'rounded-full')}
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
                    className='rounded-full'
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
                  className='text-muted-foreground/70 order-3 col-span-full mt-2 text-[11px] leading-relaxed'
                >
                  {t(
                    'Images in generation records expire automatically. Download them before expiry. Asset library copies are kept separately for video generation.'
                  )}
                </p>
              )}
              {props.compact && (
                <StudioRunSummary run={run} ownerName={props.ownerName} />
              )}
              {(result.description || run.prompt) && (
                <p
                  className={cn(
                    'text-muted-foreground line-clamp-2 text-xs break-words',
                    // Under the stage image the prompt reads before the actions.
                    !props.compact && 'order-1 self-center leading-relaxed'
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
        <DialogContent className='image-studio-theme max-h-[90vh] overflow-y-auto sm:max-w-4xl'>
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
              className='mx-auto h-auto max-h-[75svh] w-auto max-w-full object-contain'
            />
          )}
        </DialogContent>
      </Dialog>
    </section>
  )
}

function StudioRunSummary(props: { run: StudioRun; ownerName?: string }) {
  const { t } = useTranslation()
  const run = props.run
  return (
    <div className='min-w-0 flex-1 space-y-1'>
      <p className='truncate text-sm font-semibold' title={studioTaskName(run)}>
        {studioTaskName(run)}
      </p>
      <p className='text-muted-foreground truncate text-xs' title={run.id}>
        {t('Task ID')}: {run.id}
      </p>
      <div className='flex flex-wrap items-center justify-between gap-1'>
        <p className='text-muted-foreground truncate text-xs'>
          {props.ownerName || run.username} · {t('User ID')}:{' '}
          {run.owner_user_id}
        </p>
        <Badge variant={studioStatusBadgeVariant(run.status)}>
          {t(studioStatusLabelKey(run.status))}
        </Badge>
      </div>
    </div>
  )
}
