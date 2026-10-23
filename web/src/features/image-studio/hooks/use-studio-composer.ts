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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { nanoid } from 'nanoid'
import { useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod'

import type { ApiKey } from '@/features/keys/types'
import { useAuthStore } from '@/stores/auth-store'

import {
  assistStudioPrompt,
  createStudioRun,
  listStudioModels,
  uploadStudioReference,
} from '../api'
import type { StudioModelSelection } from '../components/studio-model-settings'
import { studioImageSize } from '../studio-utils'
import type {
  StudioImageSize,
  StudioReference,
  StudioRun,
  StudioSettings,
} from '../types'

export function useStudioComposer(props: {
  authenticated: boolean
  settings: StudioSettings
  keys: ApiKey[]
  runningCount: number
  onGenerated: (run: StudioRun) => void
  onNewCreation: () => void
}) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const userId = useAuthStore((state) => state.auth.user?.id ?? 0)
  const [size, setSize] = useState<StudioImageSize>('auto')
  const [references, setReferences] = useState<StudioReference[]>([])
  const userGroup = useAuthStore((state) => state.auth.user?.group || 'default')
  const maxCharacters = props.settings.max_prompt_characters
  const [text, setText] = useState<StudioModelSelection>({
    tokenId: props.keys[0]?.id ?? 0,
    model: '',
  })
  // Retrying unchanged input replays the same run instead of billing again.
  const requestKey = useRef('')
  const form = useForm({
    resolver: zodResolver(
      z.object({
        tokenId: z.string().min(1),
        model: z.string(),
        prompt: z.string().trim().min(1).max(maxCharacters),
      })
    ),
    defaultValues: {
      tokenId: String(props.keys[0]?.id ?? ''),
      model: '',
      prompt: '',
    },
  })
  const tokenId = Number(form.watch('tokenId'))
  const prompt = form.watch('prompt')
  const selectedModel = form.watch('model')
  const models = useQuery({
    queryKey: ['image-studio', 'models', tokenId],
    queryFn: () => listStudioModels(tokenId),
    enabled: props.authenticated && Boolean(tokenId),
  })
  const model = models.data?.includes(selectedModel)
    ? selectedModel
    : (models.data?.[0] ?? '')
  const textModels = useQuery({
    queryKey: ['image-studio', 'text-models', text.tokenId],
    queryFn: () => listStudioModels(text.tokenId, 'text'),
    enabled: props.authenticated && Boolean(text.tokenId),
  })
  const textModel = textModels.data?.includes(text.model)
    ? text.model
    : (textModels.data?.[0] ?? '')
  const changePrompt = (value: string) => {
    requestKey.current = ''
    form.setValue('prompt', value, { shouldDirty: true, shouldValidate: true })
  }
  const optimize = useMutation({
    mutationFn: (original: string) =>
      assistStudioPrompt({
        token_id: text.tokenId,
        model: textModel,
        prompt: original,
      }),
    onSuccess: (result) => changePrompt(result.suggested_prompt),
  })
  const upload = useMutation({
    mutationFn: async (files: (File | string)[]) => {
      if (
        files.length + references.length >
        props.settings.max_reference_images
      ) {
        throw new Error(
          t('You can add up to {{count}} reference images.', {
            count: props.settings.max_reference_images,
          })
        )
      }
      if (
        files.some((file) => {
          if (typeof file === 'string') return !/^https?:\/\//.test(file)
          return (
            !['image/png', 'image/jpeg', 'image/webp'].includes(file.type) ||
            !file.size ||
            file.size > props.settings.max_reference_bytes
          )
        })
      ) {
        throw new Error(
          t(
            'Reference images must be PNG, JPEG or WebP and no larger than 20 MB.'
          )
        )
      }
      for (const file of files) {
        const reference = await uploadStudioReference(file)
        requestKey.current = ''
        setReferences((current) => [...current, reference])
      }
    },
    onSettled: () =>
      client.invalidateQueries({
        queryKey: ['image-studio', 'references', userId],
      }),
  })
  const resetDraft = () => {
    requestKey.current = ''
    form.resetField('prompt')
    setReferences([])
    upload.reset()
    optimize.reset()
  }
  const generate = useMutation({
    mutationFn: (values: { tokenId: string; prompt: string }) => {
      requestKey.current ||= nanoid()
      const resolvedSize = studioImageSize(values.prompt, size)
      return createStudioRun({
        token_id: Number(values.tokenId),
        idempotency_key: requestKey.current,
        request: {
          model,
          prompt: values.prompt,
          ...(resolvedSize ? { size: resolvedSize } : {}),
          ...(references.length
            ? { reference_ids: references.map((item) => item.id) }
            : {}),
        },
      })
    },
    onSuccess: (run) => {
      resetDraft()
      props.onGenerated(run)
    },
  })
  const generating = generate.isPending
  const busy = generating || optimize.isPending || upload.isPending
  const editable = props.authenticated && props.settings.enabled && !busy
  const canAddReference =
    editable && references.length < props.settings.max_reference_images
  const runLimitReached = props.runningCount >= props.settings.max_running_runs
  const canGenerate =
    editable &&
    !runLimitReached &&
    Boolean(model) &&
    Boolean(prompt.trim()) &&
    prompt.length <= maxCharacters
  const submit = form.handleSubmit((values) => {
    if (canGenerate) generate.mutate(values)
  })

  return {
    form,
    size,
    setSize,
    userGroup,
    maxCharacters,
    text,
    setText,
    requestKey,
    tokenId,
    prompt,
    models,
    model,
    textModels,
    textModel,
    changePrompt,
    optimize,
    generate,
    generating,
    busy,
    canGenerate,
    runLimitReached,
    editable,
    canAddReference,
    submit,
    references,
    addReference: (reference: StudioReference) => {
      if (
        !canAddReference ||
        reference.expires_at <= Date.now() / 1000 ||
        references.some((item) => item.id === reference.id)
      ) {
        return
      }
      requestKey.current = ''
      setReferences((current) => [...current, reference])
    },
    upload,
    uploadReferences: (files: (File | string)[]) => {
      if (!files.length || !editable) {
        return
      }
      upload.mutate(files)
    },
    removeReference: (id: string) => {
      if (busy) return
      requestKey.current = ''
      setReferences((current) => current.filter((item) => item.id !== id))
      upload.reset()
    },
    reusePrompt: (value: string) => {
      if (busy || value.length > maxCharacters) return
      changePrompt(value)
      optimize.reset()
    },
    newCreation: () => {
      if (busy) return
      resetDraft()
      generate.reset()
      props.onNewCreation()
    },
  }
}
