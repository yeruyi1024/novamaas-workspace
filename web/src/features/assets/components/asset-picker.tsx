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
import {
  ArrowLeft01Icon,
  ArrowRight01Icon,
  Search01Icon,
  Tick02Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { ChevronsUpDown } from 'lucide-react'
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { cn } from '@/lib/utils'

type AssetPickerOption = {
  key: string
  label: string
  /** Full accessible name, e.g. including IDs that are shown abbreviated. */
  accessibleName?: string
  description?: ReactNode
  badge?: ReactNode
  icon: ReactNode
  selected: boolean
  onSelect: () => void
}

/**
 * Searchable popover list shared by the user and group filters: an icon tile
 * per row, a secondary line, a count badge and a check on the chosen row.
 */
export function AssetPicker(props: {
  label: string
  open: boolean
  onOpenChange: (open: boolean) => void
  disabled?: boolean
  triggerClassName?: string
  trigger: ReactNode
  searchLabel: string
  search: string
  onSearchChange: (value: string) => void
  allOption?: {
    label: string
    icon: ReactNode
    selected: boolean
    onSelect: () => void
  }
  options: AssetPickerOption[]
  loading: boolean
  error?: string
  emptyLabel: string
  page: number
  totalPages: number
  paging: boolean
  onPageChange: (page: number) => void
}) {
  const { t } = useTranslation()
  const allOption = props.allOption
  const choose = (onSelect: () => void) => {
    onSelect()
    props.onOpenChange(false)
  }

  return (
    <Popover open={props.open} onOpenChange={props.onOpenChange}>
      <PopoverTrigger
        render={
          <Button
            type='button'
            disabled={props.disabled}
            variant='outline'
            role='combobox'
            aria-label={props.label}
            aria-expanded={props.open}
            className={cn(
              'min-h-12 w-full justify-between gap-2 px-2.5 py-2',
              props.triggerClassName
            )}
          />
        }
      >
        {props.trigger}
        <ChevronsUpDown
          className='size-4 shrink-0 opacity-50'
          aria-hidden='true'
        />
      </PopoverTrigger>
      <PopoverContent
        align='start'
        className='w-[min(24rem,calc(100vw-2rem))] min-w-[var(--anchor-width)] gap-0 overflow-hidden p-0'
      >
        <div className='border-b p-2'>
          <div className='bg-muted/60 focus-within:ring-ring/40 flex h-9 items-center gap-2 rounded-lg px-2.5 focus-within:ring-2'>
            <HugeiconsIcon
              icon={Search01Icon}
              className='text-muted-foreground size-4 shrink-0'
            />
            <input
              autoFocus
              aria-label={props.searchLabel}
              placeholder={props.searchLabel}
              value={props.search}
              className='placeholder:text-muted-foreground min-w-0 flex-1 bg-transparent text-sm outline-none'
              onChange={(event) => props.onSearchChange(event.target.value)}
            />
          </div>
        </div>
        <div
          role='listbox'
          aria-label={props.label}
          className='max-h-[min(60vh,26rem)] overflow-y-auto p-1.5'
        >
          {allOption && (
            <button
              type='button'
              role='option'
              aria-selected={allOption.selected}
              className={cn(
                'hover:bg-accent flex w-full items-center gap-2.5 rounded-lg px-2 py-1.5 text-left text-sm',
                allOption.selected && 'bg-accent/70'
              )}
              onClick={() => choose(allOption.onSelect)}
            >
              {allOption.icon}
              <span className='flex-1 font-medium'>{allOption.label}</span>
              {allOption.selected && (
                <HugeiconsIcon
                  icon={Tick02Icon}
                  className='text-primary size-4'
                />
              )}
            </button>
          )}
          {props.loading && (
            <p className='text-muted-foreground px-2 py-3 text-sm'>
              {t('Loading...')}
            </p>
          )}
          {!props.loading && props.error && (
            <p role='alert' className='text-destructive px-2 py-3 text-sm'>
              {props.error}
            </p>
          )}
          {!props.loading && !props.error && props.options.length === 0 && (
            <p className='text-muted-foreground px-2 py-3 text-sm'>
              {props.emptyLabel}
            </p>
          )}
          {!props.loading &&
            props.options.map((option) => (
              <button
                key={option.key}
                type='button'
                role='option'
                aria-selected={option.selected}
                aria-label={option.accessibleName ?? option.label}
                className={cn(
                  'hover:bg-accent flex w-full items-center gap-2.5 rounded-lg px-2 py-1.5 text-left',
                  option.selected && 'bg-accent/70'
                )}
                onClick={() => choose(option.onSelect)}
              >
                {option.icon}
                <span className='flex min-w-0 flex-1 flex-col leading-tight'>
                  <span className='truncate text-sm font-medium'>
                    {option.label}
                  </span>
                  {option.description && (
                    <span className='text-muted-foreground mt-0.5 truncate text-[11px]'>
                      {option.description}
                    </span>
                  )}
                </span>
                {option.badge}
                <HugeiconsIcon
                  icon={Tick02Icon}
                  aria-hidden='true'
                  className={cn(
                    'text-primary size-4 shrink-0',
                    !option.selected && 'invisible'
                  )}
                />
              </button>
            ))}
        </div>
        {props.totalPages > 1 && (
          <div className='text-muted-foreground flex items-center justify-between border-t px-2 py-1.5 text-xs'>
            <Button
              type='button'
              size='icon-sm'
              variant='ghost'
              aria-label={t('Previous')}
              disabled={props.page <= 1 || props.paging}
              onClick={() => props.onPageChange(props.page - 1)}
            >
              <HugeiconsIcon icon={ArrowLeft01Icon} />
            </Button>
            <span className='tabular-nums' aria-live='polite'>
              {t('Page {{page}} of {{total}}', {
                page: props.page,
                total: props.totalPages,
              })}
            </span>
            <Button
              type='button'
              size='icon-sm'
              variant='ghost'
              aria-label={t('Next')}
              disabled={props.page >= props.totalPages || props.paging}
              onClick={() => props.onPageChange(props.page + 1)}
            >
              <HugeiconsIcon icon={ArrowRight01Icon} />
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}

/** Square icon tile used at the start of picker rows and triggers. */
export function AssetPickerTile(props: {
  children: ReactNode
  tone?: 'primary' | 'muted'
  size?: 'sm' | 'md'
}) {
  return (
    <span
      aria-hidden='true'
      className={cn(
        'grid shrink-0 place-items-center rounded-md text-xs font-semibold uppercase',
        props.size === 'sm' ? 'size-5' : 'size-8',
        props.tone === 'muted'
          ? 'bg-muted text-muted-foreground'
          : 'bg-primary/10 text-primary'
      )}
    >
      {props.children}
    </span>
  )
}
