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
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, test, vi } from 'vitest'

import { updateSystemOption } from '../../api'
import { SettingsPageProvider } from '../../components/settings-page-context'
import { HEADER_NAV_DEFAULT, serializeHeaderNavModules } from '../config'
import { HeaderNavigationSection } from '../header-navigation-section'

vi.mock('../../api', () => ({ updateSystemOption: vi.fn() }))

let client: QueryClient
let actions: HTMLDivElement

afterEach(() => {
  client?.clear()
  actions?.remove()
  vi.resetAllMocks()
})

test.each([true, false])(
  'saving the image creation toggle from %s updates the existing feature switch without changing navigation configuration',
  async (enabled) => {
    vi.mocked(updateSystemOption).mockResolvedValue({
      success: true,
      message: '',
    })
    client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    })
    actions = document.createElement('div')
    document.body.append(actions)
    render(
      <QueryClientProvider client={client}>
        <SettingsPageProvider actionsContainer={actions}>
          <HeaderNavigationSection
            config={HEADER_NAV_DEFAULT}
            initialSerialized={serializeHeaderNavModules(HEADER_NAV_DEFAULT)}
            imageStudioEnabled={enabled}
          />
        </SettingsPageProvider>
      </QueryClientProvider>
    )
    const user = userEvent.setup()
    const toggle = screen.getByRole('switch', { name: 'Image creation' })
    expect(toggle).toHaveAttribute('aria-checked', String(enabled))
    await user.click(toggle)
    await user.click(screen.getByRole('button', { name: 'Save navigation' }))
    await waitFor(() =>
      expect(updateSystemOption).toHaveBeenCalledWith({
        key: 'image_studio_setting.enabled',
        value: !enabled,
      })
    )
    expect(updateSystemOption).toHaveBeenCalledOnce()
  }
)
