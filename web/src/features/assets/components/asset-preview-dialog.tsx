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
import {
  AudioLines,
  ChevronLeft,
  ChevronRight,
  Download,
  Folder,
  Trash2,
  User,
} from 'lucide-react'
import type { KeyboardEvent } from 'react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'
import { formatTimestampToDate } from '@/lib/format'

import { getMediaAssetPreview } from '../api'
import {
  assertAssetSuccess,
  assetErrorMessage,
  formatAssetBytes,
} from '../asset-utils'
import type { MediaAsset } from '../types'

const TYPE_LABEL_KEYS = {
  image: 'Image',
  video: 'Video',
  audio: 'Audio',
} as const

function unavailableReasonKey(asset: MediaAsset): string {
  if (asset.status === 'failed') return 'Request failed'
  if (asset.unavailable_reason === 'real_person') {
    return 'Real-person content was rejected by the upstream provider.'
  }
  if (asset.unavailable_reason === 'sensitive_content') {
    return 'Sensitive content was rejected by the upstream provider.'
  }
  return 'The upstream provider rejected this asset.'
}

/**
 * Lightbox for one asset: original media on a dark stage, arrow-key and
 * button navigation through the current page, and its complete audit details.
 */
export function AssetPreviewDialog(props: {
  asset: MediaAsset
  isAdmin?: boolean
  position?: { index: number; total: number }
  onPrevious?: () => void
  onNext?: () => void
  onOpenChange: (open: boolean) => void
  onDownload: (asset: MediaAsset) => void
  onDelete?: (asset: MediaAsset) => void
  onReupload?: (asset: MediaAsset) => void
  downloading: boolean
}) {
  const { t } = useTranslation()
  const asset = props.asset
  const failed = asset.status === 'failed'
  const unavailable = asset.status === 'unavailable' || failed
  const statusLabel = failed ? t('Failed') : t('Asset unavailable')
  const reference = `asset://${asset.id}`
  const previewQuery = useQuery({
    queryKey: [
      'asset-library',
      'preview',
      asset.owner_user_id,
      asset.id,
      'original',
    ],
    queryFn: async () =>
      assertAssetSuccess(
        props.isAdmin
          ? await getMediaAssetPreview(
              asset.id,
              'original',
              asset.owner_user_id
            )
          : await getMediaAssetPreview(asset.id, 'original')
      ),
    staleTime: 5 * 60 * 1000,
  })
  const url = previewQuery.data?.url
  const navigate = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.target instanceof HTMLMediaElement) return
    if (event.key === 'ArrowLeft' && props.onPrevious) {
      event.preventDefault()
      props.onPrevious()
    } else if (event.key === 'ArrowRight' && props.onNext) {
      event.preventDefault()
      props.onNext()
    }
  }

  let stage = <Skeleton className='h-[50dvh] w-full rounded-none opacity-20' />
  if (!previewQuery.isLoading && !url) {
    stage = (
      <div className='flex flex-col items-center justify-center gap-1 px-4 text-center text-white/70'>
        <span>{t('Preview unavailable')}</span>
        {previewQuery.isError && (
          <span role='alert' className='text-xs text-red-300'>
            {assetErrorMessage(previewQuery.error)}
          </span>
        )}
      </div>
    )
  } else if (url && asset.type === 'image') {
    stage = (
      <img
        key={url}
        src={url}
        alt={`${asset.name} ${t('Preview')}`}
        className='max-h-[72dvh] max-w-full object-contain'
      />
    )
  } else if (url && asset.type === 'video') {
    stage = (
      <video
        key={url}
        src={url}
        controls
        playsInline
        preload='metadata'
        className='max-h-[72dvh] w-full object-contain'
      />
    )
  } else if (url) {
    stage = (
      <div className='flex w-full max-w-xl flex-col items-center gap-6 px-6'>
        <span className='grid size-18 place-items-center rounded-2xl bg-white/10 text-white ring-1 ring-white/15'>
          <AudioLines className='size-9' strokeWidth={1.6} aria-hidden='true' />
        </span>
        <audio
          key={url}
          src={url}
          controls
          preload='metadata'
          className='w-full'
        />
      </div>
    )
  }

  const summaryDetails = [
    t(TYPE_LABEL_KEYS[asset.type]),
    formatAssetBytes(asset.size),
    formatTimestampToDate(asset.created_at),
  ]

  return (
    <Dialog open onOpenChange={props.onOpenChange}>
      <DialogContent
        className='max-h-[calc(100dvh-2rem)] gap-0 overflow-y-auto p-0 sm:max-w-5xl [&>[data-slot=dialog-close]]:top-3 [&>[data-slot=dialog-close]]:right-3 [&>[data-slot=dialog-close]]:rounded-full [&>[data-slot=dialog-close]]:bg-black/45 [&>[data-slot=dialog-close]]:text-white [&>[data-slot=dialog-close]]:backdrop-blur [&>[data-slot=dialog-close]]:hover:bg-black/60'
        onKeyDown={navigate}
      >
        <div className='relative grid min-h-[45dvh] place-items-center bg-neutral-950'>
          {stage}
          {props.onPrevious && (
            <Button
              type='button'
              size='icon'
              variant='secondary'
              aria-label={t('Previous asset')}
              className='absolute top-1/2 left-3 size-10 -translate-y-1/2 rounded-full bg-white/15 text-white backdrop-blur hover:bg-white/25'
              onClick={props.onPrevious}
            >
              <ChevronLeft className='size-5' strokeWidth={1.75} />
            </Button>
          )}
          {props.onNext && (
            <Button
              type='button'
              size='icon'
              variant='secondary'
              aria-label={t('Next asset')}
              className='absolute top-1/2 right-3 size-10 -translate-y-1/2 rounded-full bg-white/15 text-white backdrop-blur hover:bg-white/25'
              onClick={props.onNext}
            >
              <ChevronRight className='size-5' strokeWidth={1.75} />
            </Button>
          )}
          {props.position && props.position.total > 1 && (
            <span className='absolute top-3 left-3 rounded-full bg-black/50 px-2.5 py-0.5 text-xs text-white tabular-nums backdrop-blur'>
              {props.position.index + 1} / {props.position.total}
            </span>
          )}
        </div>
        <div className='flex flex-col gap-3 p-4 sm:p-5'>
          {unavailable && (
            <Alert variant='destructive'>
              <AlertTitle>{statusLabel}</AlertTitle>
              <AlertDescription>
                {t(unavailableReasonKey(asset))}{' '}
                {t('Upload revised material and use its new asset ID.')}
              </AlertDescription>
            </Alert>
          )}
          <div className='flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between'>
            <div className='min-w-0 flex-1 space-y-2'>
              <DialogTitle className='text-base break-words'>
                {asset.name}
              </DialogTitle>
              <DialogDescription className='flex flex-wrap items-center gap-x-3 gap-y-1 text-xs'>
                <span>{summaryDetails.join(' · ')}</span>
                <span className='inline-flex items-center gap-1'>
                  <Folder
                    className='size-3.5 shrink-0'
                    strokeWidth={1.75}
                    aria-hidden='true'
                  />
                  <span className='text-foreground/90 font-medium'>
                    {asset.group_name || t('Unknown group')}
                  </span>
                  <span className='font-mono text-[11px] break-all'>
                    ({t('ID')}: {asset.group_id})
                  </span>
                </span>
                <span className='inline-flex items-center gap-1'>
                  <User
                    className='size-3.5 shrink-0'
                    strokeWidth={1.75}
                    aria-hidden='true'
                  />
                  <span className='text-foreground/90 font-medium'>
                    {asset.owner_name || t('Unknown uploader')}
                  </span>
                  <span className='font-mono text-[11px]'>
                    ({t('ID')}: {asset.owner_user_id})
                  </span>
                </span>
              </DialogDescription>
              <div className='bg-muted/70 ring-border/50 inline-flex max-w-full items-center gap-1.5 rounded-md py-0.5 pr-1 pl-2 ring-1'>
                <span className='text-muted-foreground text-xs'>
                  {t('Asset ID')}
                </span>
                <code className='font-mono text-xs break-all select-all'>
                  {asset.status === 'ready' ? reference : asset.id}
                </code>
                {asset.status === 'ready' && (
                  <CopyButton
                    value={reference}
                    className='size-6 shrink-0'
                    tooltip={t('Copy asset reference')}
                    successTooltip={t('Asset reference copied')}
                  />
                )}
              </div>
            </div>
            <div className='flex shrink-0 flex-wrap gap-2'>
              {props.onDelete && (
                <Button
                  type='button'
                  variant='ghost'
                  className='text-destructive hover:bg-destructive/10 hover:text-destructive'
                  onClick={() => props.onDelete?.(asset)}
                >
                  <Trash2
                    className='size-3.5'
                    strokeWidth={1.75}
                    aria-hidden='true'
                  />
                  {t('Delete')}
                </Button>
              )}
              <Button
                type='button'
                variant={unavailable ? 'outline' : 'default'}
                disabled={props.downloading}
                onClick={() => props.onDownload(asset)}
              >
                <Download
                  className='size-3.5'
                  strokeWidth={1.75}
                  aria-hidden='true'
                />
                {t('Download')}
              </Button>
              {unavailable && props.onReupload && (
                <Button type='button' onClick={() => props.onReupload?.(asset)}>
                  {t('Re-upload revised file')}
                </Button>
              )}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
