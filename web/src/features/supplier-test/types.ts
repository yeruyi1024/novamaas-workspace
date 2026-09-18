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
import type { VendorId } from './vendors'

export type SupplierTestModule = 'basic' | 'stress' | 'cache'

export type CheckStatus = 'idle' | 'running' | 'pass' | 'fail' | 'skip'

export type StressMetrics = {
  total: number
  succeeded: number
  failed: number
  error_rate: number
  elapsed_ms: number
  tokens_per_sec: number
  prompt_tokens: number
  completion_tokens: number
  ttft_avg_ms: number
  ttft_p50_ms: number
  ttft_p90_ms: number
  ttft_n: number
  tpot_avg_ms: number
  tpot_p50_ms: number
  tpot_p90_ms: number
  tpot_n: number
  rpm: number
  tpm: number
}

export type CacheMetrics = {
  warm_prompt_tokens: number
  avg_hit_rate: number
  min_hit_rate: number
  last_cached_tokens: number
  last_prompt_tokens: number
  wait_seconds: number
  rounds: number
  has_cached_tokens: boolean
}

export type SupplierTestEvent = {
  type: 'check' | 'progress' | 'stream' | 'metrics' | 'summary' | 'done' | 'error'
  module?: SupplierTestModule
  check_id?: string
  status?: CheckStatus
  title?: string
  message?: string
  summary?: string
  worker?: number
  text?: string
  completed?: number
  total?: number
  metrics?: StressMetrics
  cache?: CacheMetrics
}

export type SupplierTestRunRequest = {
  base_url: string
  api_key: string
  model: string
  vendor?: VendorId
  modules: SupplierTestModule[]
  basic: {
    prompt: string
    max_tokens: number
    temperature?: number
    top_p?: number
    stream: boolean
    checks?: string[]
  }
  cache: {
    prompt: string
    follow_up: string
    wait_seconds: number
    max_tokens: number
    rounds: number
    stream: boolean
  }
  stress: {
    concurrency: number
    rounds: number
    max_tokens: number
    prompt: string
    break_cache: boolean
    stream: boolean
  }
}

export type CheckResult = {
  id: string
  title: string
  status: CheckStatus
  message?: string
  hintKey?: string
}

export type TargetForm = {
  baseUrl: string
  apiKey: string
  model: string
  vendor: VendorId
}

export type StressForm = {
  concurrency: number
  rounds: number
  maxTokens: number
  corpus: string
  prompt: string
  breakCache: boolean
  stream: boolean
}

export type BasicForm = {
  prompt: string
  maxTokens: number
  temperature: string
  topP: string
  stream: boolean
}

export type CacheForm = {
  corpus: string
  prompt: string
  followUp: string
  waitSeconds: number
  maxTokens: number
  rounds: number
  stream: boolean
}
