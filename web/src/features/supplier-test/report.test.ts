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
import { describe, expect, test } from 'vitest'

import {
  assessCache,
  assessStress,
  getStandard,
  isInformationalRow,
} from './baselines'
import {
  buildHtmlReport,
  buildMarkdownReport,
  buildVideoHtmlReport,
  buildVideoMarkdownReport,
  formatTokenCompact,
  type ReportInput,
  type VideoReportInput,
} from './report'
import type { CacheMetrics, CheckResult, StressMetrics } from './types'

function makeTestInput(): ReportInput {
  const basicChecks: CheckResult[] = [
    { id: 'ping', title: 'Connectivity', status: 'pass' },
    { id: 'stream', title: 'Streaming', status: 'pass' },
    { id: 'json_mode', title: 'JSON mode', status: 'pass' },
  ]
  const stressMetrics: StressMetrics = {
    total: 100,
    succeeded: 100,
    failed: 0,
    error_rate: 0,
    elapsed_ms: 40383,
    tokens_per_sec: 935.0,
    prompt_tokens: 137500,
    completion_tokens: 37760,
    ttft_avg_ms: 3208,
    ttft_p50_ms: 2431,
    ttft_p90_ms: 3236,
    ttft_n: 100,
    tpot_avg_ms: 15.6,
    tpot_p50_ms: 14.4,
    tpot_p90_ms: 22.7,
    tpot_n: 100,
    rpm: 148.6,
    tpm: 260394,
  }
  const cacheMetrics: CacheMetrics = {
    warm_prompt_tokens: 3000,
    avg_hit_rate: 0.989,
    min_hit_rate: 0.95,
    last_cached_tokens: 2967,
    last_prompt_tokens: 3000,
    wait_seconds: 30,
    rounds: 5,
    has_cached_tokens: true,
    hit_count: 5,
    avg_depth_rate: 0.989,
    mode: 'static',
  }
  const std = getStandard('default')

  return {
    baseUrl: 'https://maas.ai.shilijia.xyz/v1',
    model: 'glm-5.2',
    standardLabel: 'Default standard',
    basicChecks,
    cacheChecks: [{ id: 'cache_probe', title: 'Cache probe', status: 'pass' }],
    summaries: {
      basic: 'Passed 9, skipped 1, failed 0',
      cache: 'Cache test completed successfully',
      stress: 'Stress test completed: 100 requests',
    },
    stressAssessment: assessStress(stressMetrics, std),
    cacheAssessment: assessCache(cacheMetrics, std),
    errorMessage: '',
    statusLabel: (status) => status,
    stressConfig: {
      concurrency: 50,
      rounds: 2,
      stream: true,
      breakCache: false,
      maxTokens: 1024,
    },
    stressMetrics,
    cacheMetrics,
    t: (key) => key,
  }
}

function makeVideoTestInput(): VideoReportInput {
  return {
    baseUrl: 'https://ark.cn-beijing.volces.com',
    model: 'doubao-seedance-1-0-pro',
    endpointUrl:
      'https://ark.cn-beijing.volces.com/api/v3/content/generation/tasks',
    taskId: 'cgt-20260921-test-12345',
    status: 'succeeded',
    elapsedMs: 25300,
    videoUrl: 'https://tos.volces.com/test-video.mp4',
    videoConfig: {
      prompt: 'A futuristic city with flying cars at sunset',
      hasImage: true,
      uploadMode: 'url',
      imageUrl: 'https://example.com/first-frame.png',
      role: 'first_frame',
      hasLastFrame: false,
      resolution: '720p',
      ratio: '16:9',
      duration: 5,
      watermark: false,
      generateAudio: true,
      returnLastFrame: false,
    },
    videoChecks: [
      { id: 'create_task', title: 'Task creation', status: 'pass' },
      { id: 'poll_task', title: 'Task execution polling', status: 'pass' },
    ],
    t: (key) => key,
  }
}

describe('supplier-test report', () => {
  test('formatTokenCompact formats large and small numbers', () => {
    expect(formatTokenCompact(0)).toBe('0')
    expect(formatTokenCompact(800)).toBe('800')
    expect(formatTokenCompact(1500)).toBe('1.5k')
    expect(formatTokenCompact(37760)).toBe('37.8k')
    expect(formatTokenCompact(137500)).toBe('137.5k')
    expect(formatTokenCompact(2500000)).toBe('2.5M')
  })

  test('buildHtmlReport suppresses browser header/footers with zero page margin', () => {
    const input = makeTestInput()
    const html = buildHtmlReport(input)
    expect(html).toContain('@page {')
    expect(html).toContain('margin: 0;')
    expect(html).toContain('print-color-adjust: exact')
  })

  test('buildHtmlReport includes concurrency config and execution telemetry', () => {
    const input = makeTestInput()
    const html = buildHtmlReport(input)

    // Concurrency configuration
    expect(html).toContain('50 × 2 (100)')
    expect(html).toContain('max_tokens = 1024')
    expect(html).toContain('Stream')

    // Execution metrics
    expect(html).toContain('40.38 s')
    expect(html).toContain('935.0 tok/s')
    expect(html).toContain('137.5k / 37.8k')

    // Sections and overall verdicts
    expect(html).toContain('Concurrency and stress test')
    expect(html).toContain('Prompt cache test')
    expect(html).toContain('Stress test assessment')
    expect(html).toContain('Prompt cache assessment')

    // Benchmark tables must NOT contain informational rows like duration or tokens
    const benchmarkRows = input.stressAssessment?.rows.filter(
      (r) => !isInformationalRow(r)
    )
    expect(benchmarkRows).toBeDefined()
    expect(benchmarkRows?.length).toBeGreaterThan(0)
    for (const r of benchmarkRows ?? []) {
      expect(html).toContain(r.label)
    }
  })

  test('reports missing upstream token usage without showing a zero token rate', () => {
    const input = makeTestInput()
    if (!input.stressMetrics) throw new Error('missing stress fixture')
    input.stressMetrics = {
      ...input.stressMetrics,
      usage_n: 0,
      prompt_tokens: 0,
      completion_tokens: 0,
      tokens_per_sec: 0,
      request_tokens_per_sec: 0,
      tpm: 0,
    }
    input.stressAssessment = assessStress(
      input.stressMetrics,
      getStandard('default')
    )
    const markdown = buildMarkdownReport(input)
    const html = buildHtmlReport(input)
    expect(markdown).toContain('Batch output throughput: —')
    expect(markdown).toContain('Per-request output rate: —')
    expect(html).not.toContain('0.0 tok/s')
  })

  test('exports grouped request errors with status and sample timing', () => {
    const input = makeTestInput()
    if (!input.stressMetrics) throw new Error('missing stress fixture')
    input.stressMetrics = {
      ...input.stressMetrics,
      succeeded: 98,
      failed: 2,
      issues: [
        {
          status_code: 429,
          message: 'rate <limit> exceeded',
          count: 2,
          worker: 2,
          round: 1,
          elapsed_ms: 1250,
        },
      ],
    }
    const markdown = buildMarkdownReport(input)
    const html = buildHtmlReport(input)
    expect(markdown).toContain('2 × 429: rate <limit> exceeded (W2/R1, 1.25 s)')
    expect(html).toContain('rate &lt;limit&gt; exceeded')
    expect(html).toContain('W2/R1, 1.25 s')
  })

  test('buildMarkdownReport formats structured sections without Cannot compare in tables', () => {
    const input = makeTestInput()
    const md = buildMarkdownReport(input)

    expect(md).toContain('# Supplier Test Report')
    expect(md).toContain('## 1. Connectivity and protocol')
    expect(md).toContain('## 2. Concurrency and stress test')
    expect(md).toContain('## 3. Prompt cache test')
    expect(md).toContain('50 × 2 (100)')
    expect(md).toContain('40.38 s')
    expect(md).toContain('935.0 tok/s')

    // Informational rows (throughput, duration, tokens) are in summary bullets, not benchmark table
    expect(md).not.toContain('| Throughput |')
    expect(md).not.toContain('| Total duration |')
  })

  test('buildVideoMarkdownReport generates dedicated Doubao video report decoupled from text metrics', () => {
    const videoInput = makeVideoTestInput()
    const md = buildVideoMarkdownReport(videoInput)

    expect(md).toContain('# Doubao Video Generation Test Report')
    expect(md).toContain('## 1. Video Generation Parameters')
    expect(md).toContain('A futuristic city with flying cars at sunset')
    expect(md).toContain('720p')
    expect(md).toContain('16:9')
    expect(md).toContain('cgt-20260921-test-12345')
    expect(md).toContain('https://tos.volces.com/test-video.mp4')
    expect(md).toContain('## 2. Execution Pipeline')
    expect(md).toContain('## 3. Output Result')

    // Zero contamination from text model tests
    expect(md).not.toContain('Concurrency and stress test')
    expect(md).not.toContain('Prompt cache test')
    expect(md).not.toContain('TTFT')
    expect(md).not.toContain('TPOT')
  })

  test('buildVideoHtmlReport generates print-ready Doubao video report without text model sections', () => {
    const videoInput = makeVideoTestInput()
    const html = buildVideoHtmlReport(videoInput)

    expect(html).toContain('Doubao Video Generation Test Report')
    expect(html).toContain('Volcano Ark Seedance Video Model Verification')
    expect(html).toContain('cgt-20260921-test-12345')
    expect(html).toContain('https://tos.volces.com/test-video.mp4')
    expect(html).toContain('@page { size: A4; margin: 0; }')

    // Confirms text model report does not leak into video report
    expect(html).not.toContain('Stress test assessment')
    expect(html).not.toContain('Prompt cache assessment')

    // And text model report does not contain video fields
    const textHtml = buildHtmlReport(makeTestInput())
    expect(textHtml).not.toContain('Doubao Video Generation Test Report')
    expect(textHtml).not.toContain('Video Generation Parameters')
  })
})
