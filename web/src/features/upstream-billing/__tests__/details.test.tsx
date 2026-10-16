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
import { fireEvent, screen, waitFor, within } from '@testing-library/react'
import { beforeEach, expect, test, vi } from 'vitest'

import { LOG_TYPE_ENUM } from '@/features/usage-logs/constants'

import { getUpstreamUsageDetails } from '../api'
import { UsageDetails } from '../components/usage-details'
import { renderUpstream, report, row } from './fixtures'

vi.mock('../api', async (original) => ({
  ...(await original<typeof import('../api')>()),
  getUpstreamUsageDetails: vi.fn(),
}))
beforeEach(() => vi.resetAllMocks())

const zeroUsageRecord = {
  ...row,
  sequence: 1,
  source_log_id: 93,
  occurred_at: 1791936000,
  request_id: 'zero-usage-request',
  channel_id: 19,
  model_name: 'video-model',
  log_type: LOG_TYPE_ENUM.CONSUME,
  cost_quota: 0,
  amount: '0.000000',
  cost_discount: '0.5',
  cost_missing: false,
  usage_missing: false,
  usage_estimated: false,
  task_pending: false,
  usage_source: 'task_final_usage',
  requests: 0,
  input_tokens: 0,
  output_tokens: 0,
  cache_read_tokens: 0,
  cache_write_tokens: 0,
  cache_write_5m_tokens: 0,
  cache_write_1h_tokens: 0,
  video_tokens: 0,
  video_seconds: 0,
}

test('detail pagination keeps the selected day and model and recovers from a page failure', async () => {
  vi.mocked(getUpstreamUsageDetails)
    .mockResolvedValueOnce({
      records: [],
      currency: report.currency,
      next_cursor: 91,
    })
    .mockRejectedValueOnce(new Error('Unavailable'))
    .mockResolvedValueOnce({ records: [], currency: report.currency })
  renderUpstream(
    <UsageDetails
      groupId={1}
      range={{ month: '2026-09' }}
      filters={{ day: '2026-09-03', model: 'gpt-5' }}
      symbol='$'
      onClose={vi.fn()}
    />
  )
  await screen.findByText('No usage records')
  fireEvent.click(screen.getByRole('button', { name: 'Next' }))
  await screen.findByRole('alert')
  expect(getUpstreamUsageDetails).toHaveBeenLastCalledWith(
    1,
    { month: '2026-09' },
    { day: '2026-09-03', model: 'gpt-5', cursor: 91 }
  )
  expect(screen.getByRole('button', { name: 'Next' })).toBeDisabled()
  fireEvent.click(screen.getByRole('button', { name: 'Refresh' }))
  await screen.findByText('No usage records')
  fireEvent.click(screen.getByRole('button', { name: 'Previous' }))
  await waitFor(() =>
    expect(getUpstreamUsageDetails).toHaveBeenLastCalledWith(
      1,
      { month: '2026-09' },
      { day: '2026-09-03', model: 'gpt-5', cursor: undefined }
    )
  )
})

test('request evidence preserves cache durations and displays missing video telemetry as unknown', async () => {
  const record = {
    ...row,
    sequence: 1,
    source_log_id: 91,
    occurred_at: 1791936000,
    request_id: 'cache-request',
    channel_id: 19,
    model_name: 'gpt-5',
    log_type: 2,
    cost_quota: 500000,
    cost_discount: '0.5',
    cost_missing: false,
    usage_missing: false,
    usage_estimated: false,
    task_pending: false,
    usage_source: 'upstream',
  }
  vi.mocked(getUpstreamUsageDetails).mockResolvedValue({
    currency: report.currency,
    records: [
      record,
      {
        ...record,
        sequence: 2,
        source_log_id: 92,
        request_id: 'unknown-video-request',
        usage_missing: true,
      },
    ],
  })
  renderUpstream(
    <UsageDetails
      groupId={1}
      range={{ month: '2026-09' }}
      filters={{}}
      symbol='$'
      onClose={vi.fn()}
    />
  )
  expect(await screen.findByText('5 minutes: 7 · 1 hour: 3')).toBeVisible()
  expect(screen.getAllByRole('cell', { name: 'Unknown' })).toHaveLength(2)
  expect(screen.getByText('Incomplete usage')).toBeVisible()
})

test('a failed task whose cost was fully refunded marks all usage as inapplicable and explains the source', async () => {
  const record = {
    ...zeroUsageRecord,
    request_id: 'fully-refunded-task-request',
    usage_source: 'task_failed_refunded',
  }
  vi.mocked(getUpstreamUsageDetails).mockResolvedValue({
    currency: report.currency,
    records: [record],
  })
  renderUpstream(
    <UsageDetails
      groupId={1}
      range={{ month: '2026-09' }}
      filters={{}}
      symbol='$'
      onClose={vi.fn()}
    />
  )
  const evidence = await screen.findByRole('row', {
    name: /fully-refunded-task-request/,
  })
  const cells = within(evidence).getAllByRole('cell')
  for (const cell of cells.slice(4, 10)) {
    expect(cell).toHaveTextContent(/^—$/)
  }
  expect(cells[10]).toHaveTextContent(
    'Source: Task failed and its upstream cost was fully refunded. Usage does not apply.'
  )
  expect(
    within(evidence).queryByText('Incomplete usage')
  ).not.toBeInTheDocument()
})

test('an ordinary refund adjustment preserves zero usage and its refund marker', async () => {
  const record = {
    ...zeroUsageRecord,
    request_id: 'ordinary-refund-request',
    log_type: LOG_TYPE_ENUM.REFUND,
    usage_source: 'refund_adjustment',
    cost_quota: -500000,
    amount: '-1.000000',
  }
  vi.mocked(getUpstreamUsageDetails).mockResolvedValue({
    currency: report.currency,
    records: [record],
  })
  renderUpstream(
    <UsageDetails
      groupId={1}
      range={{ month: '2026-09' }}
      filters={{}}
      symbol='$'
      onClose={vi.fn()}
    />
  )
  const evidence = await screen.findByRole('row', {
    name: /ordinary-refund-request/,
  })
  for (const cell of within(evidence).getAllByRole('cell').slice(4, 10)) {
    expect(cell).toHaveTextContent(/^0$/)
  }
  expect(within(evidence).getByText('Refund')).toBeVisible()
  expect(within(evidence).getByText('$ -1.000000')).toBeVisible()
})

test('request evidence exposes original request, source log, task and upstream request identifiers together', async () => {
  const record = {
    ...zeroUsageRecord,
    request_id: 'client-request-original-001',
    task_id: 'upstream-task-original-009',
    upstream_request_id: 'provider-request-original-041',
  }
  vi.mocked(getUpstreamUsageDetails).mockResolvedValue({
    currency: report.currency,
    records: [record],
  })
  renderUpstream(
    <UsageDetails
      groupId={1}
      range={{ month: '2026-09' }}
      filters={{}}
      symbol='$'
      onClose={vi.fn()}
    />
  )
  const evidence = await screen.findByRole('row', {
    name: /client-request-original-001/,
  })
  const identifiers = within(evidence).getAllByRole('cell')[11]
  expect(identifiers).toHaveTextContent('client-request-original-001')
  expect(identifiers).toHaveTextContent('Source log ID: 93')
  expect(identifiers).toHaveTextContent('Task ID: upstream-task-original-009')
  expect(identifiers).toHaveTextContent(
    'Upstream Request ID: provider-request-original-041'
  )
})

test.each(['supplier_private_reason_219', 'constructor'])(
  'the unknown usage source %s falls back safely without presenting its raw code as a known reason',
  async (source) => {
    const record = {
      ...zeroUsageRecord,
      request_id: 'unknown-source-request',
      usage_source: source,
    }
    vi.mocked(getUpstreamUsageDetails).mockResolvedValue({
      currency: report.currency,
      records: [record],
    })
    renderUpstream(
      <UsageDetails
        groupId={1}
        range={{ month: '2026-09' }}
        filters={{}}
        symbol='$'
        onClose={vi.fn()}
      />
    )
    const evidence = await screen.findByRole('row', {
      name: /unknown-source-request/,
    })
    expect(within(evidence).getAllByRole('cell')[10]).toHaveTextContent(
      'Source: Unknown'
    )
    expect(evidence).not.toHaveTextContent(source)
  }
)

test.each([
  ['task_missing_id', 'The source log has no task ID.'],
  ['task_not_found', 'The matching task was not found.'],
  ['task_usage_missing', 'The task’s final usage evidence is missing.'],
])(
  'the missing usage source %s explains its gap while preserving unknown telemetry and the source log identifier',
  async (source, explanation) => {
    const record = {
      ...zeroUsageRecord,
      request_id: 'missing-task-usage-request',
      usage_missing: true,
      usage_source: source,
    }
    vi.mocked(getUpstreamUsageDetails).mockResolvedValue({
      currency: report.currency,
      records: [record],
    })
    renderUpstream(
      <UsageDetails
        groupId={1}
        range={{ month: '2026-09' }}
        filters={{}}
        symbol='$'
        onClose={vi.fn()}
      />
    )
    const evidence = await screen.findByRole('row', {
      name: /missing-task-usage-request/,
    })
    const cells = within(evidence).getAllByRole('cell')
    for (const cell of cells.slice(4, 8)) {
      expect(cell).toHaveTextContent(/^—$/)
    }
    for (const cell of cells.slice(8, 10)) {
      expect(cell).toHaveTextContent(/^Unknown$/)
    }
    expect(cells[10]).toHaveTextContent(`Source: ${explanation}`)
    expect(within(cells[10]).getByText('Incomplete usage')).toBeVisible()
    expect(cells[11]).toHaveTextContent('Source log ID: 93')
  }
)

test('reviewed historical task links show recovered final token evidence and its task identity', async () => {
  const record = {
    ...zeroUsageRecord,
    request_id: 'historical-final-usage-request',
    task_id: 'reviewed-historical-task-871',
    usage_source: 'task_historical_final_usage',
    requests: 1,
    video_tokens: 8717745,
  }
  vi.mocked(getUpstreamUsageDetails).mockResolvedValue({
    currency: report.currency,
    records: [record],
  })
  renderUpstream(
    <UsageDetails
      groupId={1}
      range={{ month: '2026-09' }}
      filters={{}}
      symbol='$'
      onClose={vi.fn()}
    />
  )
  const evidence = await screen.findByRole('row', {
    name: /historical-final-usage-request/,
  })
  const cells = within(evidence).getAllByRole('cell')
  expect(cells[8]).toHaveTextContent(/^8717745$/)
  expect(cells[10]).toHaveTextContent(
    'Source: Final task usage recovered through a reviewed historical task link.'
  )
  expect(cells[11]).toHaveTextContent('Task ID: reviewed-historical-task-871')
})

test('reviewed historical failed tasks with a full refund keep all usage inapplicable and show the historical audit explanation', async () => {
  const record = {
    ...zeroUsageRecord,
    request_id: 'historical-failed-refunded-request',
    task_id: 'reviewed-historical-task-refunded',
    usage_source: 'task_historical_failed_refunded',
  }
  vi.mocked(getUpstreamUsageDetails).mockResolvedValue({
    currency: report.currency,
    records: [record],
  })
  renderUpstream(
    <UsageDetails
      groupId={1}
      range={{ month: '2026-09' }}
      filters={{}}
      symbol='$'
      onClose={vi.fn()}
    />
  )
  const evidence = await screen.findByRole('row', {
    name: /historical-failed-refunded-request/,
  })
  const cells = within(evidence).getAllByRole('cell')
  for (const cell of cells.slice(4, 10)) {
    expect(cell).toHaveTextContent(/^—$/)
  }
  expect(cells[10]).toHaveTextContent(
    'Source: Historical task link reviewed: task failed and its upstream cost was fully refunded. Usage does not apply.'
  )
  expect(
    within(evidence).queryByText('Incomplete usage')
  ).not.toBeInTheDocument()
})
