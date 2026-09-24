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
import type { StatusVariant } from '@/components/status-badge'

import type { Verdict } from './baselines'
import type { CheckStatus } from './types'

export function statusVariant(status: CheckStatus): StatusVariant {
  if (status === 'pass') return 'success'
  if (status === 'fail') return 'danger'
  if (status === 'skip') return 'warning'
  if (status === 'running') return 'info'
  return 'neutral'
}

export function statusLabel(status: CheckStatus): string {
  if (status === 'pass') return 'Passed'
  if (status === 'fail') return 'Failed'
  if (status === 'skip') return 'Skipped'
  if (status === 'running') return 'Running'
  return 'Idle'
}

export function verdictClass(verdict: Verdict): string {
  if (verdict === 'ok') return 'text-success font-medium'
  if (verdict === 'slow') return 'text-warning font-medium'
  if (verdict === 'abnormal') return 'text-destructive font-medium'
  return 'text-muted-foreground'
}
