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
import type { BasicForm, CacheForm, CheckResult, StressForm } from './types'
import longText from './corpora/long.txt?raw'
import mediumText from './corpora/medium.txt?raw'
import minQpsText from './corpora/min-qps.txt?raw'
import shortText from './corpora/short.txt?raw'
import threeKText from './corpora/3k.txt?raw'
import veryLongText from './corpora/very-long.txt?raw'

export const API_ENDPOINTS = {
  MODELS: '/api/supplier-test/models',
  RUNS: '/api/supplier-test/runs',
} as const

export function estimateTokens(text: string): number {
  const raw = text.trim()
  if (!raw) return 0
  let cjk = 0
  let other = 0
  for (const ch of raw) {
    const code = ch.codePointAt(0) ?? 0
    if (
      (code >= 0x4e00 && code <= 0x9fff) ||
      (code >= 0x3400 && code <= 0x4dbf) ||
      (code >= 0x3040 && code <= 0x30ff)
    ) {
      cjk += 1
    } else if (ch.trim() !== '') {
      other += 1
    }
  }
  return Math.max(1, Math.round(cjk + other / 4))
}

export const CORPORA: Array<{
  id: string
  labelKey: string
  prompt: string
  tokens: number
}> = [
  { id: 'min-qps', labelKey: 'Min QPS', prompt: minQpsText.trim() },
  { id: 'short', labelKey: 'Short text', prompt: shortText.trim() },
  { id: 'medium', labelKey: 'Medium text', prompt: mediumText.trim() },
  { id: '3k', labelKey: '3k text', prompt: threeKText.trim() },
  { id: 'long', labelKey: 'Long text', prompt: longText.trim() },
  {
    id: 'very-long',
    labelKey: 'Very long text',
    prompt: veryLongText.trim(),
  },
  { id: 'custom', labelKey: 'Custom prompt', prompt: '' },
].map((item) => ({ ...item, tokens: estimateTokens(item.prompt) }))

export function thisPlatformBaseURL(): string {
  const fromEnv = String(
    import.meta.env.VITE_REACT_APP_SERVER_URL ?? ''
  )
    .trim()
    .replace(/\/$/, '')
  if (fromEnv) return fromEnv
  if (typeof window !== 'undefined' && window.location.origin) {
    return window.location.origin
  }
  return ''
}

export function resolveCorpusPrompt(form: {
  corpus: string
  prompt: string
}): string {
  if (form.corpus === 'custom') {
    const custom = form.prompt.trim()
    if (custom) return custom
  }
  const corpus = CORPORA.find((item) => item.id === form.corpus)
  if (corpus?.prompt) return corpus.prompt
  return shortText.trim()
}

export const DEFAULT_STRESS_FORM: StressForm = {
  concurrency: 10,
  rounds: 1,
  maxTokens: 256,
  corpus: 'short',
  prompt: '',
  breakCache: false,
  stream: true,
}

export const LOAD_PRESETS: Array<{
  id: string
  labelKey: string
  concurrency: number
  rounds: number
  maxTokens: number
}> = [
  { id: 'light', labelKey: 'Light load', concurrency: 1, rounds: 1, maxTokens: 8 },
  { id: 'standard', labelKey: 'Standard load', concurrency: 10, rounds: 1, maxTokens: 64 },
  { id: 'heavy', labelKey: 'Heavy load', concurrency: 20, rounds: 2, maxTokens: 256 },
]

export const CACHE_WAIT_PRESETS = [0, 5, 30, 60]
export const CACHE_ROUND_PRESETS = [1, 3, 5, 10, 20]

export const DEFAULT_BASIC_PROMPT = '请用一句话介绍你自己。'

export const DEFAULT_BASIC_FORM: BasicForm = {
  prompt: DEFAULT_BASIC_PROMPT,
  maxTokens: 64,
  temperature: '',
  topP: '',
  stream: true,
}

export const DEFAULT_CACHE_FOLLOW_UP =
  '根据前面的内容，只用一个词回复：pong'

export const DEFAULT_CACHE_FORM: CacheForm = {
  corpus: '3k',
  prompt: '',
  followUp: DEFAULT_CACHE_FOLLOW_UP,
  waitSeconds: 30,
  maxTokens: 16,
  rounds: 5,
  stream: true,
}

export const BASIC_CHECKS: CheckResult[] = [
  {
    id: 'connectivity',
    title: 'Connectivity',
    status: 'idle',
    hintKey:
      'First chat request. Pass means HTTP 200 with text, thinking, or a tool call.',
  },
  {
    id: 'stream_format',
    title: 'Stream format',
    status: 'idle',
    hintKey:
      'Streaming frames are valid. Missing finish_reason is skipped, not failed.',
  },
  {
    id: 'usage',
    title: 'Usage fields',
    status: 'idle',
    hintKey:
      'Response includes prompt_tokens and completion_tokens. Needed for billing, not intelligence.',
  },
  {
    id: 'request_id',
    title: 'Request id',
    status: 'idle',
    hintKey:
      'Response body or common headers include a request id, for troubleshooting.',
  },
  {
    id: 'sampling',
    title: 'Sampling parameters',
    status: 'idle',
    hintKey:
      'Only checks max_tokens and the temperature / top_p you filled. Vendor 4xx is skipped.',
  },
  {
    id: 'auth_error',
    title: 'Auth error',
    status: 'idle',
    hintKey: 'A bad key should return 401 or 403. Other codes are skipped.',
  },
  {
    id: 'bad_request',
    title: 'Bad request',
    status: 'idle',
    hintKey:
      'A request missing required fields should return 4xx. Other codes are skipped.',
  },
  {
    id: 'json_mode',
    title: 'JSON mode',
    status: 'idle',
    hintKey:
      'Asks the vendor to return JSON. No JSON mode is skipped, not failed.',
  },
  {
    id: 'tool_call',
    title: 'Tool call',
    status: 'idle',
    hintKey:
      'Asks the vendor to call a tool. No tools API is skipped, not failed.',
  },
  {
    id: 'thinking',
    title: 'Thinking mode',
    status: 'idle',
    hintKey:
      'Asks the vendor for thinking / reasoning. No thinking API is skipped, not failed.',
  },
]

export const SHALLOW_BASIC_IDS = [
  'connectivity',
  'stream_format',
  'usage',
  'request_id',
  'sampling',
  'auth_error',
  'bad_request',
] as const

export const PROTOCOL_BASIC_IDS = [
  'json_mode',
  'tool_call',
  'thinking',
] as const

export const CACHE_CHECKS: CheckResult[] = [
  {
    id: 'cache_warm',
    title: 'Cache warm',
    status: 'idle',
    hintKey:
      'First request with the same corpus prefix, to fill the vendor cache.',
  },
  {
    id: 'cache_probe',
    title: 'Cache probe',
    status: 'idle',
    hintKey:
      'Later rounds reuse that prefix plus a follow-up, to see if cache hits.',
  },
  {
    id: 'cache_tokens',
    title: 'Cached tokens',
    status: 'idle',
    hintKey:
      'Whether usage.cached_tokens is present. Missing field is skipped, not failed.',
  },
  {
    id: 'cache_hit_rate',
    title: 'Cache hit rate',
    status: 'idle',
    hintKey:
      'Hits divided by probe rounds. Judged by the Cache hit (%) ruler above.',
  },
  {
    id: 'cache_ttl',
    title: 'Cache TTL',
    status: 'idle',
    hintKey:
      'Hit rate after waiting. Wait less than TTL wait (s) and this cannot be compared.',
  },
]

export const MAX_CONCURRENCY = 1000
export const MAX_ROUNDS = 10000
export const MAX_TOKENS_CAP = 256000
export const MAX_CACHE_ROUNDS = 50
export const MAX_CACHE_WAIT_SECONDS = 600
export const STRESS_WARN_TOTAL = 50
