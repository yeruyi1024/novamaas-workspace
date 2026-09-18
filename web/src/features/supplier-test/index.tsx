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
import { useMutation } from '@tanstack/react-query'
import {
  ChevronDown,
  ClipboardCheck,
  Copy,
  Download,
  Loader2,
  SlidersHorizontal,
  Square,
} from 'lucide-react'
import { useEffect, useState, type Dispatch, type SetStateAction } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { PasswordInput } from '@/components/password-input'
import { StatusBadge, type StatusVariant } from '@/components/status-badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Progress } from '@/components/ui/progress'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { TitledCard } from '@/components/ui/titled-card'
import { cn } from '@/lib/utils'

import { fetchSupplierModels } from './api'
import {
  applyStandardEditorValue,
  assessCache,
  assessmentGroup,
  assessStress,
  DEFAULT_STANDARD,
  displayMeasured,
  displayThreshold,
  getStandard,
  matchingStandardId,
  overallLabel,
  sanitizeStandard,
  STANDARD_EDITOR_FIELDS,
  standardEditorValue,
  SUPPLIER_STANDARDS,
  VERDICT_LABEL,
  type Assessment,
  type SupplierStandard,
  type Verdict,
} from './baselines'
import {
  CACHE_ROUND_PRESETS,
  CACHE_WAIT_PRESETS,
  CORPORA,
  DEFAULT_BASIC_FORM,
  DEFAULT_CACHE_FORM,
  DEFAULT_STRESS_FORM,
  LOAD_PRESETS,
  MAX_CACHE_ROUNDS,
  MAX_CACHE_WAIT_SECONDS,
  MAX_CONCURRENCY,
  MAX_ROUNDS,
  MAX_TOKENS_CAP,
  PROTOCOL_BASIC_IDS,
  SHALLOW_BASIC_IDS,
  STRESS_WARN_TOTAL,
  estimateTokens,
  resolveCorpusPrompt,
  thisPlatformBaseURL,
} from './constants'
import { useSupplierTestRun } from './hooks/use-supplier-test-run'
import { vendorHintKey, VENDOR_OPTIONS, type VendorId } from './vendors'
import {
  buildHtmlReport,
  buildMarkdownReport,
  downloadFile,
  exportPdfReport,
  stampFileName,
  type ReportInput,
} from './report'
import type {
  BasicForm,
  CacheForm,
  CheckResult,
  CheckStatus,
  StressForm,
  SupplierTestModule,
  SupplierTestRunRequest,
  TargetForm,
} from './types'

function statusVariant(status: CheckStatus): StatusVariant {
  if (status === 'pass') return 'success'
  if (status === 'fail') return 'danger'
  if (status === 'skip') return 'warning'
  if (status === 'running') return 'info'
  return 'neutral'
}

function axiosErrorMessage(error: unknown, fallback: string): string {
  if (error && typeof error === 'object' && 'response' in error) {
    const message = (
      error as { response?: { data?: { message?: string } } }
    ).response?.data?.message
    if (message) return message
  }
  if (error instanceof Error && error.message) return error.message
  return fallback
}

function optionalNumber(raw: string): number | undefined {
  const trimmed = raw.trim()
  if (trimmed === '') return undefined
  const value = Number(trimmed)
  if (!Number.isFinite(value)) return undefined
  return value
}

const STANDARD_STORAGE_KEY = 'supplier-test-standard'

function readStoredStandard(): SupplierStandard {
  if (typeof sessionStorage === 'undefined') {
    return { ...DEFAULT_STANDARD }
  }
  try {
    const raw = sessionStorage.getItem(STANDARD_STORAGE_KEY)
    if (!raw) return { ...DEFAULT_STANDARD }
    return sanitizeStandard(JSON.parse(raw) as Partial<SupplierStandard>)
  } catch {
    return { ...DEFAULT_STANDARD }
  }
}

export function SupplierTest() {
  const { t } = useTranslation()
  const [target, setTarget] = useState<TargetForm>({
    baseUrl: '',
    apiKey: '',
    model: '',
    vendor: 'generic',
  })
  const [models, setModels] = useState<string[]>([])
  const [basic, setBasic] = useState<BasicForm>(DEFAULT_BASIC_FORM)
  const [cache, setCache] = useState<CacheForm>(DEFAULT_CACHE_FORM)
  const [stress, setStress] = useState<StressForm>(DEFAULT_STRESS_FORM)
  const [standard, setStandard] = useState<SupplierStandard>(readStoredStandard)
  const [moduleTab, setModuleTab] = useState<SupplierTestModule>('basic')
  const run = useSupplierTestRun()
  const matchedStandardId = matchingStandardId(standard)
  const standardLabel = t(
    matchedStandardId
      ? getStandard(matchedStandardId).labelKey
      : 'Custom standard'
  )

  useEffect(() => {
    try {
      sessionStorage.setItem(STANDARD_STORAGE_KEY, JSON.stringify(standard))
    } catch {
      // ignore quota or private-mode failures
    }
  }, [standard])

  useEffect(() => {
    if (run.runningModule) setModuleTab(run.runningModule)
  }, [run.runningModule])
  const stressTotal = stress.concurrency * stress.rounds
  const busy = run.runningModule !== null
  const hasReport =
    Boolean(run.summaries.basic || run.summaries.cache || run.summaries.stress) ||
    run.basicChecks.some((check) => check.status !== 'idle') ||
    run.cacheChecks.some((check) => check.status !== 'idle') ||
    run.metrics !== null
  const progressValue =
    run.progress.total > 0
      ? Math.min(100, (run.progress.completed / run.progress.total) * 100)
      : 0
  let runLabel = t('Run all basic checks')
  if (moduleTab === 'cache') {
    runLabel = t('Run cache test')
  } else if (moduleTab === 'stress') {
    runLabel = t('Run stress test')
  }

  const modelsMutation = useMutation({
    mutationFn: fetchSupplierModels,
    onSuccess: (ids) => {
      setModels(ids)
      if (ids.length > 0 && !target.model.trim()) {
        setTarget((current) => ({ ...current, model: ids[0] ?? '' }))
      }
      toast.success(t('Fetched {{count}} models', { count: ids.length }))
    },
    onError: (error) => {
      toast.error(axiosErrorMessage(error, t('Failed to fetch models')))
    },
  })

  const validateTarget = (): boolean => {
    if (!target.baseUrl.trim()) {
      toast.error(t('Enter a base URL first'))
      return false
    }
    if (!target.model.trim()) {
      toast.error(t('Enter a model ID first'))
      return false
    }
    return true
  }

  const buildPayload = (
    module: SupplierTestModule,
    checks?: string[]
  ): SupplierTestRunRequest => {
    const temperature = optionalNumber(basic.temperature)
    const topP = optionalNumber(basic.topP)
    return {
      base_url: target.baseUrl.trim(),
      api_key: target.apiKey,
      model: target.model.trim(),
      vendor: target.vendor,
      modules: [module],
      basic: {
        prompt: basic.prompt,
        max_tokens: basic.maxTokens,
        stream: basic.stream,
        ...(temperature === undefined ? {} : { temperature }),
        ...(topP === undefined ? {} : { top_p: topP }),
        ...(checks && checks.length > 0 ? { checks } : {}),
      },
      cache: {
        prompt: resolveCorpusPrompt(cache),
        follow_up: cache.followUp,
        wait_seconds: cache.waitSeconds,
        max_tokens: cache.maxTokens,
        rounds: cache.rounds,
        stream: cache.stream,
      },
      stress: {
        concurrency: stress.concurrency,
        rounds: stress.rounds,
        max_tokens: stress.maxTokens,
        prompt: resolveCorpusPrompt(stress),
        break_cache: stress.breakCache,
        stream: stress.stream,
      },
    }
  }

  const startModule = (module: SupplierTestModule, checks?: string[]) => {
    if (!validateTarget()) return
    if (module === 'basic') {
      if (
        !Number.isInteger(basic.maxTokens) ||
        basic.maxTokens < 1 ||
        basic.maxTokens > MAX_TOKENS_CAP
      ) {
        toast.error(
          t('Max tokens must be between 1 and {{max}}', { max: MAX_TOKENS_CAP })
        )
        return
      }
    }
    if (module === 'cache') {
      if (cache.waitSeconds < 0 || cache.waitSeconds > MAX_CACHE_WAIT_SECONDS) {
        toast.error(
          t('Cache wait must be between 0 and {{max}} seconds', {
            max: MAX_CACHE_WAIT_SECONDS,
          })
        )
        return
      }
      if (
        !Number.isInteger(cache.rounds) ||
        cache.rounds < 1 ||
        cache.rounds > MAX_CACHE_ROUNDS
      ) {
        toast.error(
          t('Cache rounds must be between 1 and {{max}}', {
            max: MAX_CACHE_ROUNDS,
          })
        )
        return
      }
      if (
        !Number.isInteger(cache.maxTokens) ||
        cache.maxTokens < 1 ||
        cache.maxTokens > MAX_TOKENS_CAP
      ) {
        toast.error(
          t('Cache max tokens must be between 1 and {{max}}', {
            max: MAX_TOKENS_CAP,
          })
        )
        return
      }
    }
    if (module === 'stress') {
      if (
        !Number.isInteger(stress.concurrency) ||
        stress.concurrency < 1 ||
        stress.concurrency > MAX_CONCURRENCY
      ) {
        toast.error(
          t('Concurrency must be between 1 and {{max}}', {
            max: MAX_CONCURRENCY,
          })
        )
        return
      }
      if (
        !Number.isInteger(stress.rounds) ||
        stress.rounds < 1 ||
        stress.rounds > MAX_ROUNDS
      ) {
        toast.error(t('Rounds must be between 1 and {{max}}', { max: MAX_ROUNDS }))
        return
      }
      if (
        !Number.isInteger(stress.maxTokens) ||
        stress.maxTokens < 1 ||
        stress.maxTokens > MAX_TOKENS_CAP
      ) {
        toast.error(
          t('Max tokens must be between 1 and {{max}}', { max: MAX_TOKENS_CAP })
        )
        return
      }
    }
    void run.start(buildPayload(module, checks))
  }

  const stressAssessment = run.metrics
    ? assessStress(run.metrics, standard)
    : null
  const cacheAssessment = run.cacheMetrics
    ? assessCache(run.cacheMetrics, standard)
    : null

  const reportInput = (): ReportInput => ({
    baseUrl: target.baseUrl.trim(),
    model: target.model.trim(),
    standardLabel,
    basicChecks: run.basicChecks,
    cacheChecks: run.cacheChecks,
    summaries: run.summaries,
    stressAssessment,
    cacheAssessment,
    errorMessage: run.errorMessage,
    statusLabel,
    t: (key, options) => t(key, options),
  })

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Supplier Test')}</SectionPageLayout.Title>
      <SectionPageLayout.Actions>
        {busy ? (
          <Button variant='outline' onClick={run.stop}>
            <Square />
            {t('Stop')}
          </Button>
        ) : null}
        <Button
          variant='outline'
          disabled={busy || !hasReport}
          onClick={async () => {
            try {
              await navigator.clipboard.writeText(
                buildMarkdownReport(reportInput())
              )
              toast.success(t('Report copied to clipboard'))
            } catch {
              toast.error(t('Failed to copy report'))
            }
          }}
        >
          <Copy />
          {t('Copy report')}
        </Button>
        <Button
          variant='outline'
          disabled={busy || !hasReport}
          onClick={() => {
            exportPdfReport(reportInput())
            toast.success(t('PDF report ready'))
          }}
        >
          <Download />
          {t('Export PDF')}
        </Button>
        <Button
          variant='outline'
          disabled={busy || !hasReport}
          onClick={() => {
            downloadFile(
              `supplier-test-${stampFileName()}.html`,
              buildHtmlReport(reportInput()),
              'text/html'
            )
            toast.success(t('HTML report exported'))
          }}
        >
          <Download />
          {t('Export HTML')}
        </Button>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <div className='space-y-4'>
          <TitledCard
            title={t('Target')}
            description={t(
              'Fill Base URL, key, and model yourself. Pick GLM, Kimi, or DeepSeek to apply that vendor field rules.'
            )}
            icon={<ClipboardCheck />}
            action={
              <div className='flex flex-wrap gap-2'>
                <Button
                  variant='outline'
                  disabled={busy}
                  onClick={() => {
                    const baseUrl = thisPlatformBaseURL()
                    if (!baseUrl) {
                      toast.error(t('Enter a base URL first'))
                      return
                    }
                    setTarget((current) => ({ ...current, baseUrl }))
                    toast.success(
                      t(
                        'Filled this platform Base URL. Paste a token from Keys, then fetch models.'
                      )
                    )
                  }}
                >
                  {t('Use this platform')}
                </Button>
                <Button
                  variant='outline'
                  onClick={() => {
                    if (!target.baseUrl.trim()) {
                      toast.error(t('Enter a base URL first'))
                      return
                    }
                    modelsMutation.mutate({
                      base_url: target.baseUrl.trim(),
                      api_key: target.apiKey,
                    })
                  }}
                  disabled={busy || modelsMutation.isPending}
                >
                  {modelsMutation.isPending ? (
                    <Loader2 className='animate-spin' />
                  ) : null}
                  {t('Fetch models')}
                </Button>
              </div>
            }
          >
            <div className='grid gap-3 md:grid-cols-2 xl:grid-cols-4'>
              <FieldSelect
                label={t('Vendor')}
                value={target.vendor}
                disabled={busy}
                items={VENDOR_OPTIONS.map((item) => ({
                  value: item.id,
                  label: t(item.labelKey),
                }))}
                onChange={(value) => {
                  const vendor = value as VendorId
                  setTarget((current) => ({ ...current, vendor }))
                }}
              />
              <div className='space-y-2'>
                <Label htmlFor='supplier-base-url'>{t('Base URL')}</Label>
                <Input
                  id='supplier-base-url'
                  placeholder='https://api.example.com'
                  value={target.baseUrl}
                  disabled={busy}
                  onChange={(event) =>
                    setTarget((current) => ({
                      ...current,
                      baseUrl: event.target.value,
                    }))
                  }
                />
              </div>
              <div className='space-y-2'>
                <Label htmlFor='supplier-api-key'>{t('API Key')}</Label>
                <PasswordInput
                  id='supplier-api-key'
                  placeholder={t('Supplier API key')}
                  value={target.apiKey}
                  disabled={busy}
                  onChange={(event) =>
                    setTarget((current) => ({
                      ...current,
                      apiKey: event.target.value,
                    }))
                  }
                />
              </div>
              <div className='space-y-2'>
                <Label htmlFor='supplier-model'>{t('Model')}</Label>
                <Input
                  id='supplier-model'
                  list='supplier-model-options'
                  placeholder={t('Type a model ID')}
                  value={target.model}
                  disabled={busy}
                  onChange={(event) =>
                    setTarget((current) => ({
                      ...current,
                      model: event.target.value,
                    }))
                  }
                />
                <datalist id='supplier-model-options'>
                  {models.map((id) => (
                    <option key={id} value={id} />
                  ))}
                </datalist>
              </div>
            </div>
            <p className='text-muted-foreground mt-3 text-sm'>
              {t(vendorHintKey(target.vendor))}
            </p>
          </TitledCard>

          <Tabs
            value={moduleTab}
            onValueChange={(value) => {
              const next = String(value)
              if (next === 'basic' || next === 'cache' || next === 'stress') {
                setModuleTab(next)
              }
            }}
          >
            <TitledCard
              title={t('Tests')}
              description={t(
                'Run one module at a time. Results stay when you switch tabs.'
              )}
              icon={<ClipboardCheck />}
              action={
                <Button
                  onClick={() => startModule(moduleTab)}
                  disabled={busy}
                >
                  {run.runningModule === moduleTab ? (
                    <Loader2 className='animate-spin' />
                  ) : null}
                  {runLabel}
                </Button>
              }
            >
              <TabsList className='mb-4 grid h-auto w-full grid-cols-3 sm:w-fit'>
                <TabsTrigger value='basic'>{t('Basic acceptance')}</TabsTrigger>
                <TabsTrigger value='cache'>{t('Cache test')}</TabsTrigger>
                <TabsTrigger value='stress'>{t('Stress test')}</TabsTrigger>
              </TabsList>
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
                value={basic.maxTokens}
                disabled={busy}
                min={1}
                max={MAX_TOKENS_CAP}
                presets={[
                  { label: '64', value: 64 },
                  { label: '256', value: 256 },
                  { label: '1k', value: 1024 },
                  { label: '4k', value: 4096 },
                ]}
                onChange={(value) =>
                  setBasic((current) => ({ ...current, maxTokens: value }))
                }
              />
              <OptionalNumberField
                id='basic-temperature'
                label={t('Temperature')}
                value={basic.temperature}
                disabled={busy}
                min={0}
                max={2}
                step={0.1}
                placeholder={t('Leave empty to omit')}
                onChange={(value) =>
                  setBasic((current) => ({ ...current, temperature: value }))
                }
              />
              <OptionalNumberField
                id='basic-top-p'
                label={t('Top P')}
                value={basic.topP}
                disabled={busy}
                min={0}
                max={1}
                step={0.05}
                placeholder={t('Leave empty to omit')}
                onChange={(value) =>
                  setBasic((current) => ({ ...current, topP: value }))
                }
              />
              <StreamSwitch
                id='basic-stream'
                checked={basic.stream}
                disabled={busy}
                onChange={(checked) =>
                  setBasic((current) => ({ ...current, stream: checked }))
                }
              />
            </div>
            <div className='mt-4 space-y-2'>
              <Label htmlFor='basic-prompt'>{t('Prompt')}</Label>
              <Textarea
                id='basic-prompt'
                rows={3}
                value={basic.prompt}
                disabled={busy}
                onChange={(event) =>
                  setBasic((current) => ({
                    ...current,
                    prompt: event.target.value,
                  }))
                }
              />
            </div>
            {run.basicStreamText ? (
              <pre className='bg-muted mt-4 max-h-48 overflow-auto rounded-lg p-3 text-sm whitespace-pre-wrap'>
                {run.basicStreamText}
              </pre>
            ) : null}
            {run.summaries.basic ? (
              <p className='text-muted-foreground mt-4 text-sm'>
                {run.summaries.basic}
              </p>
            ) : null}
            <div className='space-y-4'>
              <div>
                <p className='mb-2 text-sm font-medium'>
                  {t('Shallow · connectivity')}
                </p>
                <CheckTable
                  checks={run.basicChecks.filter((check) =>
                    (SHALLOW_BASIC_IDS as readonly string[]).includes(check.id)
                  )}
                  busy={busy}
                  onRun={(id) => startModule('basic', [id])}
                />
              </div>
              <div>
                <p className='mb-2 text-sm font-medium'>
                  {t('Deep · protocol')}
                </p>
                <p className='text-muted-foreground mb-2 text-sm'>
                  {t(
                    'Protocol checks are skipped when the vendor has no matching API. That is incomplete protocol, not a broken supplier.'
                  )}
                </p>
                <CheckTable
                  checks={run.basicChecks.filter((check) =>
                    (PROTOCOL_BASIC_IDS as readonly string[]).includes(check.id)
                  )}
                  busy={busy}
                  onRun={(id) => startModule('basic', [id])}
                />
              </div>
            </div>
              </TabsContent>
              <TabsContent value='cache' className='space-y-4'>
            <p className='text-muted-foreground text-sm'>
              {t(
                'Corpus is the input prefix, sent as-is. Max tokens only caps the reply. The first request warms cache; later rounds check the hit.'
              )}
            </p>
            <div className='grid gap-4 md:grid-cols-4'>
              <NumberField
                id='cache-wait'
                label={t('Wait seconds')}
                value={cache.waitSeconds}
                disabled={busy}
                min={0}
                max={MAX_CACHE_WAIT_SECONDS}
                presets={CACHE_WAIT_PRESETS.map((s) => ({
                  label: `${s}s`,
                  value: s,
                }))}
                onChange={(value) =>
                  setCache((current) => ({ ...current, waitSeconds: value }))
                }
              />
              <NumberField
                id='cache-rounds'
                label={t('Probe rounds')}
                value={cache.rounds}
                disabled={busy}
                min={1}
                max={MAX_CACHE_ROUNDS}
                presets={CACHE_ROUND_PRESETS.map((r) => ({
                  label: String(r),
                  value: r,
                }))}
                onChange={(value) =>
                  setCache((current) => ({
                    ...current,
                    rounds: value,
                  }))
                }
              />
              <NumberField
                id='cache-max-tokens'
                label={t('Max tokens')}
                value={cache.maxTokens}
                disabled={busy}
                min={1}
                max={MAX_TOKENS_CAP}
                presets={[
                  { label: '16', value: 16 },
                  { label: '64', value: 64 },
                  { label: '256', value: 256 },
                ]}
                onChange={(value) =>
                  setCache((current) => ({ ...current, maxTokens: value }))
                }
              />
              <StreamSwitch
                id='cache-stream'
                checked={cache.stream}
                disabled={busy}
                onChange={(checked) =>
                  setCache((current) => ({ ...current, stream: checked }))
                }
              />
            </div>
            <div className='mt-4'>
              <CorpusPicker
                id='cache-prompt'
                form={cache}
                disabled={busy}
                onChange={(next) =>
                  setCache((current) => ({
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
                value={cache.followUp}
                disabled={busy}
                onChange={(event) =>
                  setCache((current) => ({
                    ...current,
                    followUp: event.target.value,
                  }))
                }
              />
            </div>
            {run.summaries.cache ? (
              <p className='text-muted-foreground mt-4 text-sm'>
                {run.summaries.cache}
              </p>
            ) : null}
            <div className='mt-4'>
              <CheckTable checks={run.cacheChecks} busy={busy} />
            </div>
            {cacheAssessment && cacheAssessment.rows.length > 0 ? (
              <AssessmentTable
                title={t('Deep · performance')}
                assessment={cacheAssessment}
              />
            ) : null}
              </TabsContent>
              <TabsContent value='stress' className='space-y-4'>
            <p className='text-muted-foreground text-sm'>
              {t(
                'Corpus is sent as-is. Allow cache reuses it; break cache puts a random prefix in front of each request. Max tokens only caps the reply.'
              )}
            </p>
            <div className='flex flex-wrap gap-2'>
              {LOAD_PRESETS.map((preset) => (
                <Button
                  key={preset.id}
                  type='button'
                  variant='outline'
                  size='sm'
                  disabled={busy}
                  onClick={() =>
                    setStress((current) => ({
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
                variant={stress.breakCache ? 'outline' : 'secondary'}
                size='sm'
                disabled={busy}
                onClick={() =>
                  setStress((current) => ({ ...current, breakCache: false }))
                }
              >
                {t('Allow prompt cache')}
              </Button>
              <Button
                type='button'
                variant={stress.breakCache ? 'secondary' : 'outline'}
                size='sm'
                disabled={busy}
                onClick={() =>
                  setStress((current) => ({ ...current, breakCache: true }))
                }
              >
                {t('Break prompt cache')}
              </Button>
            </div>
            <div className='mt-4 grid gap-4 md:grid-cols-4'>
              <NumberField
                id='stress-concurrency'
                label={t('Concurrency')}
                value={stress.concurrency}
                disabled={busy}
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
                  setStress((current) => ({ ...current, concurrency: value }))
                }
              />
              <NumberField
                id='stress-rounds'
                label={t('Rounds')}
                value={stress.rounds}
                disabled={busy}
                min={1}
                max={MAX_ROUNDS}
                presets={[
                  { label: '1', value: 1 },
                  { label: '2', value: 2 },
                  { label: '5', value: 5 },
                  { label: '10', value: 10 },
                ]}
                onChange={(value) =>
                  setStress((current) => ({ ...current, rounds: value }))
                }
              />
              <NumberField
                id='stress-max-tokens'
                label={t('Max tokens')}
                value={stress.maxTokens}
                disabled={busy}
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
                  setStress((current) => ({ ...current, maxTokens: value }))
                }
              />
              <StreamSwitch
                id='stress-stream'
                checked={stress.stream}
                disabled={busy}
                onChange={(checked) =>
                  setStress((current) => ({ ...current, stream: checked }))
                }
              />
            </div>
            <div className='mt-4'>
              <CorpusPicker
                id='stress-prompt'
                form={stress}
                disabled={busy}
                onChange={(next) =>
                  setStress((current) => ({
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
            {run.runningModule === 'stress' && run.progress.total > 0 ? (
              <div className='mt-4 space-y-2'>
                <div className='text-muted-foreground text-sm'>
                  {t('Progress')}: {run.progress.completed}/{run.progress.total}
                </div>
                <Progress value={progressValue} />
              </div>
            ) : null}
            {run.runningModule === 'stress' && run.streamText ? (
              <pre className='bg-muted mt-4 max-h-64 overflow-auto rounded-lg p-3 text-sm whitespace-pre-wrap'>
                {run.streamText}
              </pre>
            ) : null}
            {run.summaries.stress ? (
              <p className='text-muted-foreground mt-4 text-sm'>
                {run.summaries.stress}
              </p>
            ) : null}
            {stressAssessment ? (
              <>
                <AssessmentTable
                  title={t('Shallow · connectivity')}
                  assessment={assessmentGroup(stressAssessment, 'shallow')}
                />
                <AssessmentTable
                  title={t('Deep · performance')}
                  assessment={assessmentGroup(stressAssessment, 'perf')}
                />
              </>
            ) : null}
              </TabsContent>
            </TitledCard>
          </Tabs>

          <JudgmentStandardCard
            standard={standard}
            matchedStandardId={matchedStandardId}
            onChange={setStandard}
          />

          {run.errorMessage ? (
            <Alert variant='destructive'>
              <AlertDescription>{run.errorMessage}</AlertDescription>
            </Alert>
          ) : null}
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}

function CorpusPicker(props: {
  id: string
  form: { corpus: string; prompt: string }
  disabled: boolean
  onChange: (next: { corpus: string; prompt: string }) => void
}) {
  const { t } = useTranslation()
  return (
    <div className='space-y-4'>
      <FieldSelect
        label={t('Corpus')}
        value={props.form.corpus}
        disabled={props.disabled}
        items={CORPORA.map((item) => ({
          value: item.id,
          label:
            item.tokens > 0
              ? t('{{label}} ({{tokens}} tokens)', {
                  label: t(item.labelKey),
                  tokens: item.tokens,
                })
              : t(item.labelKey),
        }))}
        onChange={(corpus) =>
          props.onChange({ corpus, prompt: props.form.prompt })
        }
      />
      {props.form.corpus === 'custom' ? (
        <div className='space-y-2'>
          <Label htmlFor={props.id}>{t('Prompt')}</Label>
          <Textarea
            id={props.id}
            rows={3}
            placeholder={t('Write a short prompt, or pick a built-in corpus')}
            value={props.form.prompt}
            disabled={props.disabled}
            onChange={(event) =>
              props.onChange({
                corpus: props.form.corpus,
                prompt: event.target.value,
              })
            }
          />
          <p className='text-muted-foreground text-sm'>
            {t('{{tokens}} tokens', {
              tokens: estimateTokens(props.form.prompt),
            })}
          </p>
        </div>
      ) : (
        <p className='text-muted-foreground text-sm'>
          {t(
            'Using built-in corpus ({{tokens}} tokens, {{chars}} characters). Replace files in supplier-test/corpora to change the text.',
            {
              tokens: estimateTokens(resolveCorpusPrompt(props.form)),
              chars: resolveCorpusPrompt(props.form).length,
            }
          )}
        </p>
      )}
    </div>
  )
}

function FieldSelect(props: {
  label: string
  value: string
  disabled: boolean
  items: Array<{ value: string; label: string }>
  onChange: (value: string) => void
}) {
  return (
    <div className='space-y-2'>
      <Label>{props.label}</Label>
      <Select
        items={props.items}
        value={props.value}
        disabled={props.disabled}
        onValueChange={(value) => {
          if (value) props.onChange(value)
        }}
      >
        <SelectTrigger className='w-full'>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {props.items.map((item) => (
            <SelectItem key={item.value} value={item.value}>
              {item.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}

function StreamSwitch(props: {
  id: string
  checked: boolean
  disabled: boolean
  onChange: (checked: boolean) => void
}) {
  const { t } = useTranslation()
  return (
    <div className='space-y-1.5'>
      <Label htmlFor={props.id}>{t('Stream')}</Label>
      <div className='flex h-8 items-center'>
        <Switch
          id={props.id}
          checked={props.checked}
          disabled={props.disabled}
          onCheckedChange={(checked) => props.onChange(Boolean(checked))}
        />
      </div>
    </div>
  )
}

function JudgmentStandardCard(props: {
  standard: SupplierStandard
  matchedStandardId: string | null
  onChange: Dispatch<SetStateAction<SupplierStandard>>
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <div className='bg-card rounded-xl border'>
        <CollapsibleTrigger
          render={
            <button
              type='button'
              className='hover:bg-muted/40 flex w-full items-center justify-between gap-3 rounded-xl px-4 py-3 text-left transition-colors'
            />
          }
        >
          <span className='flex min-w-0 items-center gap-3'>
            <span className='bg-muted text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-md'>
              <SlidersHorizontal className='size-4' aria-hidden='true' />
            </span>
            <span className='min-w-0'>
              <span className='block text-sm font-medium'>
                {t('Judgment standard')}
              </span>
              <span className='text-muted-foreground block text-xs'>
                {t(
                  'These numbers are the ruler after a run. They do not judge whether the model is smart. Change them for this vendor; tables update immediately.'
                )}
              </span>
            </span>
          </span>
          <span className='text-muted-foreground flex shrink-0 items-center gap-1 text-sm'>
            {open ? t('Collapse') : t('Expand')}
            <ChevronDown
              className={cn(
                'size-4 transition-transform',
                open && 'rotate-180'
              )}
              aria-hidden='true'
            />
          </span>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <div className='space-y-5 border-t px-4 py-4'>
          <div className='max-w-sm'>
            <FieldSelect
              label={t('Load preset')}
              value={props.matchedStandardId ?? 'custom'}
              disabled={false}
              items={[
                ...SUPPLIER_STANDARDS.map((item) => ({
                  value: item.id,
                  label: t(item.labelKey),
                })),
                ...(props.matchedStandardId
                  ? []
                  : [
                      {
                        value: 'custom',
                        label: t('Custom standard'),
                      },
                    ]),
              ]}
              onChange={(id) => {
                if (id === 'custom') return
                props.onChange({ ...getStandard(id) })
              }}
            />
          </div>
          <div className='mt-4 space-y-5'>
            {(
              [
                {
                  group: 'stress' as const,
                  title: t('From stress test'),
                  note: t(
                    'Error rate is the share of failed HTTP/API requests in that short run (timeouts, 5xx, refused). It is not wrong model answers.'
                  ),
                },
                {
                  group: 'cache' as const,
                  title: t('From cache test'),
                  note: t(
                    'These rulers score Cache test only. They do not use stress numbers.'
                  ),
                },
              ]
            ).map((section) => (
              <div key={section.group}>
                <p className='mb-1 text-sm font-medium'>{section.title}</p>
                <p className='text-muted-foreground mb-3 text-xs'>
                  {section.note}
                </p>
                <div className='grid gap-4 md:grid-cols-2 lg:grid-cols-3'>
                  {STANDARD_EDITOR_FIELDS.filter(
                    (field) => field.group === section.group
                  ).map((field) => (
                    <NumberField
                      key={field.id}
                      id={field.id}
                      label={t(field.label)}
                      hint={t(field.hint)}
                      value={standardEditorValue(props.standard, field)}
                      disabled={false}
                      min={field.min}
                      max={field.max}
                      step={field.step}
                      onChange={(value) =>
                        props.onChange((current) =>
                          applyStandardEditorValue(current, field, value)
                        )
                      }
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
          </div>
        </CollapsibleContent>
      </div>
    </Collapsible>
  )
}

function NumberField(props: {
  id: string
  label: string
  hint?: string
  value: number
  min: number
  max: number
  step?: number
  disabled: boolean
  presets?: Array<{ label: string; value: number }>
  onChange: (value: number) => void
}) {
  return (
    <div className='space-y-1.5'>
      <Label htmlFor={props.id}>{props.label}</Label>
      <Input
        id={props.id}
        type='number'
        min={props.min}
        max={props.max}
        step={props.step}
        value={props.value}
        disabled={props.disabled}
        onChange={(event) => {
          const raw = event.target.value
          const next =
            props.step && props.step < 1
              ? Number.parseFloat(raw)
              : Number.parseInt(raw, 10)
          props.onChange(Number.isNaN(next) ? 0 : next)
        }}
      />
      {props.hint ? (
        <p className='text-muted-foreground text-xs leading-snug'>{props.hint}</p>
      ) : null}
      {props.presets && props.presets.length > 0 ? (
        <div className='flex flex-wrap gap-1 pt-0.5'>
          {props.presets.map((preset) => {
            const active = props.value === preset.value
            return (
              <Button
                key={preset.label}
                type='button'
                variant={active ? 'secondary' : 'outline'}
                size='xs'
                disabled={props.disabled}
                className='h-5.5 px-1.5 text-[11px] font-normal'
                onClick={() => props.onChange(preset.value)}
              >
                {preset.label}
              </Button>
            )
          })}
        </div>
      ) : null}
    </div>
  )
}

function OptionalNumberField(props: {
  id: string
  label: string
  value: string
  min: number
  max: number
  step?: number
  disabled: boolean
  placeholder?: string
  onChange: (value: string) => void
}) {
  return (
    <div className='space-y-2'>
      <Label htmlFor={props.id}>{props.label}</Label>
      <Input
        id={props.id}
        type='number'
        min={props.min}
        max={props.max}
        step={props.step}
        value={props.value}
        disabled={props.disabled}
        placeholder={props.placeholder}
        onChange={(event) => props.onChange(event.target.value)}
      />
    </div>
  )
}

function CheckTable(props: {
  checks: CheckResult[]
  busy: boolean
  onRun?: (id: string) => void
}) {
  const { t } = useTranslation()
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{t('Check')}</TableHead>
          <TableHead>{t('Result')}</TableHead>
          <TableHead>{t('Detail')}</TableHead>
          {props.onRun ? <TableHead className='w-24'>{t('Action')}</TableHead> : null}
        </TableRow>
      </TableHeader>
      <TableBody>
        {props.checks.map((check) => (
          <TableRow key={check.id}>
            <TableCell>
              <div>{t(check.title)}</div>
              {check.hintKey ? (
                <p className='text-muted-foreground mt-0.5 max-w-prose text-xs leading-snug'>
                  {t(check.hintKey)}
                </p>
              ) : null}
            </TableCell>
            <TableCell>
              <StatusBadge
                variant={statusVariant(check.status)}
                copyable={false}
                label={t(statusLabel(check.status))}
              />
            </TableCell>
            <TableCell className='text-muted-foreground max-w-xl whitespace-pre-wrap'>
              {check.message ?? ''}
            </TableCell>
            {props.onRun ? (
              <TableCell>
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  disabled={props.busy}
                  onClick={() => props.onRun?.(check.id)}
                >
                  {t('Run this check')}
                </Button>
              </TableCell>
            ) : null}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

function statusLabel(status: CheckStatus): string {
  if (status === 'pass') return 'Passed'
  if (status === 'fail') return 'Failed'
  if (status === 'skip') return 'Skipped'
  if (status === 'running') return 'Running'
  return 'Idle'
}

function verdictClass(verdict: Verdict): string {
  if (verdict === 'ok') return 'text-success font-medium'
  if (verdict === 'slow') return 'text-warning font-medium'
  if (verdict === 'abnormal') return 'text-destructive font-medium'
  return 'text-muted-foreground'
}

function AssessmentTable(props: { assessment: Assessment; title?: string }) {
  const { t } = useTranslation()
  if (props.assessment.rows.length === 0) return null
  return (
    <div className='mt-4 space-y-2 overflow-x-auto'>
      {props.title ? (
        <p className='text-sm font-medium'>{props.title}</p>
      ) : null}
      <div className={verdictClass(props.assessment.overall)}>
        {t(overallLabel(props.assessment.overall))}
      </div>
      <Table className='min-w-[640px]'>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Metric')}</TableHead>
            <TableHead>{t('Measured')}</TableHead>
            <TableHead>{t('Threshold')}</TableHead>
            <TableHead>{t('Verdict')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {props.assessment.rows.map((row) => (
            <TableRow key={row.id}>
              <TableCell className='whitespace-nowrap'>{t(row.label)}</TableCell>
              <TableCell className='font-medium whitespace-nowrap'>
                {displayMeasured(row, t)}
              </TableCell>
              <TableCell className='text-muted-foreground'>
                {displayThreshold(row, t)}
              </TableCell>
              <TableCell className={verdictClass(row.verdict)}>
                {t(VERDICT_LABEL[row.verdict])}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
