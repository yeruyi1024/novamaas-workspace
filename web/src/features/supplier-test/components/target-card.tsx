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
import { ClipboardCheck, Loader2 } from 'lucide-react'
import type { Dispatch, SetStateAction } from 'react'
import { useTranslation } from 'react-i18next'

import { PasswordInput } from '@/components/password-input'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { TitledCard } from '@/components/ui/titled-card'

import type { TargetForm } from '../types'
import { vendorHintKey, VENDOR_OPTIONS, type VendorId } from '../vendors'
import { FieldSelect } from './form-controls'

export function TargetCard(props: {
  target: TargetForm
  busy: boolean
  models: string[]
  isFetchingModels: boolean
  onTargetChange: Dispatch<SetStateAction<TargetForm>>
  onFetchModels: () => void
  onUseThisPlatform: () => void
}) {
  const { t } = useTranslation()

  return (
    <TitledCard
      title={t('Target')}
      description={t(
        'Fill Base URL, key, and model yourself. Pick GLM, Kimi, or DeepSeek to apply that vendor field rules.'
      )}
      icon={<ClipboardCheck />}
      action={
        <div className='flex flex-wrap gap-2'>
          <Button
            variant='outline'
            disabled={props.busy}
            onClick={props.onUseThisPlatform}
          >
            {t('Use this platform')}
          </Button>
          <Button
            variant='outline'
            onClick={props.onFetchModels}
            disabled={props.busy || props.isFetchingModels}
          >
            {props.isFetchingModels ? (
              <Loader2 className='animate-spin' />
            ) : null}
            {t('Fetch models')}
          </Button>
        </div>
      }
    >
      <div className='grid gap-3 md:grid-cols-2 xl:grid-cols-4'>
        <FieldSelect
          label={t('Vendor')}
          value={props.target.vendor}
          disabled={props.busy}
          items={VENDOR_OPTIONS.map((item) => ({
            value: item.id,
            label: t(item.labelKey),
          }))}
          onChange={(value) => {
            const vendor = value as VendorId
            props.onTargetChange((current) => ({ ...current, vendor }))
          }}
        />
        <div className='space-y-2'>
          <Label htmlFor='supplier-base-url'>{t('Base URL')}</Label>
          <Input
            id='supplier-base-url'
            placeholder='https://api.example.com'
            value={props.target.baseUrl}
            disabled={props.busy}
            onChange={(event) =>
              props.onTargetChange((current) => ({
                ...current,
                baseUrl: event.target.value,
              }))
            }
          />
        </div>
        <div className='space-y-2'>
          <Label htmlFor='supplier-api-key'>{t('API Key')}</Label>
          <PasswordInput
            id='supplier-api-key'
            placeholder={t('Supplier API key')}
            value={props.target.apiKey}
            disabled={props.busy}
            onChange={(event) =>
              props.onTargetChange((current) => ({
                ...current,
                apiKey: event.target.value,
              }))
            }
          />
        </div>
        <div className='space-y-2'>
          <Label htmlFor='supplier-model'>{t('Model')}</Label>
          <Input
            id='supplier-model'
            list='supplier-model-options'
            placeholder={t('Type a model ID')}
            value={props.target.model}
            disabled={props.busy}
            onChange={(event) =>
              props.onTargetChange((current) => ({
                ...current,
                model: event.target.value,
              }))
            }
          />
          <datalist id='supplier-model-options'>
            {props.models.map((id) => (
              <option key={id} value={id} />
            ))}
          </datalist>
        </div>
      </div>
      <p className='text-muted-foreground mt-3 text-sm'>
        {t(vendorHintKey(props.target.vendor))}
      </p>
    </TitledCard>
  )
}
