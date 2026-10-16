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
import type { BillingCurrency, BillingRow } from '@/features/billing/types'

export interface UpstreamGroup {
  id: number
  name: string
  version: number
  channels: UpstreamChannel[]
}

export interface UpstreamChannel {
  id: number
  name: string
  tag?: string
  status: number
  cost_discount: string
  group_id?: number
  eligible: boolean
  missing?: boolean
}

export interface UpstreamUsage {
  requests: number
  input_tokens: number
  output_tokens: number
  cache_read_tokens: number
  cache_write_tokens: number
  cache_write_5m_tokens: number
  cache_write_1h_tokens: number
  video_tokens: number
  video_seconds: number
}

export interface UpstreamDayRow extends BillingRow, UpstreamUsage {
  charge_count: number
  refund_count: number
  missing_cost_records: number
  missing_usage_records: number
  estimated_usage_records: number
  pending_tasks: number
}
export interface UpstreamModelRow extends UpstreamDayRow {
  model_name: string
  upstream_model_name?: string
  billing_group?: string
  active_days?: number
  first_posted_at?: number
  last_posted_at?: number
}
export interface UpstreamChannelUserRow extends UpstreamDayRow {
  channel_id: number
  channel_name: string
  user_id: number
  username: string
  display_name: string
  channel_amount: string
  share_percent: string | null
}

export interface UpstreamCoverage {
  complete: boolean
  missing_cost_records: number
  missing_usage_records: number
  estimated_usage_records: number
  pending_tasks: number
  excluded_channels: UpstreamChannel[]
  missing_channels: number
}

export interface UpstreamReport {
  schema_version: number
  accounting_basis: string
  group: UpstreamGroup
  month?: string
  start_date: string
  end_date: string
  start_at: number
  end_at: number
  timezone: string
  currency: BillingCurrency
  days: UpstreamDayRow[]
  models: UpstreamModelRow[]
  day_models: (UpstreamModelRow & { date: string })[]
  channel_users?: UpstreamChannelUserRow[]
  total: UpstreamDayRow
  coverage: UpstreamCoverage
  updated_at: number
  rounding_difference: string
  source_sha256: string
  issuer: string
}

export type UpstreamRange =
  | { month: string }
  | { start_date: string; end_date: string }

export interface UpstreamUsageRecord extends UpstreamUsage {
  sequence: number
  source_log_id: number
  occurred_at: number
  request_id: string
  upstream_request_id?: string
  task_id?: string
  channel_id: number
  model_name: string
  upstream_model_name?: string
  billing_group?: string
  log_type: number
  cost_quota: number
  amount: string
  cost_discount: string
  cost_missing: boolean
  usage_missing: boolean
  usage_estimated: boolean
  task_pending: boolean
  usage_source: string
}
export interface UpstreamUsageDetails {
  records: UpstreamUsageRecord[]
  next_cursor?: number
  currency: BillingCurrency
}

export type UpstreamStatementStatus =
  | 'preparing'
  | 'draft'
  | 'confirmed'
  | 'void'
  | 'failed'
export interface UpstreamStatement {
  id: string
  group_id: number
  month: string
  revision: number
  status: UpstreamStatementStatus
  manifest_sha256: string
  created_at: number
  confirmed_at?: number
  last_error?: string
}

export interface UpstreamStatementDetail {
  statement: UpstreamStatement
  snapshot: UpstreamReport | null
  events: {
    id: number
    actor_id: number
    actor_username?: string
    action: string
    note?: string
    created_at: number
  }[]
  artifacts: {
    id: number
    kind: string
    ordinal: number
    rows: number
    sha256: string
  }[]
  detail_count: number
  attachments: UpstreamAttachment[]
  evidence_sha256: string
  pending_uploads: number
}

export interface UpstreamAttachment {
  id: string
  statement_id: string
  original_filename: string
  content_type: string
  size: number
  sha256: string
  storage_profile_id: number
  uploaded_by: number
  uploaded_username?: string
  uploaded_at: number
  status: string
}

export interface UpstreamStorageProfile {
  id: number
  name: string
}
