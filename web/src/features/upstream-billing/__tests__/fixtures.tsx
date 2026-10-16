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
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render } from '@testing-library/react'
import type { ReactNode } from 'react'

import type {
  UpstreamChannel,
  UpstreamChannelUserRow,
  UpstreamDayRow,
  UpstreamGroup,
  UpstreamReport,
  UpstreamStatementDetail,
} from '../types'

export const channels: UpstreamChannel[] = [
  {
    id: 19,
    name: 'Primary channel',
    tag: 'demo',
    status: 1,
    cost_discount: '0.5',
    eligible: true,
  },
  {
    id: 26,
    name: 'Zero-cost channel',
    tag: 'demo',
    status: 1,
    cost_discount: '0',
    eligible: true,
  },
  {
    id: 27,
    name: 'Unconfigured channel',
    tag: 'legacy',
    status: 2,
    cost_discount: '',
    eligible: false,
  },
  {
    id: 28,
    name: 'Assigned channel',
    tag: 'other',
    status: 1,
    cost_discount: '0.8',
    eligible: true,
    group_id: 2,
  },
]
export const group: UpstreamGroup = {
  id: 1,
  name: 'Demo upstream',
  version: 3,
  channels: channels.slice(0, 3),
}
export const row: UpstreamDayRow = {
  label: '2026-09-03',
  charge: '1.230000',
  refund: '0.230000',
  amount: '1.000000',
  count: 3,
  requests: 2,
  charge_count: 2,
  refund_count: 1,
  input_tokens: 100,
  output_tokens: 20,
  cache_read_tokens: 40,
  cache_write_tokens: 10,
  cache_write_5m_tokens: 7,
  cache_write_1h_tokens: 3,
  video_tokens: 0,
  video_seconds: 0,
  missing_cost_records: 0,
  missing_usage_records: 0,
  estimated_usage_records: 0,
  pending_tasks: 0,
}
export const channelUserRow: UpstreamChannelUserRow = {
  ...row,
  channel_id: 19,
  channel_name: 'Primary channel',
  user_id: 5,
  username: 'customer-five',
  display_name: 'Customer Five',
  channel_amount: '1.000000',
  share_percent: '100.0000',
}
export const report: UpstreamReport = {
  schema_version: 2,
  accounting_basis: 'recorded_supplier_cost',
  group,
  month: '2026-09',
  start_date: '2026-09-01',
  end_date: '2026-09-30',
  start_at: 1788192000,
  end_at: 1790784000,
  timezone: 'Asia/Shanghai',
  currency: { code: 'USD', symbol: '$', rate: '1', quota_per_unit: '500000' },
  days: [row],
  models: [
    {
      ...row,
      model_name: 'gpt-5',
      upstream_model_name: 'upstream-gpt-5',
      billing_group: 'default',
    },
  ],
  channel_users: [channelUserRow],
  day_models: [{ ...row, date: row.label, model_name: 'gpt-5' }],
  total: { ...row, label: 'Total' },
  coverage: {
    complete: true,
    missing_cost_records: 0,
    missing_usage_records: 0,
    estimated_usage_records: 0,
    pending_tasks: 0,
    excluded_channels: [channels[2]],
    missing_channels: 0,
  },
  updated_at: 1791936000,
  rounding_difference: '0.000000',
  source_sha256: 'source',
  issuer: 'Issuer',
}
export const detail: UpstreamStatementDetail = {
  statement: {
    id: 'frozen-statement',
    group_id: 1,
    month: '2026-09',
    revision: 1,
    status: 'draft',
    manifest_sha256: 'frozen-manifest',
    created_at: 1791936000,
  },
  snapshot: report,
  events: [
    {
      id: 1,
      actor_id: 1,
      actor_username: 'accountant',
      action: 'prepare',
      note: '',
      created_at: 1791936000,
    },
  ],
  artifacts: [
    { id: 1, kind: 'pdf', ordinal: 0, rows: 3, sha256: 'pdf' },
    { id: 2, kind: 'xlsx', ordinal: 0, rows: 3, sha256: 'excel' },
  ],
  detail_count: 3,
  attachments: [],
  evidence_sha256: 'e'.repeat(64),
  pending_uploads: 0,
}

export function renderUpstream(node: ReactNode) {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  })
  const result = render(
    <QueryClientProvider client={client}>{node}</QueryClientProvider>
  )
  return { ...result, client }
}
