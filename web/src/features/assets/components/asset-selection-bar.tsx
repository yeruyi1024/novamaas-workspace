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
import { Check, CheckCheck, Copy, Trash2, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Kbd } from '@/components/ui/kbd'
import { Separator } from '@/components/ui/separator'

/** Floating netdisk batch-action bar shown while assets are selected. */
export function AssetSelectionBar(props: {
  count: number
  totalMatching: number
  allOnPageSelected: boolean
  allMatching: boolean
  canSelectAllMatching: boolean
  copyableCount: number
  deleting: boolean
  onTogglePage: () => void
  onSelectAllMatching: () => void
  onCopyReferences: () => void
  onDelete: () => void
  onClear: () => void
}) {
  const { t } = useTranslation()
  if (props.count === 0) return null

  return (
    <div
      role='toolbar'
      aria-label={t('Selected assets')}
      className='bg-background/90 ring-border/70 animate-in fade-in-0 slide-in-from-bottom-4 zoom-in-95 sticky bottom-3 z-30 mx-auto flex w-fit max-w-full flex-wrap items-center gap-1 rounded-2xl px-2.5 py-1.5 shadow-2xl ring-1 backdrop-blur-xl duration-200 motion-reduce:animate-none'
    >
      <span className='flex items-center gap-1.5 px-2 text-sm font-medium'>
        <span className='bg-primary text-primary-foreground relative grid h-6 min-w-6 place-items-center overflow-hidden rounded-full px-1.5 text-xs tabular-nums'>
          {props.count}
        </span>
        {props.allMatching ? t('matching selected') : t('selected')}
      </span>
      <span className='text-muted-foreground hidden items-center gap-1 text-[11px] xl:flex'>
        <Kbd>Shift</Kbd>
        {t('range')}
        <Kbd className='ml-1.5'>Del</Kbd>
        {t('Delete')}
        <Kbd className='ml-1.5'>Esc</Kbd>
        {t('clear')}
      </span>
      <Separator orientation='vertical' className='mx-1 h-5' />
      <Button
        type='button'
        size='sm'
        variant='ghost'
        onClick={props.onTogglePage}
      >
        <Check className='size-3.5' strokeWidth={1.85} aria-hidden='true' />
        {props.allOnPageSelected ? t('Deselect page') : t('Select page')}
      </Button>
      {props.canSelectAllMatching && (
        <Button
          type='button'
          size='sm'
          variant='ghost'
          onClick={props.onSelectAllMatching}
        >
          <CheckCheck
            className='size-3.5'
            strokeWidth={1.85}
            aria-hidden='true'
          />
          {t('Select all {{count}}', { count: props.totalMatching })}
        </Button>
      )}
      {!props.allMatching && (
        <Button
          type='button'
          size='sm'
          variant='ghost'
          disabled={props.copyableCount === 0}
          onClick={props.onCopyReferences}
        >
          <Copy className='size-3.5' strokeWidth={1.75} aria-hidden='true' />
          {t('Copy references')}
        </Button>
      )}
      <Button
        type='button'
        size='sm'
        variant='destructive'
        disabled={props.deleting}
        onClick={props.onDelete}
      >
        <Trash2 className='size-3.5' strokeWidth={1.75} aria-hidden='true' />
        {t('Delete')}
      </Button>
      <Button
        type='button'
        size='icon-sm'
        variant='ghost'
        aria-label={t('Clear selection')}
        onClick={props.onClear}
      >
        <X className='size-4' strokeWidth={1.75} aria-hidden='true' />
      </Button>
    </div>
  )
}
