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
import { useQuery } from '@tanstack/react-query'
import { ChevronsUpDown } from 'lucide-react'
import { useDeferredValue, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Separator } from '@/components/ui/separator'

import { listAssetGroupsPage } from '../api'
import { assertAssetSuccess, assetErrorMessage } from '../asset-utils'
import type { AssetGroup } from '../types'

const GROUP_PAGE_SIZE = 20

type AssetGroupSelectProps = {
  value: string
  selectedGroup?: AssetGroup
  isAdmin: boolean
  onValueChange: (value: string, group?: AssetGroup) => void
}

function AssetGroupOptionLabel(props: {
  group: AssetGroup
  isAdmin: boolean
  selected?: boolean
}) {
  const { t } = useTranslation()
  const creator = props.group.owner_name || `#${props.group.owner_user_id}`
  const metadata = props.isAdmin
    ? `${t('ID')}: ${props.group.id} · ${t('Creator')}: ${creator}`
    : `${t('ID')}: ${props.group.id}`

  return (
    <span className='flex w-full min-w-0 flex-1 flex-col items-start text-left leading-tight'>
      <span
        className='block w-full truncate font-medium'
        data-testid={props.selected ? 'selected-asset-group-name' : undefined}
      >
        {props.group.name}
      </span>
      <span
        aria-label={metadata}
        className='text-muted-foreground mt-0.5 flex w-full min-w-0 items-baseline gap-1 overflow-hidden text-[11px] leading-4 whitespace-nowrap'
        data-testid={
          props.selected ? 'selected-asset-group-metadata' : undefined
        }
        title={metadata}
      >
        <span className='shrink-0'>{t('ID')}: </span>
        <span
          className='min-w-0 truncate font-mono'
          data-testid={props.selected ? 'selected-asset-group-id' : undefined}
        >
          {props.group.id}
        </span>
        {props.isAdmin && (
          <>
            <span className='shrink-0'> · </span>
            <span className='shrink-0'>{t('Creator')}: </span>
            <span className='max-w-[35%] shrink-0 truncate'>{creator}</span>
          </>
        )}
      </span>
    </span>
  )
}

export function AssetGroupSelect(props: AssetGroupSelectProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const deferredSearch = useDeferredValue(search.trim())
  const searchPending = search.trim() !== deferredSearch
  const groupsQuery = useQuery({
    queryKey: ['asset-library', 'groups', props.isAdmin, page, deferredSearch],
    queryFn: async () =>
      assertAssetSuccess(
        await listAssetGroupsPage({
          includeAllOwners: props.isAdmin,
          page,
          pageSize: GROUP_PAGE_SIZE,
          search: deferredSearch,
        })
      ),
    enabled: open,
  })
  const groups = groupsQuery.data?.items ?? []
  const totalPages = Math.max(
    1,
    Math.ceil((groupsQuery.data?.total ?? 0) / GROUP_PAGE_SIZE)
  )

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button
            type='button'
            variant='outline'
            role='combobox'
            aria-label={t('Asset group')}
            aria-expanded={open}
            className='min-h-12 w-full justify-between gap-2 px-3 py-2'
          />
        }
      >
        {props.value && props.selectedGroup ? (
          <AssetGroupOptionLabel
            group={props.selectedGroup}
            isAdmin={props.isAdmin}
            selected
          />
        ) : (
          <span className='truncate'>{t('All groups')}</span>
        )}
        <ChevronsUpDown
          className='size-4 shrink-0 opacity-50'
          aria-hidden='true'
        />
      </PopoverTrigger>
      <PopoverContent
        align='start'
        className='w-[var(--anchor-width)] min-w-80 gap-2 p-2'
      >
        <Input
          autoFocus
          aria-label={t('Search groups...')}
          placeholder={t('Search groups...')}
          value={search}
          onChange={(event) => {
            setSearch(event.target.value)
            setPage(1)
          }}
        />
        <div
          role='listbox'
          aria-label={t('Asset group')}
          className='max-h-72 space-y-0.5 overflow-y-auto'
        >
          <button
            type='button'
            role='option'
            aria-selected={!props.value}
            className='hover:bg-accent w-full rounded-md px-2 py-2 text-left text-sm'
            onClick={() => {
              props.onValueChange('')
              setOpen(false)
            }}
          >
            {t('All groups')}
          </button>
          {(groupsQuery.isLoading || searchPending) && (
            <p className='text-muted-foreground px-2 py-3 text-sm'>
              {t('Loading...')}
            </p>
          )}
          {groupsQuery.isError && (
            <p role='alert' className='text-destructive px-2 py-3 text-sm'>
              {assetErrorMessage(groupsQuery.error)}
            </p>
          )}
          {!groupsQuery.isLoading &&
            !searchPending &&
            !groupsQuery.isError &&
            groups.length === 0 && (
              <p className='text-muted-foreground px-2 py-3 text-sm'>
                {t('No groups match your search')}
              </p>
            )}
          {!searchPending &&
            groups.map((group) => (
              <button
                key={group.id}
                type='button'
                role='option'
                aria-selected={props.value === group.id}
                aria-label={`${group.name} · ${t('ID')}: ${group.id}${props.isAdmin ? ` · ${t('Creator')}: ${group.owner_name || `#${group.owner_user_id}`}` : ''}`}
                className='hover:bg-accent w-full rounded-md px-2 py-2'
                onClick={() => {
                  props.onValueChange(group.id, group)
                  setOpen(false)
                }}
              >
                <AssetGroupOptionLabel group={group} isAdmin={props.isAdmin} />
              </button>
            ))}
        </div>
        {totalPages > 1 && (
          <>
            <Separator />
            <div className='flex items-center justify-between pt-1'>
              <Button
                type='button'
                variant='outline'
                size='sm'
                disabled={page <= 1 || groupsQuery.isFetching}
                onClick={() => setPage((current) => current - 1)}
              >
                {t('Previous')}
              </Button>
              <span
                className='text-muted-foreground text-xs tabular-nums'
                aria-live='polite'
              >
                {t('Page {{page}} of {{total}}', { page, total: totalPages })}
              </span>
              <Button
                type='button'
                variant='outline'
                size='sm'
                disabled={page >= totalPages || groupsQuery.isFetching}
                onClick={() => setPage((current) => current + 1)}
              >
                {t('Next')}
              </Button>
            </div>
          </>
        )}
      </PopoverContent>
    </Popover>
  )
}
