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
import { useCallback, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { SSE } from 'sse.js'

import { getFreshAuthHeaders } from '@/lib/api'

import type { QueryVideoTaskResult } from '../api'
import {
  API_ENDPOINTS,
  BASIC_CHECKS,
  CACHE_CHECKS,
  VIDEO_CHECKS,
} from '../constants'
import type {
  CacheMetrics,
  CheckResult,
  StressMetrics,
  SupplierTestEvent,
  SupplierTestModule,
  SupplierTestRunRequest,
  VideoMetrics,
} from '../types'

type StreamHandle = {
  close: () => void
  stream: () => void
  xhr?: XMLHttpRequest | null
  readyState?: number
  addEventListener: (
    type: string,
    listener: (event: Event & { data?: string; responseCode?: number }) => void
  ) => void
}

function errorMessageFromStream(
  event: Event & { data?: string },
  source: StreamHandle,
  fallback: string
): string {
  const chunks = [event.data, source.xhr?.responseText]
  for (const chunk of chunks) {
    const text = chunk?.trim()
    if (!text) continue
    try {
      const parsed = JSON.parse(text) as { message?: string; type?: string }
      if (parsed.message) return parsed.message
    } catch {
      const line = text.split('\n').find((entry) => entry.startsWith('data:'))
      if (line) {
        try {
          const parsed = JSON.parse(line.slice(5).trim()) as {
            message?: string
          }
          if (parsed.message) return parsed.message
        } catch {
          // keep looking
        }
      }
    }
  }
  return fallback
}

function resetChecks(current: CheckResult[], ids?: string[]): CheckResult[] {
  const selected = ids && ids.length > 0 ? new Set(ids) : null
  return current.map((check) => {
    if (selected && !selected.has(check.id)) {
      return check
    }
    return { ...check, status: 'idle', message: undefined }
  })
}

export function useSupplierTestRun() {
  const { t } = useTranslation()
  const sourceRef = useRef<StreamHandle | null>(null)
  const runIdRef = useRef(0)
  const [runningModule, setRunningModule] = useState<SupplierTestModule | null>(
    null
  )
  const [basicChecks, setBasicChecks] = useState<CheckResult[]>(BASIC_CHECKS)
  const [cacheChecks, setCacheChecks] = useState<CheckResult[]>(CACHE_CHECKS)
  const [videoChecks, setVideoChecks] = useState<CheckResult[]>(VIDEO_CHECKS)
  const [streamText, setStreamText] = useState('')
  const [basicStreamText, setBasicStreamText] = useState('')
  const [progress, setProgress] = useState({ completed: 0, total: 0 })
  const [metrics, setMetrics] = useState<StressMetrics | null>(null)
  const [cacheMetrics, setCacheMetrics] = useState<CacheMetrics | null>(null)
  const [videoMetrics, setVideoMetrics] = useState<VideoMetrics | null>(null)
  const [summaries, setSummaries] = useState({
    basic: '',
    cache: '',
    stress: '',
  })
  const [errorMessage, setErrorMessage] = useState('')

  const stop = useCallback(() => {
    runIdRef.current += 1
    const source = sourceRef.current
    sourceRef.current = null
    source?.close()
    setRunningModule(null)
  }, [])

  const start = useCallback(
    async (payload: SupplierTestRunRequest) => {
      const runId = ++runIdRef.current
      const previousSource = sourceRef.current
      sourceRef.current = null
      previousSource?.close()
      const module = payload.modules[0] ?? null
      setRunningModule(module)
      setErrorMessage('')
      setProgress({ completed: 0, total: 0 })
      if (module === 'basic') {
        setBasicChecks((current) => resetChecks(current, payload.basic.checks))
        setBasicStreamText('')
        setSummaries((current) => ({ ...current, basic: '' }))
      }
      if (module === 'cache') {
        setCacheChecks(
          CACHE_CHECKS.map((check) => ({ ...check, status: 'idle' }))
        )
        setCacheMetrics(null)
        setSummaries((current) => ({ ...current, cache: '' }))
      }
      if (module === 'stress') {
        setStreamText('')
        setMetrics(null)
        setSummaries((current) => ({ ...current, stress: '' }))
      }
      if (module === 'video') {
        const videoChecksToRun = payload.video?.checks
        setVideoChecks((current) => resetChecks(current, videoChecksToRun))
        if (!videoChecksToRun || videoChecksToRun.includes('video_submit')) {
          setVideoMetrics(null)
        }
      }

      let headers: Record<string, string>
      try {
        headers = await getFreshAuthHeaders()
      } catch (error) {
        if (runIdRef.current !== runId) return
        setRunningModule(null)
        toast.error(
          error instanceof Error ? error.message : t('Failed to start test')
        )
        return
      }
      if (runIdRef.current !== runId) return

      const source = new SSE(API_ENDPOINTS.RUNS, {
        headers: {
          ...headers,
          'Content-Type': 'application/json',
        },
        method: 'POST',
        payload: JSON.stringify(payload),
        withCredentials: true,
        start: false,
      }) as StreamHandle
      sourceRef.current = source
      let completed = false

      const finish = () => {
        if (runIdRef.current !== runId || sourceRef.current !== source) return
        setRunningModule(null)
        sourceRef.current = null
        source.close()
      }

      const handleError = (message: string) => {
        setErrorMessage(message)
        finish()
        toast.error(message)
      }

      source.addEventListener('message', (event) => {
        if (runIdRef.current !== runId || sourceRef.current !== source) return
        const data = event.data ?? ''
        if (data.trim() === '[DONE]') {
          completed = true
          finish()
          return
        }
        let parsed: SupplierTestEvent
        try {
          parsed = JSON.parse(data) as SupplierTestEvent
        } catch {
          handleError(t('Failed to parse stream'))
          return
        }
        if (parsed.video) {
          setVideoMetrics((prev) => {
            if (!prev) return parsed.video ?? null
            return {
              ...prev,
              ...parsed.video,
              raw_request_json:
                parsed.video?.raw_request_json || prev.raw_request_json,
              raw_submit_response_json:
                parsed.video?.raw_submit_response_json ||
                prev.raw_submit_response_json,
              endpoint_url: parsed.video?.endpoint_url || prev.endpoint_url,
            }
          })
        }
        if (parsed.type === 'error') {
          handleError(parsed.message || t('Failed to start test'))
          return
        }
        if (parsed.type === 'check' && parsed.check_id) {
          const applyCheck = (current: CheckResult[]) =>
            current.map((check) =>
              check.id === parsed.check_id
                ? {
                    ...check,
                    title: parsed.title || check.title,
                    status: parsed.status ?? check.status,
                    message: parsed.message,
                  }
                : check
            )
          if (parsed.module === 'video') {
            setVideoChecks(applyCheck)
          } else if (parsed.module === 'cache') {
            setCacheChecks(applyCheck)
          } else if (parsed.module !== 'stress') {
            setBasicChecks(applyCheck)
          }
          return
        }
        if (parsed.type === 'stream' && parsed.text) {
          if (parsed.module === 'basic') {
            setBasicStreamText((text) => text + parsed.text)
            return
          }
          if (parsed.module === 'stress') {
            setStreamText((text) => text + parsed.text)
          }
          return
        }
        if (parsed.type === 'progress') {
          if (parsed.check_id) {
            const applyCheck = (current: CheckResult[]) =>
              current.map((check) =>
                check.id === parsed.check_id
                  ? {
                      ...check,
                      title: parsed.title || check.title,
                      status: parsed.status ?? check.status,
                      message: parsed.message,
                    }
                  : check
              )
            if (parsed.module === 'video') {
              setVideoChecks(applyCheck)
            } else if (parsed.module === 'cache') {
              setCacheChecks(applyCheck)
            } else if (parsed.module !== 'stress') {
              setBasicChecks(applyCheck)
            }
          }
          setProgress({
            completed: parsed.completed ?? 0,
            total: parsed.total ?? 0,
          })
          return
        }
        if (parsed.type === 'metrics' && parsed.cache) {
          setCacheMetrics(parsed.cache)
          if (!parsed.metrics) {
            return
          }
        }
        if (parsed.type === 'metrics' && parsed.metrics) {
          setMetrics(parsed.metrics)
          setProgress({
            completed: parsed.completed ?? parsed.metrics.total,
            total: parsed.total ?? parsed.metrics.total,
          })
          return
        }
        if (parsed.type === 'summary' && parsed.summary) {
          const next = parsed.summary
          if (parsed.module === 'cache') {
            setSummaries((current) => ({ ...current, cache: next }))
          } else if (parsed.module === 'stress') {
            setSummaries((current) => ({ ...current, stress: next }))
          } else {
            setSummaries((current) => ({ ...current, basic: next }))
          }
          return
        }
        if (parsed.type === 'done') {
          completed = true
          setRunningModule(null)
        }
      })

      source.addEventListener('abort', () => {
        if (runIdRef.current !== runId || sourceRef.current !== source) return
        if (completed) finish()
        else handleError(t('Test stream ended unexpectedly'))
      })

      source.addEventListener('error', (event) => {
        if (runIdRef.current !== runId || sourceRef.current !== source) return
        if (completed) {
          finish()
          return
        }
        const code = event.responseCode ?? source.xhr?.status ?? 0
        if (code >= 200 && code < 300) {
          handleError(t('Test stream ended unexpectedly'))
          return
        }
        handleError(
          errorMessageFromStream(event, source, t('Failed to start test'))
        )
      })

      source.stream()
    },
    [t]
  )

  const updateVideoMetricsWithQueryResult = useCallback(
    (result: QueryVideoTaskResult) => {
      setVideoMetrics((prev) => ({
        ...prev,
        task_id: result.task_id || prev?.task_id || '',
        status: result.status || prev?.status || '',
        elapsed_ms: prev?.elapsed_ms ?? 0,
        video_url: result.video_url || prev?.video_url,
        fail_reason: result.fail_reason || prev?.fail_reason,
        raw_poll_response_json:
          result.raw_response || prev?.raw_poll_response_json,
      }))
      if (result.status) {
        const normalized = result.status.toLowerCase()
        const isSuccess = normalized === 'succeeded' || normalized === 'success'
        const isFailed = ['failed', 'failure', 'cancelled', 'expired'].includes(
          normalized
        )

        let pollStatus: CheckResult['status'] = 'running'
        let pollMessage = t('Current status: {{status}}', {
          status: result.status,
        })
        if (isSuccess) {
          pollStatus = 'pass'
          pollMessage = t('Manual query: task completed')
        } else if (isFailed) {
          pollStatus = 'fail'
          pollMessage = result.fail_reason || t('Task failed')
        }

        setVideoChecks((current) =>
          current.map((check) => {
            if (check.id === 'video_poll') {
              return {
                ...check,
                status: pollStatus,
                message: pollMessage,
              }
            }
            if (check.id === 'video_result') {
              if (isSuccess) {
                return {
                  ...check,
                  status: result.video_url ? 'pass' : 'fail',
                  message: result.video_url
                    ? t('Successfully fetched video playback URL')
                    : t('Task succeeded but no video URL returned'),
                }
              }
              if (isFailed) {
                return {
                  ...check,
                  status: 'fail',
                  message: result.fail_reason || t('Task failed'),
                }
              }
            }
            return check
          })
        )
      }
    },
    [t]
  )

  return {
    runningModule,
    basicChecks,
    cacheChecks,
    videoChecks,
    streamText,
    basicStreamText,
    progress,
    metrics,
    cacheMetrics,
    videoMetrics,
    summaries,
    errorMessage,
    start,
    stop,
    updateVideoMetricsWithQueryResult,
  }
}
