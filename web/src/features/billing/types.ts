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
export interface BillingAccount {
  user_id: number
  accounting_start_at: number
  company_title: string
  tax_id: string
  profile_version: number
}
export interface BillingRow {
  label: string
  state?: 'future' | 'in_progress' | 'outside_period'
  charge: string
  refund: string
  amount: string
  count: number
}
export interface BillingCurrency {
  code: string
  symbol: string
  rate: string
  quota_per_unit: string
}
export interface BillingSnapshot {
  user_id?: number
  username?: string
  display_name?: string
  pdf_template_version?: number
  pdf_logo_png?: string
  pdf_footer?: string
  month: string
  timezone: string
  company_title: string
  tax_id: string
  issuer: string
  accounting_start_at: number
  currency: BillingCurrency
  days: BillingRow[]
  total: BillingRow
  rounding_difference?: string
}
export interface BillingDay {
  date: string
  timezone: string
  account: BillingAccount
  currency: BillingCurrency
  hours: BillingRow[]
  total: BillingRow
  updated_at: number
  available: boolean
  source: 'usage_logs'
  last_log_at: number
  rounding_difference?: string
}
export interface BillingMonthPreview {
  reference: BillingSnapshot & {
    source: 'usage_logs'
    updated_at: number
    last_log_at: number
  }
  formal: BillingSnapshot | null
  readiness: {
    ready: boolean
    status:
      | 'ready'
      | 'not_configured'
      | 'outside_period'
      | 'data_error'
      | 'no_consumption'
      | 'historical_data_unreconciled'
      | 'blocked'
    checks: { code: string; passed: boolean }[]
    accounting_start_at: number
    period_start_at: number
    period_end_at: number
    prepare_at: number
    pending: number
    existing_statement: string
  }
}
export interface BillingUsageDetails {
  items: {
    id: number
    created_at: number
    request_id: string
    model_name: string
    kind: 'consumption' | 'refund'
    amount: string
  }[]
  next_cursor: string
  currency: BillingCurrency
}
export type StatementStatus =
  | 'preparing'
  | 'draft'
  | 'issued'
  | 'disputed'
  | 'confirmed'
  | 'void'
  | 'failed'
export interface BillingStatement {
  id: string
  user_id: number
  month: string
  revision: number
  status: StatementStatus
  snapshot: string
  manifest_sha256: string
  pdf_sha256: string
  issued_at: number
  due_at: number
  confirmed_at: number
  created_at: number
  last_error?: string
}
export interface StatementDetail {
  customer?: { id: number; username: string; display_name: string }
  source_warning?: 'historical_data_unreconciled' | ''
  detail_count?: number
  statement: BillingStatement
  events: {
    id: number
    actor_id: number
    actor_username?: string
    action: string
    note: string
    created_at: number
  }[]
  artifacts: {
    id: number
    kind: string
    ordinal: number
    rows: number
    sha256: string
  }[]
}

export interface BillingHistoryReview {
  user_id: number
  month: string
  customer: { id: number; username: string; display_name: string }
  ready: boolean
  checks: { code: string; passed: boolean }[]
  source_sha256: string
  source_count: number
  record_limit: number
  existing_statement: string
  old_start_at: number
  new_start_at: number
  snapshot: BillingSnapshot | null
}

export interface BillingHistoryImport {
  id: string
  user_id: number
  month: string
  source_sha256: string
  records: number
  confirmed_by: number
  confirmed_by_username?: string
  confirmed_at: number
  note: string
}
