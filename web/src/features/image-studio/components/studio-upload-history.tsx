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
import { Add01Icon, Tick02Icon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import type { UseQueryResult } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from '@/components/ui/empty'

import type { StudioReference, StudioReferenceList } from '../types'
import { StudioReferenceUpload } from './studio-reference-images'

export function StudioUploadHistory(props: {
  authenticated: boolean
  uploads: UseQueryResult<StudioReferenceList>
  selectedId?: string
  disabled: boolean
  onSelect: (reference: StudioReference) => void
  referenceIds: string[]
  canAddReference: boolean
  uploading: boolean
  onUpload: (files: File[]) => void
  onUseReference: (reference: StudioReference) => void
}) {
  const { t } = useTranslation()
  const [dragging, setDragging] = useState(false)
  return (
    <section
      aria-label={t('Upload history')}
      className='flex min-h-0 flex-1 flex-col gap-3 px-4 pb-4'
      onDragOver={(event) => {
        if (!event.dataTransfer.types.includes('Files')) return
        event.preventDefault()
        if (props.canAddReference) setDragging(true)
      }}
      onDragLeave={(event) => {
        if (
          !(event.relatedTarget instanceof Node) ||
          !event.currentTarget.contains(event.relatedTarget)
        ) {
          setDragging(false)
        }
      }}
      onDrop={(event) => {
        event.preventDefault()
        setDragging(false)
        if (props.canAddReference) props.onUpload([...event.dataTransfer.files])
      }}
    >
      <StudioReferenceUpload
        panel
        dragging={dragging}
        disabled={!props.canAddReference}
        uploading={props.uploading}
        onUpload={props.onUpload}
      />
      <p className='text-muted-foreground text-[11px] leading-relaxed'>
        {t('Preview a thumbnail, or use + to add it as a reference.')}
      </p>
      {props.authenticated && props.uploads.isPending && (
        <p role='status'>{t('Loading...')}</p>
      )}
      {props.uploads.isError && <p role='alert'>{t('Failed to load data')}</p>}
      {(!props.authenticated || props.uploads.data?.items.length === 0) && (
        <Empty className='px-2 py-8'>
          <EmptyHeader>
            <EmptyTitle>{t('No uploaded images')}</EmptyTitle>
            <EmptyDescription>
              {t(
                'Uploaded reference images can be reused here until they expire.'
              )}
            </EmptyDescription>
          </EmptyHeader>
          {!props.authenticated && (
            <Link to='/sign-in' search={{ redirect: '/image-studio' }}>
              {t('Sign in')}
            </Link>
          )}
        </Empty>
      )}
      <div className='grid min-h-0 grid-cols-3 content-start gap-2 overflow-y-auto overscroll-contain has-[button]:flex-1 md:grid-cols-5 min-[75rem]:grid-cols-3'>
        {props.uploads.data?.items.map((reference, index) => (
          <div key={reference.id} className='group relative min-w-0'>
            <button
              type='button'
              aria-label={t('Reference image {{index}}', { index: index + 1 })}
              aria-pressed={reference.id === props.selectedId}
              disabled={props.disabled}
              onClick={() => props.onSelect(reference)}
              className='focus-visible:outline-ring relative aspect-square w-full min-w-0 overflow-hidden rounded-lg bg-(--studio-history-item) ring-1 ring-(--studio-subtle-border) transition focus-visible:outline-2 disabled:opacity-50 aria-pressed:ring-2 aria-pressed:ring-(--studio-accent)'
            >
              <img
                src={reference.url}
                alt=''
                loading='lazy'
                className='size-full object-cover transition-transform duration-300 group-hover:scale-105'
                onDragStart={(event) =>
                  event.dataTransfer.setData('text/uri-list', reference.url)
                }
              />
              {reference.created_at && (
                <time
                  dateTime={new Date(reference.created_at * 1000).toISOString()}
                  className='absolute inset-x-0 bottom-0 bg-linear-to-t from-black/75 to-transparent px-1.5 pt-4 pb-1 text-left text-[10px] text-white/85 opacity-0 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100'
                >
                  {new Date(reference.created_at * 1000).toLocaleDateString()}
                </time>
              )}
            </button>
            <Button
              type='button'
              variant='secondary'
              size='icon-xs'
              className='absolute right-1 bottom-1 rounded-full border border-white/20 bg-black/75 text-white hover:bg-black'
              aria-label={t('Add reference image {{index}}', {
                index: index + 1,
              })}
              title={t('Use as reference')}
              disabled={
                !props.canAddReference ||
                props.referenceIds.includes(reference.id) ||
                reference.expires_at <= Date.now() / 1000
              }
              onClick={() => props.onUseReference(reference)}
            >
              <HugeiconsIcon
                icon={
                  props.referenceIds.includes(reference.id)
                    ? Tick02Icon
                    : Add01Icon
                }
                aria-hidden='true'
              />
            </Button>
          </div>
        ))}
      </div>
    </section>
  )
}

export function StudioReferencePreview(props: {
  reference: StudioReference
  disabled: boolean
  onUse: () => void
  onPreview: () => void
}) {
  const { t } = useTranslation()
  return (
    <section
      aria-label={t('Reference images')}
      className='flex min-h-0 min-w-0 flex-1 flex-col gap-3 p-4 md:p-5'
    >
      <div className='flex items-center justify-between gap-3'>
        <h2 className='text-muted-foreground text-xs'>
          {t('Reference images')}
        </h2>
        <Button
          type='button'
          variant='outline'
          size='sm'
          className='rounded-full'
          disabled={props.disabled}
          onClick={props.onUse}
        >
          {t('Use as reference')}
        </Button>
      </div>
      <Dialog>
        <DialogTrigger
          type='button'
          aria-label={t('Preview')}
          onClick={props.onPreview}
          className='focus-visible:outline-ring relative min-h-64 flex-1 cursor-zoom-in rounded-xl focus-visible:outline-2'
        >
          <img
            src={props.reference.url}
            alt={t('Reference images')}
            className='absolute inset-0 m-auto h-auto max-h-full w-auto max-w-full rounded-xl object-contain shadow-[0_30px_80px_-30px_#000]'
          />
        </DialogTrigger>
        <DialogContent className='image-studio-theme max-h-[90svh] overflow-y-auto sm:max-w-4xl'>
          <DialogHeader>
            <DialogTitle>{t('Reference images')}</DialogTitle>
          </DialogHeader>
          <img
            src={props.reference.url}
            alt={t('Reference images')}
            className='mx-auto h-auto max-h-[75svh] w-auto max-w-full object-contain'
          />
        </DialogContent>
      </Dialog>
      <p className='text-muted-foreground shrink-0 text-xs'>
        {t('Expires at')}:{' '}
        {new Date(props.reference.expires_at * 1000).toLocaleString()}
      </p>
    </section>
  )
}
