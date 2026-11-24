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
import {
  AlertCircle,
  ArrowDown,
  ArrowRight,
  CheckCheck,
  CheckCircle2,
  Copy,
  Folder,
  FolderSync,
  Info,
  Loader2,
  RefreshCw,
  RotateCcw,
  Trash2,
  X,
} from 'lucide-react'
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
import { cn } from '@/lib/utils'

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
import { AssetPickerTile } from './asset-picker'

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
      queryClient.invalidateQueries({ queryKey: ['asset-library', 'owners'] }),
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

  const hasBothUsers = Boolean(source && target)
  const failedTotal = result
    ? result.failed_count + result.unavailable_count
    : 0
  let resultStatusIcon = (
    <CheckCircle2
      className='size-4 shrink-0 text-emerald-500'
      aria-hidden='true'
    />
  )
  if (result && result.processing_count > 0) {
    resultStatusIcon = (
      <Loader2
        className='text-primary size-4 shrink-0 animate-spin'
        aria-hidden='true'
      />
    )
  } else if (failedTotal > 0) {
    resultStatusIcon = (
      <AlertCircle
        className='size-4 shrink-0 text-amber-500'
        aria-hidden='true'
      />
    )
  }

  return (
    <>
      <Dialog
        open={props.open}
        onOpenChange={(open) => !pending && props.onOpenChange(open)}
      >
        <DialogContent className='max-h-[85dvh] overflow-y-auto sm:max-w-2xl'>
          <DialogHeader className='flex-row items-start gap-3.5 text-left'>
            <span
              aria-hidden='true'
              className='bg-primary/10 text-primary ring-primary/20 mt-0.5 grid size-10 shrink-0 place-items-center rounded-xl ring-1'
            >
              <FolderSync className='size-5' strokeWidth={1.75} />
            </span>
            <div className='min-w-0 space-y-1'>
              <DialogTitle>{t('Copy asset groups')}</DialogTitle>
              <DialogDescription className='leading-relaxed'>
                {t(
                  'Copy the selected asset groups to another user while keeping the original asset and group IDs. Synchronization is handled automatically, and the source assets are kept.'
                )}
              </DialogDescription>
            </div>
          </DialogHeader>
          {result ? (
            <div className='space-y-4'>
              <div className='bg-muted/30 border-border/70 space-y-3 rounded-xl border p-4'>
                <div className='flex flex-wrap items-center justify-between gap-2'>
                  <div className='flex items-center gap-2 text-sm font-medium'>
                    {resultStatusIcon}
                    <span>
                      {t('Created {{count}} assets in {{groups}} groups', {
                        count: result.asset_count,
                        groups: result.groups.length,
                      })}
                    </span>
                  </div>
                  {source && target && (
                    <span className='bg-background/80 border-border/70 text-muted-foreground inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs'>
                      <span className='text-foreground max-w-28 truncate font-medium'>
                        {source.username}
                      </span>
                      <ArrowRight className='size-3 shrink-0' />
                      <span className='text-foreground max-w-28 truncate font-medium'>
                        {target.username}
                      </span>
                    </span>
                  )}
                </div>
                <Progress value={progress}>
                  <ProgressLabel>{t('Copy progress')}</ProgressLabel>
                  <ProgressValue />
                </Progress>
                <div className='grid grid-cols-2 gap-2 sm:grid-cols-4'>
                  <div className='bg-background/70 border-border/60 rounded-lg border px-2.5 py-1.5'>
                    <div className='text-muted-foreground text-[11px]'>
                      {t('Ready')}
                    </div>
                    <div className='text-sm font-semibold text-emerald-500 tabular-nums'>
                      {result.ready_count}
                    </div>
                  </div>
                  <div className='bg-background/70 border-border/60 rounded-lg border px-2.5 py-1.5'>
                    <div className='text-muted-foreground text-[11px]'>
                      {t('Processing')}
                    </div>
                    <div className='text-primary text-sm font-semibold tabular-nums'>
                      {result.processing_count}
                    </div>
                  </div>
                  <div className='bg-background/70 border-border/60 rounded-lg border px-2.5 py-1.5'>
                    <div className='text-muted-foreground text-[11px]'>
                      {t('Failed')}
                    </div>
                    <div className='text-destructive text-sm font-semibold tabular-nums'>
                      {result.failed_count}
                    </div>
                  </div>
                  <div className='bg-background/70 border-border/60 rounded-lg border px-2.5 py-1.5'>
                    <div className='text-muted-foreground text-[11px]'>
                      {t('Unavailable')}
                    </div>
                    <div className='text-sm font-semibold text-amber-500 tabular-nums'>
                      {result.unavailable_count}
                    </div>
                  </div>
                </div>
                <p role='status' className='text-muted-foreground text-xs'>
                  {result.processing_count === 0
                    ? t(
                        'Copy complete: {{ready}} succeeded, {{failed}} failed.',
                        {
                          ready: result.ready_count,
                          failed: failedTotal,
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
              </div>
              {receipt.isError && (
                <p
                  role='alert'
                  className='border-destructive/30 bg-destructive/10 text-destructive rounded-lg border px-3 py-2 text-xs'
                >
                  {assetErrorMessage(receipt.error)}
                </p>
              )}
              <div className='bg-muted/20 border-border/70 rounded-xl border p-3'>
                <ul className='max-h-44 space-y-1.5 overflow-y-auto pr-0.5 text-sm'>
                  {result.groups.map((group) => (
                    <li
                      className='bg-background/75 border-border/60 flex items-center justify-between gap-2.5 rounded-lg border px-3 py-2'
                      key={group.id}
                    >
                      <span className='flex min-w-0 items-center gap-2'>
                        <AssetPickerTile size='sm'>
                          <Folder className='size-3.5' strokeWidth={1.75} />
                        </AssetPickerTile>
                        <span className='min-w-0 truncate font-medium'>
                          {group.name} · {group.asset_count}
                        </span>
                      </span>
                      <span className='text-muted-foreground shrink-0 font-mono text-[11px]'>
                        {group.id}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
              <div className='bg-muted/30 border-border/60 text-muted-foreground flex items-start gap-2 rounded-lg border px-3 py-2.5 text-xs leading-relaxed'>
                <Info
                  className='mt-0.5 size-3.5 shrink-0 opacity-80'
                  aria-hidden='true'
                />
                <p>
                  {t(
                    'Once processing finishes, all sources in this batch can be deleted, even if copies failed or were deleted. Files still referenced by other assets are kept.'
                  )}
                </p>
              </div>
              <DialogFooter className='flex-wrap gap-2'>
                <Button
                  variant='outline'
                  disabled={receipt.isFetching || pending}
                  onClick={() => void receipt.refetch()}
                >
                  <RefreshCw
                    className={cn(
                      'size-3.5',
                      receipt.isFetching && 'animate-spin'
                    )}
                    aria-hidden='true'
                  />
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
                  <RotateCcw className='size-3.5' aria-hidden='true' />
                  {t('New copy')}
                </Button>
                <Button
                  variant='destructive'
                  disabled={pending || !canDelete}
                  onClick={() => setConfirmDelete(true)}
                >
                  <Trash2 className='size-3.5' aria-hidden='true' />
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
                <section className='bg-muted/30 border-border/70 space-y-3 rounded-xl border p-4'>
                  <div className='flex items-center gap-2'>
                    <span className='bg-primary/15 text-primary grid size-5 shrink-0 place-items-center rounded-full text-[11px] font-semibold tabular-nums'>
                      1
                    </span>
                    <span className='text-foreground text-xs font-semibold tracking-wide'>
                      {t('Choose source and target users')}
                    </span>
                  </div>
                  <FieldGroup className='[container-type:normal] grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] sm:items-end'>
                    <Field data-disabled={pending}>
                      <FieldLabel
                        htmlFor='copy-source'
                        className='text-muted-foreground text-xs font-medium'
                      >
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
                    <div
                      aria-hidden='true'
                      className={cn(
                        'mx-auto flex size-8 shrink-0 items-center justify-center rounded-full border transition-colors sm:mb-2',
                        hasBothUsers
                          ? 'border-primary/40 bg-primary/10 text-primary'
                          : 'border-border/70 bg-background/60 text-muted-foreground'
                      )}
                    >
                      <ArrowRight className='hidden size-4 sm:block' />
                      <ArrowDown className='size-4 sm:hidden' />
                    </div>
                    <Field data-disabled={pending}>
                      <FieldLabel
                        htmlFor='copy-target'
                        className='text-muted-foreground text-xs font-medium'
                      >
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
                </section>

                <section className='bg-muted/20 border-border/70 space-y-3 rounded-xl border p-4'>
                  <div className='flex flex-wrap items-center justify-between gap-2'>
                    <div className='flex items-center gap-2'>
                      <span
                        className={cn(
                          'grid size-5 shrink-0 place-items-center rounded-full text-[11px] font-semibold tabular-nums transition-colors',
                          source
                            ? 'bg-primary/15 text-primary'
                            : 'bg-muted text-muted-foreground'
                        )}
                      >
                        2
                      </span>
                      <span className='text-foreground text-xs font-semibold tracking-wide'>
                        {t('Asset groups to copy')}
                      </span>
                    </div>
                    {source && groups.length > 0 && (
                      <Button
                        type='button'
                        size='xs'
                        variant='ghost'
                        disabled={selectAll.isPending}
                        className='text-muted-foreground hover:text-foreground h-6 px-2 text-xs'
                        onClick={() => setGroups([])}
                      >
                        {t('Clear')}
                      </Button>
                    )}
                  </div>

                  {!source ? (
                    <div className='border-border/60 bg-background/40 text-muted-foreground flex flex-col items-center justify-center gap-1.5 rounded-lg border border-dashed px-4 py-5 text-center text-xs'>
                      <Folder
                        className='size-4 opacity-50'
                        aria-hidden='true'
                      />
                      <span>
                        {t(
                          'Select a source user first to choose asset groups.'
                        )}
                      </span>
                    </div>
                  ) : (
                    <div className='space-y-3'>
                      <div className='flex flex-col gap-2 sm:flex-row sm:items-center'>
                        <div className='min-w-0 flex-1'>
                          <AssetGroupSelect
                            key={source.id}
                            isAdmin
                            ownerUserId={source.id}
                            selectOnly
                            disabled={selectAll.isPending}
                            triggerClassName='rounded-xl border-border/80 bg-background/80 hover:border-primary/40 hover:bg-background'
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
                        <Button
                          type='button'
                          variant='secondary'
                          className='h-12 shrink-0 gap-1.5 rounded-xl px-3.5'
                          disabled={selectAll.isPending}
                          onClick={() => selectAll.mutate(source.id)}
                        >
                          <CheckCheck className='size-4' aria-hidden='true' />
                          {selectAll.isPending
                            ? t('Loading...')
                            : t('Select all asset groups')}
                        </Button>
                      </div>
                      {selectAll.isError && (
                        <p
                          role='alert'
                          className='border-destructive/30 bg-destructive/10 text-destructive rounded-lg border px-3 py-2 text-xs'
                        >
                          {assetErrorMessage(selectAll.error)}
                        </p>
                      )}
                      {groups.length === 0 ? (
                        <div className='border-border/60 bg-background/40 text-muted-foreground rounded-lg border border-dashed px-3 py-4 text-center text-xs'>
                          {t(
                            'No groups selected yet. Pick from the dropdown above or select all.'
                          )}
                        </div>
                      ) : (
                        <ul className='max-h-48 space-y-1.5 overflow-y-auto pr-0.5'>
                          {groups.map((group) => (
                            <li
                              className='bg-background/80 border-border/70 flex items-center justify-between gap-2.5 rounded-lg border px-3 py-2 text-sm'
                              key={group.id}
                            >
                              <div className='flex min-w-0 items-center gap-2.5'>
                                <AssetPickerTile size='sm'>
                                  <Folder
                                    className='size-3.5'
                                    strokeWidth={1.75}
                                  />
                                </AssetPickerTile>
                                <div className='min-w-0'>
                                  <div className='truncate font-medium'>
                                    {group.name}
                                  </div>
                                  <div className='text-muted-foreground truncate font-mono text-[11px]'>
                                    {group.id}
                                  </div>
                                </div>
                              </div>
                              <div className='flex shrink-0 items-center gap-1.5'>
                                {group.asset_count !== undefined && (
                                  <span className='bg-muted text-muted-foreground rounded-full px-2 py-0.5 text-[11px] tabular-nums'>
                                    {group.asset_count}
                                  </span>
                                )}
                                <Button
                                  type='button'
                                  disabled={selectAll.isPending}
                                  size='icon-sm'
                                  variant='ghost'
                                  className='text-muted-foreground hover:bg-destructive/10 hover:text-destructive size-7'
                                  aria-label={t('Remove {{name}}', {
                                    name: group.name,
                                  })}
                                  onClick={() =>
                                    setGroups(
                                      groups.filter((g) => g.id !== group.id)
                                    )
                                  }
                                >
                                  <X className='size-3.5' aria-hidden='true' />
                                </Button>
                              </div>
                            </li>
                          ))}
                        </ul>
                      )}
                      <p className='text-muted-foreground text-xs tabular-nums'>
                        {t('{{count}} groups selected', {
                          count: groups.length,
                        })}
                        {' · '}
                        {t('Select 1–100 groups')}
                      </p>
                    </div>
                  )}
                </section>
              </fieldset>
              {copy.isPending && (
                <Progress value={null}>
                  <ProgressLabel>{t('Copying...')}</ProgressLabel>
                </Progress>
              )}
              <DialogFooter className='gap-2'>
                <Button
                  type='button'
                  variant='outline'
                  disabled={pending}
                  onClick={() => props.onOpenChange(false)}
                >
                  {t('Cancel')}
                </Button>
                <Button type='submit' disabled={!canCopy}>
                  <Copy className='size-3.5' aria-hidden='true' />
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
