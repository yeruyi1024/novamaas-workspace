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
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import { listAssetGroups } from '@/features/assets/api'

import { saveStudioResult } from '../api'
import { assertStudioSuccess, studioErrorMessage } from '../studio-utils'
import type { StudioResult } from '../types'

export function StudioSaveDialog(props: {
  result: StudioResult
  onClose: () => void
}) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const [saveGroup, setSaveGroup] = useState('')
  const groups = useQuery({
    queryKey: ['image-studio', 'asset-groups'],
    queryFn: async () => assertStudioSuccess(await listAssetGroups(false)),
  })
  const save = useMutation({
    mutationFn: () => saveStudioResult(props.result.id, saveGroup),
    onSuccess: () => {
      toast.success(t('Saved to asset library'))
      props.onClose()
    },
    onError: (error) => toast.error(studioErrorMessage(error)),
    onSettled: () => {
      void client.invalidateQueries({ queryKey: ['image-studio'] })
      void client.invalidateQueries({ queryKey: ['asset-library'] })
    },
  })
  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !save.isPending) props.onClose()
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('Save to asset library')}</DialogTitle>
          <DialogDescription>
            {t(
              'Only add images to an asset group when you need them in the asset library for video generation or other workflows.'
            )}
          </DialogDescription>
        </DialogHeader>
        <div className='grid gap-2'>
          <Label htmlFor='studio-save-group'>{t('Asset group')}</Label>
          <NativeSelect
            id='studio-save-group'
            className='w-full'
            value={saveGroup}
            onChange={(event) => setSaveGroup(event.target.value)}
            disabled={groups.isFetching || save.isPending}
          >
            <NativeSelectOption value=''>
              {t('Select an asset group')}
            </NativeSelectOption>
            {groups.data?.map((group) => (
              <NativeSelectOption key={group.id} value={group.id}>
                {group.name}
              </NativeSelectOption>
            ))}
          </NativeSelect>
          {groups.isError && (
            <p role='alert' className='text-destructive text-sm'>
              {t('Failed to load data')}
            </p>
          )}
          {save.isError && (
            <p role='alert' className='text-destructive text-sm'>
              {studioErrorMessage(save.error)}
            </p>
          )}
        </div>
        <DialogFooter>
          <Button
            variant='outline'
            onClick={() => props.onClose()}
            disabled={save.isPending}
          >
            {t('Cancel')}
          </Button>
          <Button
            disabled={!saveGroup || groups.isFetching || save.isPending}
            onClick={() => save.mutate()}
          >
            {t('Save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
