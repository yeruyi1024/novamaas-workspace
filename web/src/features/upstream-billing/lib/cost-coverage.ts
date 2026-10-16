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
import type { UpstreamDayRow, UpstreamReport } from '../types'

export type CostCoverage = 'complete' | 'partial' | 'missing' | 'excluded'

export function upstreamCostCoverage(
  row: Pick<UpstreamDayRow, 'count' | 'missing_cost_records'>,
  scope?: Pick<UpstreamReport, 'group' | 'coverage'>,
  includeMissingChannels = true
): CostCoverage {
  if (
    scope &&
    !scope.group.channels.some(
      (channel) => channel.eligible && !channel.missing
    )
  ) {
    return 'excluded'
  }
  if (row.count > 0 && row.missing_cost_records >= row.count) return 'missing'
  if (row.missing_cost_records > 0) return 'partial'
  if (includeMissingChannels && (scope?.coverage.missing_channels ?? 0) > 0) {
    return 'partial'
  }
  return 'complete'
}
