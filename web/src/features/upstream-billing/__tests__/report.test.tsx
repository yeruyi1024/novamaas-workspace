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
import userEvent from '@testing-library/user-event'
import { beforeEach, expect, test, vi } from 'vitest'

import { getUpstreamReport, getUpstreamUsageDetails } from '../api'
import { ReportPanel } from '../components/report-panel'
import { renderUpstream, report } from './fixtures'

vi.mock('../api', async (original) => ({
  ...(await original<typeof import('../api')>()),
  getUpstreamReport: vi.fn(),
  getUpstreamUsageDetails: vi.fn(),
}))
beforeEach(() => {
  vi.resetAllMocks()
  vi.mocked(getUpstreamReport).mockResolvedValue(report)
  vi.mocked(getUpstreamUsageDetails).mockResolvedValue({
    records: [],
    currency: report.currency,
  })
})

test('report displays backend decimal costs, excludes unconfigured channels and drills into day and model evidence', async () => {
  const user = userEvent.setup()
  renderUpstream(<ReportPanel groupId={1} />)
  const day = await screen.findByRole('button', { name: '2026-09-03' })
  expect(
    screen.getByText('Excluded channels: Unconfigured channel (#27)')
  ).toBeVisible()
  expect(
    within(
      screen.getByRole('table', { name: 'Daily upstream costs' })
    ).getAllByText('$ 1.000000')
  ).toHaveLength(2)
  await user.click(day)
  await screen.findByRole('table', { name: 'Upstream request details' })
  expect(getUpstreamUsageDetails).toHaveBeenLastCalledWith(
    1,
    expect.objectContaining({ month: expect.any(String) }),
    { day: '2026-09-03', cursor: undefined }
  )
  expect(
    screen.getByRole('table', { name: 'Upstream model usage' })
  ).toBeVisible()
  await user.click(
    within(screen.getByRole('dialog')).getAllByRole('button', {
      name: 'Close',
    })[0]
  )
  await user.click(screen.getByRole('tab', { name: 'Upstream model usage' }))
  await user.click(screen.getByRole('button', { name: 'gpt-5' }))
  await waitFor(() =>
    expect(getUpstreamUsageDetails).toHaveBeenLastCalledWith(
      1,
      expect.objectContaining({ month: expect.any(String) }),
      { model: 'gpt-5', cursor: undefined }
    )
  )
})

test('custom inclusive date ranges query the selected period and invalid ranges stop requests', async () => {
  const user = userEvent.setup()
  renderUpstream(<ReportPanel groupId={1} />)
  await screen.findByRole('table', { name: 'Daily upstream costs' })
  await user.click(screen.getByRole('button', { name: 'Date range' }))
  fireEvent.change(screen.getByLabelText('Start date (Asia/Shanghai)'), {
    target: { value: '2026-08-01' },
  })
  fireEvent.change(screen.getByLabelText('End date (Asia/Shanghai)'), {
    target: { value: '2026-08-31' },
  })
  await waitFor(() =>
    expect(getUpstreamReport).toHaveBeenLastCalledWith(1, {
      start_date: '2026-08-01',
      end_date: '2026-08-31',
    })
  )
  const calls = vi.mocked(getUpstreamReport).mock.calls.length
  fireEvent.change(screen.getByLabelText('Start date (Asia/Shanghai)'), {
    target: { value: '2026-09-01' },
  })
  expect(screen.getByRole('alert')).toHaveTextContent(
    'Select a valid date range.'
  )
  expect(screen.getByRole('button', { name: 'Refresh' })).toBeDisabled()
  expect(getUpstreamReport).toHaveBeenCalledTimes(calls)
})

test('report failure removes stale data and refresh retries the same selected group', async () => {
  vi.mocked(getUpstreamReport).mockRejectedValueOnce(new Error('Unavailable'))
  renderUpstream(<ReportPanel groupId={1} />)
  await screen.findByRole('alert')
  expect(screen.queryByRole('table')).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: 'Refresh' }))
  await screen.findByRole('table', { name: 'Daily upstream costs' })
})

test('daily rows and the total remain unknown when no channel is included in cost statistics', async () => {
  const empty = {
    ...report.total,
    label: '2026-09-03',
    count: 0,
    requests: 0,
    charge: '0.000000',
    refund: '0.000000',
    amount: '0.000000',
  }
  vi.mocked(getUpstreamReport).mockResolvedValue({
    ...report,
    group: {
      ...report.group,
      channels: report.group.channels.filter((channel) => !channel.eligible),
    },
    days: [empty],
    total: { ...empty, label: 'Total' },
    coverage: { ...report.coverage, complete: false },
  })
  renderUpstream(<ReportPanel groupId={1} />)
  const table = await screen.findByRole('table', {
    name: 'Daily upstream costs',
  })
  expect(within(table).getAllByText('Unknown')).toHaveLength(6)
  expect(within(table).queryByText('$ 0.000000')).not.toBeInTheDocument()
  expect(
    within(table).getAllByText('Not included in cost statistics')
  ).toHaveLength(6)
})

test('channel user allocation uses the same selected inclusive range as the cost preview', async () => {
  const user = userEvent.setup()
  renderUpstream(<ReportPanel groupId={1} />)
  await screen.findByRole('table', { name: 'Daily upstream costs' })
  await user.click(screen.getByRole('button', { name: 'Date range' }))
  fireEvent.change(screen.getByLabelText('Start date (Asia/Shanghai)'), {
    target: { value: '2026-08-01' },
  })
  fireEvent.change(screen.getByLabelText('End date (Asia/Shanghai)'), {
    target: { value: '2026-08-31' },
  })
  await waitFor(() =>
    expect(getUpstreamReport).toHaveBeenLastCalledWith(1, {
      start_date: '2026-08-01',
      end_date: '2026-08-31',
    })
  )
  await user.click(
    await screen.findByRole('tab', { name: 'User cost allocation by channel' })
  )
  expect(
    await screen.findByRole('table', {
      name: 'User cost allocation by channel',
    })
  ).toHaveTextContent('100.0000%')
  expect(getUpstreamUsageDetails).not.toHaveBeenCalled()
})
