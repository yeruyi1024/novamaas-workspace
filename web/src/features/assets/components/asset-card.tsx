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
  AudioLines,
  Check,
  Clock,
  Download,
  Eye,
  Film,
  Folder,
  Image,
  Trash2,
  User,
} from 'lucide-react'
import { memo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { formatTimestampToDate } from '@/lib/format'
import { cn } from '@/lib/utils'

import { assetKey, formatAssetBytes } from '../asset-utils'
import type { MediaAsset } from '../types'
import { AssetAudioMedia } from './asset-audio-media'
import { AssetImageMedia } from './asset-image-media'
import { AssetVideoMedia } from './asset-video-media'

const TYPE_ICONS = {
  image: Image,
  video: Film,
  audio: AudioLines,
} as const

const TYPE_LABEL_KEYS = {
  image: 'Image',
  video: 'Video',
  audio: 'Audio',
} as const

type AssetCardProps = {
  asset: MediaAsset
  isAdmin?: boolean
  thumbnailUrl?: string
  posterUrl?: string
  priority?: 'high' | 'eager' | 'lazy'
  previewLoading?: boolean
  selected: boolean
  selectionMode: boolean
  downloading: boolean
  canReupload: boolean
  onToggleSelect: (asset: MediaAsset, extend?: boolean) => void
  onPreview: (asset: MediaAsset) => void
  onDownload: (asset: MediaAsset) => void
  onDelete: (asset: MediaAsset) => void
  onReupload: (asset: MediaAsset) => void
}

function CardAction(props: {
  label: string
  Icon: typeof Eye
  disabled?: boolean
  destructive?: boolean
  onClick: () => void
}) {
  const ActionIcon = props.Icon
  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <Button
            type='button'
            size='icon-sm'
            variant='secondary'
            aria-label={props.label}
            disabled={props.disabled}
            className={cn(
              'bg-background/85 size-7 shadow-sm backdrop-blur',
              props.destructive && 'hover:text-destructive'
            )}
            onClick={props.onClick}
          />
        }
      >
        <ActionIcon
          className='size-3.5'
          strokeWidth={1.75}
          aria-hidden='true'
        />
      </TooltipTrigger>
      <TooltipContent>{props.label}</TooltipContent>
    </Tooltip>
  )
}

export const AssetCard = memo(function AssetCard(props: AssetCardProps) {
  const { t } = useTranslation()
  const [hovered, setHovered] = useState(false)
  const asset = props.asset
  const reference = `asset://${asset.id}`
  const failed = asset.status === 'failed'
  const unavailable = asset.status === 'unavailable' || failed
  const processing = asset.status === 'processing'
  const statusLabel = failed ? t('Failed') : t('Asset unavailable')
  const TypeIcon = TYPE_ICONS[asset.type]
  const uploadedAt = formatTimestampToDate(asset.created_at)
  const openOrSelect = (event?: { shiftKey: boolean }) => {
    if (props.selectionMode) props.onToggleSelect(asset, event?.shiftKey)
    else props.onPreview(asset)
  }

  let media = (
    <div
      className='skeleton-shimmer size-full'
      data-testid='asset-media-loading'
    />
  )
  if (!props.previewLoading && asset.type === 'image') {
    media = (
      <AssetImageMedia
        asset={asset}
        url={props.thumbnailUrl}
        isAdmin={props.isAdmin}
        priority={props.priority}
        onOpen={openOrSelect}
      />
    )
  } else if (!props.previewLoading && asset.type === 'video') {
    media = (
      <AssetVideoMedia
        asset={asset}
        url={props.thumbnailUrl}
        posterUrl={props.posterUrl}
        priority={props.priority}
        active={hovered && !props.selectionMode}
        onOpen={openOrSelect}
      />
    )
  } else if (!props.previewLoading) {
    media = (
      <AssetAudioMedia
        asset={asset}
        url={props.thumbnailUrl}
        onOpen={openOrSelect}
      />
    )
  }

  return (
    <article
      aria-label={asset.name}
      data-asset-key={assetKey(asset)}
      data-selected={props.selected || undefined}
      className={cn(
        'group/card bg-card/50 ring-border/60 relative flex h-full flex-col rounded-2xl p-1.5 ring-1 transition-all duration-300 ease-out [contain-intrinsic-size:340px] [content-visibility:auto]',
        'hover:bg-card hover:ring-border hover:-translate-y-0.5 hover:shadow-xl hover:shadow-black/10 motion-reduce:transition-none motion-reduce:hover:translate-y-0 dark:hover:shadow-black/40',
        props.selected &&
          'bg-primary/5 ring-primary hover:bg-primary/10 ring-2 hover:translate-y-0'
      )}
    >
      <div
        className='bg-muted relative aspect-[4/3] overflow-hidden rounded-xl'
        onPointerEnter={() => setHovered(true)}
        onPointerLeave={() => setHovered(false)}
      >
        {media}
        <span
          aria-hidden='true'
          className='pointer-events-none absolute inset-x-0 top-0 h-14 bg-gradient-to-b from-black/35 to-transparent opacity-0 transition-opacity group-hover/card:opacity-100'
        />
        <span
          aria-hidden='true'
          className='pointer-events-none absolute inset-x-0 bottom-0 h-14 bg-gradient-to-t from-black/40 to-transparent opacity-0 transition-opacity group-hover/card:opacity-100'
        />
        <button
          type='button'
          role='checkbox'
          aria-checked={props.selected}
          aria-label={`${t('Select')} ${asset.name}`}
          className={cn(
            'focus-visible:outline-ring absolute top-2 left-2 z-10 grid size-6 place-items-center rounded-full border-2 shadow-sm transition-all focus-visible:opacity-100 focus-visible:outline-2',
            props.selected
              ? 'bg-primary border-primary text-primary-foreground scale-100 opacity-100'
              : 'scale-90 border-white/90 bg-black/25 text-transparent opacity-0 backdrop-blur-sm group-hover/card:scale-100 group-hover/card:opacity-100 [@media(hover:none)]:opacity-100',
            props.selectionMode && 'opacity-100'
          )}
          onClick={(event) => props.onToggleSelect(asset, event.shiftKey)}
        >
          <Check className='size-3.5' strokeWidth={2.75} aria-hidden='true' />
        </button>
        <span
          aria-label={t(TYPE_LABEL_KEYS[asset.type])}
          title={t(TYPE_LABEL_KEYS[asset.type])}
          className='pointer-events-none absolute top-2 right-2 z-10 inline-flex items-center gap-1 rounded-full bg-neutral-950/60 px-2 py-0.5 text-[11px] font-medium text-white/95 ring-1 ring-white/15 backdrop-blur-md'
        >
          <TypeIcon
            aria-hidden='true'
            className='size-3 shrink-0'
            strokeWidth={1.85}
          />
          <span>{t(TYPE_LABEL_KEYS[asset.type])}</span>
        </span>
        {unavailable && (
          <Badge
            variant='destructive'
            className='absolute bottom-2 left-2 z-10 gap-1.5 border-0 bg-red-600/90 text-white shadow-sm backdrop-blur hover:bg-red-600'
            render={
              <button
                type='button'
                aria-label={`${statusLabel}: ${t('View details')}`}
                onClick={() => props.onPreview(asset)}
              />
            }
          >
            <span
              aria-hidden='true'
              className='size-1.5 rounded-full bg-white'
            />
            {statusLabel}
          </Badge>
        )}
        {processing && (
          <Badge className='absolute bottom-2 left-2 z-10 gap-1.5 border-0 bg-amber-500/90 text-white shadow-sm backdrop-blur'>
            <span aria-hidden='true' className='relative flex size-1.5'>
              <span className='absolute inline-flex size-full animate-ping rounded-full bg-white/80 motion-reduce:animate-none' />
              <span className='relative inline-flex size-1.5 rounded-full bg-white' />
            </span>
            {t('Under review')}
          </Badge>
        )}
        <div
          className={cn(
            'absolute right-2 bottom-2 z-10 flex gap-1 transition-all duration-200',
            'translate-y-1 opacity-0 group-focus-within/card:translate-y-0 group-focus-within/card:opacity-100 group-hover/card:translate-y-0 group-hover/card:opacity-100 max-sm:hidden [@media(hover:none)]:hidden',
            props.selectionMode && 'hidden'
          )}
        >
          <CardAction
            label={t('Preview asset')}
            Icon={Eye}
            onClick={() => props.onPreview(asset)}
          />
          <CardAction
            label={t('Download asset')}
            Icon={Download}
            disabled={props.downloading}
            onClick={() => props.onDownload(asset)}
          />
          <CardAction
            label={t('Delete asset')}
            Icon={Trash2}
            destructive
            onClick={() => props.onDelete(asset)}
          />
        </div>
      </div>

      <div className='flex flex-1 flex-col gap-1.5 px-1.5 pt-2.5 pb-1'>
        <div className='flex items-start justify-between gap-2'>
          <h3
            className='min-w-0 truncate text-sm leading-5 font-semibold tracking-tight'
            title={asset.name}
          >
            {asset.name}
          </h3>
          {unavailable && props.canReupload && (
            <Button
              type='button'
              size='xs'
              variant='outline'
              className='-my-0.5 shrink-0'
              aria-label={t('Re-upload revised file')}
              onClick={() => props.onReupload(asset)}
            >
              {t('Re-upload')}
            </Button>
          )}
        </div>

        <div className='bg-muted/70 ring-border/40 flex items-center gap-1 rounded-md py-1 pr-0.5 pl-2 ring-1 ring-inset'>
          <code
            data-testid='asset-card-id'
            className='text-foreground/90 min-w-0 flex-1 font-mono text-[11px] leading-4 break-all select-all'
            title={`${t('Asset ID')}: ${asset.id}`}
          >
            {asset.id}
          </code>
          {asset.status === 'ready' && (
            <CopyButton
              value={reference}
              className='size-5 shrink-0 [&_svg]:size-3'
              tooltip={t('Copy asset reference')}
              successTooltip={t('Asset reference copied')}
            />
          )}
        </div>

        <div
          data-testid='asset-card-group'
          className='text-muted-foreground flex items-start gap-1.5 text-xs'
          title={`${asset.group_name || t('Unknown group')} · ${t('ID')}: ${asset.group_id}`}
        >
          <Folder
            className='text-primary/80 mt-0.5 size-3.5 shrink-0'
            strokeWidth={1.75}
            aria-hidden='true'
          />
          <span className='text-foreground/85 min-w-0 flex-1 leading-4 break-words'>
            {asset.group_name || t('Unknown group')}
          </span>
        </div>

        {props.isAdmin && (
          <div
            data-testid='asset-card-owner'
            className='text-muted-foreground flex items-start gap-1.5 text-xs'
            title={`${t('Uploader')}: ${asset.owner_name || t('Unknown uploader')} · ${t('ID')}: ${asset.owner_user_id}`}
          >
            <User
              className='text-muted-foreground/85 mt-0.5 size-3.5 shrink-0'
              strokeWidth={1.75}
              aria-hidden='true'
            />
            <span className='text-foreground/85 min-w-0 flex-1 leading-4 break-words'>
              {asset.owner_name || t('Unknown uploader')}
              <span className='text-muted-foreground ml-1.5 font-mono text-[11px]'>
                {t('ID')}: {asset.owner_user_id}
              </span>
            </span>
          </div>
        )}

        <div className='text-muted-foreground/85 mt-auto flex flex-wrap items-center justify-between gap-x-2 gap-y-0.5 pt-0.5 text-[11px] tabular-nums'>
          <span className='inline-flex items-center gap-1'>
            <Clock
              className='size-3 shrink-0 opacity-75'
              strokeWidth={1.75}
              aria-hidden='true'
            />
            <span>{uploadedAt}</span>
          </span>
          <span className='shrink-0'>{formatAssetBytes(asset.size)}</span>
        </div>
      </div>
    </article>
  )
})
