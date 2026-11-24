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
import { Check, ChevronsUpDown, UserRound } from 'lucide-react'
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
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

import { assetErrorMessage } from '../asset-utils'
import { AssetPickerTile } from './asset-picker'

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
            className='border-border/80 bg-background/80 hover:border-primary/40 hover:bg-background h-auto min-h-12 w-full justify-between gap-2.5 rounded-xl px-3 py-2 text-left transition-colors'
          />
        }
      >
        <span className='flex min-w-0 flex-1 items-center gap-2.5'>
          <AssetPickerTile size='sm' tone={props.value ? 'primary' : 'muted'}>
            <UserRound className='size-3.5' strokeWidth={1.75} />
          </AssetPickerTile>
          <span className='min-w-0 flex-1'>
            <span
              className={cn(
                'block truncate text-sm',
                props.value
                  ? 'text-foreground font-medium'
                  : 'text-muted-foreground font-normal'
              )}
            >
              {props.value?.username || t('Select a user')}
            </span>
            {props.value && (
              <span className='text-muted-foreground block truncate text-[11px] tabular-nums'>
                {props.value.display_name && `${props.value.display_name} · `}
                {t('ID')}: {props.value.id}
              </span>
            )}
          </span>
        </span>
        <ChevronsUpDown
          className='size-4 shrink-0 opacity-50'
          aria-hidden='true'
        />
      </PopoverTrigger>
      <PopoverContent
        align='start'
        className='border-border/80 w-[var(--anchor-width)] max-w-[calc(100vw-2rem)] min-w-72 rounded-xl p-1.5 shadow-xl'
      >
        <Command shouldFilter={false} label={t('Search users by name or ID')}>
          <CommandInput
            aria-label={t('Search users by name or ID')}
            placeholder={t('Search users by name or ID')}
            value={search}
            onValueChange={setSearch}
          />
          <CommandList label={props.label} className='max-h-64 p-1'>
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
              <CommandGroup className='p-0'>
                {items.map((user) => {
                  const selected = user.id === props.value?.id
                  return (
                    <CommandItem
                      key={user.id}
                      value={String(user.id)}
                      aria-label={`${user.username} · ${user.display_name} · ${t('ID')}: ${user.id}`}
                      disabled={
                        user.id === props.excludeUserId || props.disabled
                      }
                      className={cn(
                        'flex items-center gap-2.5 rounded-lg px-2.5 py-2',
                        selected && 'bg-accent/70'
                      )}
                      onSelect={() => {
                        if (user.id !== props.value?.id) {
                          props.onValueChange(user)
                        }
                        setOpen(false)
                        setSearch('')
                      }}
                    >
                      <AssetPickerTile
                        size='sm'
                        tone={selected ? 'primary' : 'muted'}
                      >
                        <UserRound className='size-3.5' strokeWidth={1.75} />
                      </AssetPickerTile>
                      <span className='min-w-0 flex-1'>
                        <span className='block truncate text-sm font-medium'>
                          {user.username}
                        </span>
                        <span className='text-muted-foreground block truncate text-[11px] tabular-nums'>
                          {user.display_name && `${user.display_name} · `}
                          {t('ID')}: {user.id}
                        </span>
                      </span>
                      <Check
                        aria-hidden='true'
                        className={cn(
                          'text-primary size-4 shrink-0',
                          !selected && 'invisible'
                        )}
                      />
                    </CommandItem>
                  )
                })}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
