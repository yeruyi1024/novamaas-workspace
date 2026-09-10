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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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

import {
  createStorageProfileSchema,
  type StorageProfileFormValues,
} from './storage-schemas'
import {
  STORAGE_AUTH_ENVIRONMENT,
  STORAGE_AUTH_STATIC,
  STORAGE_PROVIDER_ALIYUN_OSS,
  type StorageProfile,
} from './types'

type StorageProfileDialogProps = {
  open: boolean
  profile: StorageProfile | null
  isSaving: boolean
  isTesting: boolean
  onOpenChange: (open: boolean) => void
  onSave: (values: StorageProfileFormValues) => void
  onTest: (values: StorageProfileFormValues) => void
}

function getProfileDefaults(
  profile: StorageProfile | null
): StorageProfileFormValues {
  return {
    name: profile?.name ?? '',
    provider_type: STORAGE_PROVIDER_ALIYUN_OSS,
    status: profile?.status ?? 1,
    endpoint: profile?.endpoint ?? 'https://oss-cn-hangzhou.aliyuncs.com',
    region: profile?.region ?? 'cn-hangzhou',
    bucket: profile?.bucket ?? '',
    auth_type:
      profile?.auth_type === STORAGE_AUTH_ENVIRONMENT
        ? STORAGE_AUTH_ENVIRONMENT
        : STORAGE_AUTH_STATIC,
    access_key_id: '',
    access_key_secret: '',
    security_token: '',
  }
}

export function StorageProfileDialog({
  open,
  profile,
  isSaving,
  isTesting,
  onOpenChange,
  onSave,
  onTest,
}: StorageProfileDialogProps) {
  const { t } = useTranslation()
  const form = useForm<StorageProfileFormValues>({
    resolver: zodResolver(createStorageProfileSchema(t, profile === null)),
    defaultValues: getProfileDefaults(profile),
  })
  const authType = form.watch('auth_type')

  useEffect(() => {
    if (open) form.reset(getProfileDefaults(profile))
  }, [form, open, profile])

  const runTest = form.handleSubmit(onTest)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-2xl'>
        <DialogHeader>
          <DialogTitle>
            {profile ? t('Edit storage profile') : t('Add storage profile')}
          </DialogTitle>
          <DialogDescription>
            {t(
              'Storage profiles are system-wide destinations shared by relay policies and future asset libraries.'
            )}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id='storage-profile-form'
            className='grid gap-4 sm:grid-cols-2'
            onSubmit={form.handleSubmit(onSave)}
            autoComplete='off'
          >
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Profile name')}</FormLabel>
                  <FormControl>
                    <Input
                      placeholder={t('Primary media storage')}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='provider_type'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Storage provider')}</FormLabel>
                  <FormControl>
                    <NativeSelect
                      className='w-full'
                      value={field.value}
                      onChange={field.onChange}
                    >
                      <NativeSelectOption value={STORAGE_PROVIDER_ALIYUN_OSS}>
                        {t('Aliyun OSS')}
                      </NativeSelectOption>
                      <NativeSelectOption value='tencent_cos' disabled>
                        {t('Tencent COS (reserved)')}
                      </NativeSelectOption>
                      <NativeSelectOption value='s3_compatible' disabled>
                        {t('S3 compatible / MinIO (reserved)')}
                      </NativeSelectOption>
                    </NativeSelect>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='endpoint'
              render={({ field }) => (
                <FormItem className='sm:col-span-2'>
                  <FormLabel>{t('OSS endpoint')}</FormLabel>
                  <FormControl>
                    <Input type='url' inputMode='url' {...field} />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Use the official HTTPS public endpoint for the selected region.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='region'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Region')}</FormLabel>
                  <FormControl>
                    <Input placeholder='cn-hangzhou' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='bucket'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Bucket')}</FormLabel>
                  <FormControl>
                    <Input placeholder='my-private-bucket' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='auth_type'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Credential source')}</FormLabel>
                  <FormControl>
                    <NativeSelect
                      className='w-full'
                      value={field.value}
                      onChange={field.onChange}
                    >
                      <NativeSelectOption value={STORAGE_AUTH_STATIC}>
                        {t('Encrypted access key')}
                      </NativeSelectOption>
                      <NativeSelectOption value={STORAGE_AUTH_ENVIRONMENT}>
                        {t('Environment variables')}
                      </NativeSelectOption>
                    </NativeSelect>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='status'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Status')}</FormLabel>
                  <FormControl>
                    <NativeSelect
                      className='w-full'
                      value={String(field.value)}
                      onChange={(event) =>
                        field.onChange(Number(event.target.value))
                      }
                    >
                      <NativeSelectOption value='1'>
                        {t('Enabled')}
                      </NativeSelectOption>
                      <NativeSelectOption value='0'>
                        {t('Disabled')}
                      </NativeSelectOption>
                    </NativeSelect>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {authType === STORAGE_AUTH_STATIC ? (
              <>
                <FormField
                  control={form.control}
                  name='access_key_id'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('Access Key ID')}</FormLabel>
                      <FormControl>
                        <Input autoComplete='off' {...field} />
                      </FormControl>
                      {profile && (
                        <FormDescription>
                          {profile.access_key_hint
                            ? t('Current key: {{hint}}', {
                                hint: profile.access_key_hint,
                              })
                            : t(
                                'Leave both key fields blank to keep the current credential.'
                              )}
                        </FormDescription>
                      )}
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='access_key_secret'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('Access Key Secret')}</FormLabel>
                      <FormControl>
                        <Input
                          type='password'
                          autoComplete='new-password'
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='security_token'
                  render={({ field }) => (
                    <FormItem className='sm:col-span-2'>
                      <FormLabel>{t('Security token (optional)')}</FormLabel>
                      <FormControl>
                        <Input
                          type='password'
                          autoComplete='new-password'
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </>
            ) : (
              <p className='text-muted-foreground sm:col-span-2'>
                {t(
                  'The server reads OSS_ACCESS_KEY_ID, OSS_ACCESS_KEY_SECRET, and optional OSS_SESSION_TOKEN at runtime.'
                )}
              </p>
            )}
          </form>
        </Form>

        <DialogFooter>
          <Button
            type='button'
            variant='outline'
            onClick={() => onOpenChange(false)}
            disabled={isSaving || isTesting}
          >
            {t('Cancel')}
          </Button>
          {!profile && (
            <Button
              type='button'
              variant='outline'
              onClick={runTest}
              disabled={isSaving || isTesting}
            >
              {isTesting ? t('Testing...') : t('Test configuration')}
            </Button>
          )}
          <Button
            type='submit'
            form='storage-profile-form'
            disabled={isSaving || isTesting}
          >
            {isSaving ? t('Saving...') : t('Save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
