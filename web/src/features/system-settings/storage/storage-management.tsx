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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import i18next from 'i18next'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

import {
  archiveStorageProfile,
  createStorageProfile,
  getRelayMediaStoragePolicy,
  listStorageProfiles,
  testSavedStorageProfile,
  testStorageProfile,
  updateRelayMediaStoragePolicy,
  updateStorageProfile,
} from './api'
import { RelayMediaPolicyForm } from './relay-media-policy-form'
import { StorageProfileDialog } from './storage-profile-dialog'
import {
  storagePolicyToInput,
  type StoragePolicyFormValues,
  type StorageProfileFormValues,
} from './storage-schemas'
import type { StorageAPIResponse, StorageProfile } from './types'

const STORAGE_QUERY_KEY = ['storage', 'profiles'] as const
const RELAY_POLICY_QUERY_KEY = ['storage', 'relay-media-policy'] as const

function assertSuccess<T>(response: StorageAPIResponse<T>): T {
  if (!response.success) {
    throw new Error(response.message || i18next.t('Storage operation failed'))
  }
  return response.data
}

export function StorageManagement() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingProfile, setEditingProfile] = useState<StorageProfile | null>(
    null
  )
  const [archiveTarget, setArchiveTarget] = useState<StorageProfile | null>(
    null
  )

  const profilesQuery = useQuery({
    queryKey: STORAGE_QUERY_KEY,
    queryFn: async () => assertSuccess(await listStorageProfiles()),
  })
  const policyQuery = useQuery({
    queryKey: RELAY_POLICY_QUERY_KEY,
    queryFn: async () => assertSuccess(await getRelayMediaStoragePolicy()),
  })

  const saveProfileMutation = useMutation({
    mutationFn: async (values: StorageProfileFormValues) => {
      const response = editingProfile
        ? await updateStorageProfile(editingProfile.id, values)
        : await createStorageProfile(values)
      return assertSuccess(response)
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: STORAGE_QUERY_KEY })
      setDialogOpen(false)
      toast.success(t('Storage profile saved'))
    },
    onError: (error: Error) => toast.error(error.message),
  })

  const testDraftMutation = useMutation({
    mutationFn: async (values: StorageProfileFormValues) =>
      assertSuccess(await testStorageProfile(values)),
    onSuccess: () => toast.success(t('Storage test completed successfully')),
    onError: (error: Error) => toast.error(error.message),
  })

  const testSavedMutation = useMutation({
    mutationFn: async (id: number) =>
      assertSuccess(await testSavedStorageProfile(id)),
    onSuccess: () => toast.success(t('Storage test completed successfully')),
    onError: (error: Error) => toast.error(error.message),
  })

  const archiveMutation = useMutation({
    mutationFn: async (id: number) =>
      assertSuccess(await archiveStorageProfile(id)),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: STORAGE_QUERY_KEY })
      setArchiveTarget(null)
      toast.success(t('Storage profile archived'))
    },
    onError: (error: Error) => toast.error(error.message),
  })

  const policyMutation = useMutation({
    mutationFn: async (values: StoragePolicyFormValues) =>
      assertSuccess(
        await updateRelayMediaStoragePolicy(storagePolicyToInput(values))
      ),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: RELAY_POLICY_QUERY_KEY })
      toast.success(t('Storage policy saved'))
    },
    onError: (error: Error) => toast.error(error.message),
  })

  const openCreateDialog = () => {
    setEditingProfile(null)
    setDialogOpen(true)
  }

  const openEditDialog = (profile: StorageProfile) => {
    setEditingProfile(profile)
    setDialogOpen(true)
  }

  const isLoading = profilesQuery.isLoading || policyQuery.isLoading

  return (
    <>
      <SectionPageLayout>
        <SectionPageLayout.Title>{t('Object Storage')}</SectionPageLayout.Title>
        <SectionPageLayout.Actions>
          <Button size='sm' onClick={openCreateDialog}>
            <Plus />
            {t('Add profile')}
          </Button>
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          <div className='flex flex-col gap-4 pb-4'>
            <Alert>
              <AlertTitle>
                {t('Private storage with temporary access')}
              </AlertTitle>
              <AlertDescription>
                {t(
                  'Buckets stay private. The platform uploads Base64 media, sends a signed HTTPS URL upstream, and deletes the object after task completion or the retention deadline.'
                )}{' '}
                {t(
                  'Set STORAGE_CREDENTIAL_ENCRYPTION_KEY before saving static credentials so encrypted keys remain readable across restarts.'
                )}
              </AlertDescription>
            </Alert>

            {isLoading && (
              <div className='text-muted-foreground flex min-h-40 items-center justify-center text-sm'>
                {t('Loading storage settings...')}
              </div>
            )}
            {!isLoading && (profilesQuery.isError || policyQuery.isError) && (
              <Alert variant='destructive'>
                <AlertTitle>{t('Unable to load storage settings')}</AlertTitle>
                <AlertDescription>
                  {(profilesQuery.error || policyQuery.error)?.message}
                </AlertDescription>
              </Alert>
            )}
            {!isLoading && !profilesQuery.isError && !policyQuery.isError && (
              <>
                <Card>
                  <CardHeader>
                    <CardTitle>{t('Storage profiles')}</CardTitle>
                    <CardDescription>
                      {t(
                        'Provider-specific connection settings are separated from policies so more destinations can be added later.'
                      )}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className='grid gap-3 lg:grid-cols-2'>
                    {profilesQuery.data?.length ? (
                      profilesQuery.data.map((profile) => (
                        <Card key={profile.id} size='sm'>
                          <CardHeader>
                            <CardTitle className='flex flex-wrap items-center gap-2'>
                              {profile.name}
                              <Badge
                                variant={
                                  profile.status === 1 ? 'default' : 'secondary'
                                }
                              >
                                {profile.status === 1
                                  ? t('Enabled')
                                  : t('Disabled')}
                              </Badge>
                            </CardTitle>
                            <CardDescription>
                              {t('Aliyun OSS')} · {profile.region} ·{' '}
                              {profile.bucket}
                            </CardDescription>
                            <CardAction className='flex gap-1'>
                              <Button
                                size='icon-sm'
                                variant='ghost'
                                aria-label={t('Edit storage profile')}
                                onClick={() => openEditDialog(profile)}
                              >
                                <Pencil />
                              </Button>
                              <Button
                                size='icon-sm'
                                variant='ghost'
                                aria-label={t('Archive storage profile')}
                                onClick={() => setArchiveTarget(profile)}
                              >
                                <Trash2 />
                              </Button>
                            </CardAction>
                          </CardHeader>
                          <CardContent className='flex items-center justify-between gap-3'>
                            <div className='text-muted-foreground min-w-0 truncate text-xs'>
                              {profile.auth_type === 'environment'
                                ? t('Environment credentials')
                                : profile.access_key_hint ||
                                  t('Credential not configured')}
                            </div>
                            <Button
                              size='sm'
                              variant='outline'
                              disabled={
                                testSavedMutation.isPending ||
                                !profile.credential_configured
                              }
                              onClick={() =>
                                testSavedMutation.mutate(profile.id)
                              }
                            >
                              {t('Test')}
                            </Button>
                          </CardContent>
                        </Card>
                      ))
                    ) : (
                      <div className='text-muted-foreground rounded-lg border border-dashed p-8 text-center text-sm lg:col-span-2'>
                        {t('No storage profiles configured yet.')}
                      </div>
                    )}
                  </CardContent>
                </Card>

                {policyQuery.data && (
                  <RelayMediaPolicyForm
                    policy={policyQuery.data}
                    profiles={profilesQuery.data ?? []}
                    isSaving={policyMutation.isPending}
                    onSave={async (values) => {
                      await policyMutation.mutateAsync(values)
                    }}
                  />
                )}
              </>
            )}
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <StorageProfileDialog
        open={dialogOpen}
        profile={editingProfile}
        isSaving={saveProfileMutation.isPending}
        isTesting={testDraftMutation.isPending}
        onOpenChange={setDialogOpen}
        onSave={async (values) => {
          await saveProfileMutation.mutateAsync(values)
        }}
        onTest={async (values) => {
          await testDraftMutation.mutateAsync(values)
        }}
      />

      <AlertDialog
        open={archiveTarget !== null}
        onOpenChange={(open) => !open && setArchiveTarget(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('Archive storage profile?')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t(
                'Profiles referenced by a policy or retained object cannot be archived. Existing data is never deleted by this action.'
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('Cancel')}</AlertDialogCancel>
            <AlertDialogAction
              disabled={archiveMutation.isPending}
              onClick={() => {
                if (archiveTarget) archiveMutation.mutate(archiveTarget.id)
              }}
            >
              {t('Archive')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
