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
import { SSE } from 'sse.js'
import { toast } from 'sonner'

import { getFreshAuthHeaders } from '@/lib/api'

import { API_ENDPOINTS, BASIC_CHECKS, CACHE_CHECKS } from '../constants'
import type {
  CacheMetrics,
  CheckResult,
  StressMetrics,
  SupplierTestEvent,
  SupplierTestModule,
  SupplierTestRunRequest,
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
      const line = text
        .split('\n')
        .find((entry) => entry.startsWith('data:'))
      if (line) {
        try {
          const parsed = JSON.parse(line.slice(5).trim()) as { message?: string }
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
  const [runningModule, setRunningModule] = useState<SupplierTestModule | null>(
    null
  )
  const [basicChecks, setBasicChecks] = useState<CheckResult[]>(BASIC_CHECKS)
  const [cacheChecks, setCacheChecks] = useState<CheckResult[]>(CACHE_CHECKS)
  const [streamText, setStreamText] = useState('')
  const [basicStreamText, setBasicStreamText] = useState('')
  const [progress, setProgress] = useState({ completed: 0, total: 0 })
  const [metrics, setMetrics] = useState<StressMetrics | null>(null)
  const [cacheMetrics, setCacheMetrics] = useState<CacheMetrics | null>(null)
  const [summaries, setSummaries] = useState({
    basic: '',
    cache: '',
    stress: '',
  })
  const [errorMessage, setErrorMessage] = useState('')

  const stop = useCallback(() => {
    sourceRef.current?.close()
    sourceRef.current = null
    setRunningModule(null)
  }, [])

  const start = useCallback(
    async (payload: SupplierTestRunRequest) => {
      sourceRef.current?.close()
      sourceRef.current = null
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
        setCacheChecks(CACHE_CHECKS.map((check) => ({ ...check, status: 'idle' })))
        setCacheMetrics(null)
        setSummaries((current) => ({ ...current, cache: '' }))
      }
      if (module === 'stress') {
        setStreamText('')
        setMetrics(null)
        setSummaries((current) => ({ ...current, stress: '' }))
      }

      let headers: Record<string, string>
      try {
        headers = await getFreshAuthHeaders()
      } catch (error) {
        setRunningModule(null)
        toast.error(
          error instanceof Error ? error.message : t('Failed to start test')
        )
        return
      }

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

      const finish = () => {
        setRunningModule(null)
        source.close()
        if (sourceRef.current === source) {
          sourceRef.current = null
        }
      }

      const handleError = (message: string) => {
        setErrorMessage(message)
        finish()
        toast.error(message)
      }

      source.addEventListener('message', (event) => {
        const data = event.data ?? ''
        if (data.trim() === '[DONE]') {
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
          if (parsed.module === 'cache') {
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
          setRunningModule(null)
        }
      })

      source.addEventListener('abort', () => {
        if (sourceRef.current !== source) return
        finish()
      })

      source.addEventListener('error', (event) => {
        if (sourceRef.current !== source) return
        if (source.readyState === 2) {
          finish()
          return
        }
        const code = event.responseCode ?? source.xhr?.status ?? 0
        if (code >= 200 && code < 300) {
          finish()
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

  return {
    runningModule,
    basicChecks,
    cacheChecks,
    streamText,
    basicStreamText,
    progress,
    metrics,
    cacheMetrics,
    summaries,
    errorMessage,
    start,
    stop,
  }
}
