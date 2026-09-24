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
import { ClipboardCheck, Copy, Download, Loader2, Square } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { TitledCard } from '@/components/ui/titled-card'
import { copyToClipboard } from '@/lib/copy-to-clipboard'

import { fetchSupplierModels } from './api'
import {
  assessCache,
  assessStress,
  DEFAULT_STANDARD,
  getStandard,
  matchingStandardId,
  sanitizeStandard,
  type SupplierStandard,
} from './baselines'
import { BasicPanel } from './components/basic-panel'
import { CachePanel } from './components/cache-panel'
import { JudgmentStandardCard } from './components/judgment-standard-card'
import { StressPanel } from './components/stress-panel'
import { TargetCard } from './components/target-card'
import { VideoPanel } from './components/video-panel'
import {
  DEFAULT_BASIC_FORM,
  DEFAULT_CACHE_FORM,
  DEFAULT_STRESS_FORM,
  DEFAULT_VIDEO_FORM,
  MAX_CACHE_ROUNDS,
  MAX_CACHE_WAIT_SECONDS,
  MAX_CONCURRENCY,
  MAX_ROUNDS,
  MAX_STRESS_REQUESTS,
  MAX_TOKENS_CAP,
  PROTOCOL_BASIC_IDS,
  resolveCorpusPrompt,
  SHALLOW_BASIC_IDS,
  thisPlatformBaseURL,
} from './constants'
import { statusLabel } from './formatters'
import { useSupplierTestRun } from './hooks/use-supplier-test-run'
import {
  buildHtmlReport,
  buildMarkdownReport,
  buildVideoHtmlReport,
  buildVideoMarkdownReport,
  downloadFile,
  exportPdfReport,
  exportVideoPdfReport,
  stampFileName,
  type ReportInput,
  type VideoReportInput,
} from './report'
import type {
  BasicForm,
  CacheForm,
  StressForm,
  SupplierTestModule,
  SupplierTestRunRequest,
  TargetForm,
  VideoForm,
} from './types'
import { buildVideoRequestPayload } from './video-json'

export { AssessmentTable } from './components/assessment-table'
export { BasicPanel } from './components/basic-panel'
export { CachePanel } from './components/cache-panel'
export { CheckTable } from './components/check-table'
export { JudgmentStandardCard } from './components/judgment-standard-card'
export { RawJsonDialog } from './components/raw-json-dialog'
export { StressPanel } from './components/stress-panel'
export { TargetCard } from './components/target-card'
export { VideoPanel } from './components/video-panel'

function axiosErrorMessage(error: unknown, fallback: string): string {
  if (error && typeof error === 'object' && 'response' in error) {
    const message = (error as { response?: { data?: { message?: string } } })
      .response?.data?.message
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
  const [video, setVideo] = useState<VideoForm>(DEFAULT_VIDEO_FORM)
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

  const busy = run.runningModule !== null
  const isVideoTab = moduleTab === 'video'
  const hasTextReport =
    Boolean(
      run.summaries.basic || run.summaries.cache || run.summaries.stress
    ) ||
    (run.basicChecks ?? []).some((check) => check.status !== 'idle') ||
    (run.cacheChecks ?? []).some((check) => check.status !== 'idle') ||
    run.metrics !== null ||
    run.cacheMetrics !== null
  const hasVideoReport =
    (run.videoChecks ?? []).some((check) => check.status !== 'idle') ||
    run.videoMetrics !== null
  const hasReport = isVideoTab ? hasVideoReport : hasTextReport

  const progressValue =
    run.progress.total > 0
      ? Math.min(100, (run.progress.completed / run.progress.total) * 100)
      : 0

  let runLabel = t('Run all basic checks')
  if (moduleTab === 'cache') {
    runLabel = t('Run cache test')
  } else if (moduleTab === 'stress') {
    runLabel = t('Run stress test')
  } else if (moduleTab === 'video') {
    runLabel = t('Start Video Test')
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

  const validateTarget = () => {
    if (!target.baseUrl.trim()) {
      toast.error(t('Enter a base URL first'))
      return false
    }
    if (!target.model.trim()) {
      toast.error(t('Select a model first'))
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
    const defaultBasicChecks =
      target.vendor === 'kimi'
        ? [...SHALLOW_BASIC_IDS, ...PROTOCOL_BASIC_IDS]
        : [
            ...SHALLOW_BASIC_IDS,
            ...PROTOCOL_BASIC_IDS.filter((id) => id !== 'kimi_kvv'),
          ]
    const resolvedBasicChecks =
      module === 'basic' && checks && checks.length > 0
        ? checks
        : defaultBasicChecks
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
        checks: resolvedBasicChecks,
      },
      cache: {
        prompt: resolveCorpusPrompt(cache),
        follow_up: cache.followUp,
        wait_seconds: cache.waitSeconds,
        max_tokens: cache.maxTokens,
        rounds: cache.rounds,
        stream: cache.stream,
        mode: cache.mode,
      },
      stress: {
        concurrency: stress.concurrency,
        rounds: stress.rounds,
        max_tokens: stress.maxTokens,
        prompt: resolveCorpusPrompt(stress),
        break_cache: stress.breakCache,
        stream: stress.stream,
      },
      video: {
        ...(module === 'video' && checks && checks.length > 0
          ? { checks }
          : {}),
        ...(video.taskId || run.videoMetrics?.task_id
          ? { task_id: (video.taskId || run.videoMetrics?.task_id)?.trim() }
          : {}),
        ...(() => {
          const edited = video.rawPayload?.trim() ?? ''
          if (edited) return { raw_payload: edited }
          if (module !== 'video') return {}
          const preview = buildVideoRequestPayload(target.model, video)
          return preview ? { raw_payload: JSON.stringify(preview) } : {}
        })(),
        prompt: video.prompt.trim(),
        ...(video.hasImage
          ? {
              upload_mode: video.uploadMode,
              ...(video.uploadMode === 'url'
                ? { image_url: video.imageUrl.trim() }
                : {}),
              ...(video.uploadMode === 'base64'
                ? { base64_data: video.base64Data.trim() }
                : {}),
              role: video.role,
            }
          : {}),
        ...(video.hasLastFrame
          ? {
              last_frame_mode: video.lastFrameMode,
              ...(video.lastFrameMode === 'url'
                ? { last_frame_url: video.lastFrameUrl.trim() }
                : {}),
              ...(video.lastFrameMode === 'base64'
                ? { last_frame_base64: video.lastFrameBase64.trim() }
                : {}),
            }
          : {}),
        ...(video.hasResolution && video.resolution
          ? { resolution: video.resolution }
          : {}),
        ...(video.hasRatio && video.ratio ? { ratio: video.ratio } : {}),
        ...(video.hasDuration && video.duration > 0
          ? { duration: video.duration }
          : {}),
        ...(video.hasWatermark ? { watermark: video.watermark } : {}),
        ...(video.hasSeed && video.seed.trim() !== ''
          ? { seed: Number.parseInt(video.seed, 10) }
          : {}),
        ...(video.hasGenerateAudio
          ? { generate_audio: video.generateAudio }
          : {}),
        ...(video.hasReturnLastFrame
          ? { return_last_frame: video.returnLastFrame }
          : {}),
        ...(video.hasCustomJson && video.customJson.trim()
          ? { custom_json: video.customJson.trim() }
          : {}),
        ...(video.customPath.trim()
          ? { custom_path: video.customPath.trim() }
          : {}),
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
        toast.error(
          t('Rounds must be between 1 and {{max}}', { max: MAX_ROUNDS })
        )
        return
      }
      if (stress.concurrency * stress.rounds > MAX_STRESS_REQUESTS) {
        toast.error(
          t('Concurrency × rounds must be at most {{max}} requests', {
            max: MAX_STRESS_REQUESTS,
          })
        )
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
    if (module === 'video') {
      const isQueryOnly =
        checks &&
        (checks.includes('video_poll') || checks.includes('video_result')) &&
        !checks.includes('video_submit')
      if (isQueryOnly) {
        const currentTaskId =
          video.taskId?.trim() || run.videoMetrics?.task_id?.trim()
        if (!currentTaskId) {
          toast.error(t('Please submit task first or provide a Task ID'))
          return
        }
      } else if (video.rawPayload && video.rawPayload.trim()) {
        try {
          JSON.parse(video.rawPayload.trim())
        } catch {
          toast.error(t('Invalid JSON format'))
          return
        }
      } else {
        if (!video.prompt.trim()) {
          toast.error(t('Please enter a prompt for video generation'))
          return
        }
        if (video.hasImage) {
          if (video.uploadMode === 'url' && !video.imageUrl.trim()) {
            toast.error(t('Please enter an image URL'))
            return
          }
          if (video.uploadMode === 'base64' && !video.base64Data.trim()) {
            toast.error(t('Please select an image file or provide Base64 data'))
            return
          }
        }
        if (video.hasLastFrame) {
          if (video.lastFrameMode === 'url' && !video.lastFrameUrl.trim()) {
            toast.error(t('Please enter an end frame image URL'))
            return
          }
          if (
            video.lastFrameMode === 'base64' &&
            !video.lastFrameBase64.trim()
          ) {
            toast.error(
              t('Please select an end frame image file or provide Base64 data')
            )
            return
          }
        }
        if (video.hasCustomJson && video.customJson.trim()) {
          try {
            JSON.parse(video.customJson)
          } catch {
            toast.error(t('Custom extra parameters must be valid JSON'))
            return
          }
        }
      }
    }
    void run.start(buildPayload(module, checks))
  }

  const handleSendRawJson = (rawJson: string) => {
    if (!validateTarget()) return
    const payload = buildPayload('video')
    if (!payload.video) {
      payload.video = { prompt: DEFAULT_VIDEO_FORM.prompt }
    }
    payload.video.raw_payload = rawJson.trim()
    void run.start(payload)
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
    vendor: target.vendor,
    standardLabel,
    basicChecks:
      target.vendor === 'kimi'
        ? run.basicChecks
        : run.basicChecks.filter((check) => check.id !== 'kimi_kvv'),
    cacheChecks: run.cacheChecks,
    summaries: run.summaries,
    stressAssessment,
    cacheAssessment,
    errorMessage: run.errorMessage,
    statusLabel,
    stressConfig: {
      concurrency: stress.concurrency,
      rounds: stress.rounds,
      stream: stress.stream,
      breakCache: stress.breakCache,
      maxTokens: stress.maxTokens,
      corpus: stress.corpus,
    },
    stressMetrics: run.metrics,
    cacheMetrics: run.cacheMetrics,
    t: (key, options) => t(key, options),
  })

  const videoReportInput = (): VideoReportInput => ({
    baseUrl: target.baseUrl.trim(),
    model: target.model.trim(),
    endpointUrl: run.videoMetrics?.endpoint_url,
    taskId: run.videoMetrics?.task_id || video.taskId?.trim() || undefined,
    status: run.videoMetrics?.status,
    elapsedMs: run.videoMetrics?.elapsed_ms ?? 0,
    failReason: run.videoMetrics?.fail_reason,
    videoUrl: run.videoMetrics?.video_url,
    videoConfig: {
      prompt: video.prompt.trim(),
      hasImage: video.hasImage,
      uploadMode: video.uploadMode,
      imageUrl: video.imageUrl.trim(),
      role: video.role,
      hasLastFrame: video.hasLastFrame,
      lastFrameMode: video.lastFrameMode,
      lastFrameUrl: video.lastFrameUrl.trim(),
      resolution: video.hasResolution ? video.resolution : undefined,
      ratio: video.hasRatio ? video.ratio : undefined,
      duration: video.hasDuration ? video.duration : undefined,
      watermark: video.hasWatermark ? video.watermark : undefined,
      seed: video.hasSeed ? video.seed.trim() : undefined,
      generateAudio: video.hasGenerateAudio ? video.generateAudio : undefined,
      returnLastFrame: video.hasReturnLastFrame
        ? video.returnLastFrame
        : undefined,
      customJson: video.hasCustomJson ? video.customJson.trim() : undefined,
      customPath: video.customPath.trim() || undefined,
    },
    videoChecks: run.videoChecks,
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
            const report = isVideoTab
              ? buildVideoMarkdownReport(videoReportInput())
              : buildMarkdownReport(reportInput())
            const ok = await copyToClipboard(report)
            if (!ok) {
              toast.error(t('Failed to copy report'))
              return
            }
            toast.success(
              isVideoTab
                ? t('Video report copied to clipboard')
                : t('Report copied to clipboard')
            )
          }}
        >
          <Copy />
          {isVideoTab ? t('Copy Video Report') : t('Copy report')}
        </Button>
        <Button
          variant='outline'
          disabled={busy || !hasReport}
          onClick={() => {
            if (isVideoTab) {
              exportVideoPdfReport(videoReportInput())
              toast.success(t('Video PDF report ready'))
            } else {
              exportPdfReport(reportInput())
              toast.success(t('PDF report ready'))
            }
          }}
        >
          <Download />
          {isVideoTab ? t('Export Video PDF') : t('Export PDF')}
        </Button>
        <Button
          variant='outline'
          disabled={busy || !hasReport}
          onClick={() => {
            if (isVideoTab) {
              downloadFile(
                `doubao-video-test-${stampFileName()}.html`,
                buildVideoHtmlReport(videoReportInput()),
                'text/html'
              )
              toast.success(t('Video HTML report exported'))
            } else {
              downloadFile(
                `supplier-test-${stampFileName()}.html`,
                buildHtmlReport(reportInput()),
                'text/html'
              )
              toast.success(t('HTML report exported'))
            }
          }}
        >
          <Download />
          {isVideoTab ? t('Export Video HTML') : t('Export HTML')}
        </Button>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <div className='space-y-4'>
          <TargetCard
            target={target}
            busy={busy}
            models={models}
            isFetchingModels={modelsMutation.isPending}
            onTargetChange={setTarget}
            onFetchModels={() => {
              if (!target.baseUrl.trim()) {
                toast.error(t('Enter a base URL first'))
                return
              }
              modelsMutation.mutate({
                base_url: target.baseUrl.trim(),
                api_key: target.apiKey,
              })
            }}
            onUseThisPlatform={() => {
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
          />

          <Tabs
            value={moduleTab}
            onValueChange={(value) => {
              const next = String(value)
              if (
                next === 'basic' ||
                next === 'cache' ||
                next === 'stress' ||
                next === 'video'
              ) {
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
                <Button onClick={() => startModule(moduleTab)} disabled={busy}>
                  {run.runningModule === moduleTab ? (
                    <Loader2 className='animate-spin' />
                  ) : null}
                  {runLabel}
                </Button>
              }
            >
              <TabsList className='mb-4 grid h-auto w-full grid-cols-2 sm:w-fit sm:grid-cols-4'>
                <TabsTrigger value='basic'>{t('Basic acceptance')}</TabsTrigger>
                <TabsTrigger value='cache'>{t('Cache test')}</TabsTrigger>
                <TabsTrigger value='stress'>{t('Stress test')}</TabsTrigger>
                <TabsTrigger value='video'>{t('Doubao Video')}</TabsTrigger>
              </TabsList>

              <BasicPanel
                basic={basic}
                busy={busy}
                basicChecks={run.basicChecks}
                basicStreamText={run.basicStreamText}
                basicSummary={run.summaries.basic}
                vendor={target.vendor}
                onBasicChange={setBasic}
                onRunCheck={(id) => startModule('basic', [id])}
              />

              <CachePanel
                cache={cache}
                busy={busy}
                cacheChecks={run.cacheChecks}
                cacheSummary={run.summaries.cache}
                cacheAssessment={cacheAssessment}
                onCacheChange={setCache}
              />

              <StressPanel
                stress={stress}
                busy={busy}
                runningModule={run.runningModule}
                progress={run.progress}
                progressValue={progressValue}
                streamText={run.streamText}
                stressSummary={run.summaries.stress}
                stressAssessment={stressAssessment}
                metrics={run.metrics}
                onStressChange={setStress}
              />

              <VideoPanel
                video={video}
                busy={busy}
                model={target.model}
                baseUrl={target.baseUrl}
                apiKey={target.apiKey}
                videoChecks={run.videoChecks}
                videoMetrics={run.videoMetrics}
                onVideoChange={setVideo}
                onModelChange={(m) => setTarget((c) => ({ ...c, model: m }))}
                onSendRawJson={handleSendRawJson}
                onRunCheck={(checkId) => {
                  startModule('video', checkId ? [checkId] : undefined)
                }}
                onManualQueryResult={run.updateVideoMetricsWithQueryResult}
                onExportPdf={() => {
                  exportVideoPdfReport(videoReportInput())
                  toast.success(t('Video PDF report ready'))
                }}
              />
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
