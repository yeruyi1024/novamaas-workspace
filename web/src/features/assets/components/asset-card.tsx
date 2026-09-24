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
  AudioWaveformIcon,
  Delete02Icon,
  Download01Icon,
  FileVideoIcon,
  Image01Icon,
  ViewIcon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { formatTimestampToDate } from '@/lib/format'

import { getMediaAssetPreview } from '../api'
import { assertAssetSuccess } from '../asset-utils'
import type { MediaAsset } from '../types'

function AssetPreview(props: {
  asset: MediaAsset
  onPreview: (asset: MediaAsset) => void
}) {
  const { t } = useTranslation()
  const previewRef = useRef<HTMLDivElement>(null)
  const [useOriginalPreview, setUseOriginalPreview] = useState(false)
  const [originalPreviewFailed, setOriginalPreviewFailed] = useState(false)
  const [visible, setVisible] = useState(
    () => typeof IntersectionObserver === 'undefined'
  )
  useEffect(() => {
    if (visible || typeof IntersectionObserver === 'undefined') return
    const target = previewRef.current
    if (!target) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          setVisible(true)
          observer.disconnect()
        }
      },
      { rootMargin: '300px' }
    )
    observer.observe(target)
    return () => observer.disconnect()
  }, [visible])
  const previewVariant = useOriginalPreview ? 'original' : 'thumbnail'
  const previewQuery = useQuery({
    queryKey: ['asset-library', 'preview', props.asset.id, previewVariant],
    queryFn: async () =>
      assertAssetSuccess(
        await getMediaAssetPreview(props.asset.id, previewVariant)
      ),
    enabled: visible,
    staleTime: 5 * 60 * 1000,
  })

  if (!visible || previewQuery.isLoading) {
    return (
      <div ref={previewRef} className='h-36 w-full'>
        <Skeleton className='h-full w-full rounded-none' />
      </div>
    )
  }
  if (!previewQuery.data?.url || originalPreviewFailed) {
    return (
      <div className='bg-muted text-muted-foreground flex h-36 items-center justify-center text-xs'>
        {t('Preview unavailable')}
      </div>
    )
  }
  if (props.asset.type === 'image') {
    return (
      <button
        type='button'
        aria-label={`${t('Preview')} ${props.asset.name}`}
        className='bg-muted/30 focus-visible:outline-ring flex h-36 w-full cursor-zoom-in items-center justify-center p-2 focus-visible:outline-2 focus-visible:outline-offset-[-2px]'
        onClick={() => props.onPreview(props.asset)}
      >
        <img
          src={previewQuery.data.url}
          alt={props.asset.name}
          loading='lazy'
          decoding='async'
          className='max-h-full max-w-full object-contain'
          onError={() => {
            if (useOriginalPreview) setOriginalPreviewFailed(true)
            else setUseOriginalPreview(true)
          }}
        />
      </button>
    )
  }
  if (props.asset.type === 'video') {
    return (
      <video
        src={previewQuery.data.url}
        controls
        preload='metadata'
        className='bg-muted h-36 w-full object-contain'
      />
    )
  }
  return (
    <div className='bg-muted/40 flex h-36 flex-col items-center justify-center gap-3 px-3'>
      <HugeiconsIcon
        icon={AudioWaveformIcon}
        className='text-muted-foreground size-8'
      />
      <audio
        src={previewQuery.data.url}
        controls
        preload='metadata'
        className='h-8 w-full'
      />
    </div>
  )
}

function assetTypeIcon(type: MediaAsset['type']) {
  if (type === 'video') return FileVideoIcon
  if (type === 'audio') return AudioWaveformIcon
  return Image01Icon
}

function assetTypeLabelKey(type: MediaAsset['type']) {
  if (type === 'video') return 'Video'
  if (type === 'audio') return 'Audio'
  return 'Image'
}

export function AssetCard(props: {
  asset: MediaAsset
  groupName: string
  onDelete: (asset: MediaAsset) => void
  onReupload?: (asset: MediaAsset) => void
  onPreview: (asset: MediaAsset) => void
  onDownload: (asset: MediaAsset) => void
  downloading: boolean
}) {
  const { t } = useTranslation()
  const reference = `asset://${props.asset.id}`
  const unavailable = props.asset.status === 'unavailable'
  let unavailableReason = t('The upstream provider rejected this asset.')
  if (props.asset.unavailable_reason === 'real_person') {
    unavailableReason = t(
      'Real-person content was rejected by the upstream provider.'
    )
  } else if (props.asset.unavailable_reason === 'sensitive_content') {
    unavailableReason = t(
      'Sensitive content was rejected by the upstream provider.'
    )
  }
  const deleteButton = (
    <Button
      type='button'
      size='icon-sm'
      variant='ghost'
      aria-label={t('Delete asset')}
      onClick={() => props.onDelete(props.asset)}
    >
      <HugeiconsIcon icon={Delete02Icon} />
    </Button>
  )

  return (
    <Card
      size='sm'
      className='overflow-hidden [contain-intrinsic-size:320px] [content-visibility:auto]'
    >
      <AssetPreview asset={props.asset} onPreview={props.onPreview} />
      <CardHeader className='gap-1.5'>
        <div className='flex items-start justify-between gap-2'>
          <CardTitle className='min-w-0 truncate'>{props.asset.name}</CardTitle>
          <Badge variant='outline' className='shrink-0'>
            <HugeiconsIcon icon={assetTypeIcon(props.asset.type)} />
            {t(assetTypeLabelKey(props.asset.type))}
          </Badge>
        </div>
        <p className='text-muted-foreground truncate text-xs'>
          {props.groupName}
        </p>
        <div className='text-muted-foreground grid gap-0.5 text-xs'>
          <p className='truncate' title={props.asset.owner_name}>
            {t('Uploader')}: {props.asset.owner_name || t('Unknown uploader')}
          </p>
          <p>
            {t('Uploaded at')}:{' '}
            <time
              dateTime={new Date(props.asset.created_at * 1000).toISOString()}
            >
              {formatTimestampToDate(props.asset.created_at)}
            </time>
          </p>
        </div>
      </CardHeader>
      <CardContent>
        {unavailable && (
          <Alert variant='destructive' className='mb-3'>
            <AlertTitle>{t('Asset unavailable')}</AlertTitle>
            <AlertDescription>
              {unavailableReason}{' '}
              {t('Upload revised material and use its new asset ID.')}
            </AlertDescription>
          </Alert>
        )}
        <code className='bg-muted block truncate rounded-md px-2 py-1.5 text-[11px]'>
          {reference}
        </code>
      </CardContent>
      <CardFooter className='flex-wrap justify-end gap-1'>
        <Button
          type='button'
          size='icon-sm'
          variant='ghost'
          aria-label={t('Preview asset')}
          onClick={() => props.onPreview(props.asset)}
        >
          <HugeiconsIcon icon={ViewIcon} />
        </Button>
        <Button
          type='button'
          size='icon-sm'
          variant='ghost'
          aria-label={t('Download asset')}
          disabled={props.downloading}
          onClick={() => props.onDownload(props.asset)}
        >
          <HugeiconsIcon icon={Download01Icon} />
        </Button>
        {unavailable && props.onReupload && (
          <Button
            type='button'
            size='sm'
            variant='outline'
            onClick={() => props.onReupload?.(props.asset)}
          >
            {t('Re-upload revised file')}
          </Button>
        )}
        {!unavailable && (
          <CopyButton
            value={reference}
            size='icon'
            className='size-7'
            tooltip={t('Copy asset reference')}
            successTooltip={t('Asset reference copied')}
          />
        )}
        <Tooltip>
          <TooltipTrigger render={deleteButton} />
          <TooltipContent>{t('Delete asset')}</TooltipContent>
        </Tooltip>
      </CardFooter>
    </Card>
  )
}
