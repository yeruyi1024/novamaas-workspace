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
import { fireEvent, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'

import {
  downloadUpstreamStatement,
  getUpstreamReport,
  getUpstreamStatement,
  getUpstreamStatements,
  getUpstreamStorageProfiles,
  prepareUpstreamStatement,
  upstreamStatementAction,
} from '../api'
import { StatementDetail } from '../components/statement-detail'
import { StatementsPanel } from '../components/statements-panel'
import { detail, renderUpstream, report } from './fixtures'

vi.mock('../api', async (original) => ({
  ...(await original<typeof import('../api')>()),
  downloadUpstreamStatement: vi.fn(),
  getUpstreamReport: vi.fn(),
  getUpstreamStatement: vi.fn(),
  getUpstreamStatements: vi.fn(),
  getUpstreamStorageProfiles: vi.fn(),
  prepareUpstreamStatement: vi.fn(),
  upstreamStatementAction: vi.fn(),
}))
beforeEach(() => {
  vi.resetAllMocks()
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date('2026-10-15T12:00:00+08:00'))
  useAuthStore.getState().auth.setBundle({
    access_token: 'local-test-session',
    token_type: 'Bearer',
    access_expires_at: 9999999999,
    user: { id: 1, username: 'accountant', role: 100 },
    session: {
      sid: 'browser-session',
      current: true,
      login_method: 'password',
      ip: '',
      user_agent: '',
      created_at: 1,
      last_active_at: 1,
      expires_at: 9999999999,
    },
  })
  vi.mocked(getUpstreamStatement).mockResolvedValue(detail)
  vi.mocked(getUpstreamReport).mockResolvedValue(report)
  vi.mocked(getUpstreamStatements).mockResolvedValue([])
  vi.mocked(getUpstreamStorageProfiles).mockResolvedValue([
    { id: 4, name: 'Archive OSS' },
  ])
  vi.mocked(prepareUpstreamStatement).mockResolvedValue(detail.statement)
  vi.mocked(upstreamStatementAction).mockResolvedValue(detail.statement)
})
afterEach(() => {
  useAuthStore.getState().auth.reset()
  vi.useRealTimers()
})

test('internal confirmation requires acknowledgement, a browser session and a second confirmation bound to the frozen manifest', async () => {
  renderUpstream(<StatementDetail id='frozen-statement' onClose={vi.fn()} />)
  const confirm = await screen.findByRole('button', {
    name: 'Confirm statement',
  })
  expect(confirm).toBeDisabled()
  fireEvent.click(
    screen.getByRole('checkbox', {
      name: 'I have checked this statement and agree to confirm it.',
    })
  )
  fireEvent.click(confirm)
  expect(upstreamStatementAction).not.toHaveBeenCalled()
  fireEvent.click(screen.getByRole('button', { name: 'Confirm' }))
  await waitFor(() =>
    expect(upstreamStatementAction).toHaveBeenCalledWith('frozen-statement', {
      action: 'confirm',
      manifest_sha256: 'frozen-manifest',
      evidence_sha256: detail.evidence_sha256,
      reason: '',
      acknowledged: true,
    })
  )
})

test('a PAT-like user without a browser session cannot acknowledge or confirm an archived draft', async () => {
  useAuthStore.getState().auth.reset()
  useAuthStore
    .getState()
    .auth.setUser({ id: 1, username: 'accountant', role: 100 })
  renderUpstream(<StatementDetail id='frozen-statement' onClose={vi.fn()} />)
  expect(
    await screen.findByRole('button', { name: 'Confirm statement' })
  ).toBeDisabled()
  expect(screen.getByRole('checkbox')).toHaveAttribute('aria-disabled', 'true')
  expect(
    screen.getByText(
      'Sign in with a browser session to confirm this statement.'
    )
  ).toBeVisible()
  expect(upstreamStatementAction).not.toHaveBeenCalled()
})

test('voiding an unconfirmed statement requires a reason and records the explicit action', async () => {
  renderUpstream(<StatementDetail id='frozen-statement' onClose={vi.fn()} />)
  const button = await screen.findByRole('button', { name: 'Void statement' })
  expect(button).toBeDisabled()
  fireEvent.change(screen.getByLabelText('Review note'), {
    target: { value: ' Correct channel membership ' },
  })
  fireEvent.click(button)
  fireEvent.click(screen.getByRole('button', { name: 'Confirm' }))
  await waitFor(() =>
    expect(upstreamStatementAction).toHaveBeenCalledWith('frozen-statement', {
      action: 'void',
      manifest_sha256: 'frozen-manifest',
      reason: 'Correct channel membership',
      acknowledged: false,
    })
  )
})

test('confirmed archives are read-only and allow confirmation receipts even before the artifact exists', async () => {
  vi.mocked(getUpstreamStatement).mockResolvedValue({
    ...detail,
    statement: {
      ...detail.statement,
      status: 'confirmed',
      confirmed_at: 1791936500,
    },
  })
  renderUpstream(<StatementDetail id='frozen-statement' onClose={vi.fn()} />)
  const receipt = await screen.findByRole('button', {
    name: 'Download confirmation receipt',
  })
  expect(
    screen.queryByRole('button', { name: 'Void statement' })
  ).not.toBeInTheDocument()
  expect(screen.queryByRole('checkbox')).not.toBeInTheDocument()
  fireEvent.click(receipt)
  await waitFor(() =>
    expect(downloadUpstreamStatement).toHaveBeenCalledWith(
      'frozen-statement',
      'receipt',
      undefined
    )
  )
})

test('failed archive jobs offer a retry without changing the frozen statement identifier', async () => {
  vi.mocked(getUpstreamStatement).mockResolvedValue({
    ...detail,
    statement: {
      ...detail.statement,
      status: 'failed',
      last_error: 'storage unavailable',
    },
  })
  renderUpstream(<StatementDetail id='frozen-statement' onClose={vi.fn()} />)
  fireEvent.click(await screen.findByRole('button', { name: 'Retry' }))
  await waitFor(() =>
    expect(upstreamStatementAction).toHaveBeenCalledWith('frozen-statement', {
      action: 'retry',
      manifest_sha256: 'frozen-manifest',
      reason: '',
      acknowledged: false,
    })
  )
})

test.each([
  { field: 'missing_cost_records', label: 'missing historical cost' },
  { field: 'missing_usage_records', label: 'incomplete usage' },
  { field: 'estimated_usage_records', label: 'estimated usage' },
  { field: 'pending_tasks', label: 'pending task' },
  { field: 'missing_channels', label: 'missing channel' },
] as const)(
  '$label remains a warning while the selected storage permits draft creation',
  async (scenario) => {
    const user = userEvent.setup()
    vi.mocked(getUpstreamReport).mockResolvedValue({
      ...report,
      coverage: { ...report.coverage, complete: false, [scenario.field]: 1 },
    })
    renderUpstream(<StatementsPanel groupId={1} />)
    await screen.findByText('Reconciliation warnings')
    expect(
      screen.getByRole('button', { name: 'Create archived draft' })
    ).toBeDisabled()
    await user.click(screen.getByRole('combobox', { name: 'Archive storage' }))
    await user.click(screen.getByRole('option', { name: 'Archive OSS' }))
    expect(
      screen.getByRole('button', { name: 'Create archived draft' })
    ).toBeEnabled()
    expect(screen.getByText(/Warnings do not block/)).toBeVisible()
    await user.click(
      screen.getByRole('button', { name: 'Create archived draft' })
    )
    await waitFor(() =>
      expect(prepareUpstreamStatement).toHaveBeenCalledWith(1, report.month, 4)
    )
  }
)

test('warnings still require an enabled archive storage profile before creating a draft', async () => {
  vi.mocked(getUpstreamStorageProfiles).mockResolvedValue([])
  vi.mocked(getUpstreamReport).mockResolvedValue({
    ...report,
    coverage: { ...report.coverage, complete: false, missing_usage_records: 2 },
  })
  renderUpstream(<StatementsPanel groupId={1} />)
  await screen.findByText('Reconciliation warnings')
  expect(
    screen.getByText('Configure an enabled Aliyun OSS storage profile first.')
  ).toBeVisible()
  expect(
    screen.getByRole('button', { name: 'Create archived draft' })
  ).toBeDisabled()
  expect(prepareUpstreamStatement).not.toHaveBeenCalled()
})

test('a closed-month report with no eligible usage can archive its channel warnings and empty evidence', async () => {
  const user = userEvent.setup()
  vi.mocked(getUpstreamReport).mockResolvedValue({
    ...report,
    days: [],
    models: [],
    day_models: [],
    channel_users: [],
    total: { ...report.total, count: 0, requests: 0, amount: '0.000000' },
    coverage: { ...report.coverage, complete: false, missing_channels: 1 },
  })
  renderUpstream(<StatementsPanel groupId={1} />)
  await screen.findByText('Reconciliation warnings')
  await user.click(screen.getByRole('combobox', { name: 'Archive storage' }))
  await user.click(screen.getByRole('option', { name: 'Archive OSS' }))
  await user.click(
    screen.getByRole('button', { name: 'Create archived draft' })
  )
  await waitFor(() =>
    expect(prepareUpstreamStatement).toHaveBeenCalledWith(1, report.month, 4)
  )
})

test('a reviewer can confirm a frozen statement with coverage warnings using the existing acknowledgement and digests', async () => {
  vi.mocked(getUpstreamStatement).mockResolvedValue({
    ...detail,
    snapshot: {
      ...report,
      coverage: {
        ...report.coverage,
        complete: false,
        missing_usage_records: 2,
      },
    },
  })
  renderUpstream(<StatementDetail id={detail.statement.id} onClose={vi.fn()} />)
  await screen.findByText('Reconciliation warnings')
  const acknowledged = screen.getByRole('checkbox', {
    name: 'I have checked this statement and agree to confirm it.',
  })
  expect(acknowledged).not.toHaveAttribute('aria-disabled', 'true')
  fireEvent.click(acknowledged)
  fireEvent.click(screen.getByRole('button', { name: 'Confirm statement' }))
  expect(
    screen.getByText(
      'Reconciliation warnings remain in this archive after confirmation.'
    )
  ).toBeVisible()
  fireEvent.click(screen.getByRole('button', { name: 'Confirm' }))
  await waitFor(() =>
    expect(upstreamStatementAction).toHaveBeenCalledWith(detail.statement.id, {
      action: 'confirm',
      manifest_sha256: detail.statement.manifest_sha256,
      evidence_sha256: detail.evidence_sha256,
      reason: '',
      acknowledged: true,
    })
  )
})

test('confirmed archives retain their frozen coverage warnings and download controls', async () => {
  vi.mocked(getUpstreamStatement).mockResolvedValue({
    ...detail,
    statement: { ...detail.statement, status: 'confirmed' },
    snapshot: {
      ...report,
      total: { ...report.total, estimated_usage_records: 98 },
      coverage: {
        ...report.coverage,
        complete: false,
        estimated_usage_records: 98,
      },
    },
  })
  renderUpstream(<StatementDetail id={detail.statement.id} onClose={vi.fn()} />)
  const warning = await screen.findByRole('alert')
  expect(warning).toHaveTextContent('Reconciliation warnings')
  expect(warning).toHaveTextContent('Estimated usage: 98')
  expect(
    screen.getByRole('button', { name: 'Download confirmation receipt' })
  ).toBeEnabled()
  expect(screen.queryByRole('checkbox')).not.toBeInTheDocument()
})

test('a complete closed-month preview creates a draft with the selected accounting group and OSS profile', async () => {
  const user = userEvent.setup()
  renderUpstream(<StatementsPanel groupId={1} />)
  fireEvent.change(screen.getByLabelText('Statement month'), {
    target: { value: '2026-09' },
  })
  await screen.findByText('Coverage complete')
  await user.click(screen.getByRole('combobox', { name: 'Archive storage' }))
  await user.click(screen.getByRole('option', { name: 'Archive OSS' }))
  await user.click(
    screen.getByRole('button', { name: 'Create archived draft' })
  )
  await waitFor(() =>
    expect(prepareUpstreamStatement).toHaveBeenCalledWith(1, '2026-09', 4)
  )
  await screen.findByRole('dialog', { name: 'Upstream statement' })
})

test('the first day of a new month blocks archival until the month has been closed for 24 hours', async () => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date('2026-10-01T12:00:00+08:00'))
  renderUpstream(<StatementsPanel groupId={1} />)
  await screen.findByText(
    'Select a month that has been closed for at least 24 hours to create an archived draft.'
  )
  expect(
    screen.getByRole('button', { name: 'Create archived draft' })
  ).toBeDisabled()
  expect(getUpstreamReport).not.toHaveBeenCalled()
})

test('archives with more than 200 detail chunks allow authorized downloads beyond the metadata window', async () => {
  vi.mocked(getUpstreamStatement).mockResolvedValue({
    ...detail,
    detail_count: 205,
    artifacts: [
      { id: 9, kind: 'details', ordinal: 0, rows: 500, sha256: 'part-1' },
    ],
  })
  renderUpstream(<StatementDetail id='frozen-statement' onClose={vi.fn()} />)
  const part = await screen.findByRole('spinbutton', {
    name: 'Archived detail files (205)',
  })
  expect(part).toHaveAttribute('max', '205')
  fireEvent.change(part, { target: { value: '201' } })
  fireEvent.click(screen.getByRole('button', { name: 'Download' }))
  await waitFor(() =>
    expect(downloadUpstreamStatement).toHaveBeenCalledWith(
      'frozen-statement',
      'details',
      200
    )
  )
  fireEvent.change(part, { target: { value: '206' } })
  expect(screen.getByRole('button', { name: 'Download' })).toBeDisabled()
})

test('preparing statements cannot be voided and their creation audit event has a readable action', async () => {
  vi.mocked(getUpstreamStatement).mockResolvedValue({
    ...detail,
    statement: { ...detail.statement, status: 'preparing' },
    events: [
      {
        id: 3,
        actor_id: 1,
        actor_username: 'accountant',
        action: 'create',
        created_at: 1791936000,
      },
    ],
  })
  renderUpstream(<StatementDetail id='frozen-statement' onClose={vi.fn()} />)
  await screen.findByRole('dialog', { name: 'Upstream statement' })
  await screen.findByText(/Create archived draft/)
  expect(
    screen.queryByRole('button', { name: 'Void statement' })
  ).not.toBeInTheDocument()
})
