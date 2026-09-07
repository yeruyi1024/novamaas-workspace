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
import type { Control } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

import type { ChannelFormValues } from '../../lib/channel-form'

type DoubaoVideoContentModeFieldProps = {
  control: Control<ChannelFormValues>
}

export function DoubaoVideoContentModeField(
  props: DoubaoVideoContentModeFieldProps
) {
  const { t } = useTranslation()

  return (
    <FormField
      control={props.control}
      name='video_content_delivery_mode'
      render={({ field }) => {
        const currentMode = field.value === 'redirect' ? 'redirect' : 'proxy'
        return (
          <FormItem className='px-4 py-3'>
            <FormLabel>{t('Video delivery mode')}</FormLabel>
            <FormControl>
              <ToggleGroup
                value={[currentMode]}
                onValueChange={(values) => {
                  const nextMode = values.find((value) => value !== currentMode)
                  if (nextMode === 'proxy' || nextMode === 'redirect') {
                    field.onChange(nextMode)
                  }
                }}
                aria-label={t('Video delivery mode')}
                variant='outline'
                spacing={2}
                className='grid w-full grid-cols-2'
              >
                <ToggleGroupItem value='proxy' className='w-full'>
                  {t('Server proxy')}
                </ToggleGroupItem>
                <ToggleGroupItem value='redirect' className='w-full'>
                  {t('Redirect')}
                </ToggleGroupItem>
              </ToggleGroup>
            </FormControl>
            <FormDescription>
              {t(
                'Redirect exposes the public upstream video URL to clients. Server proxy keeps the existing behavior with a 600-second timeout.'
              )}
            </FormDescription>
            <FormMessage />
          </FormItem>
        )
      }}
    />
  )
}
