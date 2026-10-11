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
import { Image01Icon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { nanoid } from 'nanoid'
import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod'

import { listMediaAssets } from '@/features/assets/api'
import type { ApiKey } from '@/features/keys/types'
import { useAuthStore } from '@/stores/auth-store'

import { assistStudioPrompt, createStudioRun, listStudioModels } from '../api'
import { assertStudioSuccess } from '../studio-utils'
import type { CreateStudioRunInput, StudioRun, StudioSettings } from '../types'
import { StudioCreationForm } from './studio-creation-form'
import {
  StudioModelSettings,
  type StudioModelSelection,
} from './studio-model-settings'

type StudioComposerProps = {
  settings: StudioSettings
  keys: ApiKey[]
  disabled: boolean
  onGenerated: (run: StudioRun) => void
  children: (
    generating: boolean,
    reusePrompt: (prompt: string) => void
  ) => ReactNode
}

function useStudioCreation(props: StudioComposerProps) {
  const userGroup = useAuthStore((state) => state.auth.user?.group || 'default')
  const [textSelection, setTextSelection] = useState<StudioModelSelection>({
    tokenId: props.keys[0]?.id ?? 0,
    model: '',
  })
  const requestKey = useRef('')
  const form = useForm({
    resolver: zodResolver(
      z.object({
        tokenId: z.string().min(1),
        model: z.string(),
        prompt: z
          .string()
          .trim()
          .min(1)
          .max(props.settings.max_prompt_characters),
        n: z.number().int().min(1).max(props.settings.max_images),
        size: z.string().trim(),
        references: z
          .array(z.string())
          .max(props.settings.max_reference_images),
      })
    ),
    defaultValues: {
      tokenId: String(props.keys[0]?.id ?? ''),
      model: '',
      prompt: '',
      n: 1,
      size: '',
      references: [],
    },
  })
  const tokenId = Number(form.watch('tokenId'))
  const libraryEnabled = props.settings.asset_library_enabled === true
  const references = libraryEnabled ? form.watch('references') : []
  const models = useQuery({
    queryKey: ['image-studio', 'models', tokenId],
    queryFn: async () => assertStudioSuccess(await listStudioModels(tokenId)),
    enabled: Boolean(tokenId),
  })
  const selectedModel = form.watch('model')
  const model = models.data?.includes(selectedModel)
    ? selectedModel
    : (models.data?.[0] ?? '')
  const gptImage =
    model.startsWith('gpt-image-') || model === 'chatgpt-image-latest'
  const dalle2 = model === 'dall-e-2' || model === 'dall-e'
  const sizeOptions = useMemo(() => {
    if (model === 'dall-e-3') {
      return ['1024x1024', '1792x1024', '1024x1792']
    }
    if (dalle2) {
      return ['1024x1024']
    }
    if (model.includes('seedream-4-5') || model.includes('seedream-5-0-lite')) {
      return ['2048x2048', '2496x1664', '1664x2496']
    }
    return ['1024x1024', '1536x1024', '1024x1536']
  }, [model, dalle2])
  let modelMaxImages = props.settings.max_images
  if (model === 'dall-e-3') {
    modelMaxImages = 1
  } else if (dalle2) {
    modelMaxImages = 10
  } else if (gptImage) {
    modelMaxImages = 10
  }
  const size = form.watch('size')
  const previousModel = useRef(model)
  const maxImages = Math.min(props.settings.max_images, modelMaxImages)
  useEffect(() => {
    if (form.getValues('n') > maxImages) {
      form.setValue('n', maxImages, { shouldValidate: true })
      requestKey.current = ''
    }
    if (previousModel.current !== model) {
      previousModel.current = model
      form.setValue('size', '')
      requestKey.current = ''
    }
  }, [form, maxImages, model])
  const textModels = useQuery({
    queryKey: ['image-studio', 'text-models', textSelection.tokenId],
    queryFn: async () =>
      assertStudioSuccess(
        await listStudioModels(textSelection.tokenId, 'text')
      ),
    enabled: Boolean(textSelection.tokenId),
  })
  const textModel = textModels.data?.includes(textSelection.model)
    ? textSelection.model
    : (textModels.data?.[0] ?? '')
  const prompt = form.watch('prompt')
  const changePrompt = (value: string) => {
    requestKey.current = ''
    form.setValue('prompt', value, { shouldDirty: true, shouldValidate: true })
  }
  const optimize = useMutation({
    mutationFn: async (originalPrompt: string) =>
      assertStudioSuccess(
        await assistStudioPrompt({
          token_id: textSelection.tokenId,
          model: textModel,
          prompt: originalPrompt,
        })
      ),
    onSuccess: (result) => changePrompt(result.suggested_prompt),
  })
  const assets = useQuery({
    queryKey: ['image-studio', 'assets'],
    queryFn: async () =>
      assertStudioSuccess(
        await listMediaAssets({ type: 'image', page: 1, pageSize: 30 })
      ),
    enabled: libraryEnabled,
  })
  const generate = useMutation({
    mutationFn: async (
      input: Omit<CreateStudioRunInput, 'idempotency_key'>
    ) => {
      requestKey.current ||= nanoid()
      return assertStudioSuccess(
        await createStudioRun({ ...input, idempotency_key: requestKey.current })
      )
    },
    onSuccess: (run) => {
      requestKey.current = ''
      props.onGenerated(run)
    },
  })
  const generating = generate.isPending || props.disabled
  const submit = form.handleSubmit((values) => {
    if (!props.settings.enabled || !model || generating || optimize.isPending) {
      return
    }
    generate.mutate({
      token_id: Number(values.tokenId),
      request: {
        model,
        prompt: values.prompt,
        ...(values.n > 1 ? { n: values.n } : {}),
        ...(values.size ? { size: values.size } : {}),
      },
      input_asset_ids: references,
    })
  })

  return {
    form,
    userGroup,
    tokenId,
    textSelection,
    setTextSelection,
    prompt,
    model,
    models,
    size,
    sizeOptions,
    maxImages,
    optimize,
    generating,
    libraryEnabled,
    assets,
    references,
    changePrompt,
    submit,
    generate,
    textModel,
    textModels,
    onChange: () => {
      requestKey.current = ''
    },
  }
}

export function StudioComposer(props: StudioComposerProps) {
  const { t } = useTranslation()
  const creation = useStudioCreation(props)
  return (
    <>
      <header className='flex shrink-0 flex-wrap items-center justify-between gap-3 border-b px-4 py-3 sm:px-6'>
        <div className='flex min-w-0 items-center gap-3'>
          <span className='bg-primary/10 text-primary flex size-10 shrink-0 items-center justify-center rounded-xl'>
            <HugeiconsIcon
              icon={Image01Icon}
              className='size-5'
              aria-hidden='true'
            />
          </span>
          <div>
            <h1 className='text-lg font-semibold tracking-tight'>
              {t('Image generation')}
            </h1>
            <p className='text-muted-foreground mt-1 text-xs'>
              {t('Describe an idea. Create something new.')}
            </p>
          </div>
        </div>
        <StudioModelSettings
          keys={props.keys}
          userGroup={creation.userGroup}
          image={{ tokenId: creation.tokenId, model: creation.model }}
          text={{ ...creation.textSelection, model: creation.textModel }}
          imageModels={creation.models}
          textModels={creation.textModels}
          disabled={creation.generating || creation.optimize.isPending}
          onImageChange={(selection) => {
            creation.onChange()
            creation.form.setValue('tokenId', String(selection.tokenId))
            creation.form.setValue('model', selection.model)
          }}
          onTextChange={creation.setTextSelection}
        />
      </header>
      <div className='grid shrink-0 items-start gap-3 p-3 sm:p-4 xl:min-h-0 xl:flex-1 xl:grid-cols-[17rem_minmax(0,1fr)_16rem] xl:grid-rows-[minmax(0,1fr)] xl:items-stretch 2xl:grid-cols-[18rem_minmax(0,1fr)_17rem]'>
        <StudioCreationForm
          settings={props.settings}
          hasKeys={props.keys.length > 0}
          creation={creation}
        />
        {props.children(
          creation.generating || creation.optimize.isPending,
          (value) => {
            if (
              creation.generating ||
              creation.optimize.isPending ||
              value.length > props.settings.max_prompt_characters
            ) {
              return
            }
            creation.changePrompt(value)
            creation.optimize.reset()
            creation.form.setFocus('prompt')
          }
        )}
      </div>
    </>
  )
}

export type StudioCreationState = ReturnType<typeof useStudioCreation>
