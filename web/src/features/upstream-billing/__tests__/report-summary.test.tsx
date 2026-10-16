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

import { ReportSummary } from '../components/report-summary'
import { report } from './fixtures'

test('a preview with every historical cost missing shows unknown rather than a zero net cost', () => {
  render(
    <ReportSummary
      report={{
        ...report,
        total: {
          ...report.total,
          amount: '0.000000',
          count: 11864,
          missing_cost_records: 11864,
        },
        coverage: {
          ...report.coverage,
          complete: false,
          missing_cost_records: 11864,
        },
      }}
    />
  )
  expect(screen.getByText('Unknown')).toBeVisible()
  expect(screen.queryByText('$ 0.000000')).not.toBeInTheDocument()
})

test('a partial cost preview keeps the exact server subtotal and marks it incomplete', () => {
  render(
    <ReportSummary
      report={{
        ...report,
        total: {
          ...report.total,
          amount: '0.123456',
          count: 3,
          missing_cost_records: 1,
        },
        coverage: {
          ...report.coverage,
          complete: false,
          missing_cost_records: 1,
        },
      }}
    />
  )
  expect(screen.getByText('$ 0.123456')).toBeVisible()
  expect(screen.getByText('Known cost subtotal (incomplete)')).toBeVisible()
})

test('a configured zero cost stays zero when costs are complete even if usage has gaps', () => {
  render(
    <ReportSummary
      report={{
        ...report,
        total: {
          ...report.total,
          amount: '0.000000',
          missing_usage_records: 9,
        },
        coverage: {
          ...report.coverage,
          complete: false,
          missing_usage_records: 9,
        },
      }}
    />
  )
  expect(screen.getByText('$ 0.000000')).toBeVisible()
  expect(screen.queryByText('Unknown')).not.toBeInTheDocument()
  expect(
    screen.queryByText('Known cost subtotal (incomplete)')
  ).not.toBeInTheDocument()
})

test('a statement without included cost channels shows unknown without inventing missing cost records', () => {
  render(
    <ReportSummary
      report={{
        ...report,
        group: {
          ...report.group,
          channels: report.group.channels.filter(
            (channel) => !channel.eligible
          ),
        },
        total: { ...report.total, count: 0, requests: 0, amount: '0.000000' },
        coverage: { ...report.coverage, complete: false },
      }}
    />
  )
  expect(screen.getByText('Unknown')).toBeVisible()
  expect(screen.getByText('Not included in cost statistics')).toBeVisible()
  expect(screen.getByText('Missing cost: 0')).toBeVisible()
  expect(screen.queryByText('$ 0.000000')).not.toBeInTheDocument()
})

test('a missing group channel marks the available amount as a known subtotal', () => {
  render(
    <ReportSummary
      report={{
        ...report,
        coverage: { ...report.coverage, complete: false, missing_channels: 1 },
      }}
    />
  )
  expect(screen.getByText('$ 1.000000')).toBeVisible()
  expect(screen.getByText('Known cost subtotal (incomplete)')).toBeVisible()
})

test('included channels with no requests retain a real zero cost', () => {
  render(
    <ReportSummary
      report={{
        ...report,
        total: { ...report.total, count: 0, requests: 0, amount: '0.000000' },
      }}
    />
  )
  expect(screen.getByText('$ 0.000000')).toBeVisible()
  expect(screen.queryByText('Unknown')).not.toBeInTheDocument()
})
