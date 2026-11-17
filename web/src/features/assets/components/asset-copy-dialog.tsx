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
import { useRef, useState } from 'react'
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
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import {
  Progress,
  ProgressLabel,
  ProgressValue,
} from '@/components/ui/progress'
import type { User } from '@/features/users/types'

import {
  copyAssetGroups,
  deleteAssetCopySources,
  getAssetCopy,
  listAssetGroupsPage,
} from '../api'
import { assertAssetSuccess, assetErrorMessage } from '../asset-utils'
import type { AssetCopyInput, AssetCopyResult, AssetGroup } from '../types'
import { AssetCopyUserSelect } from './asset-copy-user-select'
import { AssetGroupSelect } from './asset-group-select'

export function AssetCopyDialog(props: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [source, setSource] = useState<User>()
  const [target, setTarget] = useState<User>()
  const [groups, setGroups] = useState<AssetGroup[]>([])
  const [submitted, setSubmitted] = useState<AssetCopyResult | null>(null)
  const [confirmDelete, setConfirmDelete] = useState(false)
  // A retry of the same selection reuses its request ID, so the server
  // returns the committed batch instead of copying twice.
  const request = useRef({ signature: '', id: '' })
  const sourceId = useRef(0)

  const receipt = useQuery({
    queryKey: ['asset-copy-receipt', submitted?.id],
    queryFn: async () =>
      assertAssetSuccess(await getAssetCopy(submitted?.id ?? '')),
    enabled: props.open && submitted !== null,
    initialData: submitted ?? undefined,
    refetchInterval: (query) => {
      const data = query.state.data
      return data && !data.sources_deleted_at && data.processing_count > 0
        ? 3000
        : false
    },
  })
  const result = receipt.data ?? submitted
  const refreshLibrary = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: ['asset-library', 'groups'] }),
      queryClient.invalidateQueries({ queryKey: ['asset-library', 'assets'] }),
    ])

  const selectAll = useMutation({
    mutationFn: async (ownerUserId: number) => {
      const data = assertAssetSuccess(
        await listAssetGroupsPage({
          includeAllOwners: true,
          ownerUserId,
          page: 1,
          pageSize: 100,
          search: '',
        })
      )
      if (data.total > 100) {
        throw new Error(
          t(
            'This user has more than 100 groups. Select up to 100 groups per copy.'
          )
        )
      }
      return data.items
    },
    onSuccess: (items, ownerUserId) => {
      if (ownerUserId === sourceId.current) setGroups(items)
    },
  })
  const copy = useMutation({
    mutationFn: async (input: Omit<AssetCopyInput, 'request_id'>) => {
      const signature = JSON.stringify(input)
      if (request.current.signature !== signature) {
        request.current = { signature, id: crypto.randomUUID() }
      }
      return assertAssetSuccess(
        await copyAssetGroups({ ...input, request_id: request.current.id })
      )
    },
    onSuccess: async (data) => {
      setSubmitted(data)
      await refreshLibrary()
      toast.success(t('Asset groups copied'))
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })
  const remove = useMutation({
    mutationFn: async (id: string) =>
      assertAssetSuccess(await deleteAssetCopySources(id)),
    onSuccess: async (data) => {
      queryClient.setQueryData(['asset-copy-receipt', data.id], data)
      setSubmitted(data)
      setConfirmDelete(false)
      await refreshLibrary()
      toast.success(t('Copied source assets deleted'))
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })

  const pending = copy.isPending || remove.isPending
  const canDelete = result?.can_delete_sources && !receipt.isError
  const canCopy =
    source &&
    target &&
    groups.length > 0 &&
    groups.length <= 100 &&
    !selectAll.isPending &&
    !pending
  const progress =
    result && result.asset_count > 0
      ? Math.floor(
          ((result.asset_count - result.processing_count) /
            result.asset_count) *
            100
        )
      : 100

  return (
    <>
      <Dialog
        open={props.open}
        onOpenChange={(open) => !pending && props.onOpenChange(open)}
      >
        <DialogContent className='max-h-[85dvh] overflow-y-auto sm:max-w-xl'>
          <DialogHeader>
            <DialogTitle>{t('Copy asset groups')}</DialogTitle>
            <DialogDescription>
              {t(
                'Copy the selected asset groups to another user. Synchronization is handled automatically, and the source assets are kept.'
              )}
            </DialogDescription>
          </DialogHeader>
          {result ? (
            <div className='space-y-4'>
              <p className='text-sm'>
                {t('Created {{count}} assets in {{groups}} groups', {
                  count: result.asset_count,
                  groups: result.groups.length,
                })}
              </p>
              <Progress value={progress}>
                <ProgressLabel>{t('Copy progress')}</ProgressLabel>
                <ProgressValue />
              </Progress>
              <p role='status' className='text-sm'>
                {result.processing_count === 0
                  ? t(
                      'Copy complete: {{ready}} succeeded, {{failed}} failed.',
                      {
                        ready: result.ready_count,
                        failed: result.failed_count + result.unavailable_count,
                      }
                    )
                  : t(
                      'Synchronization: {{ready}} ready, {{processing}} processing, {{failed}} failed, {{unavailable}} unavailable',
                      {
                        ready: result.ready_count,
                        processing: result.processing_count,
                        failed: result.failed_count,
                        unavailable: result.unavailable_count,
                      }
                    )}
              </p>
              {receipt.isError && (
                <p role='alert' className='text-destructive text-sm'>
                  {assetErrorMessage(receipt.error)}
                </p>
              )}
              <ul className='max-h-48 space-y-1 overflow-y-auto text-sm'>
                {result.groups.map((group) => (
                  <li className='break-words' key={group.id}>
                    {group.name} · {group.asset_count}
                  </li>
                ))}
              </ul>
              <p className='text-muted-foreground text-sm'>
                {t(
                  'Once processing finishes, all sources in this batch can be deleted, even if copies failed or were deleted. Files still referenced by other assets are kept.'
                )}
              </p>
              <DialogFooter className='flex-wrap'>
                <Button
                  variant='outline'
                  disabled={receipt.isFetching || pending}
                  onClick={() => void receipt.refetch()}
                >
                  {t('Refresh')}
                </Button>
                <Button
                  variant='outline'
                  disabled={pending}
                  onClick={() => {
                    setSubmitted(null)
                    request.current = { signature: '', id: '' }
                  }}
                >
                  {t('New copy')}
                </Button>
                <Button
                  variant='destructive'
                  disabled={pending || !canDelete}
                  onClick={() => setConfirmDelete(true)}
                >
                  {result.sources_deleted_at > 0
                    ? t('Copied source assets deleted')
                    : t('Delete copied sources')}
                </Button>
              </DialogFooter>
            </div>
          ) : (
            <form
              onSubmit={(event) => {
                event.preventDefault()
                if (!canCopy) return
                copy.mutate({
                  source_user_id: source.id,
                  target_user_id: target.id,
                  group_ids: groups.map((group) => group.id),
                })
              }}
              className='space-y-4'
            >
              <fieldset disabled={pending} className='space-y-4'>
                <FieldGroup className='[container-type:normal] grid gap-4 sm:grid-cols-2'>
                  <Field data-disabled={pending}>
                    <FieldLabel htmlFor='copy-source'>
                      {t('Source user')}
                    </FieldLabel>
                    <AssetCopyUserSelect
                      id='copy-source'
                      label={t('Source user')}
                      value={source}
                      excludeUserId={target?.id ?? 0}
                      disabled={pending}
                      onValueChange={(user) => {
                        sourceId.current = user.id
                        selectAll.reset()
                        setSource(user)
                        setGroups([])
                      }}
                    />
                  </Field>
                  <Field data-disabled={pending}>
                    <FieldLabel htmlFor='copy-target'>
                      {t('Target user')}
                    </FieldLabel>
                    <AssetCopyUserSelect
                      id='copy-target'
                      label={t('Target user')}
                      value={target}
                      excludeUserId={source?.id ?? 0}
                      enabledOnly
                      disabled={pending}
                      onValueChange={setTarget}
                    />
                  </Field>
                </FieldGroup>
                {source && (
                  <div className='space-y-2'>
                    <div className='flex flex-col gap-2 sm:flex-row'>
                      <Button
                        type='button'
                        variant='outline'
                        className='min-h-12 shrink-0'
                        disabled={selectAll.isPending}
                        onClick={() => selectAll.mutate(source.id)}
                      >
                        {selectAll.isPending
                          ? t('Loading...')
                          : t('Select all asset groups')}
                      </Button>
                      <div className='min-w-0 flex-1'>
                        <AssetGroupSelect
                          key={source.id}
                          isAdmin
                          ownerUserId={source.id}
                          selectOnly
                          disabled={selectAll.isPending}
                          value=''
                          onValueChange={(_value, group) => {
                            if (
                              group &&
                              !groups.some((g) => g.id === group.id)
                            ) {
                              setGroups([...groups, group])
                            }
                          }}
                        />
                      </div>
                    </div>
                    {selectAll.isError && (
                      <p role='alert' className='text-destructive text-sm'>
                        {assetErrorMessage(selectAll.error)}
                      </p>
                    )}
                    <ul className='max-h-40 space-y-1 overflow-y-auto'>
                      {groups.map((group) => (
                        <li
                          className='flex items-center justify-between gap-2 text-sm'
                          key={group.id}
                        >
                          <span className='min-w-0 truncate'>{group.name}</span>
                          <Button
                            type='button'
                            disabled={selectAll.isPending}
                            size='sm'
                            variant='ghost'
                            aria-label={t('Remove {{name}}', {
                              name: group.name,
                            })}
                            onClick={() =>
                              setGroups(groups.filter((g) => g.id !== group.id))
                            }
                          >
                            {t('Remove')}
                          </Button>
                        </li>
                      ))}
                    </ul>
                    <p className='text-muted-foreground text-xs'>
                      {t('{{count}} groups selected', { count: groups.length })}
                      {' · '}
                      {t('Select 1–100 groups')}
                    </p>
                  </div>
                )}
              </fieldset>
              {copy.isPending && (
                <Progress value={null}>
                  <ProgressLabel>{t('Copying...')}</ProgressLabel>
                </Progress>
              )}
              <DialogFooter>
                <Button type='submit' disabled={!canCopy}>
                  {pending ? t('Copying...') : t('Copy asset groups')}
                </Button>
              </DialogFooter>
            </form>
          )}
        </DialogContent>
      </Dialog>
      <AlertDialog
        open={confirmDelete}
        onOpenChange={(open) => !remove.isPending && setConfirmDelete(open)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('Delete copied sources?')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t(
                'Delete all sources in this batch, including sources whose destination copy failed or was deleted. Moved sources and later uploads are kept. Unused files are reclaimed after upstream cleanup. This cannot be undone.'
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={remove.isPending}>
              {t('Cancel')}
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={remove.isPending || !canDelete}
              onClick={(event) => {
                event.preventDefault()
                if (result && canDelete) remove.mutate(result.id)
              }}
            >
              {t('Delete copied sources')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
