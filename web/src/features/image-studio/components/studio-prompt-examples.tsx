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
  Camera01Icon,
  CubeIcon,
  Image01Icon,
  PaintBrush01Icon,
  SparklesIcon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useTranslation } from 'react-i18next'

export function StudioPromptExamples(props: {
  maxCharacters: number
  disabled: boolean
  onChange: (prompt: string) => void
}) {
  const { t } = useTranslation()
  const examples = [
    {
      name: t('Cinematic photography'),
      icon: Camera01Icon,
      prompt: t(
        'A ginger cat on a sunny windowsill, warm afternoon light, cinematic photography, soft shadows and a calm composition.'
      ),
    },
    {
      name: t('Anime illustration'),
      icon: Image01Icon,
      prompt: t(
        'A small seaside town at sunset, an anime illustration with glowing clouds, colorful rooftops and a gentle breeze.'
      ),
    },
    {
      name: t('3D product'),
      icon: CubeIcon,
      prompt: t(
        'A pastel ceramic coffee cup on a studio pedestal, a playful 3D product render with soft lighting and subtle reflections.'
      ),
    },
    {
      name: t('Ink painting'),
      icon: PaintBrush01Icon,
      prompt: t(
        'A mountain lake in morning mist, traditional ink painting, flowing brushwork and generous negative space.'
      ),
    },
  ]
  return (
    <section aria-label={t('Try an example')} className='grid gap-2'>
      <h3 className='flex items-center gap-2 text-sm font-medium'>
        <HugeiconsIcon
          icon={SparklesIcon}
          className='text-primary size-4'
          aria-hidden='true'
        />
        {t('Try an example')}
      </h3>
      <div className='grid grid-cols-2 gap-2'>
        {examples.map((example) => (
          <button
            key={example.name}
            type='button'
            aria-label={example.name}
            title={example.prompt}
            disabled={
              props.disabled || example.prompt.length > props.maxCharacters
            }
            onClick={() => props.onChange(example.prompt)}
            className='bg-muted/20 hover:border-primary/40 hover:bg-primary/5 focus-visible:outline-ring flex min-w-0 items-start gap-1.5 rounded-lg border p-2 text-left transition-colors focus-visible:outline-2 disabled:opacity-50'
          >
            <HugeiconsIcon
              icon={example.icon}
              className='text-primary mt-0.5 size-4 shrink-0'
              aria-hidden='true'
            />
            <span className='grid min-w-0 gap-0.5'>
              <span className='truncate text-xs font-medium'>
                {example.name}
              </span>
              <span className='text-muted-foreground line-clamp-1 text-[11px] leading-relaxed'>
                {example.prompt}
              </span>
            </span>
          </button>
        ))}
      </div>
    </section>
  )
}
