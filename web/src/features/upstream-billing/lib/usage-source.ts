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
export const upstreamUsageSourceKeys: Record<string, string> = {
  task_historical_final_usage:
    'Final task usage recovered through a reviewed historical task link.',
  task_historical_failed_refunded:
    'Historical task link reviewed: task failed and its upstream cost was fully refunded. Usage does not apply.',
  task_failed_refunded:
    'Task failed and its upstream cost was fully refunded. Usage does not apply.',
  task_missing_id: 'The source log has no task ID.',
  task_not_found: 'The matching task was not found.',
  task_ambiguous: 'Multiple matching tasks were found.',
  task_pending: 'Task is awaiting a final result.',
  task_invalid_usage: 'Task response usage data is invalid.',
  task_usage_missing: 'The task’s final usage evidence is missing.',
  task_duration_usage: 'Final task usage reported in seconds.',
  task_final_usage: 'Final usage recorded by the task.',
  refund_adjustment: 'Refund adjustment; usage is not counted again.',
  missing_log: 'Source usage log is missing.',
  invalid_metadata: 'Usage log metadata is invalid.',
  invalid_usage: 'Saved usage values are invalid.',
  usage_logs: 'Usage recorded in the request log.',
  upstream: 'Usage reported by the upstream provider.',
  local: 'Usage estimated locally.',
  'billing-usage-openai': 'Usage reported by the upstream provider.',
  'billing-usage-openai-estimated': 'Estimated usage recorded for billing.',
  'billing-usage-anthropic': 'Usage reported by the upstream provider.',
  'billing-usage-anthropic-estimated': 'Estimated usage recorded for billing.',
  'billing-usage-gemini': 'Usage reported by the upstream provider.',
  'billing-usage-gemini-estimated': 'Estimated usage recorded for billing.',
}
