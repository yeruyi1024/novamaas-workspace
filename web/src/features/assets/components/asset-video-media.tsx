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
import { Film, Play } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

import { formatMediaDuration } from '../asset-utils'
import {
  playExclusively,
  stopPlayback,
  useNearViewport,
} from '../lib/media-playback'
import type { MediaAsset } from '../types'

const POSTER_TIME = 0.1

/**
 * Renders a lightweight OSS first-frame snapshot poster at rest so video
 * galleries do not fetch MP4 headers up front, and mounts the muted video
 * stream only when hovered (or as a fallback if snapshot generation fails).
 */
export function AssetVideoMedia(props: {
  asset: MediaAsset
  url?: string
  posterUrl?: string
  priority?: 'high' | 'eager' | 'lazy'
  active: boolean
  onOpen: (event?: { shiftKey: boolean }) => void
}) {
  const { t } = useTranslation()
  const viewport = useNearViewport<HTMLDivElement>()
  const videoRef = useRef<HTMLVideoElement>(null)
  const [duration, setDuration] = useState(0)
  const [progress, setProgress] = useState(0)
  const [playing, setPlaying] = useState(false)
  const [failed, setFailed] = useState(false)
  const [posterFailed, setPosterFailed] = useState(false)

  useEffect(() => {
    setFailed(false)
  }, [props.url])

  useEffect(() => {
    setPosterFailed(false)
  }, [props.posterUrl])

  const usePoster = Boolean(props.posterUrl) && !posterFailed
  const mountVideo =
    viewport.near &&
    Boolean(props.url) &&
    !failed &&
    (props.active || !usePoster)

  useEffect(() => {
    const video = videoRef.current
    if (!video) return
    if (props.active) {
      playExclusively(video)
      return () => stopPlayback(video)
    }
    stopPlayback(video)
    if (video.readyState > 0) video.currentTime = POSTER_TIME
    return undefined
  }, [props.active, mountVideo])

  const loadingMode =
    props.priority === 'high' || props.priority === 'eager' ? 'eager' : 'lazy'

  return (
    <div
      ref={viewport.ref}
      className='relative size-full bg-gradient-to-br from-slate-900 via-zinc-900 to-slate-800'
    >
      {usePoster && (
        <img
          src={props.posterUrl}
          alt={props.asset.name}
          loading={loadingMode}
          fetchPriority={props.priority === 'high' ? 'high' : 'auto'}
          decoding='async'
          draggable={false}
          className='size-full object-cover transition-transform duration-700 ease-out group-hover/card:scale-[1.04] motion-reduce:transition-none motion-reduce:group-hover/card:scale-100'
          onError={() => setPosterFailed(true)}
        />
      )}
      {mountVideo ? (
        <video
          ref={videoRef}
          src={`${props.url}#t=${POSTER_TIME}`}
          poster={usePoster ? props.posterUrl : undefined}
          muted
          loop
          playsInline
          preload='metadata'
          aria-hidden='true'
          className={cn(
            'size-full object-cover transition-transform duration-700 ease-out group-hover/card:scale-[1.04] motion-reduce:transition-none motion-reduce:group-hover/card:scale-100',
            usePoster && 'absolute inset-0'
          )}
          onLoadedMetadata={(event) =>
            setDuration(event.currentTarget.duration)
          }
          onTimeUpdate={(event) => {
            const video = event.currentTarget
            if (video.duration > 0) {
              setProgress(video.currentTime / video.duration)
            }
          }}
          onPlay={() => setPlaying(true)}
          onPause={() => {
            setPlaying(false)
            setProgress(0)
          }}
          onError={() => setFailed(true)}
        />
      ) : (
        !usePoster && (
          <div className='grid size-full place-items-center text-white/45'>
            <Film className='size-8' strokeWidth={1.5} aria-hidden='true' />
          </div>
        )
      )}
      <button
        type='button'
        aria-label={`${t('Preview')} ${props.asset.name}`}
        className='focus-visible:outline-ring absolute inset-0 grid cursor-pointer place-items-center focus-visible:outline-2 focus-visible:-outline-offset-2'
        onClick={(event) => props.onOpen(event)}
      >
        <span
          className={cn(
            'grid size-11 place-items-center rounded-full bg-black/50 text-white ring-1 ring-white/25 backdrop-blur-sm transition-all duration-300',
            playing ? 'scale-75 opacity-0' : 'opacity-100'
          )}
        >
          <Play
            className='ml-0.5 size-4 fill-current'
            strokeWidth={1.75}
            aria-hidden='true'
          />
        </span>
      </button>
      {duration > 0 && (
        <span className='pointer-events-none absolute right-2 bottom-2 rounded-full bg-black/60 px-2 py-0.5 text-[11px] font-medium text-white tabular-nums backdrop-blur-sm transition-opacity group-hover/card:opacity-0'>
          {formatMediaDuration(duration)}
        </span>
      )}
      {playing && (
        <span
          aria-hidden='true'
          className='pointer-events-none absolute inset-x-0 bottom-0 h-0.5 bg-white/20'
        >
          <span
            className='bg-primary block h-full transition-[width] duration-200 ease-linear'
            style={{ width: `${Math.round(progress * 100)}%` }}
          />
        </span>
      )}
    </div>
  )
}
