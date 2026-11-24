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
import { CheckSquare, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'

/**
 * Compact netdisk-style selection strip right above the gallery grid: exposes
 * an obvious "select page" tri-state checkbox, cross-page "select all matching"
 * promotion, and a persistent batch-manage toggle for touch and rapid audits.
 */
export function AssetSelectionStrip(props: {
  pageCount: number
  totalCount: number
  selectedCount: number
  allOnPageSelected: boolean
  someOnPageSelected: boolean
  allMatching: boolean
  canSelectAllMatching: boolean
  batchMode: boolean
  onTogglePage: () => void
  onSelectAllMatching: () => void
  onClearSelection: () => void
  onToggleBatchMode: () => void
}) {
  const { t } = useTranslation()
  if (props.pageCount === 0) return null

  let selectionLabel = t('Select page')
  if (props.selectedCount > 0) {
    selectionLabel = props.allMatching
      ? t('All {{count}} matching assets selected', {
          count: props.selectedCount,
        })
      : t('{{count}} selected', { count: props.selectedCount })
  }

  return (
    <div
      data-testid='asset-selection-strip'
      className='flex min-h-8 flex-wrap items-center justify-between gap-x-3 gap-y-1.5 px-1 text-xs'
    >
      <div className='flex flex-wrap items-center gap-2.5'>
        <div className='text-foreground/90 inline-flex items-center gap-2 font-medium select-none'>
          <Checkbox
            id='asset-select-page-checkbox'
            className='border-muted-foreground/60'
            checked={props.allOnPageSelected}
            indeterminate={props.someOnPageSelected}
            onCheckedChange={() => props.onTogglePage()}
          />
          <Label
            htmlFor='asset-select-page-checkbox'
            className='cursor-pointer text-xs font-medium'
          >
            {selectionLabel}
          </Label>
        </div>

        {props.selectedCount === 0 && (
          <span className='text-muted-foreground tabular-nums'>
            {t('{{count}} assets', { count: props.totalCount })}
          </span>
        )}

        {props.canSelectAllMatching && (
          <Button
            type='button'
            size='xs'
            variant='link'
            className='h-6 px-1 text-xs'
            onClick={props.onSelectAllMatching}
          >
            {t('Select all {{count}} matching assets', {
              count: props.totalCount,
            })}
          </Button>
        )}

        {props.selectedCount > 0 && (
          <Button
            type='button'
            size='xs'
            variant='ghost'
            className='text-muted-foreground h-6 px-1.5 text-xs'
            onClick={props.onClearSelection}
          >
            <X className='size-3' strokeWidth={2} aria-hidden='true' />
            {t('Clear selection')}
          </Button>
        )}
      </div>

      <div className='flex items-center gap-3'>
        <span className='text-muted-foreground hidden text-[11px] xl:inline'>
          {t('Drag to select · Shift range · Delete to remove')}
        </span>
        <Button
          type='button'
          size='xs'
          variant={props.batchMode ? 'secondary' : 'ghost'}
          aria-pressed={props.batchMode}
          className='h-7 gap-1.5 px-2 text-xs'
          onClick={props.onToggleBatchMode}
        >
          <CheckSquare
            className='size-3.5'
            strokeWidth={1.75}
            aria-hidden='true'
          />
          {props.batchMode ? t('Done selecting') : t('Batch manage')}
        </Button>
      </div>
    </div>
  )
}
