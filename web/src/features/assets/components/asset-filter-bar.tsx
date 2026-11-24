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
import { Search, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from '@/components/ui/input-group'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

import type {
  AssetGroup,
  AssetLibraryFilters,
  AssetOwner,
  AssetSortOrder,
  AssetStatusFilter,
  AssetSummary,
} from '../types'
import { AssetGroupSelect } from './asset-group-select'
import { AssetOwnerSelect } from './asset-owner-select'
import { AssetTypeTabs } from './asset-type-tabs'

const ALL = 'all'

const STATUS_OPTIONS = [
  { value: ALL, labelKey: 'All statuses' },
  { value: 'ready', labelKey: 'Ready' },
  { value: 'processing', labelKey: 'Under review' },
  { value: 'unavailable', labelKey: 'Unavailable' },
] as const

const SORT_OPTIONS = [
  { value: 'newest', labelKey: 'Newest first' },
  { value: 'oldest', labelKey: 'Oldest first' },
] as const

const TYPE_LABEL_KEYS = {
  image: 'Images',
  video: 'Videos',
  audio: 'Audio',
} as const

function FilterChip(props: { label: string; onRemove: () => void }) {
  const { t } = useTranslation()
  return (
    <span className='bg-primary/10 text-primary inline-flex h-6 max-w-56 items-center gap-1 rounded-full pr-1 pl-2.5 text-xs font-medium'>
      <span className='truncate'>{props.label}</span>
      <button
        type='button'
        aria-label={`${t('Remove filter')} ${props.label}`}
        className='hover:bg-primary/15 grid size-4 place-items-center rounded-full'
        onClick={props.onRemove}
      >
        <X className='size-3' strokeWidth={2} aria-hidden='true' />
      </button>
    </span>
  )
}

export function AssetFilterBar(props: {
  filters: AssetLibraryFilters
  summary?: AssetSummary
  total: number
  isAdmin: boolean
  search: string
  selectedGroup?: AssetGroup
  selectedGroupLabel?: string
  ownerLabel?: string
  onSearchChange: (value: string) => void
  onOwnerChange: (owner?: AssetOwner) => void
  onGroupChange: (value: string, group?: AssetGroup) => void
  onFiltersChange: (patch: Partial<AssetLibraryFilters>) => void
  onClearAll: () => void
}) {
  const { t } = useTranslation()
  const filters = props.filters
  const groupLabel = props.selectedGroup?.name || props.selectedGroupLabel
  const statusLabel = STATUS_OPTIONS.find(
    (option) => option.value === filters.status
  )?.labelKey
  const hasFilters = Boolean(
    filters.owner ||
    filters.group ||
    filters.type ||
    filters.status ||
    filters.q
  )
  const statusItems = STATUS_OPTIONS.map((option) => ({
    value: option.value,
    label: t(option.labelKey),
  }))
  const sortItems = SORT_OPTIONS.map((option) => ({
    value: option.value,
    label: t(option.labelKey),
  }))

  return (
    <section
      aria-label={t('Asset filters')}
      className='bg-card/70 flex flex-col gap-2 rounded-xl border px-2.5 py-2'
    >
      <div className='flex flex-wrap items-center gap-2'>
        {props.isAdmin && (
          <div className='w-full min-w-0 sm:w-36'>
            <AssetOwnerSelect
              value={filters.owner}
              selectedLabel={props.ownerLabel}
              triggerClassName='h-8 min-h-8 py-0.5 text-xs'
              onValueChange={props.onOwnerChange}
            />
          </div>
        )}
        <div className='w-full min-w-0 sm:w-44'>
          <AssetGroupSelect
            value={filters.group || ''}
            selectedGroup={props.selectedGroup}
            selectedLabel={props.selectedGroupLabel}
            isAdmin={props.isAdmin}
            ownerUserId={props.isAdmin ? filters.owner : undefined}
            triggerClassName='h-8 min-h-8 py-0.5 text-xs'
            onValueChange={props.onGroupChange}
          />
        </div>
        <div className='max-w-full shrink-0 overflow-x-auto'>
          <AssetTypeTabs
            value={filters.type}
            summary={props.summary}
            onChange={(type) => props.onFiltersChange({ type })}
          />
        </div>
        <InputGroup className='h-8 w-full min-w-44 sm:w-auto sm:flex-1'>
          <InputGroupAddon>
            <Search
              className='text-muted-foreground size-3.5'
              strokeWidth={1.75}
              aria-hidden='true'
            />
          </InputGroupAddon>
          <InputGroupInput
            value={props.search}
            aria-label={t('Search assets')}
            placeholder={t('Search by name, asset ID, or uploader')}
            className='text-xs'
            onChange={(event) => props.onSearchChange(event.target.value)}
          />
          {props.search && (
            <InputGroupAddon align='inline-end'>
              <InputGroupButton
                size='icon-xs'
                aria-label={t('Clear search')}
                onClick={() => props.onSearchChange('')}
              >
                <X className='size-3' strokeWidth={2} aria-hidden='true' />
              </InputGroupButton>
            </InputGroupAddon>
          )}
        </InputGroup>
        <div className='flex w-full shrink-0 gap-2 sm:w-auto'>
          <Select
            items={statusItems}
            value={filters.status || ALL}
            onValueChange={(value) =>
              props.onFiltersChange({
                status:
                  value && value !== ALL
                    ? (value as AssetStatusFilter)
                    : undefined,
              })
            }
          >
            <SelectTrigger
              aria-label={t('Asset status')}
              className='h-8 flex-1 text-xs sm:w-28 sm:flex-none'
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {statusItems.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select
            items={sortItems}
            value={filters.sort || 'newest'}
            onValueChange={(value) =>
              props.onFiltersChange({
                sort:
                  value === 'oldest' ? ('oldest' as AssetSortOrder) : undefined,
              })
            }
          >
            <SelectTrigger
              aria-label={t('Sort order')}
              className='h-8 flex-1 text-xs sm:w-26 sm:flex-none'
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {sortItems.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
      {hasFilters && (
        <div className='border-border/60 flex flex-wrap items-center gap-1.5 border-t pt-2 text-xs'>
          {filters.owner && (
            <FilterChip
              label={`${t('User')}: ${props.ownerLabel || `#${filters.owner}`}`}
              onRemove={() => props.onOwnerChange(undefined)}
            />
          )}
          {filters.group && groupLabel && (
            <FilterChip
              label={`${t('Group')}: ${groupLabel}`}
              onRemove={() => props.onGroupChange('')}
            />
          )}
          {filters.type && (
            <FilterChip
              label={t(TYPE_LABEL_KEYS[filters.type])}
              onRemove={() => props.onFiltersChange({ type: undefined })}
            />
          )}
          {filters.status && statusLabel && (
            <FilterChip
              label={t(statusLabel)}
              onRemove={() => props.onFiltersChange({ status: undefined })}
            />
          )}
          {filters.q && (
            <FilterChip
              label={`“${filters.q}”`}
              onRemove={() => props.onSearchChange('')}
            />
          )}
          <span
            className='text-muted-foreground ml-auto text-xs tabular-nums'
            aria-live='polite'
          >
            {t('{{count}} assets', { count: props.total })}
          </span>
          <Button
            type='button'
            size='xs'
            variant='ghost'
            className='text-muted-foreground h-6'
            onClick={props.onClearAll}
          >
            {t('Clear filters')}
          </Button>
        </div>
      )}
    </section>
  )
}
