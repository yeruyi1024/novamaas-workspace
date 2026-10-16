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
import { beforeEach, expect, test, vi } from 'vitest'

import { getUpstreamChannels, saveUpstreamGroup } from '../api'
import { GroupEditor } from '../components/group-editor'
import { channels, group, renderUpstream } from './fixtures'

vi.mock('../api', async (original) => ({
  ...(await original<typeof import('../api')>()),
  getUpstreamChannels: vi.fn(),
  saveUpstreamGroup: vi.fn(),
}))
beforeEach(() => {
  vi.resetAllMocks()
  vi.mocked(getUpstreamChannels).mockResolvedValue(channels)
  vi.mocked(saveUpstreamGroup).mockResolvedValue(group)
})

test('creating an accounting group combines selected channels including zero-cost channels and blocks occupied members', async () => {
  const close = vi.fn()
  const saved = vi.fn()
  renderUpstream(<GroupEditor group={null} onClose={close} onSaved={saved} />)
  fireEvent.change(screen.getByLabelText('Upstream name'), {
    target: { value: ' Shared supplier ' },
  })
  expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled()
  const primary = await screen.findByRole('checkbox', {
    name: /Primary channel/,
  })
  const zero = screen.getByRole('checkbox', { name: /Zero-cost channel/ })
  expect(zero).toBeEnabled()
  expect(
    screen.getByRole('checkbox', { name: /Assigned channel/ })
  ).toHaveAttribute('aria-disabled', 'true')
  expect(
    screen.getByRole('checkbox', { name: /Unconfigured channel/ })
  ).toBeEnabled()
  fireEvent.click(primary)
  fireEvent.click(zero)
  fireEvent.click(screen.getByRole('button', { name: 'Save' }))
  await waitFor(() =>
    expect(saveUpstreamGroup).toHaveBeenCalledWith({
      id: undefined,
      name: 'Shared supplier',
      version: 0,
      channel_ids: [19, 26],
    })
  )
  expect(saved).toHaveBeenCalledWith(1)
  expect(close).toHaveBeenCalled()
})

test('tag filtering preserves existing selections and updates using the group version', async () => {
  const user = userEvent.setup()
  renderUpstream(
    <GroupEditor group={group} onClose={vi.fn()} onSaved={vi.fn()} />
  )
  await screen.findByRole('checkbox', { name: /Primary channel/ })
  await user.click(screen.getByRole('combobox', { name: 'Tag' }))
  await user.click(screen.getByRole('option', { name: 'demo' }))
  expect(
    screen.queryByRole('checkbox', { name: /Unconfigured channel/ })
  ).not.toBeInTheDocument()
  expect(
    screen.getByRole('checkbox', { name: /Zero-cost channel/ })
  ).toBeChecked()
  await user.click(screen.getByRole('button', { name: 'Save' }))
  await waitFor(() =>
    expect(saveUpstreamGroup).toHaveBeenCalledWith({
      id: 1,
      name: 'Demo upstream',
      version: 3,
      channel_ids: [19, 26, 27],
    })
  )
})

test('a channel list failure preserves the edit and disables saving until retry succeeds', async () => {
  vi.mocked(getUpstreamChannels).mockRejectedValueOnce(new Error('Unavailable'))
  renderUpstream(
    <GroupEditor group={group} onClose={vi.fn()} onSaved={vi.fn()} />
  )
  await screen.findByRole('alert')
  expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled()
  fireEvent.click(screen.getByRole('button', { name: 'Retry' }))
  await screen.findByRole('checkbox', { name: /Zero-cost channel/ })
  expect(screen.getByRole('button', { name: 'Save' })).toBeEnabled()
})
