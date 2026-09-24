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
import { render, screen } from '@testing-library/react'
import { expect, test } from 'vitest'

import type { StressMetrics } from '../types'
import { StressResults } from './stress-results'

test('shows request failures and separates batch time from request timing', () => {
  const metrics: StressMetrics = {
    total: 2,
    attempted: 2,
    succeeded: 1,
    failed: 1,
    error_rate: 0.5,
    elapsed_ms: 56000,
    request_avg_ms: 55000,
    request_p50_ms: 55000,
    request_p90_ms: 55000,
    tokens_per_sec: 1.14,
    request_tokens_per_sec: 1.16,
    prompt_tokens: 31,
    completion_tokens: 64,
    usage_n: 1,
    ttft_avg_ms: 54000,
    ttft_p50_ms: 54000,
    ttft_p90_ms: 54000,
    ttft_n: 1,
    tpot_avg_ms: 16,
    tpot_p50_ms: 16,
    tpot_p90_ms: 16,
    tpot_n: 1,
    rpm: 1,
    tpm: 100,
    issues: [
      {
        status_code: 429,
        message: 'rate limit exceeded',
        count: 1,
        worker: 2,
        round: 1,
        elapsed_ms: 1250,
      },
    ],
  }
  render(<StressResults metrics={metrics} assessment={null} stream />)

  expect(screen.getByText('Batch wall time')).toBeInTheDocument()
  expect(screen.getByText('56.0 s')).toBeInTheDocument()
  expect(screen.getByText('rate limit exceeded')).toBeInTheDocument()
  expect(screen.getByText('W2 / R1')).toBeInTheDocument()
  expect(screen.getByText('1.1 tok/s')).toBeInTheDocument()
  expect(screen.getByText('1.2 tok/s')).toBeInTheDocument()
})
