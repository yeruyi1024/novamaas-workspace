import { zodResolver } from '@hookform/resolvers/zod'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useMemo } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import * as z from 'zod'

import { Checkbox } from '@/components/ui/checkbox'
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
import { getPricing } from '@/features/pricing/api'

import { SettingsForm } from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

const mediaSchema = z.object({
  image_input_models: z.array(z.string()),
  voice_input_models: z.array(z.string()),
  voice_output_models: z.array(z.string()),
  transcription_model: z.string(),
  speech_model: z.string(),
  speech_voice: z.string(),
})
type MediaSettings = z.infer<typeof mediaSchema>
const emptySettings: MediaSettings = {
  image_input_models: [],
  voice_input_models: [],
  voice_output_models: [],
  transcription_model: '',
  speech_model: '',
  speech_voice: '',
}

function parseSettings(raw: string): MediaSettings {
  try {
    return mediaSchema.parse(JSON.parse(raw))
  } catch {
    return emptySettings
  }
}

type ModelListField =
  | 'image_input_models'
  | 'voice_input_models'
  | 'voice_output_models'

export function MiniappMediaCard({ value }: { value: string }) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const pricing = useQuery({
    queryKey: ['pricing', 'miniapp-media'],
    queryFn: getPricing,
  })
  const defaults = useMemo(() => parseSettings(value), [value])
  const form = useForm<MediaSettings>({
    resolver: zodResolver(mediaSchema),
    defaultValues: defaults,
  })
  const modelNames = useMemo(
    () =>
      [
        ...new Set(pricing.data?.data?.map((item) => item.model_name) ?? []),
      ].sort(),
    [pricing.data]
  )

  useEffect(() => {
    form.reset(defaults)
  }, [defaults, form])

  const onSubmit = async (settings: MediaSettings) => {
    if (settings.voice_input_models.length && !settings.transcription_model) {
      form.setError('transcription_model', {
        message: t('Select a transcription model first'),
      })
      return
    }
    if (
      settings.voice_output_models.length &&
      (!settings.speech_model || !settings.speech_voice)
    ) {
      form.setError('speech_model', {
        message: t('Select a speech model and voice first'),
      })
      return
    }
    const normalized = {
      ...settings,
      image_input_models: [...settings.image_input_models].sort(),
      voice_input_models: [...settings.voice_input_models].sort(),
      voice_output_models: [...settings.voice_output_models].sort(),
    }
    if (
      JSON.stringify(normalized) ===
      JSON.stringify({
        ...defaults,
        image_input_models: [...defaults.image_input_models].sort(),
        voice_input_models: [...defaults.voice_input_models].sort(),
        voice_output_models: [...defaults.voice_output_models].sort(),
      })
    ) {
      toast.info(t('No changes to save'))
      return
    }
    const result = await updateOption.mutateAsync({
      key: 'miniapp.playground_media',
      value: JSON.stringify(normalized),
    })
    if (result.success) form.reset(normalized)
  }

  const lists: Array<{ name: ModelListField; title: string; help: string }> = [
    {
      name: 'image_input_models',
      title: 'Models accepting image attachments',
      help: 'Select only chat models whose upstream channel accepts image input.',
    },
    {
      name: 'voice_input_models',
      title: 'Chat models with voice input',
      help: 'Recorded speech is transcribed to text before it reaches the chat model.',
    },
    {
      name: 'voice_output_models',
      title: 'Chat models with voice output',
      help: 'Assistant text is converted to speech by the selected speech model.',
    },
  ]

  return (
    <SettingsSection title={t('Mini Program Media')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)}>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending}
          />
          <p className='text-muted-foreground text-sm'>
            {t(
              'Media features are disabled until configured. Select models only after confirming their upstream channel supports the corresponding endpoint.'
            )}
          </p>
          {lists.map(({ name, title, help }) => (
            <FormField
              key={name}
              control={form.control}
              name={name}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t(title)}</FormLabel>
                  <FormDescription>{t(help)}</FormDescription>
                  <div className='border-input max-h-44 space-y-2 overflow-y-auto rounded-md border p-3'>
                    {modelNames.length === 0 ? (
                      <p className='text-muted-foreground text-sm'>
                        {t('No available models')}
                      </p>
                    ) : (
                      modelNames.map((model) => (
                        <label
                          className='flex cursor-pointer items-center gap-2 text-sm'
                          key={model}
                        >
                          <Checkbox
                            checked={field.value.includes(model)}
                            onCheckedChange={(checked) =>
                              field.onChange(
                                checked
                                  ? [...field.value, model]
                                  : field.value.filter((item) => item !== model)
                              )
                            }
                          />
                          <span>{model}</span>
                        </label>
                      ))
                    )}
                  </div>
                  <FormMessage />
                </FormItem>
              )}
            />
          ))}
          <datalist id='miniapp-media-models'>
            {modelNames.map((model) => (
              <option value={model} key={model} />
            ))}
          </datalist>
          <FormField
            control={form.control}
            name='transcription_model'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Speech transcription model')}</FormLabel>
                <FormControl>
                  <Input
                    list='miniapp-media-models'
                    placeholder='whisper-1'
                    {...field}
                  />
                </FormControl>
                <FormDescription>
                  {t('Must support the audio transcriptions endpoint.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='speech_model'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Speech synthesis model')}</FormLabel>
                <FormControl>
                  <Input
                    list='miniapp-media-models'
                    placeholder='tts-1'
                    {...field}
                  />
                </FormControl>
                <FormDescription>
                  {t('Must support the audio speech endpoint.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='speech_voice'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Speech voice')}</FormLabel>
                <FormControl>
                  <Input placeholder='alloy' {...field} />
                </FormControl>
                <FormDescription>
                  {t('Use a voice name accepted by the selected speech model.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
