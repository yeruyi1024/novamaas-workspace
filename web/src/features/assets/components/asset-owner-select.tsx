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
import { User, Users } from 'lucide-react'
import { useDeferredValue, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { listAssetOwners } from '../api'
import { assertAssetSuccess, assetErrorMessage } from '../asset-utils'
import type { AssetOwner } from '../types'
import { AssetPicker, AssetPickerTile } from './asset-picker'

const OWNER_PAGE_SIZE = 20

/**
 * Admin filter for whose assets to browse. It lists only users who own asset
 * groups, with their group and asset counts.
 */
export function AssetOwnerSelect(props: {
  value?: number
  /** Name to show when the chosen user's details are not loaded. */
  selectedLabel?: string
  triggerClassName?: string
  onValueChange: (owner?: AssetOwner) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const deferredSearch = useDeferredValue(search.trim())
  const searchPending = search.trim() !== deferredSearch
  const ownersQuery = useQuery({
    queryKey: ['asset-library', 'owners', page, deferredSearch],
    queryFn: async () =>
      assertAssetSuccess(
        await listAssetOwners({
          search: deferredSearch,
          page,
          pageSize: OWNER_PAGE_SIZE,
        })
      ),
    enabled: open,
  })
  const owners = ownersQuery.data?.items ?? []
  const totalPages = Math.max(
    1,
    Math.ceil((ownersQuery.data?.total ?? 0) / OWNER_PAGE_SIZE)
  )
  const label = props.selectedLabel || (props.value ? `#${props.value}` : '')

  const trigger = props.value ? (
    <span className='flex min-w-0 flex-1 items-center gap-2 text-left'>
      <AssetPickerTile size='sm'>
        <User className='size-3.5' strokeWidth={1.75} />
      </AssetPickerTile>
      <span
        className='min-w-0 truncate font-medium'
        data-testid='selected-asset-owner'
      >
        {label}
      </span>
    </span>
  ) : (
    <span className='flex min-w-0 flex-1 items-center gap-2 text-left'>
      <AssetPickerTile size='sm' tone='muted'>
        <Users className='size-3.5' strokeWidth={1.75} />
      </AssetPickerTile>
      <span className='truncate'>{t('All users')}</span>
    </span>
  )

  return (
    <AssetPicker
      label={t('Asset owner')}
      open={open}
      onOpenChange={setOpen}
      triggerClassName={props.triggerClassName}
      trigger={trigger}
      searchLabel={t('Search users by name or ID')}
      search={search}
      onSearchChange={(value) => {
        setSearch(value)
        setPage(1)
      }}
      allOption={{
        label: t('All users'),
        icon: (
          <AssetPickerTile tone='muted'>
            <Users className='size-4' strokeWidth={1.75} />
          </AssetPickerTile>
        ),
        selected: !props.value,
        onSelect: () => props.onValueChange(undefined),
      }}
      options={
        searchPending
          ? []
          : owners.map((owner) => {
              const name = owner.username || `#${owner.user_id}`
              return {
                key: String(owner.user_id),
                label: name,
                accessibleName: `${name} · ${t('ID')}: ${owner.user_id}`,
                description: t(
                  'ID {{id}} · {{groups}} groups · {{assets}} assets',
                  {
                    id: owner.user_id,
                    groups: owner.group_count,
                    assets: owner.asset_count,
                  }
                ),
                icon: (
                  <AssetPickerTile>
                    <User className='size-4' strokeWidth={1.75} />
                  </AssetPickerTile>
                ),
                selected: props.value === owner.user_id,
                onSelect: () => props.onValueChange(owner),
              }
            })
      }
      loading={ownersQuery.isLoading || searchPending}
      error={
        ownersQuery.isError ? assetErrorMessage(ownersQuery.error) : undefined
      }
      emptyLabel={t('No matching results')}
      page={page}
      totalPages={totalPages}
      paging={ownersQuery.isFetching}
      onPageChange={setPage}
    />
  )
}
