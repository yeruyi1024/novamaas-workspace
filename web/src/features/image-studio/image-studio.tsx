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
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { getApiKeys } from '@/features/keys/api'
import type { ApiKey } from '@/features/keys/types'
import { useAuthStore } from '@/stores/auth-store'

import { getStudioRun, getStudioSettings, listStudioRuns } from './api'
import { StudioComposer } from './components/studio-composer'
import { StudioGallery } from './components/studio-gallery'
import { StudioHistory } from './components/studio-history'
import { StudioSaveDialog } from './components/studio-save-dialog'
import { assertStudioSuccess } from './studio-utils'
import type {
  StudioResult,
  StudioRun,
  StudioRunList,
  StudioSettings,
} from './types'

export function ImageStudio() {
  const { t } = useTranslation()
  const settings = useQuery({
    queryKey: ['image-studio', 'settings'],
    queryFn: async () => assertStudioSuccess(await getStudioSettings()),
  })
  const keys = useQuery({
    queryKey: ['api-keys', 'image-studio'],
    queryFn: async () => {
      const items: ApiKey[] = []
      for (let page = 1; ; page++) {
        const response = await getApiKeys({ p: page, size: 100 })
        if (!response.success) {
          throw new Error(response.message || 'Request failed')
        }
        const batch = response.data?.items ?? []
        items.push(...batch)
        if (!batch.length || items.length >= (response.data?.total ?? 0)) break
      }
      return items.filter((key) => key.status === 1)
    },
  })

  return (
    <div className='flex min-h-0 flex-1 flex-col overflow-y-auto xl:overflow-hidden'>
      {(settings.isPending || keys.isPending) && (
        <p role='status' className='p-6'>
          {t('Loading...')}
        </p>
      )}
      {(settings.isError || keys.isError) && (
        <div role='alert' className='p-6'>
          <p>{t('Failed to load data')}</p>
          <Button
            variant='outline'
            onClick={() => {
              void settings.refetch()
              void keys.refetch()
            }}
          >
            {t('Retry')}
          </Button>
        </div>
      )}
      {settings.data && keys.data && (
        <StudioWorkspace settings={settings.data} keys={keys.data} />
      )}
    </div>
  )
}

function StudioWorkspace(props: { settings: StudioSettings; keys: ApiKey[] }) {
  const { t } = useTranslation()
  const ownerName = useAuthStore((state) => state.auth.user?.username || '')
  const client = useQueryClient()
  const [page, setPage] = useState(1)
  const [runId, setActiveId] = useState('')
  const [activeResultId, setActiveResultId] = useState('')
  const [saveTarget, setSaveTarget] = useState<StudioResult>()
  const libraryEnabled = props.settings.asset_library_enabled === true
  const runs = useQuery({
    queryKey: ['image-studio', 'runs', page],
    queryFn: async () => assertStudioSuccess(await listStudioRuns(page)),
  })
  const current = useQuery({
    queryKey: ['image-studio', 'run', runId],
    queryFn: async () => assertStudioSuccess(await getStudioRun(runId)),
    enabled: Boolean(runId),
    staleTime: 30_000,
    gcTime: 0,
    refetchInterval: (query) =>
      ['pending', 'running'].includes(query.state.data?.status ?? '')
        ? 1500
        : false,
  })
  const resultId =
    current.data?.results.find((result) => result.id === activeResultId)?.id ||
    current.data?.results[0]?.id
  const selectRun = (run: StudioRun, selectedId = run.results[0]?.id || '') => {
    setActiveId(run.id)
    setActiveResultId(selectedId)
  }
  const removed = async () => {
    await client.invalidateQueries({ queryKey: ['image-studio', 'runs'] })
    const refreshed = client.getQueryData<StudioRunList>([
      'image-studio',
      'runs',
      page,
    ])
    if (!runId) {
      if (!refreshed?.items.length && page > 1) setPage(page - 1)
      return
    }
    const selected =
      refreshed?.items.find((run) => run.id === runId) || refreshed?.items[0]
    if (selected) {
      selectRun(selected)
      await client.invalidateQueries({
        queryKey: ['image-studio', 'run', selected.id],
      })
    } else {
      setActiveId('')
      setActiveResultId('')
      if (page > 1) setPage(page - 1)
    }
  }
  return (
    <StudioComposer
      settings={props.settings}
      keys={props.keys}
      disabled={['pending', 'running'].includes(current.data?.status ?? '')}
      onGenerated={(run) => {
        client.setQueryData(['image-studio', 'run', run.id], run)
        selectRun(run)
        void client.invalidateQueries({ queryKey: ['image-studio', 'runs'] })
      }}
    >
      {(generating, reusePrompt) => (
        <>
          <div className='bg-card/60 flex min-h-0 min-w-0 flex-col overflow-y-auto rounded-2xl border p-4 shadow-sm'>
            {current.isError && <p role='alert'>{t('Failed to load data')}</p>}
            {current.isPending && runId && (
              <p role='status'>{t('Loading...')}</p>
            )}
            <StudioGallery
              run={current.data}
              resultId={resultId}
              ownerName={ownerName}
              onReusePrompt={reusePrompt}
              reuseDisabled={
                !props.settings.enabled || generating || Boolean(saveTarget)
              }
              onSave={libraryEnabled ? setSaveTarget : undefined}
              saving={Boolean(saveTarget) || generating}
            />
          </div>
          <StudioHistory
            runs={runs}
            page={page}
            runId={runId}
            resultId={resultId}
            ownerName={ownerName}
            disabled={generating || Boolean(saveTarget)}
            onSelect={(run, id) => {
              selectRun(run, id)
              void client.invalidateQueries({
                queryKey: ['image-studio', 'run', run.id],
              })
            }}
            onPageChange={setPage}
            onRemoved={removed}
          />
          {libraryEnabled && saveTarget && (
            <StudioSaveDialog
              result={saveTarget}
              onClose={() => setSaveTarget(undefined)}
            />
          )}
        </>
      )}
    </StudioComposer>
  )
}
