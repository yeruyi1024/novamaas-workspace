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
  AiChat02Icon,
  Image01Icon,
  Settings02Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import type { UseQueryResult } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Field,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from '@/components/ui/field'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import type { ApiKey } from '@/features/keys/types'

export type StudioModelSelection = { tokenId: number; model: string }

export function StudioModelSettings(props: {
  keys: ApiKey[]
  userGroup: string
  image: StudioModelSelection
  text: StudioModelSelection
  imageModels: UseQueryResult<string[], Error>
  textModels: UseQueryResult<string[], Error>
  disabled: boolean
  onImageChange: (selection: StudioModelSelection) => void
  onTextChange: (selection: StudioModelSelection) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <Button
        type='button'
        variant='outline'
        size='sm'
        disabled={props.disabled}
        onClick={() => setOpen(true)}
      >
        <HugeiconsIcon icon={Settings02Icon} aria-hidden='true' />
        {t('Model settings')}
      </Button>
      <DialogContent className='image-studio-theme max-h-[90vh] overflow-y-auto sm:max-w-2xl'>
        <DialogHeader>
          <DialogTitle>{t('Model settings')}</DialogTitle>
          <DialogDescription>
            {t(
              'Choose an API key for each model. Its group and permissions are applied automatically.'
            )}
          </DialogDescription>
        </DialogHeader>
        <div className='grid gap-4 sm:grid-cols-2'>
          <ModelField
            kind='image'
            keys={props.keys}
            userGroup={props.userGroup}
            selection={props.image}
            models={props.imageModels}
            disabled={props.disabled}
            onChange={props.onImageChange}
          />
          <ModelField
            kind='text'
            keys={props.keys}
            userGroup={props.userGroup}
            selection={props.text}
            models={props.textModels}
            disabled={props.disabled}
            onChange={props.onTextChange}
          />
        </div>
        <DialogFooter>
          <Button type='button' onClick={() => setOpen(false)}>
            {t('Done')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function ModelField(props: {
  kind: 'image' | 'text'
  keys: ApiKey[]
  userGroup: string
  selection: StudioModelSelection
  models: UseQueryResult<string[], Error>
  disabled: boolean
  onChange: (selection: StudioModelSelection) => void
}) {
  const { t } = useTranslation()
  const prefix = `studio-${props.kind}`
  const label = props.kind === 'image' ? t('Image model') : t('Text model')
  const selectedKey = props.keys.find(
    (key) => key.id === props.selection.tokenId
  )
  const group = selectedKey?.group || props.userGroup
  return (
    <FieldSet aria-label={label} className='bg-muted/20 rounded-xl border p-4'>
      <FieldLegend className='flex items-center gap-2'>
        <HugeiconsIcon
          icon={props.kind === 'image' ? Image01Icon : AiChat02Icon}
          className='text-primary size-4'
          aria-hidden='true'
        />
        {label}
      </FieldLegend>
      <FieldGroup className='gap-4'>
        <Field data-disabled={props.disabled || !props.keys.length}>
          <FieldLabel htmlFor={`${prefix}-key`}>{t('API key')}</FieldLabel>
          <NativeSelect
            id={`${prefix}-key`}
            className='w-full'
            value={props.selection.tokenId}
            disabled={props.disabled || !props.keys.length}
            onChange={(event) =>
              props.onChange({ tokenId: Number(event.target.value), model: '' })
            }
          >
            {props.keys.map((key) => {
              let keyGroup = key.group || props.userGroup
              if (keyGroup === 'auto') {
                keyGroup =
                  `${t('Automatic')} ${key.auto_groups?.join(' / ') || ''}`.trim()
              }
              return (
                <NativeSelectOption key={key.id} value={key.id}>
                  {key.name} · {keyGroup}
                </NativeSelectOption>
              )
            })}
          </NativeSelect>
          <p className='text-muted-foreground text-xs'>
            {t('Group')}: {group === 'auto' ? t('Automatic') : group}
          </p>
        </Field>
        <Field
          data-disabled={
            props.disabled || props.models.isFetching || !props.selection.model
          }
        >
          <FieldLabel htmlFor={`${prefix}-model`}>{t('Model')}</FieldLabel>
          <NativeSelect
            id={`${prefix}-model`}
            className='w-full'
            value={props.selection.model}
            disabled={
              props.disabled ||
              props.models.isFetching ||
              !props.selection.model
            }
            onChange={(event) =>
              props.onChange({ ...props.selection, model: event.target.value })
            }
          >
            {props.models.data?.map((name) => (
              <NativeSelectOption key={name} value={name}>
                {name}
              </NativeSelectOption>
            ))}
          </NativeSelect>
          {props.models.isPending && props.selection.tokenId > 0 && (
            <p role='status' className='text-muted-foreground text-xs'>
              {t('Loading...')}
            </p>
          )}
          {props.models.isSuccess && !props.selection.model && (
            <p role='status' className='text-muted-foreground text-xs'>
              {t('No models available')}
            </p>
          )}
          {props.models.isError && (
            <div role='alert' className='text-destructive text-xs'>
              {t('Failed to load enabled models')}
              <Button
                type='button'
                variant='ghost'
                size='sm'
                onClick={() => void props.models.refetch()}
              >
                {t('Retry')}
              </Button>
            </div>
          )}
        </Field>
      </FieldGroup>
      {props.kind === 'text' && (
        <p className='text-muted-foreground text-xs'>
          {t('Text model calls are billed normally.')}
        </p>
      )}
    </FieldSet>
  )
}
