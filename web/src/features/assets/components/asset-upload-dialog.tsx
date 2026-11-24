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
import { zodResolver } from '@hookform/resolvers/zod'
import { Cancel01Icon, FileUploadIcon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { type DragEvent, useEffect, useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import * as z from 'zod'

import { Badge } from '@/components/ui/badge'
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
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import { Progress } from '@/components/ui/progress'
import { Spinner } from '@/components/ui/spinner'
import { cn } from '@/lib/utils'

import { formatAssetBytes, inferAssetType } from '../asset-utils'
import type { AssetGroup, MediaAsset } from '../types'

type UploadValues = {
  group_id: string
  name: string
  type: MediaAsset['type']
  file: File
}

function typeLabelKey(type: MediaAsset['type']) {
  if (type === 'video') return 'Video'
  if (type === 'audio') return 'Audio'
  return 'Image'
}

export function AssetUploadDialog(props: {
  open: boolean
  onOpenChange: (open: boolean) => void
  groups: AssetGroup[]
  selectedGroup: string
  initialName?: string
  initialFiles?: File[]
  replacing?: boolean
  /** One form per file, uploaded in order. */
  onSubmit: (forms: FormData[]) => void
  pending: boolean
  progress: number
  /** Position of the file being uploaded when several are queued. */
  queue?: { index: number; total: number }
}) {
  const { t } = useTranslation()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [dragging, setDragging] = useState(false)
  const [previewUrl, setPreviewUrl] = useState('')
  const [files, setFiles] = useState<File[]>([])
  const multiple = files.length > 1
  const queueState = (index: number) => {
    if (!props.pending || !props.queue) return 'waiting'
    if (index < props.queue.index) return 'done'
    return index === props.queue.index ? 'active' : 'waiting'
  }
  const schema = z.object({
    group_id: z.string().min(1, t('Select an asset group')),
    name: z.string().trim().max(255),
    type: z.enum(['image', 'video', 'audio']),
    file: z
      .instanceof(File, { message: t('Select a file') })
      .refine(
        (file) => inferAssetType(file) !== null,
        t('Select an image, video, or audio file')
      ),
  })
  const form = useForm<UploadValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      group_id: props.selectedGroup,
      name: props.initialName || '',
      type: 'image',
    },
  })
  const selectedFile = form.watch('file')

  const wasOpenRef = useRef(false)
  useEffect(() => {
    if (!props.open) {
      wasOpenRef.current = false
      return
    }
    const justOpened = !wasOpenRef.current
    wasOpenRef.current = true
    if (!justOpened) return

    const initial = props.initialFiles ?? []
    const dropped = initial.filter((file) => inferAssetType(file) !== null)
    const first = dropped[0]
    setFiles(dropped)
    form.reset({
      group_id: props.selectedGroup || props.groups[0]?.id || '',
      name: props.initialName || '',
      type: (first && inferAssetType(first)) || 'image',
      file: first,
    })
    if (initial.length > 0 && dropped.length === 0) {
      form.setError('file', {
        type: 'validate',
        message: t('Select an image, video, or audio file'),
      })
    }
  }, [
    form,
    props.groups,
    props.initialFiles,
    props.initialName,
    props.open,
    props.selectedGroup,
    t,
  ])

  // Populate the default group once async group options arrive, without
  // resetting files or inputs the user has already entered.
  useEffect(() => {
    if (!props.open || form.getValues('group_id')) return
    const fallbackGroup = props.selectedGroup || props.groups[0]?.id || ''
    if (fallbackGroup) {
      form.setValue('group_id', fallbackGroup)
    }
  }, [form, props.groups, props.open, props.selectedGroup])

  useEffect(() => {
    if (!selectedFile || !selectedFile.type.startsWith('image/')) {
      setPreviewUrl('')
      return
    }
    const url = URL.createObjectURL(selectedFile)
    setPreviewUrl(url)
    return () => URL.revokeObjectURL(url)
  }, [selectedFile])

  const selectFiles = (picked: Iterable<File>) => {
    const candidates = props.replacing ? [...picked].slice(0, 1) : [...picked]
    if (candidates.length === 0) return
    const accepted = candidates.filter((file) => inferAssetType(file) !== null)
    const first = accepted[0]
    if (!first) {
      form.setError('file', {
        type: 'validate',
        message: t('Select an image, video, or audio file'),
      })
      return
    }
    form.clearErrors('file')
    setFiles(accepted)
    form.setValue('file', first, { shouldDirty: true, shouldValidate: true })
    form.setValue('type', inferAssetType(first) ?? 'image', {
      shouldDirty: true,
    })
  }

  const removeFile = (target: File) => {
    const remaining = files.filter((file) => file !== target)
    setFiles(remaining)
    if (fileInputRef.current) fileInputRef.current.value = ''
    const first = remaining[0]
    if (!first) {
      form.resetField('file')
      return
    }
    form.setValue('file', first, { shouldValidate: true })
    form.setValue('type', inferAssetType(first) ?? 'image')
  }

  const handleDrop = (event: DragEvent<HTMLButtonElement>) => {
    event.preventDefault()
    setDragging(false)
    selectFiles(event.dataTransfer.files)
  }

  const submit = (values: UploadValues) => {
    const queued = files.length > 0 ? files : [values.file]
    const forms = queued.map((file) => {
      const formData = new FormData()
      formData.append('group_id', values.group_id)
      formData.append('name', queued.length > 1 ? '' : values.name)
      formData.append('type', inferAssetType(file) ?? values.type)
      formData.append('file', file)
      return formData
    })
    props.onSubmit(forms)
  }

  return (
    <Dialog
      open={props.open}
      onOpenChange={(open) => !props.pending && props.onOpenChange(open)}
    >
      <DialogContent className='sm:max-w-xl'>
        <DialogHeader>
          <DialogTitle>
            {props.replacing ? t('Re-upload revised file') : t('Upload asset')}
          </DialogTitle>
          <DialogDescription>
            {props.replacing
              ? t(
                  'A new asset ID will be created. The rejected asset remains unavailable; update your video requests to use the new ID.'
                )
              : t(
                  'Your file is saved to permanent object storage. Channel delivery runs automatically in the background.'
                )}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(submit)}
            className='flex flex-col gap-4'
          >
            <FormField
              control={form.control}
              name='group_id'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Asset group')}</FormLabel>
                  <FormControl>
                    <NativeSelect className='w-full' {...field}>
                      {props.groups.map((group) => (
                        <NativeSelectOption key={group.id} value={group.id}>
                          {group.name} · {t('ID')}: {group.id}
                        </NativeSelectOption>
                      ))}
                    </NativeSelect>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            {!multiple && (
              <FormField
                control={form.control}
                name='name'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Name')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t('Uses the file name when empty')}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            <FormField
              control={form.control}
              name='file'
              render={() => (
                <FormItem>
                  <FormLabel>{t('File')}</FormLabel>
                  <FormControl>
                    <div>
                      <input
                        ref={fileInputRef}
                        type='file'
                        aria-label={t('File')}
                        className='sr-only'
                        accept='image/*,video/*,audio/*'
                        multiple={!props.replacing}
                        onChange={(event) =>
                          selectFiles(event.target.files ?? [])
                        }
                      />
                      <button
                        type='button'
                        disabled={props.pending}
                        onClick={() => fileInputRef.current?.click()}
                        onDragEnter={(event) => {
                          event.preventDefault()
                          setDragging(true)
                        }}
                        onDragOver={(event) => event.preventDefault()}
                        onDragLeave={() => setDragging(false)}
                        onDrop={handleDrop}
                        className={cn(
                          'border-input bg-muted/20 hover:bg-muted/40 focus-visible:border-ring focus-visible:ring-ring/50 flex min-h-36 w-full flex-col items-center justify-center gap-2 rounded-lg border border-dashed px-4 py-6 text-center transition-colors outline-none focus-visible:ring-3 disabled:pointer-events-none disabled:opacity-50',
                          dragging && 'border-primary bg-muted/50'
                        )}
                      >
                        {previewUrl ? (
                          <img
                            src={previewUrl}
                            alt={selectedFile?.name || ''}
                            className='max-h-24 max-w-full object-contain'
                          />
                        ) : (
                          <HugeiconsIcon
                            icon={FileUploadIcon}
                            className='text-muted-foreground size-7'
                          />
                        )}
                        <span className='font-medium'>
                          {props.replacing
                            ? t('Drop a file here, or click to browse')
                            : t('Drop files here, or click to browse')}
                        </span>
                        <span className='text-muted-foreground text-xs'>
                          {t('Images, videos, and audio files are supported')}
                        </span>
                      </button>
                    </div>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            {selectedFile && !multiple && (
              <div className='bg-muted/40 flex items-center justify-between gap-3 rounded-lg border px-3 py-2'>
                <div className='min-w-0'>
                  <p className='truncate text-sm font-medium'>
                    {selectedFile.name}
                  </p>
                  <p className='text-muted-foreground text-xs'>
                    {formatAssetBytes(selectedFile.size)}
                  </p>
                </div>
                <Badge variant='outline'>
                  {t(typeLabelKey(form.getValues('type')))}
                </Badge>
              </div>
            )}
            {multiple && (
              <div className='flex flex-col gap-1.5'>
                <p className='text-muted-foreground text-xs'>
                  {t('{{count}} files selected · each uses its file name', {
                    count: files.length,
                  })}
                </p>
                <ul className='bg-muted/30 max-h-48 divide-y overflow-y-auto rounded-lg border'>
                  {files.map((file, index) => {
                    const type = inferAssetType(file) ?? 'image'
                    const state = queueState(index)
                    return (
                      <li
                        key={`${file.name}-${file.size}-${file.lastModified}`}
                        className='flex items-center gap-3 px-3 py-2'
                      >
                        <div className='min-w-0 flex-1'>
                          <p className='truncate text-sm font-medium'>
                            {file.name}
                          </p>
                          <p className='text-muted-foreground text-xs'>
                            {formatAssetBytes(file.size)} ·{' '}
                            {t(typeLabelKey(type))}
                          </p>
                        </div>
                        {state === 'done' && (
                          <Badge variant='secondary'>{t('Uploaded')}</Badge>
                        )}
                        {state === 'active' && <Spinner />}
                        {!props.pending && (
                          <Button
                            type='button'
                            size='icon-xs'
                            variant='ghost'
                            aria-label={`${t('Remove')} ${file.name}`}
                            onClick={() => removeFile(file)}
                          >
                            <HugeiconsIcon icon={Cancel01Icon} />
                          </Button>
                        )}
                      </li>
                    )
                  })}
                </ul>
              </div>
            )}
            {props.pending && (
              <div className='flex flex-col gap-2' aria-live='polite'>
                <div className='flex items-center justify-between text-xs'>
                  <span>
                    {props.queue && props.queue.total > 1
                      ? t('Uploading {{current}} of {{total}}', {
                          current: props.queue.index + 1,
                          total: props.queue.total,
                        })
                      : t('Uploading')}
                  </span>
                  <span>{props.progress}%</span>
                </div>
                <Progress value={props.progress} />
              </div>
            )}
            <DialogFooter>
              <Button
                type='button'
                variant='outline'
                disabled={props.pending}
                onClick={() => props.onOpenChange(false)}
              >
                {t('Cancel')}
              </Button>
              <Button type='submit' disabled={props.pending || !selectedFile}>
                {props.pending && <Spinner data-icon='inline-start' />}
                {multiple
                  ? t('Upload {{count}} files', { count: files.length })
                  : t('Upload')}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
