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
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'

import {
  getImageStudioStoragePolicies,
  listStorageProfiles,
  updateImageStudioStoragePolicy,
} from './api'
import { StoragePolicyForm } from './storage-policy-form'
import {
  storagePolicyToInput,
  type StoragePolicyFormValues,
} from './storage-schemas'
import { assertStorageSuccess, storageErrorMessage } from './storage-utils'

export function ImageStudioStorageSettings() {
  const { t } = useTranslation()
  const client = useQueryClient()
  const query = useQuery({
    queryKey: ['storage', 'image-studio'],
    queryFn: async () => {
      const [profiles, policies] = await Promise.all([
        listStorageProfiles(),
        getImageStudioStoragePolicies(),
      ])
      return {
        profiles: assertStorageSuccess(profiles),
        policies: assertStorageSuccess(policies),
      }
    },
  })
  const save = useMutation({
    mutationFn: async (input: {
      purpose: string
      values: StoragePolicyFormValues
    }) => {
      return assertStorageSuccess(
        await updateImageStudioStoragePolicy(
          input.purpose,
          storagePolicyToInput(input.values)
        )
      )
    },
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['storage', 'image-studio'] })
      toast.success(t('Storage policy saved'))
    },
    onError: (error) => toast.error(storageErrorMessage(error)),
  })
  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Image creation')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='flex flex-col gap-4 pb-4'>
          <p className='text-muted-foreground text-sm'>
            {t(
              'Choose separate storage profiles and retention periods for generated images and references. Asset library copies use their own policy.'
            )}
          </p>
          {query.isPending && (
            <p role='status'>{t('Loading storage settings...')}</p>
          )}
          {query.isError && (
            <div role='alert'>
              <p>{storageErrorMessage(query.error)}</p>
              <Button variant='outline' onClick={() => void query.refetch()}>
                {t('Retry')}
              </Button>
            </div>
          )}
          {query.data?.policies.map((policy) => (
            <StoragePolicyForm
              key={policy.key}
              policy={policy}
              purpose={policy.purpose}
              profiles={query.data?.profiles ?? []}
              isSaving={save.isPending}
              onSave={(values) =>
                save.mutate({ purpose: policy.purpose, values })
              }
            />
          ))}
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
