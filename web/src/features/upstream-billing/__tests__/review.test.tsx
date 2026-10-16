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
import { act, fireEvent, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'

import {
  downloadUpstreamAttachment,
  getUpstreamStatement,
  uploadUpstreamAttachment,
  upstreamStatementAction,
} from '../api'
import { StatementDetail } from '../components/statement-detail'
import type { UpstreamAttachment } from '../types'
import { detail, renderUpstream, report } from './fixtures'

vi.mock('../api', async (original) => ({
  ...(await original<typeof import('../api')>()),
  getUpstreamStatement: vi.fn(),
  uploadUpstreamAttachment: vi.fn(),
  downloadUpstreamAttachment: vi.fn(),
  upstreamStatementAction: vi.fn(),
}))
const attachment: UpstreamAttachment = {
  id: 'af1557fe-2623-4f8c-9584-b78807acd540',
  statement_id: detail.statement.id,
  original_filename: 'supplier.pdf',
  content_type: 'application/pdf',
  size: 1024,
  sha256: 'a'.repeat(64),
  storage_profile_id: 4,
  uploaded_by: 1,
  uploaded_username: 'reviewer',
  uploaded_at: 1791936000,
  status: 'ready',
}

beforeEach(() => {
  vi.resetAllMocks()
  useAuthStore.getState().auth.setBundle({
    access_token: 'local-session',
    token_type: 'Bearer',
    access_expires_at: 9999999999,
    user: {
      id: 1,
      username: 'reviewer',
      role: 10,
      permissions: {
        admin_permissions: { financial_management: { view: true } },
      },
    },
    session: {
      sid: 'browser',
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
  vi.mocked(uploadUpstreamAttachment).mockResolvedValue(attachment)
})
afterEach(() => useAuthStore.getState().auth.reset())

test('accounting viewers can download evidence without upload, confirmation, void or review controls', async () => {
  useAuthStore.getState().auth.setUser({
    id: 1,
    username: 'accountant',
    role: 10,
    permissions: {
      admin_permissions: { financial_accounting: { view: true } },
    },
  })
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
    attachments: [attachment],
  })
  renderUpstream(<StatementDetail id={detail.statement.id} onClose={vi.fn()} />)
  fireEvent.click(
    await screen.findByRole('button', { name: 'Download attachment' })
  )
  await waitFor(() =>
    expect(downloadUpstreamAttachment).toHaveBeenCalledWith(
      detail.statement.id,
      attachment.id,
      'supplier.pdf'
    )
  )
  expect(
    screen.queryByRole('button', { name: 'Upload attachment' })
  ).not.toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: 'Confirm statement' })
  ).not.toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: 'Void statement' })
  ).not.toBeInTheDocument()
  expect(
    screen.queryByRole('textbox', { name: 'Review note' })
  ).not.toBeInTheDocument()
  expect(screen.queryByRole('checkbox')).not.toBeInTheDocument()
})

test('uploading evidence refreshes the attachment audit fields and clears the previous review acknowledgement', async () => {
  const user = userEvent.setup()
  vi.mocked(getUpstreamStatement)
    .mockResolvedValueOnce(detail)
    .mockResolvedValue({
      ...detail,
      attachments: [attachment],
      evidence_sha256: 'f'.repeat(64),
    })
  renderUpstream(<StatementDetail id={detail.statement.id} onClose={vi.fn()} />)
  const check = await screen.findByRole('checkbox')
  fireEvent.click(check)
  expect(
    screen.getByRole('button', { name: 'Confirm statement' })
  ).toBeEnabled()
  const file = new File(['supplier evidence'], 'supplier.pdf', {
    type: 'application/pdf',
  })
  await user.upload(screen.getByLabelText('Choose supplier attachment'), file)
  fireEvent.click(screen.getByRole('button', { name: 'Upload attachment' }))
  await waitFor(() =>
    expect(uploadUpstreamAttachment).toHaveBeenCalledWith(
      detail.statement.id,
      file
    )
  )
  await screen.findByText('supplier.pdf')
  expect(screen.getByText(/reviewer/)).toBeVisible()
  expect(screen.getByText(`SHA-256: ${attachment.sha256}`)).toBeVisible()
  expect(check).not.toBeChecked()
  expect(
    screen.getByRole('button', { name: 'Confirm statement' })
  ).toBeDisabled()
  expect(upstreamStatementAction).not.toHaveBeenCalled()
})

test('an evidence digest change during the second confirmation blocks submitting either the old or silently updated digest', async () => {
  const { client } = renderUpstream(
    <StatementDetail id={detail.statement.id} onClose={vi.fn()} />
  )
  fireEvent.click(await screen.findByRole('checkbox'))
  fireEvent.click(screen.getByRole('button', { name: 'Confirm statement' }))
  await act(async () => {
    client.setQueryData(
      ['upstream-billing', 'statement', detail.statement.id],
      { ...detail, evidence_sha256: 'f'.repeat(64) }
    )
  })
  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Supplier evidence changed.'
  )
  expect(screen.getByRole('button', { name: 'Confirm' })).toBeDisabled()
  fireEvent.click(screen.getByRole('button', { name: 'Confirm' }))
  expect(upstreamStatementAction).not.toHaveBeenCalled()
})

test('pending uploads block review confirmation even while the frozen accounting coverage is complete', async () => {
  vi.mocked(getUpstreamStatement).mockResolvedValue({
    ...detail,
    pending_uploads: 1,
  })
  renderUpstream(<StatementDetail id={detail.statement.id} onClose={vi.fn()} />)
  expect(await screen.findByRole('checkbox')).toHaveAttribute(
    'aria-disabled',
    'true'
  )
  expect(
    screen.getByRole('button', { name: 'Confirm statement' })
  ).toBeDisabled()
  expect(screen.getByRole('status')).toHaveTextContent(
    'Attachment uploads are in progress.'
  )
  fireEvent.change(screen.getByRole('textbox', { name: 'Review note' }), {
    target: { value: 'Review interrupted' },
  })
  expect(screen.getByRole('button', { name: 'Void statement' })).toBeDisabled()
})

test('confirmed statements keep supplier evidence readable and offer no upload inputs', async () => {
  vi.mocked(getUpstreamStatement).mockResolvedValue({
    ...detail,
    statement: { ...detail.statement, status: 'confirmed' },
    attachments: [attachment],
  })
  renderUpstream(<StatementDetail id={detail.statement.id} onClose={vi.fn()} />)
  await screen.findByRole('button', { name: 'Download attachment' })
  expect(
    screen.queryByLabelText('Choose supplier attachment')
  ).not.toBeInTheDocument()
})

test('oversized or unsupported files cannot be uploaded', async () => {
  const user = userEvent.setup({ applyAccept: false })
  renderUpstream(<StatementDetail id={detail.statement.id} onClose={vi.fn()} />)
  const input = await screen.findByLabelText('Choose supplier attachment')
  const oversized = new File(['x'], 'large.pdf', { type: 'application/pdf' })
  Object.defineProperty(oversized, 'size', { value: 20 * 1024 * 1024 + 1 })
  await user.upload(input, oversized)
  expect(
    screen.getByText('Each attachment must be 20 MiB or smaller.')
  ).toHaveAttribute('role', 'alert')
  expect(
    screen.getByRole('button', { name: 'Upload attachment' })
  ).toBeDisabled()
  await user.upload(
    input,
    new File(['x'], 'script.html', { type: 'text/html' })
  )
  expect(
    screen.getByText('Choose a PDF, XLSX, CSV, PNG or JPEG attachment.')
  ).toHaveAttribute('role', 'alert')
  expect(uploadUpstreamAttachment).not.toHaveBeenCalled()
})

test('the ready and pending evidence count together enforce the twenty attachment limit', async () => {
  vi.mocked(getUpstreamStatement).mockResolvedValue({
    ...detail,
    pending_uploads: 1,
    attachments: Array.from({ length: 19 }, (_, i) => ({
      ...attachment,
      id: `attachment-${i}`,
    })),
  })
  renderUpstream(<StatementDetail id={detail.statement.id} onClose={vi.fn()} />)
  expect(
    await screen.findByLabelText('Choose supplier attachment')
  ).toBeDisabled()
  expect(
    screen.getByRole('button', { name: 'Upload attachment' })
  ).toBeDisabled()
  expect(
    screen.getByText('A statement can have at most 20 supplier attachments.')
  ).toHaveAttribute('role', 'alert')
})

test('local evidence uploads block voiding before the server pending counter refreshes', async () => {
  const user = userEvent.setup()
  let finish!: (attachment: UpstreamAttachment) => void
  vi.mocked(uploadUpstreamAttachment).mockImplementation(
    () =>
      new Promise((resolve) => {
        finish = resolve
      })
  )
  renderUpstream(<StatementDetail id={detail.statement.id} onClose={vi.fn()} />)
  fireEvent.change(
    await screen.findByRole('textbox', { name: 'Review note' }),
    { target: { value: 'Replace statement' } }
  )
  const voidButton = screen.getByRole('button', { name: 'Void statement' })
  expect(voidButton).toBeEnabled()
  await user.upload(
    screen.getByLabelText('Choose supplier attachment'),
    new File(['evidence'], 'supplier.pdf', { type: 'application/pdf' })
  )
  fireEvent.click(screen.getByRole('button', { name: 'Upload attachment' }))
  await waitFor(() => expect(voidButton).toBeDisabled())
  fireEvent.click(voidButton)
  expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument()
  expect(upstreamStatementAction).not.toHaveBeenCalled()
  await act(async () => finish(attachment))
  await waitFor(() => expect(voidButton).toBeEnabled())
})
