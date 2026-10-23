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
import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { getApiKeys } from '@/features/keys/api'
import type { ApiKey } from '@/features/keys/types'
import { useAuthStore } from '@/stores/auth-store'

import '@/styles/image-studio.css'

import { getStudioRun, getStudioSettings, listStudioRuns } from './api'
import { StudioComposer } from './components/studio-composer'
import { StudioGallery } from './components/studio-gallery'
import { StudioHistory } from './components/studio-history'
import { StudioSaveDialog } from './components/studio-save-dialog'
import { StudioReferencePreview } from './components/studio-upload-history'
import { useStudioPolling } from './hooks/use-studio-polling'
import type {
  StudioResult,
  StudioRun,
  StudioRunList,
  StudioSettings,
  StudioReference,
} from './types'

export function ImageStudio() {
  const { t } = useTranslation()
  const userId = useAuthStore((state) =>
    state.auth.accessToken ? (state.auth.user?.id ?? 0) : 0
  )
  const settings = useQuery({
    queryKey: ['image-studio', 'settings'],
    queryFn: getStudioSettings,
  })
  const keys = useQuery({
    queryKey: ['api-keys', 'image-studio', userId],
    enabled: Boolean(userId),
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
    <div className='image-studio-theme studio-shell bg-background flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden'>
      {(settings.isPending || (userId > 0 && keys.isPending)) && (
        <p role='status' className='p-6'>
          {t('Loading...')}
        </p>
      )}
      {(settings.isError || (userId > 0 && keys.isError)) && (
        <div role='alert' className='p-6'>
          <p>{t('Failed to load data')}</p>
          <Button
            variant='outline'
            onClick={() => {
              void settings.refetch()
              if (userId) void keys.refetch()
            }}
          >
            {t('Retry')}
          </Button>
        </div>
      )}
      {settings.data && (!userId || keys.data) && (
        <StudioWorkspace
          key={userId}
          userId={userId}
          settings={settings.data}
          keys={keys.data ?? []}
        />
      )}
    </div>
  )
}

function StudioWorkspace(props: {
  userId: number
  settings: StudioSettings
  keys: ApiKey[]
}) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const galleryRef = useRef<HTMLDivElement>(null)
  const [page, setPage] = useState(1)
  const [historyTab, setHistoryTab] = useState('generated')
  const [selection, setSelection] = useState<{
    runId?: string
    resultId?: string
    upload?: StudioReference
  }>({})
  const {
    runId = '',
    resultId: activeResultId,
    upload: selectedUpload,
  } = selection
  const [saveTarget, setSaveTarget] = useState<StudioResult>()
  const runs = useQuery({
    queryKey: ['image-studio', 'runs', props.userId, page],
    queryFn: () => listStudioRuns(page),
    enabled: Boolean(props.userId),
  })
  const { track: trackRun, runningCount } = useStudioPolling(props.userId, runs.data?.items)
  const current = useQuery({
    queryKey: ['image-studio', 'run', props.userId, runId],
    queryFn: () => getStudioRun(runId),
    enabled: Boolean(props.userId && runId),
    staleTime: 30_000,
  })
  const resultId =
    current.data?.results.find((result) => result.id === activeResultId)?.id ||
    current.data?.results[0]?.id
  const selectRun = (run: StudioRun, resultId?: string) => {
    setSelection({ runId: run.id, resultId })
  }
  const removed = async () => {
    await client.invalidateQueries({ queryKey: ['image-studio', 'runs'] })
    const refreshed = client.getQueryData<StudioRunList>([
      'image-studio',
      'runs',
      props.userId,
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
        queryKey: ['image-studio', 'run', props.userId, selected.id],
      })
    } else {
      setSelection({})
      if (page > 1) setPage(page - 1)
    }
  }
  return (
    <StudioComposer
      authenticated={Boolean(props.userId)}
      settings={props.settings}
      keys={props.keys}
      runningCount={runningCount}
      onNewCreation={() => setSelection({})}
      onGenerated={(run) => {
        trackRun(run)
        setPage(1)
        setHistoryTab('generated')
        client.setQueryData(['image-studio', 'run', props.userId, run.id], run)
        setSelection({})
        void client.invalidateQueries({ queryKey: ['image-studio', 'runs'] })
      }}
    >
      {(actions) => ({
        canvas: (
          <div
            className='flex min-h-0 min-w-0 flex-1 flex-col outline-none'
            ref={galleryRef}
            tabIndex={-1}
            role='group'
            aria-label={t('Generation results')}
          >
            {!selectedUpload && current.isError && (
              <p role='alert' className='m-auto p-6 text-sm'>
                {t('Failed to load data')}
              </p>
            )}
            {!selectedUpload && current.isPending && runId && (
              <p role='status' className='text-muted-foreground m-auto p-6'>
                {t('Loading...')}
              </p>
            )}
            {selectedUpload ? (
              <StudioReferencePreview
                reference={selectedUpload}
                disabled={!actions.canAddReference}
                onUse={() => actions.useReference(selectedUpload)}
                onPreview={actions.preview}
              />
            ) : (
              (!runId || current.data) && (
                <StudioGallery
                  run={current.data}
                  resultId={resultId}
                  onReusePrompt={actions.reusePrompt}
                  onPreview={actions.preview}
                  reuseDisabled={
                    !props.settings.enabled ||
                    actions.generating ||
                    Boolean(saveTarget)
                  }
                  onSave={setSaveTarget}
                  saving={Boolean(saveTarget) || actions.generating}
                  maxPromptCharacters={props.settings.max_prompt_characters}
                />
              )
            )}
            {saveTarget && (
              <StudioSaveDialog
                result={saveTarget}
                onClose={() => setSaveTarget(undefined)}
              />
            )}
          </div>
        ),
        history: (
          <StudioHistory
            tab={historyTab}
            onTabChange={setHistoryTab}
            authenticated={Boolean(props.userId)}
            runs={runs}
            page={page}
            runId={runId}
            resultId={resultId}
            disabled={actions.generating || Boolean(saveTarget)}
            uploadId={selectedUpload?.id}
            referenceIds={actions.referenceIds}
            canAddReference={actions.canAddReference}
            uploading={actions.uploading}
            onUpload={actions.uploadReferences}
            onUseReference={actions.useReference}
            onSelectUpload={(reference) => {
              setSelection({ upload: reference })
              actions.preview()
              requestAnimationFrame(() => {
                galleryRef.current?.scrollIntoView({ block: 'start' })
                galleryRef.current?.focus({ preventScroll: true })
              })
            }}
            onSelect={(run, id) => {
              actions.preview()
              selectRun(run, id)
              galleryRef.current?.scrollIntoView({ block: 'start' })
              galleryRef.current?.focus({ preventScroll: true })
              void client.invalidateQueries({
                queryKey: ['image-studio', 'run', props.userId, run.id],
              })
            }}
            onPageChange={setPage}
            onRemoved={removed}
          />
        ),
      })}
    </StudioComposer>
  )
}
