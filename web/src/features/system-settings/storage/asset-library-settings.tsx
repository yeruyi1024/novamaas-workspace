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
import { ShieldKeyIcon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import axios from 'axios'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
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
  listAssetLibraryConnections,
  listStorageProfiles,
  testAssetLibraryConnection,
  saveAssetLibraryConnection,
  syncAssetLibraryMissing,
  bindAssetChannelLibrary,
  mergeAssetLibraries,
  updateAssetLibraryStoragePolicy,
  deleteAssetLibraryConnection,
} from './api'
import { AssetLibraryAccountCard } from './asset-library-account-card'
import { AssetLibraryChannelsDialog } from './asset-library-channels-dialog'
import { AssetRequestLogDialog } from './asset-request-log-dialog'
import {
  assetChannelConfigToInput,
  assetLibraryConnectionToForm,
  assetLibraryPolicyToForm,
  assetLibraryPolicyToInput,
  createAssetLibraryConnectionSchema,
  createAssetLibraryPolicySchema,
  type AssetLibraryConnectionFormValues,
  type AssetLibraryPolicyFormValues,
} from './storage-schemas'
import {
  ASSET_PROTOCOL_VOLC_ACTION,
  ASSET_PROTOCOL_YOUFANG_REST,
  type AssetLibraryConnection,
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

function LibraryConfigDialog({
  config,
  open,
  onOpenChange,
  onSave,
  onTest,
  saving,
  testing,
}: {
  config: AssetLibraryConnection | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (values: AssetLibraryConnectionFormValues) => void
  onTest: (values: AssetLibraryConnectionFormValues) => void
  saving: boolean
  testing: boolean
}) {
  const { t } = useTranslation()
  const form = useForm<AssetLibraryConnectionFormValues>({
    resolver: config
      ? zodResolver(createAssetLibraryConnectionSchema(t, config))
      : undefined,
    defaultValues: config ? assetLibraryConnectionToForm(config) : undefined,
  })
  useEffect(() => {
    if (config) form.reset(assetLibraryConnectionToForm(config))
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
      <DialogContent className='max-h-[90dvh] overflow-y-auto sm:max-w-2xl'>
        <DialogHeader>
          <DialogTitle>{t('Configure upstream asset library')}</DialogTitle>
          <DialogDescription>
            {t(
              'One resource account can be shared by multiple channels. Use a separate library for another account or project.'
            )}
          </DialogDescription>
        </DialogHeader>
        {config && (
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(onSave)}
              className='grid gap-4 sm:grid-cols-2'
            >
              <FormField
                control={form.control}
                name='provider_name'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Provider name')}</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='name'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      {t('Resource account / library name')}
                    </FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='enabled'
                render={({ field }) => (
                  <FormItem className='bg-muted/40 flex items-center justify-between rounded-lg p-3 sm:col-span-2'>
                    <div>
                      <FormLabel>{t('Enable asset synchronization')}</FormLabel>
                      <FormDescription>
                        {t(
                          'When off, no new assets are written to this account and synchronization pauses, but assets already synchronized stay usable. Turning it on again syncs the assets missed meanwhile.'
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

const EMPTY_LIBRARY: AssetLibraryConnection = {
  id: 0,
  name: '',
  provider_name: '',
  channels: [],
  missing: 0,
  summary: { total: 0, pending: 0, processing: 0, active: 0, failed: 0 },
  channel_id: 0,
  channel_name: '',
  channel_type: 0,
  enabled: true,
  protocol: 'volc_action',
  auth_type: 'ak_sk',
  base_url: 'https://ark.cn-beijing.volcengineapi.com',
  region: 'cn-beijing',
  service: 'ark',
  api_version: '2024-01-01',
  project_name: 'default',
  qpm: 60,
  access_key_hint: '',
  credential_configured: false,
  updated_at: 0,
}

const LIBRARIES_QUERY_KEY = ['storage', 'asset-library-connections'] as const

export function AssetLibrarySettings() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [editing, setEditing] = useState<AssetLibraryConnection | null>(null)
  const [managingChannels, setManagingChannels] =
    useState<AssetLibraryConnection | null>(null)
  const [mergeOpen, setMergeOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] =
    useState<AssetLibraryConnection | null>(null)
  const [logLibraryId, setLogLibraryId] = useState<number | null>(null)
  const [syncTarget, setSyncTarget] = useState<AssetLibraryConnection | null>(
    null
  )
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
  const librariesQuery = useQuery({
    queryKey: LIBRARIES_QUERY_KEY,
    queryFn: async () => assertSuccess(await listAssetLibraryConnections()),
    refetchInterval: (query) =>
      query.state.data?.some(
        (library) =>
          Boolean(library.delete_requested_at) ||
          (library.channels.length > 0 &&
            library.enabled &&
            library.summary.pending + library.summary.processing > 0)
      )
        ? 5000
        : false,
  })
  const invalidate = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: SETTINGS_QUERY_KEY }),
      queryClient.invalidateQueries({ queryKey: LIBRARIES_QUERY_KEY }),
      queryClient.invalidateQueries({ queryKey: ['asset-library', 'admin'] }),
    ])
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
  const saveLibraryMutation = useMutation({
    mutationFn: async (values: AssetLibraryConnectionFormValues) => {
      if (!editing) throw new Error('No library selected')
      return assertSuccess(
        await saveAssetLibraryConnection(editing.id, {
          ...assetChannelConfigToInput(values),
          name: values.name,
          provider_name: values.provider_name,
        })
      )
    },
    onSuccess: async (library) => {
      await invalidate()
      setEditing(null)
      toast.success(
        library.reused_existing
          ? t('Existing resource account reused')
          : t('Upstream asset library saved')
      )
    },
    onError: (error) => toast.error(errorMessage(error)),
  })
  const testMutation = useMutation({
    mutationFn: async (values: AssetLibraryConnectionFormValues) => {
      if (!editing) throw new Error('No library selected')
      return assertSuccess(
        await testAssetLibraryConnection(editing.id, {
          ...assetChannelConfigToInput(values),
          name: values.name,
          provider_name: values.provider_name,
        })
      )
    },
    onSuccess: () => toast.success(t('Asset channel connection succeeded')),
    onError: (error) => toast.error(errorMessage(error)),
  })
  const syncMutation = useMutation({
    mutationFn: async (id: number) =>
      assertSuccess(await syncAssetLibraryMissing(id)),
    onSuccess: async () => {
      await invalidate()
      setSyncTarget(null)
      toast.success(t('Missing asset synchronization queued'))
    },
    onError: (error) => toast.error(errorMessage(error)),
  })
  const deleteMutation = useMutation({
    mutationFn: async (id: number) =>
      assertSuccess(await deleteAssetLibraryConnection(id)),
    onSuccess: async () => {
      await invalidate()
      setDeleteTarget(null)
      toast.success(t('Account deletion queued'))
    },
    onError: (error) => toast.error(errorMessage(error)),
  })
  const mergeMutation = useMutation({
    mutationFn: async () => assertSuccess(await mergeAssetLibraries()),
    onSuccess: async (report) => {
      await invalidate()
      setMergeOpen(false)
      toast.success(
        t(
          'Merged {{merged}} account groups; {{deferred}} deferred while uploads are running; {{unreadable}} accounts skipped because their credentials cannot be read.',
          {
            merged: report.merged_groups,
            deferred: report.deferred_groups,
            unreadable: report.unreadable_accounts,
          }
        )
      )
    },
    onError: (error) => toast.error(errorMessage(error)),
  })
  const channelsMutation = useMutation({
    mutationFn: async (channelIDs: number[]) => {
      if (!managingChannels) throw new Error('No library selected')
      const linked = new Set(
        managingChannels.channels.map((channel) => channel.id)
      )
      const requested = new Set(channelIDs)
      // Unlink first so a channel moving between accounts is released before
      // it is claimed. Each call is one atomic ownership change.
      for (const id of linked) {
        if (!requested.has(id)) {
          assertSuccess(await bindAssetChannelLibrary(id, 0, false))
        }
      }
      for (const id of requested) {
        if (!linked.has(id)) {
          assertSuccess(await bindAssetChannelLibrary(id, managingChannels.id))
        }
      }
    },
    onSuccess: async () => {
      await invalidate()
      setManagingChannels(null)
      toast.success(t('Channel associations saved'))
    },
    onError: (error) => {
      toast.error(errorMessage(error))
      void invalidate()
    },
  })
  const manageLibraryChannels = (library: AssetLibraryConnection) => {
    channelsMutation.reset()
    setManagingChannels(library)
  }
  // An account without channels that still owns copies was retired by a merge.
  const libraries = librariesQuery.data ?? []
  const retired = libraries.filter(
    (library) =>
      library.channels.length === 0 &&
      (library.legacy_channel_id != null || library.summary.total > 0)
  )
  const current = libraries.filter((library) => !retired.includes(library))
  const queryError = settingsQuery.error || librariesQuery.error
  return (
    <>
      <SectionPageLayout>
        <SectionPageLayout.Title>{t('Asset Library')}</SectionPageLayout.Title>
        <SectionPageLayout.Content>
          <div className='flex flex-col gap-4 pb-4'>
            <Alert>
              <AlertTitle>{t('Shared upstream asset libraries')}</AlertTitle>
              <AlertDescription>
                {t(
                  'Configure each provider resource account once and link its channels. Uploads create one upstream copy per linked resource account.'
                )}
              </AlertDescription>
            </Alert>
            {(settingsQuery.isPending || librariesQuery.isPending) && (
              <div className='text-muted-foreground flex min-h-48 items-center justify-center text-sm'>
                {t('Loading asset library settings...')}
              </div>
            )}
            {queryError && (
              <Alert variant='destructive'>
                <AlertTitle>
                  {t('Unable to load asset library settings')}
                </AlertTitle>
                <AlertDescription>{errorMessage(queryError)}</AlertDescription>
              </Alert>
            )}
            {settingsQuery.data && librariesQuery.data && (
              <>
                <details className='rounded-lg border p-4'>
                  <summary className='cursor-pointer text-sm font-medium'>
                    {t('Permanent asset storage')}
                  </summary>
                  <div className='mt-4'>
                    <AssetPolicyForm
                      policy={settingsQuery.data.policy}
                      profiles={settingsQuery.data.profiles}
                      pending={policyMutation.isPending}
                      onSubmit={(values) => policyMutation.mutate(values)}
                    />
                  </div>
                </details>
                <div className='flex flex-wrap items-center justify-between gap-3'>
                  <h2 className='font-semibold'>
                    {t('Provider asset libraries')}
                  </h2>
                  <div className='flex flex-wrap gap-2'>
                    <Button
                      variant='outline'
                      onClick={() => setMergeOpen(true)}
                    >
                      {t('Merge identical accounts')}
                    </Button>
                    <Button onClick={() => setEditing({ ...EMPTY_LIBRARY })}>
                      {t('Add resource account')}
                    </Button>
                  </div>
                </div>
                {current.length === 0 && (
                  <Card>
                    <CardContent className='text-muted-foreground py-6 text-sm'>
                      {t(
                        'Add a provider resource account and link its video channels to start synchronizing.'
                      )}
                    </CardContent>
                  </Card>
                )}
                <div
                  role='group'
                  aria-label={t('Provider asset libraries')}
                  className='grid grid-cols-[repeat(auto-fill,minmax(min(100%,22rem),1fr))] gap-4'
                >
                  {current.map((library) => (
                    <AssetLibraryAccountCard
                      key={library.id}
                      library={library}
                      onConfigure={setEditing}
                      onManageChannels={manageLibraryChannels}
                      onViewLogs={setLogLibraryId}
                      onSyncMissing={setSyncTarget}
                      onDelete={setDeleteTarget}
                    />
                  ))}
                </div>
                {retired.length > 0 && (
                  <details className='rounded-lg border p-4'>
                    <summary className='cursor-pointer text-sm font-medium'>
                      {t('Unlinked / historical library')} · {retired.length}
                    </summary>
                    <div className='mt-4 grid grid-cols-[repeat(auto-fill,minmax(min(100%,22rem),1fr))] gap-4'>
                      {retired.map((library) => (
                        <AssetLibraryAccountCard
                          key={library.id}
                          library={library}
                          onConfigure={setEditing}
                          onManageChannels={manageLibraryChannels}
                          onViewLogs={setLogLibraryId}
                          onSyncMissing={setSyncTarget}
                          onDelete={setDeleteTarget}
                        />
                      ))}
                    </div>
                  </details>
                )}
              </>
            )}
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>
      <LibraryConfigDialog
        config={editing}
        open={editing !== null}
        onOpenChange={(open) => !open && setEditing(null)}
        onSave={(values) => saveLibraryMutation.mutate(values)}
        onTest={(values) => testMutation.mutate(values)}
        saving={saveLibraryMutation.isPending}
        testing={testMutation.isPending}
      />
      <AssetLibraryChannelsDialog
        library={managingChannels}
        channels={settingsQuery.data?.channels ?? []}
        saving={channelsMutation.isPending}
        error={
          channelsMutation.error
            ? errorMessage(channelsMutation.error)
            : undefined
        }
        onOpenChange={(open) => {
          if (!open) {
            setManagingChannels(null)
            channelsMutation.reset()
          }
        }}
        onSave={(ids) => channelsMutation.mutate(ids)}
        onReload={async () => {
          const result = await librariesQuery.refetch()
          const latest = result.data?.find(
            (library) => library.id === managingChannels?.id
          )
          if (latest) {
            setManagingChannels(latest)
            channelsMutation.reset()
          }
        }}
      />
      {logLibraryId !== null && (
        <AssetRequestLogDialog
          open
          libraryId={logLibraryId}
          onOpenChange={(open) => !open && setLogLibraryId(null)}
        />
      )}
      <Dialog
        open={syncTarget !== null}
        onOpenChange={(open) => !open && setSyncTarget(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('Sync missing assets')}</DialogTitle>
            <DialogDescription>
              {t(
                'Sync {{count}} missing assets to {{library}}? Existing upstream assets and failed job states will be preserved.',
                {
                  count: syncTarget?.missing ?? 0,
                  library: syncTarget?.name ?? '',
                }
              )}
            </DialogDescription>
          </DialogHeader>
          <p className='text-muted-foreground text-sm'>
            {t('This library is shared by {{count}} channels.', {
              count: syncTarget?.channels.length ?? 0,
            })}
          </p>
          <DialogFooter>
            <Button variant='outline' onClick={() => setSyncTarget(null)}>
              {t('Cancel')}
            </Button>
            <Button
              disabled={syncMutation.isPending || !syncTarget}
              onClick={() => syncTarget && syncMutation.mutate(syncTarget.id)}
            >
              {t('Confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog
        open={deleteTarget !== null}
        onOpenChange={(open) =>
          !open && !deleteMutation.isPending && setDeleteTarget(null)
        }
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('Delete account and copies')}</DialogTitle>
            <DialogDescription>
              {t(
                'Delete {{library}} and its upstream copies? Copies still used by another account are kept. Platform assets and upstream groups are retained. This cannot be undone.',
                { library: deleteTarget?.name ?? '' }
              )}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant='outline'
              disabled={deleteMutation.isPending}
              onClick={() => setDeleteTarget(null)}
            >
              {t('Cancel')}
            </Button>
            <Button
              variant='destructive'
              disabled={deleteMutation.isPending || !deleteTarget}
              onClick={() =>
                deleteTarget && deleteMutation.mutate(deleteTarget.id)
              }
            >
              {t('Confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog open={mergeOpen} onOpenChange={setMergeOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('Merge identical accounts')}</DialogTitle>
            <DialogDescription>
              {t(
                'Accounts that point at exactly the same upstream account are merged into one and their channels move to it. Existing upstream assets are kept and nothing is deleted.'
              )}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant='outline' onClick={() => setMergeOpen(false)}>
              {t('Cancel')}
            </Button>
            <Button
              disabled={mergeMutation.isPending}
              onClick={() => mergeMutation.mutate()}
            >
              {t('Confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
