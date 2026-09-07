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
import { fireEvent, render, screen } from '@testing-library/react'
import { useForm } from 'react-hook-form'
import { describe, expect, test } from 'vitest'

import { Form } from '@/components/ui/form'

import {
  CHANNEL_FORM_DEFAULT_VALUES,
  type ChannelFormValues,
} from '../../../lib/channel-form'
import { DoubaoVideoContentModeField } from '../doubao-video-content-mode-field'

function Harness() {
  const form = useForm<ChannelFormValues>({
    defaultValues: {
      ...CHANNEL_FORM_DEFAULT_VALUES,
      video_content_delivery_mode: 'proxy',
    },
  })
  const mode = form.watch('video_content_delivery_mode')

  return (
    <Form {...form}>
      <DoubaoVideoContentModeField control={form.control} />
      <output data-testid='delivery-mode'>{mode}</output>
    </Form>
  )
}

describe('DoubaoVideo content delivery mode field', () => {
  test('switches from server proxy to redirect through the visible control', () => {
    render(<Harness />)

    const proxyButton = screen.getByRole('button', { name: 'Server proxy' })
    const redirectButton = screen.getByRole('button', { name: 'Redirect' })
    expect(proxyButton).toHaveAttribute('aria-pressed', 'true')

    fireEvent.click(redirectButton)

    expect(redirectButton).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByTestId('delivery-mode')).toHaveTextContent('redirect')
  })
})
