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
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import { Switch } from '@/components/ui/switch'

import {
  createStoragePolicySchema,
  storagePolicyToForm,
  type StoragePolicyFormValues,
} from './storage-schemas'
import type { StoragePolicy, StorageProfile } from './types'

type RelayMediaPolicyFormProps = {
  policy: StoragePolicy
  profiles: StorageProfile[]
  isSaving: boolean
  onSave: (values: StoragePolicyFormValues) => Promise<void>
}

function NumberField({
  form,
  name,
  label,
  description,
}: {
  form: ReturnType<typeof useForm<StoragePolicyFormValues>>
  name:
    | 'signed_url_ttl_hours'
    | 'retention_hours'
    | 'max_file_mib'
    | 'max_total_mib'
    | 'max_files'
  label: string
  description?: string
}) {
  return (
    <FormField
      control={form.control}
      name={name}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <Input
              type='number'
              min={1}
              step={1}
              value={field.value}
              onBlur={field.onBlur}
              onChange={(event) => field.onChange(event.target.valueAsNumber)}
            />
          </FormControl>
          {description && <FormDescription>{description}</FormDescription>}
          <FormMessage />
        </FormItem>
      )}
    />
  )
}

export function RelayMediaPolicyForm({
  policy,
  profiles,
  isSaving,
  onSave,
}: RelayMediaPolicyFormProps) {
  const { t } = useTranslation()
  const form = useForm<StoragePolicyFormValues>({
    resolver: zodResolver(createStoragePolicySchema(t)),
    defaultValues: storagePolicyToForm(policy),
  })

  useEffect(() => {
    form.reset(storagePolicyToForm(policy))
  }, [form, policy])

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Relay media temporary storage')}</CardTitle>
        <CardDescription>
          {t(
            'Controls Base64 materialization for supported relay channels. Objects remain private and upstream receives a time-limited signed URL.'
          )}
        </CardDescription>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSave)}>
          <CardContent className='grid gap-4 sm:grid-cols-2'>
            <FormField
              control={form.control}
              name='enabled'
              render={({ field }) => (
                <FormItem className='bg-muted/40 flex items-center justify-between gap-4 rounded-lg p-3 sm:col-span-2'>
                  <div className='space-y-1'>
                    <FormLabel>
                      {t('Enable temporary storage policy')}
                    </FormLabel>
                    <FormDescription>
                      {t(
                        'A Volc Native channel must also enable Base64 staging before this policy is used.'
                      )}
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='storage_profile_id'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Storage profile')}</FormLabel>
                  <FormControl>
                    <NativeSelect
                      className='w-full'
                      value={String(field.value)}
                      onChange={(event) =>
                        field.onChange(Number(event.target.value))
                      }
                    >
                      <NativeSelectOption value='0'>
                        {t('Select a profile')}
                      </NativeSelectOption>
                      {profiles.map((profile) => (
                        <NativeSelectOption
                          key={profile.id}
                          value={String(profile.id)}
                          disabled={profile.status !== 1}
                        >
                          {profile.name}
                          {profile.status !== 1 ? ` (${t('Disabled')})` : ''}
                        </NativeSelectOption>
                      ))}
                    </NativeSelect>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='object_prefix'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Object prefix')}</FormLabel>
                  <FormControl>
                    <Input placeholder='temporary/relay-media' {...field} />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'User isolation and task paths are appended automatically.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <NumberField
              form={form}
              name='signed_url_ttl_hours'
              label={t('Signed URL lifetime (hours)')}
              description={t(
                'Aliyun OSS V4 signed URLs support up to 168 hours.'
              )}
            />
            <NumberField
              form={form}
              name='retention_hours'
              label={t('Maximum retention (hours)')}
              description={t(
                'Terminal tasks are cleaned immediately; this value is the failure-safe deadline.'
              )}
            />
            <NumberField
              form={form}
              name='max_file_mib'
              label={t('Maximum file size (MiB)')}
            />
            <NumberField
              form={form}
              name='max_total_mib'
              label={t('Maximum request size (MiB)')}
            />
            <NumberField
              form={form}
              name='max_files'
              label={t('Maximum files per request')}
            />
            <p className='text-muted-foreground text-xs sm:col-span-2'>
              {t(
                'For audit integrity, the complete video task request body must stay within 2 MiB; larger requests are rejected before upload.'
              )}
            </p>
            <FormItem>
              <FormLabel>{t('Allowed media types')}</FormLabel>
              <Input value='image/jpeg, image/png, image/webp' disabled />
              <FormDescription>
                {t('Declared MIME type and decoded file signature must match.')}
              </FormDescription>
            </FormItem>
          </CardContent>
          <CardFooter className='mt-4 justify-end'>
            <Button type='submit' disabled={isSaving}>
              {isSaving ? t('Saving...') : t('Save storage policy')}
            </Button>
          </CardFooter>
        </form>
      </Form>
    </Card>
  )
}
