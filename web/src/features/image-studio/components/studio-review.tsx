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
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'

import { listStudioRuns } from '../api'
import { StudioGallery } from './studio-gallery'
import { StudioRunPagination } from './studio-history'

export function StudioReview() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const [userId, setUserId] = useState(0)
  const runs = useQuery({
    queryKey: ['image-studio', 'review', userId, page],
    queryFn: () => listStudioRuns(page, 12, 'all', userId),
    placeholderData: keepPreviousData,
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
            <StudioGallery
              key={run.id}
              run={run}
              ownerName={
                runs.data?.users?.find((user) => user.id === run.owner_user_id)
                  ?.name
              }
              compact
            />
          ))}
      </div>
      <StudioRunPagination
        page={page}
        runs={runs}
        onPageChange={setPage}
        showPage
      />
    </section>
  )
}
