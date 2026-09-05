import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, test } from 'vitest'

import { HeroTerminalDemo } from '../hero-terminal-demo'

test('keyboard selection switches protocol, request and response together without claiming live service metrics', async () => {
  render(<HeroTerminalDemo />)
  const user = userEvent.setup()
  expect(screen.getByRole('button', { name: 'Chat' })).toHaveAttribute(
    'aria-pressed',
    'true'
  )
  screen.getByRole('button', { name: 'Gemini' }).focus()
  await user.keyboard('{Enter}')
  expect(screen.getByRole('button', { name: 'Gemini' })).toHaveAttribute(
    'aria-pressed',
    'true'
  )
  expect(screen.getByRole('button', { name: 'Chat' })).toHaveAttribute(
    'aria-pressed',
    'false'
  )
  const request = screen.getByRole('region', { name: 'Request' })
  expect(request).toHaveTextContent('/v1beta/models/{model}:generateContent')
  expect(request).toHaveAttribute('tabindex', '0')
  expect(
    within(screen.getByRole('region', { name: 'Response' })).getByText(
      /Gemini request served/
    )
  ).toBeVisible()
  expect(
    screen.getByText(
      'Illustrative API examples. Check the model catalog for available services and pricing.'
    )
  ).toBeVisible()
  expect(screen.queryByText(/200 ok/i)).not.toBeInTheDocument()
})
