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
import { render, screen, within } from '@testing-library/react'
import { expect, test } from 'vitest'

import { ChannelUserCosts } from '../components/channel-user-costs'
import { channelUserRow, report, row } from './fixtures'

test('channel allocation displays server cost amounts and exact shares without sales amounts or client-side recalculation', () => {
  render(
    <ChannelUserCosts
      rows={report.channel_users}
      schemaVersion={2}
      symbol='$'
    />
  )
  const user = within(screen.getByRole('row', { name: /Customer Five/ }))
  expect(user.getAllByText('$ 1.000000')).toHaveLength(2)
  expect(user.getByText('$ 1.230000')).toBeVisible()
  expect(user.getByText('100.0000%')).toBeVisible()
  expect(
    screen.getByText(/Shares use each channel’s net upstream cost/)
  ).toBeVisible()
})

test('a missing-cost or nonpositive channel share stays unavailable instead of being calculated in the browser', () => {
  render(
    <ChannelUserCosts
      rows={[
        {
          ...row,
          channel_id: 27,
          channel_name: 'No cost',
          user_id: 3,
          username: 'customer',
          display_name: '',
          channel_amount: '0.000000',
          share_percent: null,
        },
      ]}
      schemaVersion={2}
      symbol='$'
    />
  )
  expect(screen.getByRole('cell', { name: '—' })).toBeVisible()
  expect(screen.queryByText('0%')).not.toBeInTheDocument()
})

test('older frozen snapshots clearly state that allocation was not frozen', () => {
  render(<ChannelUserCosts schemaVersion={1} symbol='$' frozen />)
  expect(
    screen.getByText(
      'This older snapshot did not freeze user cost allocation by channel.'
    )
  ).toBeVisible()
  expect(screen.queryByRole('table')).not.toBeInTheDocument()
})

test('user and channel costs remain unknown when every user in the channel lacks cost evidence', () => {
  render(
    <ChannelUserCosts
      schemaVersion={2}
      symbol='$'
      rows={[
        {
          ...channelUserRow,
          count: 2,
          missing_cost_records: 2,
          charge: '0.000000',
          refund: '0.000000',
          amount: '0.000000',
          channel_amount: '0.000000',
          share_percent: null,
        },
      ]}
    />
  )
  expect(screen.getAllByText('Unknown')).toHaveLength(4)
  expect(screen.queryByText('$ 0.000000')).not.toBeInTheDocument()
})

test('a complete user amount stays known while the channel subtotal is incomplete because another user lacks cost evidence', () => {
  const known = channelUserRow
  render(
    <ChannelUserCosts
      schemaVersion={2}
      symbol='$'
      rows={[
        {
          ...known,
          count: 1,
          missing_cost_records: 0,
          channel_amount: '1.000000',
          share_percent: null,
        },
        {
          ...known,
          user_id: 6,
          display_name: 'Missing User',
          username: 'missing-user',
          count: 1,
          missing_cost_records: 1,
          charge: '0.000000',
          refund: '0.000000',
          amount: '0.000000',
          channel_amount: '1.000000',
          share_percent: null,
        },
      ]}
    />
  )
  const knownUser = within(screen.getByRole('row', { name: /Customer Five/ }))
  expect(knownUser.getAllByText('$ 1.000000')).toHaveLength(2)
  expect(knownUser.getByText('Known cost subtotal (incomplete)')).toBeVisible()
  const missingUser = within(screen.getByRole('row', { name: /Missing User/ }))
  expect(missingUser.getAllByText('Unknown')).toHaveLength(3)
  expect(missingUser.getByText('$ 1.000000')).toBeVisible()
  expect(
    missingUser.getByText('Known cost subtotal (incomplete)')
  ).toBeVisible()
})
