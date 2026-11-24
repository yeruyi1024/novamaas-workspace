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
import { Layers, User, UserX } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'

import type { AssetSummary } from '../types'

/**
 * Compact administrator context bar shown when the library is filtered to one
 * user (and not narrowed to a single group): surfaces their live counts and
 * lets admins jump straight into copying or clearing that user's assets
 * without searching for the user a second time.
 */
export function AssetOwnerHeader(props: {
  ownerId: number
  ownerName?: string
  groupCount?: number
  summary?: AssetSummary
  onCopyGroups: () => void
  onPurgeOwner: () => void
}) {
  const { t } = useTranslation()
  const displayName = props.ownerName || `#${props.ownerId}`
  const unavailableCount = props.summary?.by_status.unavailable ?? 0

  return (
    <section
      aria-label={t('User asset summary')}
      className='bg-card/80 flex flex-wrap items-center justify-between gap-3 rounded-xl border px-3.5 py-2.5'
    >
      <div className='flex min-w-0 flex-wrap items-center gap-2.5'>
        <span className='bg-muted text-foreground/80 ring-border/70 grid size-8 shrink-0 place-items-center rounded-lg ring-1'>
          <User className='size-4' strokeWidth={1.75} aria-hidden='true' />
        </span>
        <div className='flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-sm'>
          <span className='font-semibold break-words'>{displayName}</span>
          <span className='text-muted-foreground font-mono text-xs'>
            ({t('ID')}: {props.ownerId})
          </span>
          {props.groupCount !== undefined && (
            <span className='bg-muted text-muted-foreground rounded-full px-2 py-0.5 text-xs tabular-nums'>
              {t('{{count}} groups', { count: props.groupCount })}
            </span>
          )}
          {props.summary !== undefined && (
            <span className='bg-muted text-muted-foreground rounded-full px-2 py-0.5 text-xs tabular-nums'>
              {t('{{count}} assets', { count: props.summary.total })}
            </span>
          )}
          {unavailableCount > 0 && (
            <span className='bg-destructive/10 text-destructive rounded-full px-2 py-0.5 text-xs font-medium tabular-nums'>
              {t('{{count}} unavailable', { count: unavailableCount })}
            </span>
          )}
        </div>
      </div>

      <div className='flex shrink-0 flex-wrap items-center gap-2'>
        <Button size='sm' variant='outline' onClick={props.onCopyGroups}>
          <Layers className='size-3.5' strokeWidth={1.75} aria-hidden='true' />
          {t('Copy asset groups')}
        </Button>
        <Button
          size='sm'
          variant='outline'
          className='text-destructive border-destructive/30 hover:bg-destructive/10 hover:text-destructive'
          onClick={props.onPurgeOwner}
        >
          <UserX className='size-3.5' strokeWidth={1.75} aria-hidden='true' />
          {t('Clear this user’s assets')}
        </Button>
      </div>
    </section>
  )
}
