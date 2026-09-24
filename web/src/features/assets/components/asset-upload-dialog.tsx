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
import { FileUploadIcon } from '@hugeicons/core-free-icons'
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
  replacing?: boolean
  onSubmit: (formData: FormData) => void
  pending: boolean
  progress: number
}) {
  const { t } = useTranslation()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [dragging, setDragging] = useState(false)
  const [previewUrl, setPreviewUrl] = useState('')
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

  useEffect(() => {
    if (!props.open) return
    form.reset({
      group_id: props.selectedGroup || props.groups[0]?.id || '',
      name: props.initialName || '',
      type: 'image',
      file: undefined,
    })
  }, [form, props.groups, props.initialName, props.open, props.selectedGroup])

  useEffect(() => {
    if (!selectedFile || !selectedFile.type.startsWith('image/')) {
      setPreviewUrl('')
      return
    }
    const url = URL.createObjectURL(selectedFile)
    setPreviewUrl(url)
    return () => URL.revokeObjectURL(url)
  }, [selectedFile])

  const selectFile = (file?: File) => {
    if (!file) return
    const type = inferAssetType(file)
    if (!type) {
      form.setError('file', {
        type: 'validate',
        message: t('Select an image, video, or audio file'),
      })
      return
    }
    form.clearErrors('file')
    form.setValue('file', file, { shouldDirty: true, shouldValidate: true })
    form.setValue('type', type, { shouldDirty: true })
  }

  const handleDrop = (event: DragEvent<HTMLButtonElement>) => {
    event.preventDefault()
    setDragging(false)
    selectFile(event.dataTransfer.files[0])
  }

  const submit = (values: UploadValues) => {
    const formData = new FormData()
    formData.append('group_id', values.group_id)
    formData.append('name', values.name)
    formData.append('type', values.type)
    formData.append('file', values.file)
    props.onSubmit(formData)
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
                        onChange={(event) =>
                          selectFile(event.target.files?.[0])
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
                          {t('Drop a file here, or click to browse')}
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
            {selectedFile && (
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
            {props.pending && (
              <div className='flex flex-col gap-2' aria-live='polite'>
                <div className='flex items-center justify-between text-xs'>
                  <span>{t('Uploading')}</span>
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
                {t('Upload')}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
