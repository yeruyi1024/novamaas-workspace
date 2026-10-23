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

export function StudioWelcome(props: {
  disabled: boolean
  maxCharacters: number
  onPrompt?: (prompt: string) => void
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
    <section
      className='studio-welcome m-auto flex w-full max-w-4xl flex-col items-center gap-6 px-5 py-8 md:-translate-y-3 md:gap-8'
      aria-label={t('Generation results')}
    >
      <div className='flex flex-col items-center gap-4 text-center md:gap-5'>
        <p className='inline-flex items-center gap-2 rounded-full border border-(--studio-panel-border) bg-(--studio-panel)/70 px-4 py-1.5 text-xs font-medium text-(--studio-eyebrow) backdrop-blur'>
          <span
            className='size-1.5 rounded-full bg-(--studio-accent) shadow-[0_0_8px_var(--studio-accent)]'
            aria-hidden='true'
          />
          {t('A new perspective')}
        </p>
        <h2 className='text-[32px] leading-[1.15] font-bold tracking-[-0.04em] text-balance text-(--studio-title) sm:text-5xl md:text-[56px] xl:text-[60px]'>
          {t('Imagine it.')}{' '}
          <span className='studio-title-gradient'>{t('Make it real.')}</span>
        </h2>
        <p className='max-w-2xl text-sm leading-relaxed text-pretty text-(--studio-description) md:text-base'>
          {t(
            'From a spark of inspiration to a finished image. Bring every idea to life, in your own style.'
          )}
        </p>
      </div>
      {props.onPrompt && (
        <section
          aria-label={t('Try an example')}
          className='flex w-full flex-col items-center gap-3 md:gap-4'
        >
          <h3 className='text-muted-foreground flex items-center gap-2 text-sm font-normal'>
            <HugeiconsIcon
              icon={SparklesIcon}
              className='size-4 text-(--studio-accent)'
              aria-hidden='true'
            />
            {t('Try an example')}
          </h3>
          <div className='grid w-full max-w-sm grid-cols-2 gap-3 sm:flex sm:w-auto sm:max-w-none sm:flex-wrap sm:justify-center'>
            {examples.map((example) => (
              <button
                key={example.name}
                type='button'
                aria-label={example.name}
                title={example.prompt}
                disabled={
                  props.disabled || example.prompt.length > props.maxCharacters
                }
                onClick={() => props.onPrompt?.(example.prompt)}
                className='studio-example-card focus-visible:outline-ring inline-flex min-h-11 items-center justify-center gap-3 rounded-full border border-(--studio-panel-border) bg-(--studio-panel)/80 px-3 py-2 text-sm font-medium text-(--studio-soft-text) backdrop-blur transition-[border-color,background-color,transform] hover:-translate-y-px hover:border-(--studio-hover-border) hover:bg-(--studio-panel) focus-visible:outline-2 focus-visible:outline-offset-3 disabled:cursor-not-allowed disabled:opacity-45 disabled:hover:translate-y-0 sm:min-h-13 sm:min-w-37 sm:ps-2 sm:pe-5'
              >
                <span className='studio-example-icon grid size-8 shrink-0 place-items-center rounded-full'>
                  <HugeiconsIcon
                    icon={example.icon}
                    className='size-4'
                    aria-hidden='true'
                  />
                </span>
                {example.name}
              </button>
            ))}
          </div>
        </section>
      )}
    </section>
  )
}
