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

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { TabsContent } from '@/components/ui/tabs'

import type { Assessment } from '../baselines'
import {
  LOAD_PRESETS,
  MAX_CONCURRENCY,
  MAX_ROUNDS,
  MAX_STRESS_REQUESTS,
  MAX_TOKENS_CAP,
  STRESS_WARN_TOTAL,
} from '../constants'
import type { StressForm, StressMetrics } from '../types'
import { AssessmentTable } from './assessment-table'
import { CorpusPicker, NumberField, StreamSwitch } from './form-controls'
import { StressResults } from './stress-results'

export function StressPanel(props: {
  stress: StressForm
  busy: boolean
  runningModule: string | null
  progress: { completed: number; total: number }
  progressValue: number
  streamText: string
  stressSummary: string
  stressAssessment: Assessment | null
  metrics: StressMetrics | null
  onStressChange: Dispatch<SetStateAction<StressForm>>
}) {
  const { t } = useTranslation()
  const stressTotal = props.stress.concurrency * props.stress.rounds

  return (
    <TabsContent value='stress' className='space-y-4'>
      <p className='text-muted-foreground text-sm'>
        {t(
          'Corpus is sent as-is. Allow cache reuses it; break cache puts a random prefix in front of each request. Max tokens only caps the reply.'
        )}
      </p>
      <p className='text-muted-foreground text-sm'>
        {t(
          'Timing is measured by this server. First output is the first streamed content or reasoning chunk; request duration ends when response parsing finishes. Network and gateway buffering can differ from upstream logs.'
        )}
      </p>
      <div className='flex flex-wrap gap-2'>
        {LOAD_PRESETS.map((preset) => (
          <Button
            key={preset.id}
            type='button'
            variant='outline'
            size='sm'
            disabled={props.busy}
            onClick={() =>
              props.onStressChange((current) => ({
                ...current,
                concurrency: preset.concurrency,
                rounds: preset.rounds,
                maxTokens: preset.maxTokens,
              }))
            }
          >
            {t(preset.labelKey)}
          </Button>
        ))}
      </div>
      <div className='mt-3 flex flex-wrap gap-2'>
        <Button
          type='button'
          variant={props.stress.breakCache ? 'outline' : 'secondary'}
          size='sm'
          disabled={props.busy}
          onClick={() =>
            props.onStressChange((current) => ({
              ...current,
              breakCache: false,
            }))
          }
        >
          {t('Allow prompt cache')}
        </Button>
        <Button
          type='button'
          variant={props.stress.breakCache ? 'secondary' : 'outline'}
          size='sm'
          disabled={props.busy}
          onClick={() =>
            props.onStressChange((current) => ({
              ...current,
              breakCache: true,
            }))
          }
        >
          {t('Break prompt cache')}
        </Button>
      </div>
      <div className='mt-4 grid gap-4 md:grid-cols-4'>
        <NumberField
          id='stress-concurrency'
          label={t('Concurrency')}
          value={props.stress.concurrency}
          disabled={props.busy}
          min={1}
          max={MAX_CONCURRENCY}
          presets={[
            { label: '1', value: 1 },
            { label: '5', value: 5 },
            { label: '10', value: 10 },
            { label: '20', value: 20 },
            { label: '50', value: 50 },
          ]}
          onChange={(value) =>
            props.onStressChange((current) => ({
              ...current,
              concurrency: value,
            }))
          }
        />
        <NumberField
          id='stress-rounds'
          label={t('Rounds')}
          value={props.stress.rounds}
          disabled={props.busy}
          min={1}
          max={MAX_ROUNDS}
          presets={[
            { label: '1', value: 1 },
            { label: '2', value: 2 },
            { label: '5', value: 5 },
            { label: '10', value: 10 },
          ]}
          onChange={(value) =>
            props.onStressChange((current) => ({ ...current, rounds: value }))
          }
        />
        <NumberField
          id='stress-max-tokens'
          label={t('Max tokens')}
          value={props.stress.maxTokens}
          disabled={props.busy}
          min={1}
          max={MAX_TOKENS_CAP}
          presets={[
            { label: '64', value: 64 },
            { label: '256', value: 256 },
            { label: '512', value: 512 },
            { label: '1k', value: 1024 },
            { label: '4k', value: 4096 },
          ]}
          onChange={(value) =>
            props.onStressChange((current) => ({
              ...current,
              maxTokens: value,
            }))
          }
        />
        <StreamSwitch
          id='stress-stream'
          checked={props.stress.stream}
          disabled={props.busy}
          onChange={(checked) =>
            props.onStressChange((current) => ({ ...current, stream: checked }))
          }
        />
      </div>
      <div className='mt-4'>
        <CorpusPicker
          id='stress-prompt'
          form={props.stress}
          disabled={props.busy}
          onChange={(next) =>
            props.onStressChange((current) => ({
              ...current,
              corpus: next.corpus,
              prompt: next.prompt,
            }))
          }
        />
      </div>
      {stressTotal >= STRESS_WARN_TOTAL ? (
        <Alert className='mt-4'>
          <AlertDescription>
            {t(
              'This run will send {{count}} requests. That can consume a large amount of upstream quota.',
              { count: stressTotal }
            )}
          </AlertDescription>
        </Alert>
      ) : null}
      <p className='text-muted-foreground text-xs'>
        {t(
          'The stress test is capped at {{max}} requests per run to avoid accidental upstream load.',
          { max: MAX_STRESS_REQUESTS }
        )}
      </p>
      {props.runningModule === 'stress' && props.progress.total > 0 ? (
        <div className='mt-4 space-y-2'>
          <div className='text-muted-foreground text-sm'>
            {t('Progress')}: {props.progress.completed}/{props.progress.total}
          </div>
          <Progress value={props.progressValue} />
        </div>
      ) : null}
      {props.runningModule === 'stress' && props.streamText ? (
        <pre className='bg-muted mt-4 max-h-64 overflow-auto rounded-lg p-3 text-sm whitespace-pre-wrap'>
          {props.streamText}
        </pre>
      ) : null}
      {props.stressSummary ? (
        <p className='text-muted-foreground mt-4 text-sm'>
          {props.stressSummary}
        </p>
      ) : null}
      {props.metrics ? (
        <StressResults
          metrics={props.metrics}
          assessment={props.stressAssessment}
          stream={props.stress.stream}
        />
      ) : null}
      {props.stressAssessment ? (
        <AssessmentTable
          title={t('Stress test assessment')}
          assessment={props.stressAssessment}
        />
      ) : null}
    </TabsContent>
  )
}
