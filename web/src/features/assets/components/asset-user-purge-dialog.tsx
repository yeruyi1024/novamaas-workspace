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
import { useEffect, useState } from 'react'
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
import { cn } from '@/lib/utils'

import {
  deleteMediaAssetsByFilter,
  listAssetGroupsPage,
  purgeUserAssets,
  summarizeMediaAssets,
} from '../api'
import { assertAssetSuccess, assetErrorMessage } from '../asset-utils'
import { AssetCopyUserSelect } from './asset-copy-user-select'

type CleanupMode = 'all' | 'unavailable'

export type InitialPurgeOwner = {
  id: number
  username: string
}

function asStubUser(owner: InitialPurgeOwner): User {
  return {
    id: owner.id,
    username: owner.username,
    display_name: owner.username,
  } as User
}

export function AssetUserPurgeDialog(props: {
  open: boolean
  initialOwner?: InitialPurgeOwner
  onOpenChange: (open: boolean) => void
  onPurged?: (mode: CleanupMode) => void
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [target, setTarget] = useState<User | undefined>(
    props.initialOwner ? asStubUser(props.initialOwner) : undefined
  )
  const [mode, setMode] = useState<CleanupMode>('all')
  const [confirmOpen, setConfirmOpen] = useState(false)

  useEffect(() => {
    if (!props.open) return
    setMode('all')
    setConfirmOpen(false)
    setTarget(props.initialOwner ? asStubUser(props.initialOwner) : undefined)
  }, [props.initialOwner, props.open])

  const summaryQuery = useQuery({
    queryKey: ['asset-library', 'purge-preview', 'summary', target?.id],
    queryFn: async () =>
      assertAssetSuccess(
        await summarizeMediaAssets({
          ownerUserId: target?.id,
          includeAllOwners: true,
        })
      ),
    enabled: props.open && Boolean(target?.id),
    retry: false,
  })
  const groupsCountQuery = useQuery({
    queryKey: ['asset-library', 'purge-preview', 'groups', target?.id],
    queryFn: async () =>
      assertAssetSuccess(
        await listAssetGroupsPage({
          includeAllOwners: true,
          ownerUserId: target?.id,
          page: 1,
          pageSize: 1,
          search: '',
        })
      ),
    enabled: props.open && Boolean(target?.id),
    retry: false,
  })

  const summary =
    summaryQuery.data &&
    summaryQuery.data.by_status &&
    summaryQuery.data.by_type
      ? summaryQuery.data
      : undefined
  const groupCount = groupsCountQuery.data?.total
  const unavailableCount = summary?.by_status?.unavailable ?? 0
  const hasPreview = summary !== undefined || groupCount !== undefined

  const cleanup = useMutation({
    mutationFn: async (input: { userId: number; mode: CleanupMode }) => {
      if (input.mode === 'unavailable') {
        const res = assertAssetSuccess(
          await deleteMediaAssetsByFilter({
            scope: 'all',
            owner_user_id: input.userId,
            status: 'unavailable',
          })
        )
        return {
          mode: 'unavailable' as const,
          assets: res.deleted_count,
          groups: 0,
        }
      }
      const res = assertAssetSuccess(await purgeUserAssets(input.userId))
      return {
        mode: 'all' as const,
        assets: res.deleted_asset_count,
        groups: res.deleted_group_count,
      }
    },
    onSuccess: async (data) => {
      setConfirmOpen(false)
      setTarget(undefined)
      props.onOpenChange(false)
      props.onPurged?.(data.mode)
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['asset-library', 'groups'],
        }),
        queryClient.invalidateQueries({
          queryKey: ['asset-library', 'assets'],
        }),
        queryClient.invalidateQueries({
          queryKey: ['asset-library', 'owners'],
        }),
      ])
      if (data.mode === 'unavailable') {
        toast.success(
          t('{{count}} unavailable assets queued for deletion', {
            count: data.assets,
          })
        )
      } else {
        toast.success(
          t('Deleted {{assets}} assets and {{groups}} groups for this user', {
            assets: data.assets,
            groups: data.groups,
          })
        )
      }
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })

  const unavailableDisabled =
    summary !== undefined && summary.by_status.unavailable === 0
  const actionLabel =
    mode === 'unavailable'
      ? t('Clear unavailable assets')
      : t('Clear user assets')
  const confirmButtonLabel = cleanup.isPending ? t('Deleting...') : actionLabel

  return (
    <>
      <Dialog
        open={props.open}
        onOpenChange={(open) => {
          if (cleanup.isPending) return
          if (!open) setTarget(undefined)
          props.onOpenChange(open)
        }}
      >
        <DialogContent className='sm:max-w-lg'>
          <DialogHeader>
            <DialogTitle>{t('Clear user assets')}</DialogTitle>
            <DialogDescription>
              {t(
                'Delete all asset groups and assets owned by the selected user. Their account, AK/SK keys, webhook settings, and files still referenced by other users are kept.'
              )}
            </DialogDescription>
          </DialogHeader>
          <div className='space-y-4'>
            <Field data-disabled={cleanup.isPending}>
              <FieldLabel htmlFor='purge-target-user'>
                {t('Target user')}
              </FieldLabel>
              <AssetCopyUserSelect
                id='purge-target-user'
                label={t('Target user')}
                value={target}
                excludeUserId={0}
                disabled={cleanup.isPending}
                onValueChange={(next) => {
                  setTarget(next)
                  setMode('all')
                }}
              />
            </Field>

            {target && hasPreview && (
              <div
                data-testid='purge-impact-summary'
                className='bg-muted/50 flex flex-wrap items-center gap-x-4 gap-y-1.5 rounded-lg border px-3 py-2.5 text-xs'
              >
                {groupCount !== undefined && (
                  <span>
                    <strong className='tabular-nums'>{groupCount}</strong>{' '}
                    {t('groups')}
                  </span>
                )}
                {summary && (
                  <>
                    <span>
                      <strong className='tabular-nums'>{summary.total}</strong>{' '}
                      {t('assets')}
                    </span>
                    <span className='text-muted-foreground tabular-nums'>
                      {t('Images')} {summary.by_type.image} · {t('Videos')}{' '}
                      {summary.by_type.video} · {t('Audio')}{' '}
                      {summary.by_type.audio}
                    </span>
                    {unavailableCount > 0 && (
                      <span className='text-destructive font-medium tabular-nums'>
                        {t('{{count}} unavailable', {
                          count: unavailableCount,
                        })}
                      </span>
                    )}
                  </>
                )}
              </div>
            )}

            {target && (
              <div
                role='radiogroup'
                aria-label={t('Cleanup scope')}
                className='grid gap-2'
              >
                <button
                  type='button'
                  role='radio'
                  aria-checked={mode === 'all'}
                  disabled={cleanup.isPending}
                  onClick={() => setMode('all')}
                  className={cn(
                    'flex flex-col gap-0.5 rounded-lg border p-3 text-left transition-colors',
                    mode === 'all'
                      ? 'border-destructive/60 bg-destructive/5'
                      : 'hover:bg-muted/40'
                  )}
                >
                  <span className='text-sm font-medium'>
                    {t('Clear all groups and assets')}
                  </span>
                  <span className='text-muted-foreground text-xs'>
                    {t(
                      'Removes every active asset and asset group for this user.'
                    )}
                  </span>
                </button>
                <button
                  type='button'
                  role='radio'
                  aria-checked={mode === 'unavailable'}
                  disabled={cleanup.isPending || unavailableDisabled}
                  onClick={() => setMode('unavailable')}
                  className={cn(
                    'flex flex-col gap-0.5 rounded-lg border p-3 text-left transition-colors disabled:cursor-not-allowed disabled:opacity-50',
                    mode === 'unavailable'
                      ? 'border-primary bg-primary/5'
                      : 'hover:bg-muted/40'
                  )}
                >
                  <span className='text-sm font-medium'>
                    {t('Clear unavailable assets only')}
                    {summary !== undefined && ` (${unavailableCount})`}
                  </span>
                  <span className='text-muted-foreground text-xs'>
                    {t(
                      'Keeps all groups and ready assets; deletes only rejected or failed items.'
                    )}
                  </span>
                </button>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              disabled={cleanup.isPending}
              onClick={() => props.onOpenChange(false)}
            >
              {t('Cancel')}
            </Button>
            <Button
              type='button'
              variant='destructive'
              disabled={
                !target ||
                cleanup.isPending ||
                (mode === 'unavailable' && unavailableDisabled)
              }
              onClick={() => setConfirmOpen(true)}
            >
              {mode === 'unavailable'
                ? t('Clear unavailable assets')
                : t('Clear user assets')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <AlertDialog
        open={confirmOpen}
        onOpenChange={(open) => !cleanup.isPending && setConfirmOpen(open)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {mode === 'unavailable'
                ? t('Clear unavailable user assets?')
                : t('Clear all user assets?')}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {mode === 'unavailable'
                ? t(
                    'Delete {{count}} unavailable or failed assets owned by {{username}} (ID: {{id}}). Groups and ready assets will be kept. This cannot be undone.',
                    {
                      count: unavailableCount,
                      username: target?.username ?? '',
                      id: target?.id ?? 0,
                    }
                  )
                : t(
                    'Delete all asset groups and assets owned by {{username}} (ID: {{id}}). Managed upstream copies will be removed in the background, while files still referenced by other users are kept. This cannot be undone.',
                    {
                      username: target?.username ?? '',
                      id: target?.id ?? 0,
                    }
                  )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={cleanup.isPending}>
              {t('Cancel')}
            </AlertDialogCancel>
            <AlertDialogAction
              variant='destructive'
              disabled={cleanup.isPending || !target}
              onClick={(event) => {
                event.preventDefault()
                if (target) cleanup.mutate({ userId: target.id, mode })
              }}
            >
              {confirmButtonLabel}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
