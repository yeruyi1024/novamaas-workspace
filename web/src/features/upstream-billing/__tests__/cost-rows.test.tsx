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
import { screen, within } from '@testing-library/react'
import { expect, test } from 'vitest'

import { CostRows } from '../components/cost-rows'
import { renderUpstream, row } from './fixtures'

test('missing video usage is shown as unknown while a reported zero remains zero', () => {
  renderUpstream(
    <CostRows
      mode='day'
      rows={[
        { ...row, label: '2026-09-01', missing_usage_records: 1 },
        { ...row, label: '2026-09-02' },
      ]}
      symbol='$'
    />
  )
  const missing = screen.getByRole('row', { name: /2026-09-01/ })
  expect(
    within(missing).getAllByRole('cell', { name: 'Unknown' })
  ).toHaveLength(2)
  expect(within(missing).getByText(/Incomplete usage/)).toBeVisible()
  const complete = screen.getByRole('row', { name: /2026-09-02/ })
  expect(within(complete).queryByText('Unknown')).not.toBeInTheDocument()
  expect(within(complete).getAllByRole('cell', { name: '0' })).toHaveLength(2)
})

test('reported cache durations and long model identifiers stay readable as model-specific evidence', () => {
  const modelName = 'supplier/model-with-a-long-version-name-2026-09-preview'
  renderUpstream(
    <CostRows
      mode='model'
      rows={[
        {
          ...row,
          model_name: modelName,
          upstream_model_name: 'provider-exact-model',
          billing_group: 'contract-group',
        },
      ]}
      symbol='$'
    />
  )
  expect(screen.getByText(modelName)).toBeVisible()
  expect(screen.getByText('provider-exact-model')).toBeVisible()
  expect(screen.getByText('Billing group: contract-group')).toBeVisible()
  expect(screen.getByText('5 minutes: 7 · 1 hour: 3')).toBeVisible()
  expect(
    screen.getByRole('columnheader', { name: 'Video seconds' })
  ).toBeVisible()
})

test.each(['day', 'model'] as const)(
  'a %s row with all costs missing does not present charges, refunds or net as zero',
  (mode) => {
    renderUpstream(
      <CostRows
        mode={mode}
        rows={[
          {
            ...row,
            model_name: 'missing-model',
            count: 3,
            missing_cost_records: 3,
            charge: '0.000000',
            refund: '0.000000',
            amount: '0.000000',
          },
        ]}
        symbol='$'
      />
    )
    const record = within(screen.getAllByRole('row')[1])
    expect(record.getAllByText('Unknown')).toHaveLength(3)
    expect(record.queryByText('$ 0.000000')).not.toBeInTheDocument()
  }
)

test('partial daily costs stay exact and are marked as incomplete subtotals', () => {
  renderUpstream(
    <CostRows
      mode='day'
      rows={[{ ...row, missing_cost_records: 1 }]}
      symbol='$'
    />
  )
  expect(screen.getByText('$ 1.000000')).toBeVisible()
  expect(screen.getAllByText('Known cost subtotal (incomplete)')).toHaveLength(
    3
  )
})

test('empty days and configured zero-cost records remain zero without incomplete cost labels', () => {
  renderUpstream(
    <CostRows
      mode='day'
      rows={[
        {
          ...row,
          label: '2026-09-01',
          count: 0,
          requests: 0,
          charge: '0.000000',
          refund: '0.000000',
          amount: '0.000000',
        },
        {
          ...row,
          label: '2026-09-02',
          count: 2,
          requests: 2,
          charge: '0.000000',
          refund: '0.000000',
          amount: '0.000000',
        },
      ]}
      symbol='$'
    />
  )
  for (const label of ['2026-09-01', '2026-09-02']) {
    const day = within(screen.getByRole('row', { name: new RegExp(label) }))
    expect(day.getAllByText('$ 0.000000')).toHaveLength(3)
    expect(
      day.queryByText('Known cost subtotal (incomplete)')
    ).not.toBeInTheDocument()
    expect(day.queryByText('Missing cost')).not.toBeInTheDocument()
  }
})
