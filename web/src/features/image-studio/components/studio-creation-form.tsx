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
import {
  ArrowTurnBackwardIcon,
  Image01Icon,
  SparklesIcon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

import { studioErrorMessage } from '../studio-utils'
import type { StudioSettings } from '../types'
import type { StudioCreationState } from './studio-composer'
import { StudioPromptExamples } from './studio-prompt-examples'

export function StudioCreationForm(props: {
  settings: StudioSettings
  hasKeys: boolean
  creation: StudioCreationState
}) {
  const { t } = useTranslation()
  const creation = props.creation
  const form = creation.form
  const optimize = creation.optimize
  const busy = creation.generating || optimize.isPending

  return (
    <form
      noValidate
      onSubmit={creation.submit}
      onChange={creation.onChange}
      className='bg-card/90 flex min-h-0 min-w-0 flex-col rounded-2xl border shadow-sm xl:overflow-hidden'
    >
      <div className='min-h-0 p-3 xl:flex-1 xl:overflow-y-auto'>
        <FieldGroup className='gap-3'>
          <div className='grid gap-1'>
            <h2 className='font-semibold'>{t('Start with an idea')}</h2>
            <p className='text-muted-foreground flex min-w-0 items-center gap-1.5 text-xs'>
              <HugeiconsIcon
                icon={Image01Icon}
                className='size-3.5 shrink-0'
                aria-hidden='true'
              />
              <span className='truncate' title={creation.model}>
                {creation.model || t('No image models available')}
              </span>
            </p>
          </div>
          {!props.settings.enabled && (
            <p role='status' className='text-muted-foreground text-sm'>
              {t('Image generation is disabled')}
            </p>
          )}
          {!props.hasKeys && (
            <p role='status' className='text-muted-foreground text-sm'>
              {t(
                'No API keys available. Create your first API key to get started.'
              )}
            </p>
          )}
          {creation.models.isError && (
            <p role='alert' className='text-destructive text-xs'>
              {t('Failed to load enabled models')}
            </p>
          )}
          <Field
            data-disabled={busy}
            data-invalid={Boolean(form.formState.errors.prompt)}
          >
            <FieldLabel htmlFor='studio-prompt'>
              {t('Describe your image')}
            </FieldLabel>
            <Textarea
              id='studio-prompt'
              rows={4}
              className='[field-sizing:fixed] min-h-24 resize-y'
              placeholder={t(
                'Describe the subject, scene, lighting and style, or try an example below.'
              )}
              {...form.register('prompt')}
              disabled={busy}
              aria-invalid={Boolean(form.formState.errors.prompt)}
              maxLength={props.settings.max_prompt_characters}
              onKeyDown={(event) => {
                if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
                  event.preventDefault()
                  void creation.submit()
                }
              }}
            />
            {form.formState.errors.prompt && (
              <p role='alert' className='text-destructive text-xs'>
                {t('Describe your image')}
              </p>
            )}
            <div className='flex flex-wrap items-center justify-between gap-2'>
              <Button
                type='button'
                size='sm'
                variant='outline'
                disabled={
                  !props.settings.enabled ||
                  busy ||
                  !creation.textModel ||
                  creation.textModels.isFetching ||
                  !creation.prompt.trim() ||
                  creation.prompt.length > props.settings.max_prompt_characters
                }
                onClick={() => optimize.mutate(creation.prompt)}
              >
                <HugeiconsIcon icon={SparklesIcon} aria-hidden='true' />
                {optimize.isPending ? t('Optimizing...') : t('Optimize prompt')}
              </Button>
              <span className='text-muted-foreground text-[11px]'>
                {creation.prompt.length} /{' '}
                {props.settings.max_prompt_characters}
              </span>
            </div>
            {optimize.isError && (
              <p role='alert' className='text-destructive text-xs'>
                {studioErrorMessage(optimize.error)}
              </p>
            )}
            {optimize.isSuccess &&
              creation.prompt === optimize.data.suggested_prompt && (
                <div className='text-muted-foreground flex items-center justify-between text-xs'>
                  <span role='status'>{t('Prompt optimized')}</span>
                  <Button
                    type='button'
                    size='xs'
                    variant='ghost'
                    disabled={busy}
                    onClick={() => {
                      creation.changePrompt(optimize.variables)
                      optimize.reset()
                    }}
                  >
                    <HugeiconsIcon
                      icon={ArrowTurnBackwardIcon}
                      aria-hidden='true'
                    />
                    {t('Undo')}
                  </Button>
                </div>
              )}
          </Field>
          <StudioPromptExamples
            prompt={creation.prompt}
            maxCharacters={props.settings.max_prompt_characters}
            disabled={!props.settings.enabled || busy}
            onChange={creation.changePrompt}
          />
          <Field data-disabled={busy}>
            <FieldLabel>{t('Image proportions')}</FieldLabel>
            <ToggleGroup
              aria-label={t('Image proportions')}
              aria-describedby='studio-proportions-description'
              value={[creation.size || 'default']}
              onValueChange={(values) => {
                if (values[0]) {
                  form.setValue(
                    'size',
                    values[0] === 'default' ? '' : values[0],
                    {
                      shouldDirty: true,
                    }
                  )
                  creation.onChange()
                }
              }}
              disabled={busy}
              variant='outline'
              spacing={2}
              className='w-full'
            >
              <ToggleGroupItem
                value='default'
                aria-label={t('Model default')}
                className='h-14 min-w-0 flex-1 px-1 text-[11px]'
              >
                {t('Model default')}
              </ToggleGroupItem>
              {creation.sizeOptions.map((value) => {
                const [width, height] = value.split('x').map(Number)
                let name = t('Square image')
                let shape = 'size-3.5'
                if (width > height) {
                  name = t('Landscape')
                  shape = 'h-2.5 w-4'
                }
                if (width < height) {
                  name = t('Portrait')
                  shape = 'h-4 w-2.5'
                }
                return (
                  <ToggleGroupItem
                    key={value}
                    value={value}
                    aria-label={name}
                    className='h-14 min-w-0 flex-1 flex-col gap-1.5 px-1 text-[11px]'
                  >
                    <span
                      aria-hidden='true'
                      className={`rounded-[2px] border ${shape}`}
                    />
                    {name}
                  </ToggleGroupItem>
                )
              })}
            </ToggleGroup>
            <FieldDescription
              id='studio-proportions-description'
              className='text-xs'
            >
              {t(
                'Aspect ratio support depends on the selected model. Model default lets the model choose the size and aspect ratio.'
              )}
            </FieldDescription>
          </Field>
          <FieldGroup className='gap-3'>
            <Field
              data-invalid={Boolean(form.formState.errors.n)}
              data-disabled={busy}
            >
              <FieldLabel htmlFor='studio-n'>{t('Image count')}</FieldLabel>
              <NativeSelect
                id='studio-n'
                className='w-full'
                disabled={busy || creation.maxImages === 1}
                {...form.register('n', { valueAsNumber: true })}
                aria-invalid={Boolean(form.formState.errors.n)}
              >
                {Array.from({ length: creation.maxImages }, (_, index) => (
                  <NativeSelectOption key={index + 1} value={index + 1}>
                    {index + 1}
                  </NativeSelectOption>
                ))}
              </NativeSelect>
              {form.formState.errors.n && (
                <p role='alert' className='text-destructive text-xs'>
                  {t('Maximum images per request')}: {props.settings.max_images}
                </p>
              )}
            </Field>
          </FieldGroup>
          {creation.model === 'dall-e-3' && (
            <p className='text-muted-foreground text-xs'>
              {t('This model supports one image per request.')}
            </p>
          )}
          {creation.libraryEnabled && !!creation.assets.data?.items.length && (
            <fieldset className='grid gap-2'>
              <legend className='mb-2 text-sm font-medium'>
                {t('Reference images')}
              </legend>
              {creation.assets.data.items.map((asset) => (
                <label
                  key={asset.id}
                  className='flex items-center gap-2 text-sm'
                >
                  <input
                    type='checkbox'
                    value={asset.id}
                    {...form.register('references')}
                    disabled={
                      busy ||
                      (!creation.references.includes(asset.id) &&
                        creation.references.length >=
                          props.settings.max_reference_images)
                    }
                  />
                  {asset.name}
                </label>
              ))}
            </fieldset>
          )}
        </FieldGroup>
      </div>
      <div className='shrink-0 border-t p-4'>
        {creation.generate.isError && (
          <p role='alert' className='text-destructive mb-3 text-xs'>
            {studioErrorMessage(creation.generate.error)}
          </p>
        )}
        <Button
          type='submit'
          size='lg'
          className='h-10 w-full'
          disabled={
            !props.settings.enabled ||
            !creation.model ||
            !creation.prompt.trim() ||
            creation.generating ||
            optimize.isPending
          }
        >
          <HugeiconsIcon icon={SparklesIcon} aria-hidden='true' />
          {creation.generating ? t('Generating...') : t('Create images')}
        </Button>
      </div>
    </form>
  )
}
