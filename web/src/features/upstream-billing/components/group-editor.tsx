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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useDebounce } from '@/hooks/use-debounce'
import { handleServerError } from '@/lib/handle-server-error'

import { getUpstreamChannels, saveUpstreamGroup } from '../api'
import type { UpstreamGroup } from '../types'

export function GroupEditor(props: {
  group: UpstreamGroup | null
  onClose: () => void
  onSaved: (id: number) => void
}) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const [name, setName] = useState(props.group?.name || '')
  const [selected, setSelected] = useState(
    () => props.group?.channels.map((channel) => channel.id) || []
  )
  const [search, setSearch] = useState('')
  const [tag, setTag] = useState<string | null>(null)
  const keyword = useDebounce(search, 300)
  const channels = useQuery({
    queryKey: ['upstream-billing', 'channels', keyword],
    queryFn: () => getUpstreamChannels(keyword),
  })
  const save = useMutation({
    mutationFn: () =>
      saveUpstreamGroup({
        id: props.group?.id,
        name: name.trim(),
        version: props.group?.version || 0,
        channel_ids: selected,
      }),
    onSuccess: (group) => {
      void client.invalidateQueries({ queryKey: ['upstream-billing'] })
      props.onSaved(group.id)
      props.onClose()
    },
    onError: handleServerError,
  })
  const tags = [
    { value: null, label: t('All tags') },
    ...[
      ...new Set(
        (channels.data || [])
          .map((channel) => channel.tag)
          .filter((value): value is string => Boolean(value))
      ),
    ]
      .sort()
      .map((value) => ({ value, label: value })),
  ]
  const candidates = [
    ...(channels.data || []),
    ...(props.group?.channels.filter(
      (channel) =>
        channel.missing &&
        !(channels.data || []).some((current) => current.id === channel.id)
    ) || []),
  ]
  const visible = candidates.filter(
    (channel) => tag === null || channel.tag === tag
  )
  return (
    <Dialog
      open
      onOpenChange={(open) => !save.isPending && !open && props.onClose()}
    >
      <DialogContent className='max-h-[90dvh] overflow-y-auto sm:max-w-2xl'>
        <DialogHeader>
          <DialogTitle>
            {props.group
              ? t('Edit upstream group')
              : t('Create upstream group')}
          </DialogTitle>
          <DialogDescription>
            {t(
              'An upstream group combines channels for accounting only. Routing tags and billing groups remain unchanged.'
            )}
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(event) => {
            event.preventDefault()
            if (
              name.trim() &&
              selected.length > 0 &&
              !save.isPending &&
              !channels.isError
            ) {
              save.mutate()
            }
          }}
          className='flex flex-col gap-4'
        >
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor='upstream-group-name'>
                {t('Upstream name')}
              </FieldLabel>
              <Input
                id='upstream-group-name'
                value={name}
                maxLength={200}
                required
                disabled={save.isPending}
                onChange={(event) => setName(event.target.value)}
              />
            </Field>
            <div className='grid gap-4 sm:grid-cols-2'>
              <Field>
                <FieldLabel htmlFor='upstream-channel-search'>
                  {t('Search channels')}
                </FieldLabel>
                <Input
                  id='upstream-channel-search'
                  value={search}
                  onChange={(event) => setSearch(event.target.value)}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor='upstream-channel-tag'>
                  {t('Tag')}
                </FieldLabel>
                <Select items={tags} value={tag} onValueChange={setTag}>
                  <SelectTrigger id='upstream-channel-tag' className='w-full'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      {tags.map((item) => (
                        <SelectItem
                          key={item.value || 'all-tags'}
                          value={item.value}
                        >
                          {item.label}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>
            </div>
            <FieldDescription>
              {t(
                'Each channel can belong to one upstream accounting group. Channels without a cost discount are excluded from cost totals.'
              )}
            </FieldDescription>
          </FieldGroup>
          <p className='text-sm'>
            {t('Selected channels')}: {selected.length}
          </p>
          {channels.isPending && <p role='status'>{t('Loading...')}</p>}
          {channels.isError && (
            <p role='alert'>
              {t('Failed to load data')}{' '}
              <Button
                variant='link'
                type='button'
                onClick={() => void channels.refetch()}
              >
                {t('Retry')}
              </Button>
            </p>
          )}
          <div
            className='flex max-h-72 flex-col gap-3 overflow-y-auto rounded-lg border p-3'
            aria-label={t('Available channels')}
          >
            {visible.map((channel) => {
              const assigned = Boolean(
                channel.group_id && channel.group_id !== props.group?.id
              )
              let costLabel = t('Excluded')
              if (channel.missing) {
                costLabel = t('Missing channel')
              } else if (channel.eligible) {
                costLabel = `${t('Upstream cost discount')}: ${channel.cost_discount}`
              }
              return (
                <Field
                  key={channel.id}
                  orientation='horizontal'
                  data-disabled={assigned || save.isPending}
                >
                  <Checkbox
                    id={`upstream-channel-${channel.id}`}
                    checked={selected.includes(channel.id)}
                    disabled={
                      assigned ||
                      save.isPending ||
                      (channel.missing && !selected.includes(channel.id))
                    }
                    onCheckedChange={(checked) =>
                      setSelected((values) =>
                        checked === true
                          ? [...values, channel.id]
                          : values.filter((value) => value !== channel.id)
                      )
                    }
                  />
                  <FieldLabel
                    htmlFor={`upstream-channel-${channel.id}`}
                    className='flex flex-wrap items-center gap-2 break-words'
                  >
                    {channel.name} (#{channel.id})
                    {channel.tag && (
                      <Badge variant='outline'>{channel.tag}</Badge>
                    )}
                    <Badge variant='secondary'>
                      {costLabel}
                      {assigned && ` · ${t('Assigned to another upstream')}`}
                    </Badge>
                  </FieldLabel>
                </Field>
              )
            })}
            {!channels.isPending &&
              !channels.isError &&
              visible.length === 0 && (
                <p role='status'>{t('No matching items')}</p>
              )}
          </div>
          <div className='flex justify-end gap-2'>
            <Button
              type='button'
              variant='outline'
              disabled={save.isPending}
              onClick={props.onClose}
            >
              {t('Cancel')}
            </Button>
            <Button
              type='submit'
              disabled={
                !name.trim() ||
                selected.length === 0 ||
                save.isPending ||
                channels.isError ||
                channels.isPending
              }
            >
              {save.isPending ? t('Saving...') : t('Save')}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
