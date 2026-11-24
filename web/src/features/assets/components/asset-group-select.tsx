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
import { Folder, LayoutGrid } from 'lucide-react'
import { useDeferredValue, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { listAssetGroupsPage } from '../api'
import { assertAssetSuccess, assetErrorMessage } from '../asset-utils'
import type { AssetGroup } from '../types'
import { AssetPicker, AssetPickerTile } from './asset-picker'

const GROUP_PAGE_SIZE = 20

type AssetGroupSelectProps = {
  value: string
  selectedGroup?: AssetGroup
  /** Shown when the value is known but its group details are not loaded. */
  selectedLabel?: string
  isAdmin: boolean
  /** Limits the list to one user's groups. */
  ownerUserId?: number
  selectOnly?: boolean
  disabled?: boolean
  triggerClassName?: string
  onValueChange: (value: string, group?: AssetGroup) => void
}

function groupCreator(group: AssetGroup) {
  return group.owner_name || `#${group.owner_user_id}`
}

export function AssetGroupSelect(props: AssetGroupSelectProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const deferredSearch = useDeferredValue(search.trim())
  const searchPending = search.trim() !== deferredSearch
  const showCreator = props.isAdmin && !props.ownerUserId

  useEffect(() => {
    setPage(1)
    setSearch('')
  }, [props.ownerUserId])

  const groupsQuery = useQuery({
    queryKey: [
      'asset-library',
      'groups',
      props.isAdmin,
      props.ownerUserId,
      page,
      deferredSearch,
    ],
    queryFn: async () =>
      assertAssetSuccess(
        await listAssetGroupsPage({
          includeAllOwners: props.isAdmin,
          ownerUserId: props.ownerUserId,
          page,
          pageSize: GROUP_PAGE_SIZE,
          search: deferredSearch,
        })
      ),
    enabled: open && !props.disabled,
  })
  const groups = groupsQuery.data?.items ?? []
  const totalPages = Math.max(
    1,
    Math.ceil((groupsQuery.data?.total ?? 0) / GROUP_PAGE_SIZE)
  )
  const selectedLabel = props.selectedGroup?.name || props.selectedLabel
  const isSelected = (group: AssetGroup) =>
    props.selectedGroup
      ? props.selectedGroup.id === group.id &&
        props.selectedGroup.owner_user_id === group.owner_user_id
      : props.value === group.id
  const placeholder = props.selectOnly
    ? t('Select asset groups')
    : t('All groups')

  const trigger =
    props.value && selectedLabel ? (
      <span className='flex min-w-0 flex-1 items-center gap-2 text-left'>
        <AssetPickerTile size='sm'>
          <Folder className='size-3.5' strokeWidth={1.75} />
        </AssetPickerTile>
        <span
          className='min-w-0 truncate font-medium'
          data-testid='selected-asset-group-name'
          title={`${t('ID')}: ${props.value}`}
        >
          {selectedLabel}
        </span>
        {showCreator && props.selectedGroup && (
          <span
            className='bg-muted text-muted-foreground max-w-[40%] shrink-0 truncate rounded px-1.5 text-[11px]'
            data-testid='selected-asset-group-creator'
          >
            {groupCreator(props.selectedGroup)}
          </span>
        )}
      </span>
    ) : (
      <span className='flex min-w-0 flex-1 items-center gap-2 text-left'>
        <AssetPickerTile size='sm' tone='muted'>
          <Folder className='size-3.5' strokeWidth={1.75} />
        </AssetPickerTile>
        <span className='truncate'>{placeholder}</span>
      </span>
    )

  return (
    <AssetPicker
      label={t('Asset group')}
      open={open}
      onOpenChange={setOpen}
      disabled={props.disabled}
      triggerClassName={props.triggerClassName}
      trigger={trigger}
      searchLabel={t('Search groups...')}
      search={search}
      onSearchChange={(value) => {
        setSearch(value)
        setPage(1)
      }}
      allOption={
        props.selectOnly
          ? undefined
          : {
              label: t('All groups'),
              icon: (
                <AssetPickerTile tone='muted'>
                  <LayoutGrid className='size-4' strokeWidth={1.75} />
                </AssetPickerTile>
              ),
              selected: !props.value,
              onSelect: () => props.onValueChange(''),
            }
      }
      options={
        searchPending
          ? []
          : groups.map((group) => {
              const creator = groupCreator(group)
              return {
                key: `${group.owner_user_id}:${group.id}`,
                label: group.name,
                accessibleName: `${group.name} · ${t('ID')}: ${group.id}${showCreator ? ` · ${t('Creator')}: ${creator}` : ''}`,
                description: (
                  <>
                    {showCreator && (
                      <span className='text-foreground/80'>{creator} · </span>
                    )}
                    <span className='font-mono'>{group.id}</span>
                  </>
                ),
                badge:
                  group.asset_count === undefined ? undefined : (
                    <span className='bg-muted text-muted-foreground shrink-0 rounded-full px-1.5 text-[11px] tabular-nums'>
                      {group.asset_count}
                    </span>
                  ),
                icon: (
                  <AssetPickerTile>
                    <Folder className='size-4' strokeWidth={1.75} />
                  </AssetPickerTile>
                ),
                selected: isSelected(group),
                onSelect: () => props.onValueChange(group.id, group),
              }
            })
      }
      loading={groupsQuery.isLoading || searchPending}
      error={
        groupsQuery.isError ? assetErrorMessage(groupsQuery.error) : undefined
      }
      emptyLabel={t('No groups match your search')}
      page={page}
      totalPages={totalPages}
      paging={groupsQuery.isFetching}
      onPageChange={setPage}
    />
  )
}
