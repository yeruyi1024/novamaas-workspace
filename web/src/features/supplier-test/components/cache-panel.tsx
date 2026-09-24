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
import type { Dispatch, SetStateAction } from 'react'
import { useTranslation } from 'react-i18next'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { TabsContent } from '@/components/ui/tabs'

import type { Assessment } from '../baselines'
import {
  CACHE_MODES,
  CACHE_ROUND_PRESETS,
  CACHE_WAIT_PRESETS,
  MAX_CACHE_ROUNDS,
  MAX_CACHE_WAIT_SECONDS,
  MAX_TOKENS_CAP,
} from '../constants'
import type { CacheForm, CacheMode, CheckResult } from '../types'
import { AssessmentTable } from './assessment-table'
import { CheckTable } from './check-table'
import {
  CorpusPicker,
  FieldSelect,
  NumberField,
  StreamSwitch,
} from './form-controls'

export function CachePanel(props: {
  cache: CacheForm
  busy: boolean
  cacheChecks: CheckResult[]
  cacheSummary: string
  cacheAssessment: Assessment | null
  onCacheChange: Dispatch<SetStateAction<CacheForm>>
}) {
  const { t } = useTranslation()

  return (
    <TabsContent value='cache' className='space-y-4'>
      <p className='text-muted-foreground text-sm'>
        {t(
          'Corpus is the input prefix, sent as-is. Max tokens only caps the reply. The first request warms cache; later rounds check the hit.'
        )}
      </p>
      <div className='grid gap-4 md:grid-cols-2'>
        <FieldSelect
          label={t('Cache mode')}
          value={props.cache.mode}
          disabled={props.busy}
          items={CACHE_MODES.map((m) => ({
            value: m.value,
            label: t(m.labelKey),
          }))}
          onChange={(value) =>
            props.onCacheChange((current) => ({
              ...current,
              mode: value as CacheMode,
            }))
          }
        />
        <div className='flex items-end text-sm text-muted-foreground pb-2'>
          {t(
            CACHE_MODES.find((m) => m.value === props.cache.mode)?.hintKey || ''
          )}
        </div>
      </div>
      <div className='grid gap-4 md:grid-cols-4'>
        <NumberField
          id='cache-wait'
          label={t('Wait seconds')}
          value={props.cache.waitSeconds}
          disabled={props.busy}
          min={0}
          max={MAX_CACHE_WAIT_SECONDS}
          presets={CACHE_WAIT_PRESETS.map((s) => ({
            label: `${s}s`,
            value: s,
          }))}
          onChange={(value) =>
            props.onCacheChange((current) => ({ ...current, waitSeconds: value }))
          }
        />
        <NumberField
          id='cache-rounds'
          label={t('Probe rounds')}
          value={props.cache.rounds}
          disabled={props.busy}
          min={1}
          max={MAX_CACHE_ROUNDS}
          presets={CACHE_ROUND_PRESETS.map((r) => ({
            label: String(r),
            value: r,
          }))}
          onChange={(value) =>
            props.onCacheChange((current) => ({
              ...current,
              rounds: value,
            }))
          }
        />
        <NumberField
          id='cache-max-tokens'
          label={t('Max tokens')}
          value={props.cache.maxTokens}
          disabled={props.busy}
          min={1}
          max={MAX_TOKENS_CAP}
          presets={[
            { label: '16', value: 16 },
            { label: '64', value: 64 },
            { label: '256', value: 256 },
          ]}
          onChange={(value) =>
            props.onCacheChange((current) => ({ ...current, maxTokens: value }))
          }
        />
        <StreamSwitch
          id='cache-stream'
          checked={props.cache.stream}
          disabled={props.busy}
          onChange={(checked) =>
            props.onCacheChange((current) => ({ ...current, stream: checked }))
          }
        />
      </div>
      <div className='mt-4'>
        <CorpusPicker
          id='cache-prompt'
          form={props.cache}
          disabled={props.busy}
          onChange={(next) =>
            props.onCacheChange((current) => ({
              ...current,
              corpus: next.corpus,
              prompt: next.prompt,
            }))
          }
        />
      </div>
      <div className='mt-4 space-y-2'>
        <Label htmlFor='cache-follow-up'>{t('Follow-up question')}</Label>
        <Input
          id='cache-follow-up'
          value={props.cache.followUp}
          disabled={props.busy}
          onChange={(event) =>
            props.onCacheChange((current) => ({
              ...current,
              followUp: event.target.value,
            }))
          }
        />
      </div>
      {props.cacheSummary ? (
        <p className='text-muted-foreground mt-4 text-sm'>
          {props.cacheSummary}
        </p>
      ) : null}
      <div className='mt-4'>
        <CheckTable checks={props.cacheChecks} busy={props.busy} />
      </div>
      {props.cacheAssessment && props.cacheAssessment.rows.length > 0 ? (
        <AssessmentTable
          title={t('Prompt cache assessment')}
          assessment={props.cacheAssessment}
        />
      ) : null}
    </TabsContent>
  )
}
