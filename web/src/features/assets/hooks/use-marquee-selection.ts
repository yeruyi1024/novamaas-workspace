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
import { type RefObject, useEffect, useRef, useState } from 'react'

type MarqueeBox = {
  left: number
  top: number
  width: number
  height: number
}

const DRAG_THRESHOLD_PX = 6

const INTERACTIVE_SELECTOR =
  'button, a, input, textarea, select, code, [role="button"], [role="checkbox"], [role="dialog"]'

function rectsIntersect(a: MarqueeBox, b: DOMRect): boolean {
  return !(
    a.left + a.width < b.left ||
    a.left > b.right ||
    a.top + a.height < b.top ||
    a.top > b.bottom
  )
}

/**
 * Desktop netdisk rubber-band selection: dragging from whitespace inside the
 * asset grid draws a selection marquee and selects every intersected card.
 */
export function useMarqueeSelection(
  containerRef: RefObject<HTMLElement | null>,
  enabled: boolean,
  onSelectKeys: (keys: string[], additive: boolean) => void
) {
  const [box, setBox] = useState<MarqueeBox | null>(null)
  const onSelectRef = useRef(onSelectKeys)
  useEffect(() => {
    onSelectRef.current = onSelectKeys
  }, [onSelectKeys])

  useEffect(() => {
    if (!enabled) {
      setBox(null)
      return
    }
    const container = containerRef.current
    if (!container) return

    let origin: { x: number; y: number; additive: boolean } | null = null
    let dragging = false
    let suppressNextClick = false

    const handleClickCapture = (event: MouseEvent) => {
      if (!suppressNextClick) return
      suppressNextClick = false
      event.preventDefault()
      event.stopPropagation()
    }

    const handleMouseDown = (event: MouseEvent) => {
      if (event.button !== 0) return
      const target = event.target as HTMLElement | null
      if (!target || target.closest(INTERACTIVE_SELECTOR)) return
      origin = {
        x: event.clientX,
        y: event.clientY,
        additive: event.shiftKey || event.metaKey || event.ctrlKey,
      }
      dragging = false
    }

    const handleMouseMove = (event: MouseEvent) => {
      if (!origin) return
      const dx = event.clientX - origin.x
      const dy = event.clientY - origin.y
      if (!dragging && Math.hypot(dx, dy) < DRAG_THRESHOLD_PX) return
      if (!dragging) {
        dragging = true
        window.getSelection()?.removeAllRanges()
      }
      event.preventDefault()

      const nextBox: MarqueeBox = {
        left: Math.min(origin.x, event.clientX),
        top: Math.min(origin.y, event.clientY),
        width: Math.abs(dx),
        height: Math.abs(dy),
      }
      setBox(nextBox)

      const cards = container.querySelectorAll<HTMLElement>('[data-asset-key]')
      const matched: string[] = []
      for (const card of cards) {
        const key = card.dataset.assetKey
        if (key && rectsIntersect(nextBox, card.getBoundingClientRect())) {
          matched.push(key)
        }
      }
      onSelectRef.current(matched, origin.additive)
    }

    const handleMouseUp = () => {
      if (!origin) return
      if (dragging) {
        suppressNextClick = true
        window.setTimeout(() => {
          suppressNextClick = false
        }, 0)
      }
      origin = null
      dragging = false
      setBox(null)
    }

    container.addEventListener('mousedown', handleMouseDown)
    container.addEventListener('click', handleClickCapture, true)
    window.addEventListener('mousemove', handleMouseMove)
    window.addEventListener('mouseup', handleMouseUp)
    return () => {
      container.removeEventListener('mousedown', handleMouseDown)
      container.removeEventListener('click', handleClickCapture, true)
      window.removeEventListener('mousemove', handleMouseMove)
      window.removeEventListener('mouseup', handleMouseUp)
    }
  }, [containerRef, enabled])

  return box
}
