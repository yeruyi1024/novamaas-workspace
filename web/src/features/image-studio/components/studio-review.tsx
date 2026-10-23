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
import { Delete02Icon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import { hasPermission } from '@/lib/admin-permissions'
import { useAuthStore } from '@/stores/auth-store'

import { deleteStudioRun, listStudioRuns } from '../api'
import type { StudioRun, StudioRunList } from '../types'
import { StudioGallery } from './studio-gallery'
import { StudioDeleteDialog, StudioRunPagination } from './studio-history'

export function StudioReview() {
  const { t } = useTranslation()
  const client = useQueryClient()
  const canDelete = useAuthStore((state) =>
    hasPermission(state.auth.user, 'image_studio', 'write')
  )
  const [deleteTarget, setDeleteTarget] = useState<StudioRun>()
  const [page, setPage] = useState(1)
  const [userId, setUserId] = useState(0)
  const runs = useQuery({
    queryKey: ['image-studio', 'review', userId, page],
    queryFn: () => listStudioRuns(page, 12, 'all', userId),
    placeholderData: keepPreviousData,
  })

  const remove = useMutation({
    mutationFn: (id: string) => deleteStudioRun(id, undefined, 'all'),
    onSuccess: async () => {
      setDeleteTarget(undefined)
      await client.invalidateQueries({ queryKey: ['image-studio'] })
      const refreshed = client.getQueryData<StudioRunList>([
        'image-studio',
        'review',
        userId,
        page,
      ])
      if (!refreshed?.items.length && page > 1) setPage(page - 1)
    },
  })

  return (
    <section className='flex min-h-0 flex-1 flex-col gap-4 p-4 sm:p-6'>
      <header className='flex flex-wrap items-center justify-between gap-3'>
        <div>
          <h1 className='text-xl font-semibold'>{t('Image review')}</h1>
          <p className='text-muted-foreground mt-1 text-sm'>
            {t(
              'Review generated images across users. Content access is audited.'
            )}
          </p>
        </div>
        <div className='flex items-center gap-2'>
          <Label htmlFor='studio-review-user' className='sr-only'>
            {t('User')}
          </Label>
          <NativeSelect
            id='studio-review-user'
            value={userId}
            className='w-44'
            disabled={runs.isPending}
            onChange={(event) => {
              setUserId(Number(event.target.value))
              setPage(1)
            }}
          >
            <NativeSelectOption value={0}>{t('All users')}</NativeSelectOption>
            {runs.data?.users?.map((user) => (
              <NativeSelectOption key={user.id} value={user.id}>
                {user.name} · {t('User ID')}: {user.id}
              </NativeSelectOption>
            ))}
          </NativeSelect>
          <Button
            variant='outline'
            disabled={runs.isFetching}
            onClick={() => void runs.refetch()}
          >
            {t('Refresh')}
          </Button>
        </div>
      </header>
      <div
        role='region'
        aria-label={t('Image review')}
        tabIndex={0}
        className='grid min-h-0 flex-1 grid-cols-2 content-start gap-3 overflow-y-auto sm:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5'
        aria-busy={runs.isFetching}
      >
        {(runs.isPending || runs.isPlaceholderData) && (
          <p role='status'>{t('Loading...')}</p>
        )}
        {runs.isError && <p role='alert'>{t('Failed to load image review')}</p>}
        {!runs.isPlaceholderData && runs.data?.items.length === 0 && (
          <p>{t('No generation records')}</p>
        )}
        {!runs.isPlaceholderData &&
          runs.data?.items.map((run) => (
            <article key={run.id} className='flex min-w-0 flex-col gap-2'>
              <StudioGallery
                run={run}
                ownerName={
                  runs.data?.users?.find(
                    (user) => user.id === run.owner_user_id
                  )?.name
                }
                compact
              />
              {canDelete && (
                <Button
                  variant='outline'
                  size='sm'
                  className='text-destructive self-end'
                  aria-label={t('Delete record')}
                  disabled={run.status === 'running' || remove.isPending}
                  onClick={() => {
                    remove.reset()
                    setDeleteTarget(run)
                  }}
                >
                  <HugeiconsIcon icon={Delete02Icon} aria-hidden='true' />
                  {t('Delete record')}
                </Button>
              )}
            </article>
          ))}
      </div>
      <StudioDeleteDialog
        open={Boolean(deleteTarget)}
        busy={remove.isPending}
        title={t('Delete record')}
        description={t(
          'Delete this reviewed generation record? Asset library copies will be kept. This action is audited and cannot be undone.'
        )}
        error={remove.error}
        onClose={() => setDeleteTarget(undefined)}
      >
        <Button
          variant='destructive'
          disabled={remove.isPending || !deleteTarget}
          onClick={() => {
            if (deleteTarget) remove.mutate(deleteTarget.id)
          }}
        >
          {remove.isPending ? t('Deleting...') : t('Delete')}
        </Button>
      </StudioDeleteDialog>
      <StudioRunPagination
        page={page}
        runs={runs}
        onPageChange={setPage}
        showPage
      />
    </section>
  )
}
