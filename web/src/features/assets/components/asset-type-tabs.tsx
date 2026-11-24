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
import { AudioLines, Film, Image, LayoutGrid } from 'lucide-react'
import { motion, useReducedMotion } from 'motion/react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

import type { AssetSummary, MediaAssetType } from '../types'

const TABS = [
  { value: undefined, labelKey: 'All', Icon: LayoutGrid },
  { value: 'image', labelKey: 'Images', Icon: Image },
  { value: 'video', labelKey: 'Videos', Icon: Film },
  { value: 'audio', labelKey: 'Audio', Icon: AudioLines },
] as const

/** Segmented type filter whose active pill slides between options. */
export function AssetTypeTabs(props: {
  value?: MediaAssetType
  summary?: AssetSummary
  onChange: (value?: MediaAssetType) => void
}) {
  const { t } = useTranslation()
  const reduceMotion = useReducedMotion()

  return (
    <div
      role='radiogroup'
      aria-label={t('Asset type')}
      className='bg-muted/70 inline-flex h-8 shrink-0 items-center gap-0.5 rounded-lg p-0.5'
    >
      {TABS.map((tab) => {
        const active = props.value === tab.value
        const count = tab.value
          ? props.summary?.by_type[tab.value]
          : props.summary?.total
        const TabIcon = tab.Icon
        return (
          <button
            key={tab.labelKey}
            type='button'
            role='radio'
            aria-checked={active}
            className={cn(
              'focus-visible:outline-ring relative flex h-7 shrink-0 items-center gap-1 rounded-md px-2 text-xs font-medium whitespace-nowrap transition-colors focus-visible:outline-2',
              active
                ? 'text-foreground'
                : 'text-muted-foreground hover:text-foreground',
              !active && count === 0 && 'opacity-55'
            )}
            onClick={() => props.onChange(tab.value)}
          >
            {active && (
              <motion.span
                layoutId='asset-type-tab-pill'
                className='bg-background ring-border/60 absolute inset-0 rounded-md shadow-sm ring-1'
                transition={
                  reduceMotion
                    ? { duration: 0 }
                    : { type: 'spring', stiffness: 500, damping: 38 }
                }
              />
            )}
            <TabIcon
              className='relative size-3.5 shrink-0'
              strokeWidth={1.75}
              aria-hidden='true'
            />
            <span className='relative'>{t(tab.labelKey)}</span>
            {count !== undefined && (
              <span
                className={cn(
                  'relative rounded-full px-1.5 text-[11px] tabular-nums',
                  active ? 'bg-primary/10 text-primary' : 'bg-foreground/5'
                )}
              >
                {count}
              </span>
            )}
          </button>
        )
      })}
    </div>
  )
}
