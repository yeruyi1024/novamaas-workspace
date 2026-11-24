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
import { Image } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

import { getMediaAssetPreview } from '../api'
import { assertAssetSuccess } from '../asset-utils'
import type { MediaAsset } from '../types'

/**
 * Image card body. Uses the inline/cached WebP thumbnail, prioritizes
 * above-the-fold cards, shows a shimmer while bytes transfer, and falls back
 * once to the original signed image if OSS image processing fails.
 */
export function AssetImageMedia(props: {
  asset: MediaAsset
  url?: string
  isAdmin?: boolean
  priority?: 'high' | 'eager' | 'lazy'
  onOpen: (event?: { shiftKey: boolean }) => void
}) {
  const { t } = useTranslation()
  const imgRef = useRef<HTMLImageElement>(null)
  const [useOriginal, setUseOriginal] = useState(false)
  const [originalFailed, setOriginalFailed] = useState(false)
  const [loadedUrl, setLoadedUrl] = useState<string>('')

  useEffect(() => {
    setUseOriginal(false)
    setOriginalFailed(false)
  }, [props.url])

  const originalQuery = useQuery({
    queryKey: [
      'asset-library',
      'preview',
      props.asset.owner_user_id,
      props.asset.id,
      'original',
    ],
    queryFn: async () =>
      assertAssetSuccess(
        props.isAdmin
          ? await getMediaAssetPreview(
              props.asset.id,
              'original',
              props.asset.owner_user_id
            )
          : await getMediaAssetPreview(props.asset.id, 'original')
      ),
    enabled: useOriginal,
    staleTime: 5 * 60 * 1000,
  })
  const source = useOriginal ? originalQuery.data?.url : props.url
  const loaded = Boolean(source && loadedUrl === source)

  useEffect(() => {
    const img = imgRef.current
    if (source && img?.complete && img.naturalWidth > 0) {
      setLoadedUrl(source)
    }
  }, [source])

  const loadingMode =
    props.priority === 'high' || props.priority === 'eager' ? 'eager' : 'lazy'

  return (
    <button
      type='button'
      aria-label={`${t('Preview')} ${props.asset.name}`}
      className='focus-visible:outline-ring relative block size-full cursor-zoom-in focus-visible:outline-2 focus-visible:-outline-offset-2'
      onClick={(event) => props.onOpen(event)}
    >
      {source && !originalFailed ? (
        <>
          {!loaded && (
            <span
              aria-hidden='true'
              className='skeleton-shimmer pointer-events-none absolute inset-0'
            />
          )}
          <img
            ref={imgRef}
            src={source}
            alt={props.asset.name}
            loading={loadingMode}
            fetchPriority={props.priority === 'high' ? 'high' : 'auto'}
            decoding='async'
            draggable={false}
            className={cn(
              'size-full object-cover transition-[opacity,transform] duration-200 ease-out group-hover/card:scale-[1.04] motion-reduce:transition-none motion-reduce:group-hover/card:scale-100',
              loaded ? 'opacity-100' : 'opacity-0'
            )}
            onLoad={() => setLoadedUrl(source)}
            onError={() => {
              if (useOriginal) setOriginalFailed(true)
              else setUseOriginal(true)
            }}
          />
        </>
      ) : (
        <span className='text-muted-foreground flex size-full flex-col items-center justify-center gap-1.5 text-xs'>
          <Image
            className='size-7 opacity-55'
            strokeWidth={1.5}
            aria-hidden='true'
          />
          {originalFailed && t('Preview unavailable')}
        </span>
      )}
    </button>
  )
}
