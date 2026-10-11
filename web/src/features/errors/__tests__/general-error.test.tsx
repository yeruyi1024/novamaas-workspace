import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, test, vi } from 'vitest'

import { GeneralError } from '../general-error'

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
  useRouter: () => ({ history: { go: vi.fn() } }),
}))

test('an outdated chunk offers a page reload instead of reporting a server 500', async () => {
  const reload = vi.fn()
  vi.stubGlobal('location', { reload })
  try {
    render(
      <GeneralError
        error={Object.assign(new Error('Loading chunk 7704 failed.'), {
          name: 'ChunkLoadError',
        })}
      />
    )
    expect(screen.queryByText('500')).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Refresh' }))
    expect(reload).toHaveBeenCalledOnce()
  } finally {
    vi.unstubAllGlobals()
  }
})

test('an actual server error still displays its HTTP status', () => {
  render(<GeneralError error={{ response: { status: 500 } }} />)
  expect(screen.getByText('500')).toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: 'Refresh' })
  ).not.toBeInTheDocument()
})
