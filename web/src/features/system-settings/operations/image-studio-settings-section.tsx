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
import { zodResolver } from '@hookform/resolvers/zod'
import { Link } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod'

import { Form, FormControl, FormField, FormLabel } from '@/components/ui/form'
import { Switch } from '@/components/ui/switch'

import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useResetForm } from '../hooks/use-reset-form'
import { useUpdateOption } from '../hooks/use-update-option'

const schema = z.object({
  image_studio_setting: z.object({
    enabled: z.boolean(),
    visible: z.boolean(),
    asset_library_enabled: z.boolean(),
  }),
})
type Values = z.infer<typeof schema>
type Props = {
  defaultValues: {
    'image_studio_setting.enabled': boolean
    'image_studio_setting.visible'?: boolean
    'image_studio_setting.asset_library_enabled'?: boolean
  }
}

export function ImageStudioSettingsSection(props: Props) {
  const { t } = useTranslation()
  const update = useUpdateOption()
  const defaults = {
    image_studio_setting: {
      enabled: props.defaultValues['image_studio_setting.enabled'],
      visible: props.defaultValues['image_studio_setting.visible'] ?? true,
      asset_library_enabled:
        props.defaultValues['image_studio_setting.asset_library_enabled'] ??
        true,
    },
  }
  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: defaults,
  })
  useResetForm(form, defaults)
  const submit = form.handleSubmit(async (values) => {
    for (const field of Object.keys(
      defaults.image_studio_setting
    ) as (keyof Values['image_studio_setting'])[]) {
      const value = values.image_studio_setting[field]
      if (value !== defaults.image_studio_setting[field]) {
        await update.mutateAsync({
          key: `image_studio_setting.${field}`,
          value,
        })
      }
    }
  })
  const switches = [
    {
      name: 'image_studio_setting.enabled',
      label: t('Enable AI Creation Studio'),
    },
    {
      name: 'image_studio_setting.visible',
      label: t('Show image generation entry'),
    },
    {
      name: 'image_studio_setting.asset_library_enabled',
      label: t('Enable asset library integration'),
    },
  ] as const

  return (
    <SettingsSection title={t('AI Creation Studio')}>
      <div className='bg-muted/30 text-muted-foreground space-y-2 rounded-lg border p-3 text-sm'>
        <p>
          {t(
            'Image generation requires an enabled private OSS profile and permanent storage policy. Configure and test the profile in Object Storage, then select it and enable permanent storage under Asset Library.'
          )}
        </p>
        <div className='flex flex-wrap gap-x-4 gap-y-2'>
          <Link
            className='text-primary underline underline-offset-4'
            to='/system-settings/storage/$section'
            params={{ section: 'object-storage' }}
          >
            {t('Object Storage')}
          </Link>
          <Link
            className='text-primary underline underline-offset-4'
            to='/system-settings/storage/$section'
            params={{ section: 'asset-library' }}
          >
            {t('Permanent asset storage')}
          </Link>
        </div>
        <p>
          {t(
            'Generated images are stored separately from library assets. Asset library integration controls reference images and manual saving only. Missing OSS configuration blocks generation before billing.'
          )}
        </p>
        <p>
          {t(
            "Clients load signed OSS URLs directly. Configure GET CORS for the PC site and add the OSS domain to the mini program's downloadFile allowed domains."
          )}
        </p>
      </div>
      <p className='text-muted-foreground text-sm'>
        {t(
          'Root can grant image review permissions to individual administrators in user management.'
        )}{' '}
        {t(
          'Disabling generation blocks image generation and prompt optimization calls; existing images remain readable.'
        )}
      </p>
      <Form {...form}>
        <SettingsForm onSubmit={submit}>
          <SettingsPageFormActions
            onSave={submit}
            isSaving={update.isPending}
          />
          {switches.map((item) => (
            <FormField
              key={item.name}
              control={form.control}
              name={item.name}
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>{item.label}</FormLabel>
                  </SettingsSwitchContent>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </SettingsSwitchItem>
              )}
            />
          ))}
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
