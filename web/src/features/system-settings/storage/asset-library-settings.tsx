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
import { PencilEdit01Icon, ShieldKeyIcon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import axios from 'axios'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
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
import { Switch } from '@/components/ui/switch'

import {
  getAssetLibraryStoragePolicy,
  listAssetChannelConfigs,
  listStorageProfiles,
  testAssetChannelConfig,
  updateAssetChannelConfig,
  updateAssetLibraryStoragePolicy,
} from './api'
import { AssetRequestLogDialog } from './asset-request-log-dialog'
import {
  assetChannelConfigToForm,
  assetChannelConfigToInput,
  assetLibraryPolicyToForm,
  assetLibraryPolicyToInput,
  createAssetChannelConfigSchema,
  createAssetLibraryPolicySchema,
  type AssetChannelConfigFormValues,
  type AssetLibraryPolicyFormValues,
} from './storage-schemas'
import {
  ASSET_PROTOCOL_VOLC_ACTION,
  ASSET_PROTOCOL_YOUFANG_REST,
  type AssetChannelConfig,
  type StorageAPIResponse,
  type StoragePolicy,
  type StorageProfile,
} from './types'

const SETTINGS_QUERY_KEY = ['storage', 'asset-library'] as const

function assertSuccess<T>(response: StorageAPIResponse<T>): T {
  if (!response.success) throw new Error(response.message || 'Request failed')
  return response.data
}

function errorMessage(error: unknown): string {
  if (axios.isAxiosError<StorageAPIResponse>(error)) {
    return error.response?.data?.message || error.message
  }
  return error instanceof Error ? error.message : 'Request failed'
}

function AssetPolicyForm({
  policy,
  profiles,
  pending,
  onSubmit,
}: {
  policy: StoragePolicy
  profiles: StorageProfile[]
  pending: boolean
  onSubmit: (values: AssetLibraryPolicyFormValues) => void
}) {
  const { t } = useTranslation()
  const form = useForm<AssetLibraryPolicyFormValues>({
    resolver: zodResolver(createAssetLibraryPolicySchema(t)),
    defaultValues: assetLibraryPolicyToForm(policy),
  })
  useEffect(() => form.reset(assetLibraryPolicyToForm(policy)), [form, policy])

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Permanent asset storage')}</CardTitle>
        <CardDescription>
          {t(
            'Assets remain in your private object storage until the owner deletes them. Signed URLs are generated only for preview and upstream synchronization.'
          )}
        </CardDescription>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <CardContent className='grid gap-4 sm:grid-cols-2'>
            <FormField
              control={form.control}
              name='enabled'
              render={({ field }) => (
                <FormItem className='bg-muted/40 flex items-center justify-between gap-4 rounded-lg p-3 sm:col-span-2'>
                  <div className='space-y-1'>
                    <FormLabel>{t('Enable asset library storage')}</FormLabel>
                    <FormDescription>
                      {t('Uploads are rejected while this policy is disabled.')}
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
                    <Input placeholder='assets/library' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='signed_url_ttl_hours'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Signed URL lifetime (hours)')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={1}
                      max={168}
                      value={field.value}
                      onBlur={field.onBlur}
                      onChange={(event) =>
                        field.onChange(event.target.valueAsNumber)
                      }
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='max_file_mib'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Maximum file size (MiB)')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={1}
                      max={512}
                      value={field.value}
                      onBlur={field.onBlur}
                      onChange={(event) =>
                        field.onChange(event.target.valueAsNumber)
                      }
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='max_files'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Maximum assets per group')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={1}
                      max={100}
                      value={field.value}
                      onBlur={field.onBlur}
                      onChange={(event) =>
                        field.onChange(event.target.valueAsNumber)
                      }
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </CardContent>
          <CardFooter className='justify-end'>
            <Button type='submit' disabled={pending}>
              {t('Save storage policy')}
            </Button>
          </CardFooter>
        </form>
      </Form>
    </Card>
  )
}

function ChannelConfigDialog({
  config,
  open,
  onOpenChange,
  onSave,
  onTest,
  saving,
  testing,
}: {
  config: AssetChannelConfig | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (values: AssetChannelConfigFormValues) => void
  onTest: (values: AssetChannelConfigFormValues) => void
  saving: boolean
  testing: boolean
}) {
  const { t } = useTranslation()
  const form = useForm<AssetChannelConfigFormValues>({
    resolver: config
      ? zodResolver(createAssetChannelConfigSchema(t, config))
      : undefined,
    defaultValues: config ? assetChannelConfigToForm(config) : undefined,
  })
  useEffect(() => {
    if (config) form.reset(assetChannelConfigToForm(config))
  }, [config, form])
  const protocol = form.watch('protocol')
  const authType = form.watch('auth_type')
  let credentialLabel = t('API key')
  if (protocol === ASSET_PROTOCOL_YOUFANG_REST) {
    credentialLabel = t('YooFang sk key')
  } else if (authType === 'ak_sk') {
    credentialLabel = t('Secret Access Key')
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-2xl'>
        <DialogHeader>
          <DialogTitle>{t('Configure asset channel')}</DialogTitle>
          <DialogDescription>{config?.channel_name}</DialogDescription>
        </DialogHeader>
        {config && (
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(onSave)}
              className='grid gap-4 sm:grid-cols-2'
            >
              <FormField
                control={form.control}
                name='enabled'
                render={({ field }) => (
                  <FormItem className='bg-muted/40 flex items-center justify-between rounded-lg p-3 sm:col-span-2'>
                    <div>
                      <FormLabel>{t('Enable asset synchronization')}</FormLabel>
                      <FormDescription>
                        {t(
                          'Existing assets are queued automatically after saving.'
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
                name='protocol'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Provider protocol')}</FormLabel>
                    <FormControl>
                      <NativeSelect
                        className='w-full'
                        value={field.value}
                        onChange={(event) => {
                          const value = event.target.value as
                            | typeof ASSET_PROTOCOL_VOLC_ACTION
                            | typeof ASSET_PROTOCOL_YOUFANG_REST
                          field.onChange(value)
                          form.setValue(
                            'auth_type',
                            value === ASSET_PROTOCOL_YOUFANG_REST
                              ? 'bearer'
                              : 'ak_sk'
                          )
                          form.setValue(
                            'base_url',
                            value === ASSET_PROTOCOL_YOUFANG_REST
                              ? 'https://asset-inference-doubao.yoofang.com'
                              : 'https://ark.cn-beijing.volcengineapi.com'
                          )
                          form.setValue('access_key_id', '')
                          form.setValue('credential', '')
                        }}
                      >
                        <NativeSelectOption value={ASSET_PROTOCOL_VOLC_ACTION}>
                          {t('Volcengine Action compatible')}
                        </NativeSelectOption>
                        <NativeSelectOption value={ASSET_PROTOCOL_YOUFANG_REST}>
                          {t('YooFang REST')}
                        </NativeSelectOption>
                      </NativeSelect>
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
                    <FormLabel>{t('Authentication')}</FormLabel>
                    <FormControl>
                      <NativeSelect
                        className='w-full'
                        disabled={protocol === ASSET_PROTOCOL_YOUFANG_REST}
                        {...field}
                      >
                        {protocol === ASSET_PROTOCOL_VOLC_ACTION && (
                          <NativeSelectOption value='ak_sk'>
                            {t('AK/SK signature')}
                          </NativeSelectOption>
                        )}
                        <NativeSelectOption value='bearer'>
                          {protocol === ASSET_PROTOCOL_YOUFANG_REST
                            ? t('Bearer sk key')
                            : t('Bearer API key')}
                        </NativeSelectOption>
                      </NativeSelect>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {protocol === ASSET_PROTOCOL_VOLC_ACTION && (
                <FormField
                  control={form.control}
                  name='project_name'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('Project name')}</FormLabel>
                      <FormControl>
                        <Input placeholder='default' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              )}
              <FormField
                control={form.control}
                name='base_url'
                render={({ field }) => (
                  <FormItem className='sm:col-span-2'>
                    <FormLabel>{t('Asset API base URL')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder={
                          protocol === ASSET_PROTOCOL_YOUFANG_REST
                            ? 'https://asset-inference-doubao.yoofang.com'
                            : 'https://ark.cn-beijing.volcengineapi.com'
                        }
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      {protocol === ASSET_PROTOCOL_YOUFANG_REST
                        ? t(
                            'The /api/v1/assets resource path is appended to this URL.'
                          )
                        : t(
                            'Action and Version query parameters are appended to this exact path.'
                          )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {protocol === ASSET_PROTOCOL_VOLC_ACTION && (
                <>
                  <FormField
                    control={form.control}
                    name='region'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t('Region')}</FormLabel>
                        <FormControl>
                          <Input {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name='service'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t('Service')}</FormLabel>
                        <FormControl>
                          <Input {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name='api_version'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t('API version')}</FormLabel>
                        <FormControl>
                          <Input {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </>
              )}
              <FormField
                control={form.control}
                name='qpm'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Requests per minute')}</FormLabel>
                    <FormControl>
                      <Input
                        type='number'
                        min={1}
                        max={1000}
                        value={field.value}
                        onBlur={field.onBlur}
                        onChange={(event) =>
                          field.onChange(event.target.valueAsNumber)
                        }
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {protocol === ASSET_PROTOCOL_VOLC_ACTION &&
                authType === 'ak_sk' && (
                  <FormField
                    control={form.control}
                    name='access_key_id'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t('Access Key ID')}</FormLabel>
                        <FormControl>
                          <Input
                            placeholder={
                              config.access_key_hint ||
                              t('Keep current value when empty')
                            }
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}
              <FormField
                control={form.control}
                name='credential'
                render={({ field }) => (
                  <FormItem
                    className={authType === 'bearer' ? '' : 'sm:col-span-2'}
                  >
                    <FormLabel>{credentialLabel}</FormLabel>
                    <FormControl>
                      <Input
                        type='password'
                        placeholder={
                          config.credential_configured
                            ? t('Keep current value when empty')
                            : ''
                        }
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      {t(
                        'Credentials are encrypted at rest and never returned by the API.'
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <DialogFooter className='sm:col-span-2'>
                <Button
                  type='button'
                  variant='outline'
                  disabled={testing}
                  onClick={form.handleSubmit(onTest)}
                >
                  <HugeiconsIcon
                    icon={ShieldKeyIcon}
                    data-icon='inline-start'
                  />
                  {t('Test connection')}
                </Button>
                <Button type='submit' disabled={saving}>
                  {t('Save')}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        )}
      </DialogContent>
    </Dialog>
  )
}

export function AssetLibrarySettings() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [editing, setEditing] = useState<AssetChannelConfig | null>(null)
  const [logChannelId, setLogChannelId] = useState<number | null>(null)
  const settingsQuery = useQuery({
    queryKey: SETTINGS_QUERY_KEY,
    queryFn: async () => {
      const [profiles, policy, channels] = await Promise.all([
        listStorageProfiles(),
        getAssetLibraryStoragePolicy(),
        listAssetChannelConfigs(),
      ])
      return {
        profiles: assertSuccess(profiles),
        policy: assertSuccess(policy),
        channels: assertSuccess(channels),
      }
    },
  })
  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: SETTINGS_QUERY_KEY })
  const policyMutation = useMutation({
    mutationFn: async (values: AssetLibraryPolicyFormValues) =>
      assertSuccess(
        await updateAssetLibraryStoragePolicy(assetLibraryPolicyToInput(values))
      ),
    onSuccess: async () => {
      await invalidate()
      toast.success(t('Asset storage policy saved'))
    },
    onError: (error) => toast.error(errorMessage(error)),
  })
  const saveChannelMutation = useMutation({
    mutationFn: async (values: AssetChannelConfigFormValues) => {
      if (!editing) throw new Error('No channel selected')
      return assertSuccess(
        await updateAssetChannelConfig(
          editing.channel_id,
          assetChannelConfigToInput(values)
        )
      )
    },
    onSuccess: async () => {
      await invalidate()
      setEditing(null)
      toast.success(t('Asset channel saved'))
    },
    onError: (error) => toast.error(errorMessage(error)),
  })
  const testMutation = useMutation({
    mutationFn: async (values: AssetChannelConfigFormValues) => {
      if (!editing) throw new Error('No channel selected')
      return assertSuccess(
        await testAssetChannelConfig(
          editing.channel_id,
          assetChannelConfigToInput(values)
        )
      )
    },
    onSuccess: () => toast.success(t('Asset channel connection succeeded')),
    onError: (error) => toast.error(errorMessage(error)),
  })
  return (
    <>
      <SectionPageLayout>
        <SectionPageLayout.Title>{t('Asset Library')}</SectionPageLayout.Title>
        <SectionPageLayout.Content>
          <div className='flex flex-col gap-4 pb-4'>
            <Alert>
              <AlertTitle>{t('Two-layer asset ownership')}</AlertTitle>
              <AlertDescription>
                {t(
                  'The platform keeps the permanent source of truth in object storage. Each enabled channel receives its own upstream asset ID, which is mapped only after routing selects that channel.'
                )}
              </AlertDescription>
            </Alert>
            {settingsQuery.isLoading && (
              <div className='text-muted-foreground flex min-h-48 items-center justify-center text-sm'>
                {t('Loading asset library settings...')}
              </div>
            )}
            {settingsQuery.isError && (
              <Alert variant='destructive'>
                <AlertTitle>
                  {t('Unable to load asset library settings')}
                </AlertTitle>
                <AlertDescription>
                  {errorMessage(settingsQuery.error)}
                </AlertDescription>
              </Alert>
            )}
            {settingsQuery.data && (
              <>
                <AssetPolicyForm
                  policy={settingsQuery.data.policy}
                  profiles={settingsQuery.data.profiles}
                  pending={policyMutation.isPending}
                  onSubmit={(values) => policyMutation.mutate(values)}
                />
                <Card>
                  <CardHeader>
                    <CardTitle>{t('Asset channels')}</CardTitle>
                    <CardDescription>
                      {t(
                        'Choose a provider protocol independently for each DoubaoVideo or Volc Native channel.'
                      )}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className='grid gap-3 lg:grid-cols-2'>
                    {settingsQuery.data.channels.length === 0 ? (
                      <p className='text-muted-foreground text-sm'>
                        {t(
                          'Create a DoubaoVideo or Volc Native channel first.'
                        )}
                      </p>
                    ) : (
                      settingsQuery.data.channels.map((channel) => (
                        <Card key={channel.channel_id} size='sm'>
                          <CardHeader className='gap-3'>
                            <div className='flex min-w-0 flex-wrap items-center gap-2'>
                              <CardTitle className='min-w-0 truncate'>
                                {channel.channel_name}
                              </CardTitle>
                              <Badge
                                variant='outline'
                                className='shrink-0 font-mono font-normal'
                              >
                                ID {channel.channel_id}
                              </Badge>
                              <Badge
                                className='shrink-0'
                                variant={
                                  channel.enabled ? 'default' : 'secondary'
                                }
                              >
                                {channel.enabled ? t('Enabled') : t('Disabled')}
                              </Badge>
                            </div>
                            <CardDescription className='flex min-w-0 flex-col gap-1'>
                              <span className='truncate'>
                                {channel.protocol ===
                                ASSET_PROTOCOL_YOUFANG_REST
                                  ? t('YooFang REST')
                                  : t('Volcengine Action compatible')}{' '}
                                ·{' '}
                                {channel.auth_type === 'ak_sk'
                                  ? t('AK/SK signature')
                                  : t('Bearer API key')}
                              </span>
                              <span
                                className='truncate font-mono text-xs'
                                title={channel.base_url}
                              >
                                {channel.base_url}
                              </span>
                            </CardDescription>
                          </CardHeader>
                          <CardFooter className='justify-end gap-2'>
                            <Button
                              size='sm'
                              variant='outline'
                              onClick={() =>
                                setLogChannelId(channel.channel_id)
                              }
                            >
                              {t('View logs')}
                            </Button>
                            <Button
                              size='sm'
                              onClick={() => setEditing(channel)}
                            >
                              <HugeiconsIcon
                                icon={PencilEdit01Icon}
                                data-icon='inline-start'
                              />
                              {t('Configure')}
                            </Button>
                          </CardFooter>
                        </Card>
                      ))
                    )}
                  </CardContent>
                </Card>
              </>
            )}
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>
      <ChannelConfigDialog
        config={editing}
        open={editing !== null}
        onOpenChange={(open) => !open && setEditing(null)}
        onSave={(values) => saveChannelMutation.mutate(values)}
        onTest={(values) => testMutation.mutate(values)}
        saving={saveChannelMutation.isPending}
        testing={testMutation.isPending}
      />
      {logChannelId !== null && (
        <AssetRequestLogDialog
          open
          channelId={logChannelId}
          onOpenChange={(open) => !open && setLogChannelId(null)}
        />
      )}
    </>
  )
}
