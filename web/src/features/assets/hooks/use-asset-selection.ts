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
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { assetKey } from '../asset-utils'
import type { AssetRef, MediaAsset } from '../types'

const EMPTY_ITEMS = new Map<string, MediaAsset>()
const EMPTY_EXCLUDED = new Map<string, AssetRef>()

type SelectionState = {
  scope: string
  allMatching: boolean
  items: Map<string, MediaAsset>
  excluded: Map<string, AssetRef>
  anchor?: string
}

/**
 * Netdisk-style multi-select state scoped to the active filters (without page):
 * - Preserves selections while paging through the same filter results.
 * - Resets automatically when the filter scope changes so batch actions only
 *   apply to the current query.
 * - Supports Shift-click range selection, drag-box marquee selection, and
 *   "select all matching across pages" with per-item exclusions.
 */
export function useAssetSelection(
  scopeKey: string,
  pageAssets: MediaAsset[],
  totalMatching: number
) {
  const emptySelection = useMemo<SelectionState>(
    () => ({
      scope: scopeKey,
      allMatching: false,
      items: EMPTY_ITEMS,
      excluded: EMPTY_EXCLUDED,
    }),
    [scopeKey]
  )
  const [state, setState] = useState(emptySelection)
  if (state.scope !== scopeKey) setState(emptySelection)
  const { allMatching, items: selected, excluded } = state

  const pageRef = useRef(pageAssets)
  const totalRef = useRef(totalMatching)
  useEffect(() => {
    pageRef.current = pageAssets
  }, [pageAssets])
  useEffect(() => {
    totalRef.current = totalMatching
  }, [totalMatching])

  const isSelected = useCallback(
    (asset: AssetRef) => {
      const key = assetKey(asset)
      return allMatching ? !excluded.has(key) : selected.has(key)
    },
    [allMatching, excluded, selected]
  )

  const toggle = useCallback(
    (asset: MediaAsset, extend = false) =>
      setState((previous) => {
        const current = previous.scope === scopeKey ? previous : undefined
        const key = assetKey(asset)
        const page = pageRef.current
        const from = current?.anchor
          ? page.findIndex((item) => assetKey(item) === current.anchor)
          : -1
        const to = page.findIndex((item) => assetKey(item) === key)

        if (current?.allMatching) {
          const nextExcluded = new Map(current.excluded)
          if (extend && from >= 0 && to >= 0) {
            const [start, end] = from < to ? [from, to] : [to, from]
            for (const item of page.slice(start, end + 1)) {
              nextExcluded.delete(assetKey(item))
            }
          } else if (nextExcluded.has(key)) {
            nextExcluded.delete(key)
          } else {
            nextExcluded.set(key, {
              id: asset.id,
              owner_user_id: asset.owner_user_id,
            })
          }
          return {
            scope: scopeKey,
            allMatching: true,
            items: EMPTY_ITEMS,
            excluded: nextExcluded,
            anchor: key,
          }
        }

        const items = new Map(current?.items ?? EMPTY_ITEMS)
        if (extend && from >= 0 && to >= 0) {
          const [start, end] = from < to ? [from, to] : [to, from]
          for (const item of page.slice(start, end + 1)) {
            items.set(assetKey(item), item)
          }
        } else if (items.has(key)) {
          items.delete(key)
        } else {
          items.set(key, asset)
        }
        return {
          scope: scopeKey,
          allMatching: false,
          items,
          excluded: EMPTY_EXCLUDED,
          anchor: key,
        }
      }),
    [scopeKey]
  )

  const selectKeys = useCallback(
    (keys: string[], additive: boolean) => {
      setState((previous) => {
        const current = previous.scope === scopeKey ? previous : undefined
        const pageByKey = new Map(
          pageRef.current.map((item) => [assetKey(item), item])
        )
        const items = new Map(
          additive && current && !current.allMatching
            ? current.items
            : EMPTY_ITEMS
        )
        for (const key of keys) {
          const asset = pageByKey.get(key)
          if (asset) items.set(key, asset)
        }
        return {
          scope: scopeKey,
          allMatching: false,
          items,
          excluded: EMPTY_EXCLUDED,
          anchor: keys.at(-1) ?? current?.anchor,
        }
      })
    },
    [scopeKey]
  )

  const allOnPageSelected =
    pageAssets.length > 0 && pageAssets.every((asset) => isSelected(asset))
  const someOnPageSelected =
    !allOnPageSelected && pageAssets.some((asset) => isSelected(asset))

  const selectPage = useCallback(() => {
    setState((previous) => {
      const current = previous.scope === scopeKey ? previous : undefined
      if (current?.allMatching) {
        const nextExcluded = new Map(current.excluded)
        for (const asset of pageRef.current) {
          nextExcluded.delete(assetKey(asset))
        }
        return {
          scope: scopeKey,
          allMatching: true,
          items: EMPTY_ITEMS,
          excluded: nextExcluded,
        }
      }
      const items = new Map(current?.items ?? EMPTY_ITEMS)
      for (const asset of pageRef.current) items.set(assetKey(asset), asset)
      return {
        scope: scopeKey,
        allMatching: false,
        items,
        excluded: EMPTY_EXCLUDED,
      }
    })
  }, [scopeKey])

  const togglePage = useCallback(() => {
    setState((previous) => {
      const current = previous.scope === scopeKey ? previous : undefined
      const page = pageRef.current
      if (current?.allMatching) {
        const everySelected = page.every(
          (asset) => !current.excluded.has(assetKey(asset))
        )
        const nextExcluded = new Map(current.excluded)
        if (everySelected) {
          for (const asset of page) {
            nextExcluded.set(assetKey(asset), {
              id: asset.id,
              owner_user_id: asset.owner_user_id,
            })
          }
          if (nextExcluded.size >= totalRef.current) {
            return {
              scope: scopeKey,
              allMatching: false,
              items: EMPTY_ITEMS,
              excluded: EMPTY_EXCLUDED,
            }
          }
          return {
            scope: scopeKey,
            allMatching: true,
            items: EMPTY_ITEMS,
            excluded: nextExcluded,
          }
        }
        for (const asset of page) nextExcluded.delete(assetKey(asset))
        return {
          scope: scopeKey,
          allMatching: true,
          items: EMPTY_ITEMS,
          excluded: nextExcluded,
        }
      }
      const items = new Map(current?.items ?? EMPTY_ITEMS)
      const everySelected =
        page.length > 0 && page.every((asset) => items.has(assetKey(asset)))
      for (const asset of page) {
        if (everySelected) items.delete(assetKey(asset))
        else items.set(assetKey(asset), asset)
      }
      return {
        scope: scopeKey,
        allMatching: false,
        items,
        excluded: EMPTY_EXCLUDED,
      }
    })
  }, [scopeKey])

  const selectAllMatching = useCallback(() => {
    setState({
      scope: scopeKey,
      allMatching: true,
      items: EMPTY_ITEMS,
      excluded: EMPTY_EXCLUDED,
    })
  }, [scopeKey])

  const removeKeys = useCallback(
    (keys: string[]) => {
      if (keys.length === 0) return
      setState((previous) => {
        if (previous.scope !== scopeKey || previous.allMatching) return previous
        const items = new Map(previous.items)
        let changed = false
        for (const key of keys) {
          if (items.delete(key)) changed = true
        }
        if (!changed) return previous
        return {
          ...previous,
          items,
        }
      })
    },
    [scopeKey]
  )

  const clear = useCallback(() => setState(emptySelection), [emptySelection])

  const assets = useMemo(() => [...selected.values()], [selected])
  const excludedRefs = useMemo(() => [...excluded.values()], [excluded])
  const count = allMatching
    ? Math.max(0, totalMatching - excluded.size)
    : selected.size
  const canSelectAllMatching =
    !allMatching && allOnPageSelected && totalMatching > count

  return {
    assets,
    allMatching,
    excludedRefs,
    count,
    allOnPageSelected,
    someOnPageSelected,
    canSelectAllMatching,
    isSelected,
    toggle,
    selectKeys,
    removeKeys,
    selectPage,
    togglePage,
    selectAllMatching,
    clear,
  }
}
