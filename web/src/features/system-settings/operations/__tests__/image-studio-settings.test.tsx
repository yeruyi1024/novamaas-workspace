import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  createMemoryHistory,
  createRootRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, test, vi } from 'vitest'

import { updateSystemOption } from '../../api'
import { SettingsPageProvider } from '../../components/settings-page-context'
import { ImageStudioSettingsSection } from '../image-studio-settings-section'

vi.mock('../../api', () => ({ updateSystemOption: vi.fn() }))

const clients: QueryClient[] = []
const containers: HTMLElement[] = []

afterEach(() => {
  for (const client of clients) client.clear()
  for (const container of containers) container.remove()
  clients.length = 0
  containers.length = 0
  vi.clearAllMocks()
})

async function renderSettings() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  const actions = document.createElement('div')
  document.body.append(actions)
  clients.push(client)
  containers.push(actions)
  const root = createRootRoute({
    component: () => (
      <QueryClientProvider client={client}>
        <SettingsPageProvider actionsContainer={actions}>
          <ImageStudioSettingsSection
            defaultValues={{
              'image_studio_setting.enabled': true,
            }}
          />
        </SettingsPageProvider>
      </QueryClientProvider>
    ),
  })
  const router = createRouter({
    routeTree: root,
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  await router.load()
  render(<RouterProvider router={router} />)
}

describe('root image studio settings', () => {
  test('explains required OSS storage and links to the profile and permanent policy configuration', async () => {
    await renderSettings()
    expect(
      screen.getByRole('link', { name: 'Object Storage' })
    ).toHaveAttribute('href', '/system-settings/storage/object-storage')
    expect(
      screen.getByRole('link', { name: 'Permanent asset storage' })
    ).toHaveAttribute('href', '/system-settings/storage/asset-library')
    expect(
      screen.getByText(
        'Image generation requires an enabled private OSS profile and permanent storage policy. Configure and test the profile in Object Storage, then select it and enable permanent storage under Asset Library.'
      )
    ).toBeVisible()
    expect(
      screen.getByText(
        'Generated images are stored separately from library assets. Asset library integration controls reference images and manual saving only. Missing OSS configuration blocks generation before billing.'
      )
    ).toBeVisible()
  })

  test.each([
    ['Show image generation entry', 'visible'],
    ['Enable asset library integration', 'asset_library_enabled'],
    ['Enable AI Creation Studio', 'enabled'],
  ])('root can turn off %s independently', async (label, field) => {
    vi.mocked(updateSystemOption).mockResolvedValue({
      success: true,
      message: '',
    })
    const user = userEvent.setup()
    await renderSettings()
    await user.click(screen.getByRole('switch', { name: label }))
    await user.click(screen.getByRole('button', { name: 'Save Changes' }))
    await waitFor(() =>
      expect(updateSystemOption).toHaveBeenCalledWith({
        key: `image_studio_setting.${field}`,
        value: false,
      })
    )
    expect(updateSystemOption).toHaveBeenCalledOnce()
    if (field !== 'enabled') {
      expect(
        screen.getByRole('switch', { name: 'Enable AI Creation Studio' })
      ).toBeChecked()
    }
  })
})
