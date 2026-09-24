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

export type SupplierTestModule = 'basic' | 'stress' | 'cache' | 'video'

export type CheckStatus = 'idle' | 'running' | 'pass' | 'fail' | 'skip'

export type VideoMetrics = {
  task_id: string
  status: string
  video_url?: string
  endpoint_url?: string
  elapsed_ms: number
  fail_reason?: string
  raw_request_json?: string
  raw_submit_response_json?: string
  raw_poll_response_json?: string
}

export type StressMetrics = {
  total: number
  attempted?: number
  not_run?: number
  succeeded: number
  failed: number
  error_rate: number
  elapsed_ms: number
  tokens_per_sec: number
  request_tokens_per_sec?: number
  request_avg_ms?: number
  request_p50_ms?: number
  request_p90_ms?: number
  usage_n?: number
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
  issues?: StressIssue[]
  other_issue_count?: number
}

export type StressIssue = {
  status_code: number
  message: string
  count: number
  worker: number
  round: number
  elapsed_ms: number
}

export type CacheMode = 'static' | 'cumulative'

export type CacheMetrics = {
  warm_prompt_tokens: number
  avg_hit_rate: number
  min_hit_rate: number
  last_cached_tokens: number
  last_prompt_tokens: number
  wait_seconds: number
  rounds: number
  has_cached_tokens: boolean
  hit_count?: number
  avg_depth_rate?: number
  mode?: CacheMode
}

export type SupplierTestEvent = {
  type:
    | 'check'
    | 'progress'
    | 'stream'
    | 'metrics'
    | 'summary'
    | 'done'
    | 'error'
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
  video?: VideoMetrics
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
    mode?: CacheMode
  }
  stress: {
    concurrency: number
    rounds: number
    max_tokens: number
    prompt: string
    break_cache: boolean
    stream: boolean
  }
  video?: {
    prompt: string
    upload_mode?: string
    image_url?: string
    base64_data?: string
    role?: string
    last_frame_mode?: string
    last_frame_url?: string
    last_frame_base64?: string
    resolution?: string
    ratio?: string
    duration?: number
    watermark?: boolean
    seed?: number
    generate_audio?: boolean
    return_last_frame?: boolean
    custom_json?: string
    custom_path?: string
    raw_payload?: string
    task_id?: string
    checks?: string[]
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
  mode: CacheMode
}

export type VideoForm = {
  prompt: string
  hasImage: boolean
  uploadMode: 'url' | 'base64'
  imageUrl: string
  base64Data: string
  role: string
  hasLastFrame: boolean
  lastFrameMode: 'url' | 'base64'
  lastFrameUrl: string
  lastFrameBase64: string
  hasResolution: boolean
  resolution: string
  hasRatio: boolean
  ratio: string
  hasDuration: boolean
  duration: number
  hasWatermark: boolean
  watermark: boolean
  hasSeed: boolean
  seed: string
  hasGenerateAudio: boolean
  generateAudio: boolean
  hasReturnLastFrame: boolean
  returnLastFrame: boolean
  hasCustomJson: boolean
  customJson: string
  customPath: string
  rawPayload?: string
  taskId?: string
}
