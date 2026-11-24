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
import { Folder, Pencil, Trash2, Upload, User } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import { Button } from '@/components/ui/button'
import { formatDateStr } from '@/lib/format'

import type { AssetGroup } from '../types'

/**
 * Context for the chosen group: what it is, who owns it, how many assets it
 * holds, and the actions that apply to the whole group.
 */
export function AssetGroupHeader(props: {
  groupId: string
  group?: AssetGroup
  fallbackName?: string
  isAdmin: boolean
  canUpload: boolean
  onUpload: () => void
  onRename: () => void
  onDelete: () => void
}) {
  const { t } = useTranslation()
  const group = props.group
  const name = group?.name || props.fallbackName || props.groupId
  const owner = group ? group.owner_name || `#${group.owner_user_id}` : ''

  return (
    <section
      aria-label={t('Group details')}
      className='bg-card/80 relative overflow-hidden rounded-xl border p-4 shadow-xs sm:p-5'
    >
      <div className='flex flex-col gap-4 sm:flex-row sm:items-start'>
        <span className='bg-muted text-foreground/80 ring-border/70 grid size-11 shrink-0 place-items-center rounded-xl ring-1'>
          <Folder className='size-5' strokeWidth={1.75} aria-hidden='true' />
        </span>
        <div className='min-w-0 flex-1 space-y-2'>
          <div className='flex flex-wrap items-center gap-2'>
            <h2 className='text-base font-semibold tracking-tight break-words'>
              {name}
            </h2>
            {group?.asset_count !== undefined && (
              <span className='bg-muted text-muted-foreground shrink-0 rounded-full px-2 py-0.5 text-xs font-medium tabular-nums'>
                {t('{{count}} assets', { count: group.asset_count })}
              </span>
            )}
          </div>
          {group?.description && (
            <p className='text-muted-foreground text-sm break-words'>
              {group.description}
            </p>
          )}
          <div className='text-muted-foreground flex flex-wrap items-center gap-x-4 gap-y-1.5 text-xs'>
            <span className='bg-muted/60 ring-border/60 inline-flex items-center gap-1.5 rounded-md py-0.5 pr-0.5 pl-2 ring-1'>
              <span>{t('ID')}</span>
              <code className='text-foreground/90 font-mono text-[11px] break-all select-all'>
                {props.groupId}
              </code>
              <CopyButton
                value={props.groupId}
                className='size-5 [&_svg]:size-3'
                tooltip={t('Copy group ID')}
                successTooltip={t('Copied')}
              />
            </span>
            {props.isAdmin && group && (
              <span className='inline-flex items-center gap-1.5'>
                <User
                  className='size-3.5 shrink-0'
                  strokeWidth={1.75}
                  aria-hidden='true'
                />
                <span className='text-foreground/90 font-medium'>{owner}</span>
                <span className='font-mono text-[11px]'>
                  ({t('ID')}: {group.owner_user_id})
                </span>
              </span>
            )}
            {group && group.created_at > 0 && (
              <span>
                {t('Created')}{' '}
                {formatDateStr(new Date(group.created_at * 1000))}
              </span>
            )}
          </div>
        </div>
        <div className='flex shrink-0 flex-wrap gap-2'>
          {props.canUpload && (
            <Button size='sm' onClick={props.onUpload}>
              <Upload
                className='size-3.5'
                strokeWidth={1.75}
                aria-hidden='true'
              />
              {t('Upload here')}
            </Button>
          )}
          <Button size='sm' variant='outline' onClick={props.onRename}>
            <Pencil
              className='size-3.5'
              strokeWidth={1.75}
              aria-hidden='true'
            />
            {t('Rename group')}
          </Button>
          <Button
            size='sm'
            variant='outline'
            className='text-destructive hover:bg-destructive/10 hover:text-destructive'
            onClick={props.onDelete}
          >
            <Trash2
              className='size-3.5'
              strokeWidth={1.75}
              aria-hidden='true'
            />
            {t('Delete group')}
          </Button>
        </div>
      </div>
    </section>
  )
}
