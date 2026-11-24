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
import { useEffect, useRef, useState } from 'react'

let activeMedia: HTMLMediaElement | null = null

/**
 * Plays one card's media and pauses whichever card was playing before, so the
 * gallery never plays two clips at once.
 */
export function playExclusively(media: HTMLMediaElement): void {
  if (activeMedia && activeMedia !== media) activeMedia.pause()
  activeMedia = media
  const playback = media.play() as Promise<void> | undefined
  playback?.catch(() => {
    if (activeMedia === media) activeMedia = null
  })
}

export function stopPlayback(media: HTMLMediaElement | null): void {
  if (!media) return
  media.pause()
  if (activeMedia === media) activeMedia = null
}

/**
 * Reports once an element comes within `rootMargin` of the viewport. Media
 * cards use it so files are only fetched for cards a person can see.
 */
export function useNearViewport<T extends Element>(rootMargin = '300px') {
  const ref = useRef<T>(null)
  const [near, setNear] = useState(
    () => typeof IntersectionObserver === 'undefined'
  )
  useEffect(() => {
    if (near || typeof IntersectionObserver === 'undefined') return
    const target = ref.current
    if (!target) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          setNear(true)
          observer.disconnect()
        }
      },
      { rootMargin }
    )
    observer.observe(target)
    return () => observer.disconnect()
  }, [near, rootMargin])
  return { ref, near }
}
