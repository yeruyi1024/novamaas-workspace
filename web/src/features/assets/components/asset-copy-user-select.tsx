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
import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { searchUsers } from '@/features/users/api'
import type { User } from '@/features/users/types'
import { useAuthStore } from '@/stores/auth-store'

import { assetErrorMessage } from '../asset-utils'

const PAGE_SIZE = 20

export function AssetCopyUserSelect(props: {
  id: string
  label: string
  value?: User
  excludeUserId: number
  enabledOnly?: boolean
  disabled?: boolean
  onValueChange: (user: User) => void
}) {
  const { t } = useTranslation()
  const isRoot = useAuthStore((state) => state.auth.user?.role === 100)
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const keyword = useDeferredValue(search.trim())
  // Search narrows the list; the first page of matches is enough here.
  const users = useQuery({
    queryKey: ['asset-copy-users', isRoot, Boolean(props.enabledOnly), keyword],
    queryFn: async () => {
      const response = await searchUsers({
        keyword,
        p: 1,
        page_size: PAGE_SIZE,
        role: isRoot ? '' : '1',
        status: props.enabledOnly ? '1' : '',
      })
      if (!response.success || !response.data) {
        throw new Error(response.message || t('Failed to load users'))
      }
      return response.data
    },
    enabled: open && !props.disabled,
    retry: false,
  })
  const loading = users.isFetching || keyword !== search.trim()
  const items = (users.data?.items ?? []).filter(
    (user) =>
      !user.DeletedAt &&
      (isRoot || user.role === 1) &&
      (!props.enabledOnly || user.status === 1)
  )

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button
            id={props.id}
            type='button'
            variant='outline'
            role='combobox'
            aria-label={props.label}
            aria-expanded={open}
            disabled={props.disabled}
            className='h-auto min-h-12 w-full justify-between gap-2 px-3 py-2 text-left'
          />
        }
      >
        <span className='min-w-0 flex-1'>
          <span className='block truncate'>
            {props.value?.username || t('Select a user')}
          </span>
          {props.value && (
            <span className='text-muted-foreground block truncate text-xs'>
              {props.value.display_name && `${props.value.display_name} · `}
              {t('ID')}: {props.value.id}
            </span>
          )}
        </span>
        <ChevronsUpDown
          className='size-4 shrink-0 opacity-50'
          aria-hidden='true'
        />
      </PopoverTrigger>
      <PopoverContent
        align='start'
        className='w-[var(--anchor-width)] max-w-[calc(100vw-2rem)] min-w-64 p-1'
      >
        <Command shouldFilter={false} label={t('Search users by name or ID')}>
          <CommandInput
            aria-label={t('Search users by name or ID')}
            placeholder={t('Search users by name or ID')}
            value={search}
            onValueChange={setSearch}
          />
          <CommandList label={props.label}>
            {loading && (
              <p role='status' className='text-muted-foreground p-3 text-sm'>
                {t('Loading...')}
              </p>
            )}
            {!loading && users.isError && (
              <p role='alert' className='text-destructive p-3 text-sm'>
                {assetErrorMessage(users.error)}
              </p>
            )}
            {!loading && users.isSuccess && items.length === 0 && (
              <p role='status' className='text-muted-foreground p-3 text-sm'>
                {t('No matching results')}
              </p>
            )}
            {!loading && !users.isError && (
              <CommandGroup>
                {items.map((user) => (
                  <CommandItem
                    key={user.id}
                    value={String(user.id)}
                    aria-label={`${user.username} · ${user.display_name} · ${t('ID')}: ${user.id}`}
                    disabled={user.id === props.excludeUserId || props.disabled}
                    onSelect={() => {
                      if (user.id !== props.value?.id) props.onValueChange(user)
                      setOpen(false)
                      setSearch('')
                    }}
                  >
                    <span className='min-w-0 flex-1'>
                      <span className='block truncate'>{user.username}</span>
                      <span className='text-muted-foreground block truncate text-xs'>
                        {user.display_name && `${user.display_name} · `}
                        {t('ID')}: {user.id}
                      </span>
                    </span>
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
