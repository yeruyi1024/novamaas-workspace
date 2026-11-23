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
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Field, FieldLabel } from '@/components/ui/field'
import type { User } from '@/features/users/types'

import { purgeUserAssets } from '../api'
import { assertAssetSuccess, assetErrorMessage } from '../asset-utils'
import { AssetCopyUserSelect } from './asset-copy-user-select'

export function AssetUserPurgeDialog(props: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onPurged?: () => void
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [target, setTarget] = useState<User>()
  const [confirmOpen, setConfirmOpen] = useState(false)

  const purge = useMutation({
    mutationFn: async (userId: number) =>
      assertAssetSuccess(await purgeUserAssets(userId)),
    onSuccess: async (data) => {
      setConfirmOpen(false)
      setTarget(undefined)
      props.onOpenChange(false)
      props.onPurged?.()
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['asset-library', 'groups'],
        }),
        queryClient.invalidateQueries({
          queryKey: ['asset-library', 'assets'],
        }),
      ])
      toast.success(
        t(
          'Deleted {{assets}} assets and {{groups}} groups for this user',
          {
            assets: data.deleted_asset_count,
            groups: data.deleted_group_count,
          }
        )
      )
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })

  return (
    <>
      <Dialog
        open={props.open}
        onOpenChange={(open) => {
          if (purge.isPending) return
          if (!open) setTarget(undefined)
          props.onOpenChange(open)
        }}
      >
        <DialogContent className='sm:max-w-md'>
          <DialogHeader>
            <DialogTitle>{t('Clear user assets')}</DialogTitle>
            <DialogDescription>
              {t(
                'Delete all asset groups and assets owned by the selected user. Their account, AK/SK keys, webhook settings, and files still referenced by other users are kept.'
              )}
            </DialogDescription>
          </DialogHeader>
          <div className='space-y-4'>
            <Field data-disabled={purge.isPending}>
              <FieldLabel htmlFor='purge-target-user'>
                {t('Target user')}
              </FieldLabel>
              <AssetCopyUserSelect
                id='purge-target-user'
                label={t('Target user')}
                value={target}
                excludeUserId={0}
                disabled={purge.isPending}
                onValueChange={setTarget}
              />
            </Field>
          </div>
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              disabled={purge.isPending}
              onClick={() => props.onOpenChange(false)}
            >
              {t('Cancel')}
            </Button>
            <Button
              type='button'
              variant='destructive'
              disabled={!target || purge.isPending}
              onClick={() => setConfirmOpen(true)}
            >
              {t('Clear user assets')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <AlertDialog
        open={confirmOpen}
        onOpenChange={(open) => !purge.isPending && setConfirmOpen(open)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('Clear all user assets?')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t(
                'Delete all asset groups and assets owned by {{username}} (ID: {{id}}). Managed upstream copies will be removed in the background, while files still referenced by other users are kept. This cannot be undone.',
                {
                  username: target?.username ?? '',
                  id: target?.id ?? 0,
                }
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={purge.isPending}>
              {t('Cancel')}
            </AlertDialogCancel>
            <AlertDialogAction
              variant='destructive'
              disabled={purge.isPending || !target}
              onClick={(event) => {
                event.preventDefault()
                if (target) purge.mutate(target.id)
              }}
            >
              {purge.isPending ? t('Deleting...') : t('Clear user assets')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
