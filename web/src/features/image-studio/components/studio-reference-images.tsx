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
  Image01Icon,
  Cancel01Icon,
  ImageUploadIcon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useRef } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { cn } from '@/lib/utils'

import type { StudioReference } from '../types'

export function StudioReferenceImages(props: {
  references: StudioReference[]
  uploading: boolean
  disabled: boolean
  maxImages: number
  onUpload: (files: File[]) => void
  onRemove: (id: string) => void
}) {
  const { t } = useTranslation()

  return (
    <section
      aria-label={t('Reference images')}
      className='flex flex-wrap items-center gap-2'
    >
      {props.references.map((reference, index) => (
        <div key={reference.id} className='relative size-11 shrink-0'>
          <img
            src={reference.url}
            alt={t('Reference image {{index}}', { index: index + 1 })}
            className='bg-background size-full rounded-xl border border-(--studio-panel-border) object-cover'
          />
          <Button
            type='button'
            variant='secondary'
            size='icon-xs'
            className='absolute -top-1.5 -right-1.5 size-5 rounded-full border border-(--studio-panel-border)'
            aria-label={t('Remove reference image {{index}}', {
              index: index + 1,
            })}
            disabled={props.disabled}
            onClick={() => props.onRemove(reference.id)}
          >
            <HugeiconsIcon icon={Cancel01Icon} aria-hidden='true' />
          </Button>
        </div>
      ))}
      <StudioReferenceUpload
        disabled={props.disabled || props.references.length >= props.maxImages}
        uploading={props.uploading}
        onUpload={props.onUpload}
      />
    </section>
  )
}

// The composer and history share file selection, formats and repeat selection.
export function StudioReferenceUpload(props: {
  disabled: boolean
  uploading: boolean
  onUpload: (files: File[]) => void
  panel?: boolean
  dragging?: boolean
}) {
  const { t } = useTranslation()
  const input = useRef<HTMLInputElement>(null)
  const icon = props.panel ? ImageUploadIcon : Image01Icon
  const label = props.panel
    ? t('Upload reference images')
    : t('Add reference images')
  return (
    <>
      <input
        ref={input}
        type='file'
        accept='image/png,image/jpeg,image/webp'
        multiple
        className='sr-only'
        tabIndex={-1}
        aria-label={t('Upload reference images')}
        disabled={props.disabled}
        onChange={(event) => {
          props.onUpload([...(event.currentTarget.files || [])])
          event.currentTarget.value = ''
        }}
      />
      <Button
        type='button'
        variant='outline'
        size='sm'
        className={cn(
          'border-dashed',
          props.panel
            ? 'h-auto w-full flex-col gap-1 rounded-xl py-3 text-wrap data-[dragging=true]:border-(--studio-accent) data-[dragging=true]:bg-(--studio-accent-soft)'
            : 'rounded-full'
        )}
        disabled={props.disabled}
        aria-busy={props.uploading}
        data-dragging={props.dragging}
        onClick={() => input.current?.click()}
      >
        <span className='flex items-center gap-2'>
          {props.uploading ? (
            <Spinner />
          ) : (
            <HugeiconsIcon icon={icon} aria-hidden='true' />
          )}
          {props.uploading ? t('Uploading...') : label}
        </span>
        {props.panel && (
          <span className='text-muted-foreground text-[11px] font-normal'>
            {t('Click to upload or drop images here')}
          </span>
        )}
      </Button>
    </>
  )
}
