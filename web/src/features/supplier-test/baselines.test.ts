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
  applyStandardEditorValue,
  assessCache,
  assessmentGroup,
  assessStress,
  getStandard,
  matchingStandardId,
  overallLabel,
  sanitizeStandard,
  STANDARD_EDITOR_FIELDS,
  standardEditorValue,
} from './baselines'
import { CORPORA, estimateTokens } from './constants'
import type { CacheMetrics, StressMetrics } from './types'

function stress(partial: Partial<StressMetrics>): StressMetrics {
  return {
    total: 10,
    succeeded: 10,
    failed: 0,
    error_rate: 0,
    elapsed_ms: 1200,
    tokens_per_sec: 40,
    prompt_tokens: 800,
    completion_tokens: 200,
    ttft_avg_ms: 1200,
    ttft_p50_ms: 1100,
    ttft_p90_ms: 1800,
    ttft_n: 10,
    tpot_avg_ms: 40,
    tpot_p50_ms: 38,
    tpot_p90_ms: 55,
    tpot_n: 10,
    rpm: 500,
    tpm: 50000,
    ...partial,
  }
}

describe('supplier-test verdicts', () => {
  test('a slightly slower vendor is still normal', () => {
    const assessment = assessStress(
      stress({
        error_rate: 0.08,
        succeeded: 92,
        failed: 8,
        total: 100,
        ttft_avg_ms: 4500,
        ttft_p50_ms: 4200,
        ttft_p90_ms: 7000,
        tpot_avg_ms: 72,
        tpot_p50_ms: 68,
        tpot_p90_ms: 140,
      })
    )
    expect(assessment.overall).toBe('ok')
    expect(overallLabel(assessment.overall)).toBe('Overall: normal')
  })

  test('marks slow latency runs as slow', () => {
    const assessment = assessStress(
      stress({
        error_rate: 0,
        succeeded: 10,
        failed: 0,
        ttft_avg_ms: 12000,
        ttft_p50_ms: 11000,
        ttft_p90_ms: 18000,
        tpot_avg_ms: 200,
        tpot_p50_ms: 180,
        tpot_p90_ms: 260,
      })
    )
    expect(assessment.overall).toBe('slow')
    expect(overallLabel(assessment.overall)).toBe('Overall: slow')
    expect(assessment.rows.some((row) => row.verdict === 'slow')).toBe(true)
  })

  test('marks high error rate runs as abnormal and gates perf assessment', () => {
    const assessment = assessStress(
      stress({
        error_rate: 0.95,
        succeeded: 10,
        failed: 190,
        total: 200,
      })
    )
    expect(assessment.overall).toBe('abnormal')
    expect(overallLabel(assessment.overall)).toBe('Overall: abnormal')
    expect(assessment.rows.find((row) => row.id === 'error_rate')?.verdict).toBe('abnormal')
    expect(assessment.rows.find((row) => row.id === 'success')?.verdict).toBe('abnormal')
    // Deep performance should be 'na' (cannot compare) when shallow failed with abnormal error rate
    const perf = assessmentGroup(assessment, 'perf')
    expect(perf.overall).toBe('na')
  })

  test('treats a 60% cache hit as normal', () => {
    const metrics: CacheMetrics = {
      warm_prompt_tokens: 3000,
      avg_hit_rate: 0.6,
      min_hit_rate: 0.55,
      last_cached_tokens: 1800,
      last_prompt_tokens: 3000,
      wait_seconds: 5,
      rounds: 3,
      has_cached_tokens: true,
    }
    const assessment = assessCache(metrics)
    expect(assessment.overall).toBe('ok')
    expect(assessment.rows.find((row) => row.id === 'hit')?.verdict).toBe('ok')
  })

  test('switching to the tight standard can mark the same run slow', () => {
    const metrics = stress({
      error_rate: 0.02,
      succeeded: 98,
      failed: 2,
      total: 100,
      ttft_avg_ms: 4500,
      ttft_p50_ms: 4200,
      ttft_p90_ms: 7000,
      tpot_avg_ms: 72,
      tpot_p50_ms: 68,
      tpot_p90_ms: 140,
    })
    expect(assessStress(metrics, getStandard('default')).overall).toBe('ok')
    expect(assessStress(metrics, getStandard('tight')).overall).toBe('slow')
  })

  test('sanitizeStandard keeps editable vendor numbers in range', () => {
    const custom = sanitizeStandard({
      id: 'vendor-x',
      errorSlow: 0.07,
      ttftShortOkMs: 4500,
      cacheHitOk: 0.65,
    })
    expect(custom.errorSlow).toBe(0.07)
    expect(custom.ttftShortOkMs).toBe(4500)
    expect(custom.cacheHitOk).toBe(0.65)
    expect(matchingStandardId(custom)).toBeNull()
    expect(
      matchingStandardId(sanitizeStandard(getStandard('default')))
    ).toBe('default')
  })

  test('editor percent and seconds round-trip on the standard form', () => {
    const errorField = STANDARD_EDITOR_FIELDS.find((item) => item.key === 'errorSlow')
    const ttftField = STANDARD_EDITOR_FIELDS.find(
      (item) => item.key === 'ttftShortOkMs'
    )
    if (!errorField || !ttftField) {
      throw new Error('standard editor fields missing')
    }
    const withError = applyStandardEditorValue(
      getStandard('default'),
      errorField,
      7
    )
    expect(withError.errorSlow).toBeCloseTo(0.07)
    expect(standardEditorValue(withError, errorField)).toBe(7)
    const withTtft = applyStandardEditorValue(withError, ttftField, 4.5)
    expect(withTtft.ttftShortOkMs).toBe(4500)
    expect(standardEditorValue(withTtft, ttftField)).toBe(4.5)
  })

  test('splits stress rows into shallow connectivity and deep performance', () => {
    const assessment = assessStress(stress({}))
    const shallow = assessmentGroup(assessment, 'shallow')
    const perf = assessmentGroup(assessment, 'perf')
    expect(shallow.rows.map((row) => row.id)).toEqual([
      'error_rate',
      'success',
      'duration',
    ])
    expect(perf.rows.some((row) => row.id === 'ttft_avg')).toBe(true)
    expect(perf.rows.some((row) => row.id === 'tokens')).toBe(true)
    expect(assessment.rows.find((row) => row.id === 'ttft_p90')?.threshold).toBe(
      'Normal ≤ {{seconds}}s and ≤ avg × {{times}}; slower above that'
    )
  })

  test('a 20% cache hit is abnormal on the default ruler, not a failed check', () => {
    const assessment = assessCache({
      warm_prompt_tokens: 3000,
      avg_hit_rate: 0.2,
      min_hit_rate: 0.2,
      last_cached_tokens: 400,
      last_prompt_tokens: 2000,
      wait_seconds: 30,
      rounds: 1,
      has_cached_tokens: true,
    })
    expect(assessment.rows.find((row) => row.id === 'hit')?.verdict).toBe('abnormal')
    expect(assessment.overall).toBe('abnormal')
  })

  test('built-in corpora show a token estimate', () => {
    for (const item of CORPORA) {
      if (item.id === 'custom') {
        expect(item.tokens).toBe(0)
        continue
      }
      expect(item.tokens).toBeGreaterThan(0)
      expect(item.tokens).toBe(estimateTokens(item.prompt))
    }
  })
})
