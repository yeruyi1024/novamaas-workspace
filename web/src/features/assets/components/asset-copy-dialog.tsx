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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { z } from 'zod'

import { Alert, AlertDescription } from '@/components/ui/alert'
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
import { getUser } from '@/features/users/api'

import {
  copyAssetGroups,
  deleteAssetCopySources,
  getAssetCopy,
  listAssetGroupsPage,
} from '../api'
import { assertAssetSuccess, assetErrorMessage } from '../asset-utils'
import type { AssetCopyResult, AssetGroup } from '../types'
import { AssetCopyUserSelect } from './asset-copy-user-select'
import { AssetGroupSelect } from './asset-group-select'

const copySchema = z
  .object({
    source_user_id: z.number().int().positive(),
    target_user_id: z.number().int().positive(),
    group_ids: z.array(z.string()).min(1).max(100),
  })
  .refine((value) => value.source_user_id !== value.target_user_id)
type CopyForm = z.infer<typeof copySchema>

export function AssetCopyDialog(props: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const form = useForm<CopyForm>({
    resolver: zodResolver(copySchema),
    defaultValues: {
      source_user_id: 0,
      target_user_id: 0,
      group_ids: [],
    },
  })
  const values = form.watch()
  const [groups, setGroups] = useState<AssetGroup[]>([])
  const [submitted, setResult] = useState<AssetCopyResult | null>(null)
  const receipt = useQuery({
    queryKey: ['asset-copy-receipt', submitted?.id],
    queryFn: async () => {
      if (!submitted) {
        throw new Error('No copy receipt selected')
      }
      return assertAssetSuccess(await getAssetCopy(submitted.id))
    },
    enabled: props.open && submitted !== null,
    initialData: submitted ?? undefined,
    refetchInterval: (query) =>
      query.state.data?.processing_count ? 3000 : false,
  })
  const result = receipt.data ?? submitted
  const [confirmDelete, setConfirmDelete] = useState(false)
  const request = useRef({ signature: '', id: '' })
  const sourceSelectionVersion = useRef(0)
  const source = useQuery({
    queryKey: ['asset-copy-user', values.source_user_id],
    queryFn: () => getUser(values.source_user_id),
    enabled: props.open && !result && values.source_user_id > 0,
    retry: false,
  })
  const target = useQuery({
    queryKey: ['asset-copy-user', values.target_user_id],
    queryFn: () => getUser(values.target_user_id),
    enabled: props.open && !result && values.target_user_id > 0,
    retry: false,
  })
  const selectAll = useMutation({
    mutationFn: async (input: { ownerUserId: number; version: number }) => {
      const data = assertAssetSuccess(
        await listAssetGroupsPage({
          includeAllOwners: true,
          ownerUserId: input.ownerUserId,
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
      if (data.items.length !== data.total) {
        throw new Error(
          t('The group list changed. Please select all groups again.')
        )
      }
      return data.items
    },
    onSuccess: (items, input) => {
      if (
        input.ownerUserId !== form.getValues('source_user_id') ||
        input.version !== sourceSelectionVersion.current
      ) {
        return
      }
      request.current = { signature: '', id: '' }
      setGroups(items)
      form.setValue(
        'group_ids',
        items.map((group) => group.id)
      )
    },
  })
  const copy = useMutation({
    mutationFn: async (input: CopyForm) => {
      const signature = JSON.stringify(input)
      if (request.current.signature !== signature) {
        request.current = { signature, id: crypto.randomUUID() }
      }
      return assertAssetSuccess(
        await copyAssetGroups({ ...input, request_id: request.current.id })
      )
    },
    onSuccess: async (data) => {
      setResult(data)
      await queryClient.invalidateQueries({ queryKey: ['asset-library'] })
      toast.success(t('Asset groups copied'))
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })
  const remove = useMutation({
    mutationFn: async (id: string) =>
      assertAssetSuccess(await deleteAssetCopySources(id)),
    onSuccess: async (data) => {
      queryClient.setQueryData(['asset-copy-receipt', data.id], data)
      setResult(data)
      setConfirmDelete(false)
      await queryClient.invalidateQueries({ queryKey: ['asset-library'] })
      toast.success(t('Copied source assets deleted'))
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })
  const pending = copy.isPending || remove.isPending
  const canCopy =
    copySchema.safeParse(values).success &&
    !selectAll.isPending &&
    source.data?.success &&
    target.data?.success &&
    !source.isFetching &&
    !target.isFetching &&
    !source.isError &&
    !target.isError

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
              <Alert>
                <AlertDescription>
                  {t('Created {{count}} assets in {{groups}} groups', {
                    count: result.asset_count,
                    groups: result.groups.length,
                  })}
                </AlertDescription>
              </Alert>
              <p role='status' className='text-sm'>
                {t(
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
              <Button
                variant='outline'
                disabled={receipt.isFetching || pending}
                onClick={() => void receipt.refetch()}
              >
                {t('Refresh')}
              </Button>
              <p className='text-sm'>
                {t('Source user ID')}: {result.source_user_id} →{' '}
                {t('Target user ID')}: {result.target_user_id}
              </p>
              <ul className='max-h-48 space-y-1 overflow-y-auto text-sm'>
                {result.groups.map((group) => (
                  <li className='break-words' key={group.id}>
                    {group.name} · {group.asset_count}
                  </li>
                ))}
              </ul>
              <p className='text-muted-foreground text-sm'>
                {t(
                  'Deleting the copied sources keeps the destination copies, the source account and later uploads.'
                )}
              </p>
              <DialogFooter className='flex-wrap'>
                <Button
                  variant='outline'
                  disabled={pending}
                  onClick={() => {
                    setResult(null)
                    request.current = { signature: '', id: '' }
                    copy.reset()
                    remove.reset()
                  }}
                >
                  {t('New copy')}
                </Button>
                <Button
                  variant='destructive'
                  disabled={pending || result.sources_deleted_at > 0}
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
              onSubmit={form.handleSubmit((input) => {
                if (canCopy && !pending) copy.mutate(input)
              })}
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
                      value={source.data?.data}
                      excludeUserId={values.target_user_id}
                      disabled={pending}
                      onValueChange={(user) => {
                        request.current = { signature: '', id: '' }
                        queryClient.setQueryData(['asset-copy-user', user.id], {
                          success: true,
                          data: user,
                        })
                        sourceSelectionVersion.current++
                        selectAll.reset()
                        form.setValue('source_user_id', user.id)
                        form.setValue('group_ids', [])
                        setGroups([])
                      }}
                    />
                    {(source.isError || source.data?.success === false) && (
                      <p role='alert' className='text-destructive text-sm'>
                        {t('User unavailable')}
                      </p>
                    )}
                  </Field>
                  <Field data-disabled={pending}>
                    <FieldLabel htmlFor='copy-target'>
                      {t('Target user')}
                    </FieldLabel>
                    <AssetCopyUserSelect
                      id='copy-target'
                      label={t('Target user')}
                      value={target.data?.data}
                      excludeUserId={values.source_user_id}
                      enabledOnly
                      disabled={pending}
                      onValueChange={(user) => {
                        request.current = { signature: '', id: '' }
                        queryClient.setQueryData(['asset-copy-user', user.id], {
                          success: true,
                          data: user,
                        })
                        form.setValue('target_user_id', user.id)
                      }}
                    />
                    {(target.isError || target.data?.success === false) && (
                      <p role='alert' className='text-destructive text-sm'>
                        {t('User unavailable')}
                      </p>
                    )}
                  </Field>
                </FieldGroup>
                {source.data?.success && values.source_user_id > 0 && (
                  <div className='space-y-2'>
                    <div className='flex flex-col gap-2 sm:flex-row'>
                      <Button
                        type='button'
                        variant='outline'
                        className='min-h-12 shrink-0'
                        disabled={selectAll.isPending || source.isFetching}
                        onClick={() =>
                          selectAll.mutate({
                            ownerUserId: values.source_user_id,
                            version: sourceSelectionVersion.current,
                          })
                        }
                      >
                        {selectAll.isPending
                          ? t('Loading...')
                          : t('Select all asset groups')}
                      </Button>
                      <div className='min-w-0 flex-1'>
                        <AssetGroupSelect
                          key={values.source_user_id}
                          isAdmin
                          ownerUserId={values.source_user_id}
                          selectOnly
                          disabled={selectAll.isPending}
                          value=''
                          onValueChange={(_value, group) => {
                            if (group && !values.group_ids.includes(group.id)) {
                              request.current = { signature: '', id: '' }
                              setGroups((current) => [...current, group])
                              form.setValue('group_ids', [
                                ...values.group_ids,
                                group.id,
                              ])
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
                    {selectAll.isSuccess && selectAll.data.length === 0 && (
                      <p
                        role='status'
                        className='text-muted-foreground text-sm'
                      >
                        {t('This user has no asset groups.')}
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
                            onClick={() => {
                              request.current = { signature: '', id: '' }
                              setGroups((current) =>
                                current.filter((item) => item.id !== group.id)
                              )
                              form.setValue(
                                'group_ids',
                                values.group_ids.filter((id) => id !== group.id)
                              )
                            }}
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

              <DialogFooter>
                <Button type='submit' disabled={!canCopy || pending}>
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
                'Delete the source assets saved in this copy. Copies are kept, and so are sources whose copy was deleted, sources moved to other groups, and later uploads. This cannot be undone.'
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={remove.isPending}>
              {t('Cancel')}
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={remove.isPending}
              onClick={(event) => {
                event.preventDefault()
                if (result) remove.mutate(result.id)
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
