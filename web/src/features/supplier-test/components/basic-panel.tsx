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
import { ShieldCheck } from 'lucide-react'
import type { Dispatch, SetStateAction } from 'react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { TabsContent } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'

import {
  MAX_TOKENS_CAP,
  PROTOCOL_BASIC_IDS,
  SHALLOW_BASIC_IDS,
} from '../constants'
import type { BasicForm, CheckResult } from '../types'
import type { VendorId } from '../vendors'
import { CheckTable } from './check-table'
import { NumberField, OptionalNumberField, StreamSwitch } from './form-controls'

export function BasicPanel(props: {
  basic: BasicForm
  busy: boolean
  basicChecks: CheckResult[]
  basicStreamText: string
  basicSummary: string
  vendor?: VendorId
  onBasicChange: Dispatch<SetStateAction<BasicForm>>
  onRunCheck: (id: string) => void
}) {
  const { t } = useTranslation()

  return (
    <TabsContent value='basic' className='space-y-4'>
      <p className='text-muted-foreground text-sm'>
        {t(
          'Shallow checks ask whether the door opens. Protocol checks are skipped when the vendor has no matching API. Empty temperature / top_p are not sent.'
        )}
      </p>
      <div className='grid gap-4 md:grid-cols-4'>
        <NumberField
          id='basic-max-tokens'
          label={t('Max tokens')}
          value={props.basic.maxTokens}
          disabled={props.busy}
          min={1}
          max={MAX_TOKENS_CAP}
          presets={[
            { label: '64', value: 64 },
            { label: '256', value: 256 },
            { label: '1k', value: 1024 },
            { label: '4k', value: 4096 },
          ]}
          onChange={(value) =>
            props.onBasicChange((current) => ({ ...current, maxTokens: value }))
          }
        />
        <OptionalNumberField
          id='basic-temperature'
          label={t('Temperature')}
          value={props.basic.temperature}
          disabled={props.busy}
          min={0}
          max={2}
          step={0.1}
          placeholder={t('Leave empty to omit')}
          onChange={(value) =>
            props.onBasicChange((current) => ({
              ...current,
              temperature: value,
            }))
          }
        />
        <OptionalNumberField
          id='basic-top-p'
          label={t('Top P')}
          value={props.basic.topP}
          disabled={props.busy}
          min={0}
          max={1}
          step={0.05}
          placeholder={t('Leave empty to omit')}
          onChange={(value) =>
            props.onBasicChange((current) => ({ ...current, topP: value }))
          }
        />
        <StreamSwitch
          id='basic-stream'
          checked={props.basic.stream}
          disabled={props.busy}
          onChange={(checked) =>
            props.onBasicChange((current) => ({ ...current, stream: checked }))
          }
        />
      </div>
      <div className='mt-4 space-y-2'>
        <Label htmlFor='basic-prompt'>{t('Prompt')}</Label>
        <Textarea
          id='basic-prompt'
          rows={3}
          value={props.basic.prompt}
          disabled={props.busy}
          onChange={(event) =>
            props.onBasicChange((current) => ({
              ...current,
              prompt: event.target.value,
            }))
          }
        />
      </div>
      {props.basicStreamText ? (
        <pre className='bg-muted mt-4 max-h-48 overflow-auto rounded-lg p-3 text-sm whitespace-pre-wrap'>
          {props.basicStreamText}
        </pre>
      ) : null}
      {props.basicSummary ? (
        <p className='text-muted-foreground mt-4 text-sm'>
          {props.basicSummary}
        </p>
      ) : null}
      <div className='space-y-4'>
        <div>
          <p className='mb-2 text-sm font-medium'>
            {t('Shallow · connectivity')}
          </p>
          <CheckTable
            checks={props.basicChecks.filter((check) =>
              (SHALLOW_BASIC_IDS as readonly string[]).includes(check.id)
            )}
            busy={props.busy}
            onRun={props.onRunCheck}
          />
        </div>
        <div>
          <p className='mb-2 text-sm font-medium'>{t('Deep · protocol')}</p>
          <p className='text-muted-foreground mb-2 text-sm'>
            {t(
              'Protocol checks are skipped when the vendor has no matching API. That is incomplete protocol, not a broken supplier.'
            )}
          </p>
          <CheckTable
            checks={props.basicChecks.filter((check) => {
              if (check.id === 'kimi_kvv' && props.vendor !== 'kimi') {
                return false
              }
              return (PROTOCOL_BASIC_IDS as readonly string[]).includes(
                check.id
              )
            })}
            busy={props.busy}
            onRun={props.onRunCheck}
          />
          {props.vendor === 'kimi' && (
            <Card className='mt-3 space-y-2.5 border-amber-500/30 bg-amber-500/5 p-3.5'>
              <div className='flex items-center justify-between gap-2'>
                <div className='flex items-center gap-1.5'>
                  <ShieldCheck className='size-4 text-amber-500' />
                  <span className='text-xs font-semibold text-amber-700 dark:text-amber-400'>
                    {t('KVV preflight')}
                  </span>
                </div>
                <Badge
                  variant='outline'
                  className='border-amber-500/40 text-[10px] text-amber-600'
                >
                  {t('Not official certification')}
                </Badge>
              </div>
              <p className='text-muted-foreground text-[11px] leading-relaxed'>
                {t(
                  'Kimi K3-only preflight for request compatibility, tool_choice, response_format, and reasoning_content. A pass is not official Kimi KVV certification.'
                )}
              </p>
              <div className='grid gap-1.5 pt-1 text-[11px] sm:grid-cols-2'>
                <div className='border-muted bg-background/60 flex items-start gap-1.5 rounded border p-2'>
                  <span className='text-primary font-mono font-semibold'>
                    [1]
                  </span>
                  <div>
                    <div className='text-foreground font-medium'>
                      {t('K3 request protocol')}
                    </div>
                    <div className='text-muted-foreground text-[10px]'>
                      {t(
                        'reasoning_effort=low is accepted and the unsupported thinking field is never sent.'
                      )}
                    </div>
                  </div>
                </div>
                <div className='border-muted bg-background/60 flex items-start gap-1.5 rounded border p-2'>
                  <span className='text-primary font-mono font-semibold'>
                    [2]
                  </span>
                  <div>
                    <div className='text-foreground font-medium'>
                      {t('tool_choice contract')}
                    </div>
                    <div className='text-muted-foreground text-[10px]'>
                      {t(
                        'auto, none, and required are checked. none does not call a tool; required does.'
                      )}
                    </div>
                  </div>
                </div>
                <div className='border-muted bg-background/60 flex items-start gap-1.5 rounded border p-2'>
                  <span className='text-primary font-mono font-semibold'>
                    [3]
                  </span>
                  <div>
                    <div className='text-foreground font-medium'>
                      {t('response_format contract')}
                    </div>
                    <div className='text-muted-foreground text-[10px]'>
                      {t(
                        'text, json_object, and strict or non-strict json_schema responses are checked.'
                      )}
                    </div>
                  </div>
                </div>
                <div className='border-muted bg-background/60 flex items-start gap-1.5 rounded border p-2'>
                  <span className='text-primary font-mono font-semibold'>
                    [4]
                  </span>
                  <div>
                    <div className='text-foreground font-medium'>
                      {t('Thinking contract')}
                    </div>
                    <div className='text-muted-foreground text-[10px]'>
                      {t(
                        'K3 returns reasoning_content with reasoning_effort=low, including streamed chunks.'
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </Card>
          )}
        </div>
      </div>
    </TabsContent>
  )
}
