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
import { Pause, Play } from 'lucide-react'
import { motion, useReducedMotion } from 'motion/react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

import { assetWaveform, formatMediaDuration } from '../asset-utils'
import {
  playExclusively,
  stopPlayback,
  useNearViewport,
} from '../lib/media-playback'
import type { MediaAsset } from '../types'

const WAVEFORM_BARS = 32

/**
 * Audio card body: a checksum-shaped decorative waveform that fills with the
 * playback position, an inline play/pause button, and a clickable stage that
 * opens the preview dialog on desktop and touch devices alike.
 */
export function AssetAudioMedia(props: {
  asset: MediaAsset
  url?: string
  onOpen?: (event?: { shiftKey: boolean }) => void
}) {
  const { t } = useTranslation()
  const reduceMotion = useReducedMotion()
  const viewport = useNearViewport<HTMLDivElement>()
  const audioRef = useRef<HTMLAudioElement>(null)
  const [duration, setDuration] = useState(0)
  const [progress, setProgress] = useState(0)
  const [playing, setPlaying] = useState(false)
  const [failed, setFailed] = useState(false)
  const bars = useMemo(
    () => assetWaveform(props.asset.sha256 || props.asset.id, WAVEFORM_BARS),
    [props.asset.id, props.asset.sha256]
  )
  const canLoad = viewport.near && Boolean(props.url) && !failed
  const playedBars = Math.round(progress * WAVEFORM_BARS)

  useEffect(() => {
    setFailed(false)
  }, [props.url])

  useEffect(() => {
    const audio = audioRef.current
    return () => stopPlayback(audio)
  }, [])

  return (
    <div
      ref={viewport.ref}
      className='relative flex size-full items-center justify-center bg-gradient-to-br from-slate-900 via-zinc-900 to-slate-800'
    >
      {canLoad && (
        <audio
          ref={audioRef}
          src={props.url}
          preload='none'
          onLoadedMetadata={(event) =>
            setDuration(event.currentTarget.duration)
          }
          onTimeUpdate={(event) => {
            const audio = event.currentTarget
            if (audio.duration > 0) {
              setProgress(audio.currentTime / audio.duration)
            }
          }}
          onPlay={() => setPlaying(true)}
          onPause={() => setPlaying(false)}
          onEnded={() => setProgress(0)}
          onError={() => {
            setPlaying(false)
            setFailed(true)
          }}
        />
      )}
      {props.onOpen && (
        <button
          type='button'
          aria-label={`${t('Preview')} ${props.asset.name}`}
          className='focus-visible:outline-ring absolute inset-0 cursor-pointer focus-visible:outline-2 focus-visible:-outline-offset-2'
          onClick={(event) => props.onOpen?.(event)}
        />
      )}
      <div
        aria-hidden='true'
        className='pointer-events-none flex h-1/2 w-[82%] items-center gap-[3px]'
      >
        {bars.map((bar, index) => (
          <motion.span
            key={bar.id}
            className={cn(
              'w-full origin-center rounded-full transition-colors duration-300',
              index < playedBars ? 'bg-white' : 'bg-white/25'
            )}
            style={{ height: `${Math.round(bar.height * 100)}%` }}
            animate={
              playing && !reduceMotion
                ? { scaleY: [1, 0.55, 1] }
                : { scaleY: 1 }
            }
            transition={
              playing && !reduceMotion
                ? {
                    duration: 0.9,
                    repeat: Infinity,
                    delay: (index % 8) * 0.07,
                    ease: 'easeInOut',
                  }
                : { duration: 0.2 }
            }
          />
        ))}
      </div>
      <button
        type='button'
        disabled={!canLoad}
        aria-label={`${playing ? t('Pause') : t('Play')} ${props.asset.name}`}
        aria-pressed={playing}
        className='focus-visible:outline-ring relative z-10 grid size-11 place-items-center rounded-full bg-white/95 text-neutral-950 shadow-lg ring-1 ring-white/25 transition-transform hover:scale-105 focus-visible:outline-2 disabled:opacity-55 motion-reduce:transition-none'
        onClick={(event) => {
          event.stopPropagation()
          const audio = audioRef.current
          if (!audio) return
          if (playing) stopPlayback(audio)
          else playExclusively(audio)
        }}
      >
        {playing ? (
          <Pause
            className='size-4 fill-current'
            strokeWidth={1.75}
            aria-hidden='true'
          />
        ) : (
          <Play
            className='ml-0.5 size-4 fill-current'
            strokeWidth={1.75}
            aria-hidden='true'
          />
        )}
      </button>
      {duration > 0 && (
        <span className='pointer-events-none absolute right-2 bottom-2 rounded-full bg-black/60 px-2 py-0.5 text-[11px] font-medium text-white tabular-nums backdrop-blur-sm transition-opacity group-hover/card:opacity-0'>
          {formatMediaDuration(duration)}
        </span>
      )}
    </div>
  )
}
