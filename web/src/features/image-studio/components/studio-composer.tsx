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
  Add01Icon,
  Clock01Icon,
  Image01Icon,
  SparklesIcon,
  ArrowDown01Icon,
  ArrowUp01Icon,
  FullScreenIcon,
  MinimizeScreenIcon,
  PanelRightCloseIcon,
  PanelRightOpenIcon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { Link } from '@tanstack/react-router'
import { useRef, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { SystemBrand } from '@/components/layout/components/system-brand'
import { Button } from '@/components/ui/button'
import { Field, FieldLabel } from '@/components/ui/field'
import { Label } from '@/components/ui/label'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import { Spinner } from '@/components/ui/spinner'
import { Textarea } from '@/components/ui/textarea'
import type { ApiKey } from '@/features/keys/types'
import { cn } from '@/lib/utils'

import { useStudioComposer } from '../hooks/use-studio-composer'
import { studioAspectRatios, studioErrorMessage } from '../studio-utils'
import type {
  StudioImageSize,
  StudioReference,
  StudioRun,
  StudioSettings,
} from '../types'
import { StudioModelSettings } from './studio-model-settings'
import { StudioParticles } from './studio-particles'
import { StudioReferenceImages } from './studio-reference-images'

export function StudioComposer(props: {
  authenticated: boolean
  settings: StudioSettings
  keys: ApiKey[]
  runningCount: number
  onGenerated: (run: StudioRun) => void
  onNewCreation: () => void
  children: (actions: {
    generating: boolean
    reusePrompt: (prompt: string) => void
    preview: () => void
    useReference: (reference: StudioReference) => void
    canAddReference: boolean
    referenceIds: string[]
    uploading: boolean
    uploadReferences: (files: File[]) => void
  }) => { canvas: ReactNode; history: ReactNode }
}) {
  const { t } = useTranslation()
  const [editorMode, setEditorMode] = useState<
    'normal' | 'collapsed' | 'expanded'
  >('normal')
  const studio = useStudioComposer({
    ...props,
    onGenerated: (run) => {
      openEditor()
      setHistoryOpen(true)
      props.onGenerated(run)
    },
  })
  const [dragging, setDragging] = useState(false)
  // The records sidebar only collapses on wide screens; narrow layouts stack it.
  const [historyOpen, setHistoryOpen] = useState(false)
  const historyRef = useRef<HTMLDivElement>(null)
  const historyLabel = historyOpen
    ? t('Hide generation records')
    : t('Show generation records')
  const openEditor = () => {
    setEditorMode('normal')
    requestAnimationFrame(() => studio.form.setFocus('prompt'))
  }
  const views = props.children({
    generating: studio.busy,
    reusePrompt: (value) => {
      if (studio.busy) return
      studio.reusePrompt(value)
      openEditor()
    },
    preview: () => setEditorMode('collapsed'),
    useReference: (reference) => {
      studio.addReference(reference)
      openEditor()
    },
    referenceIds: studio.references.map((reference) => reference.id),
    uploading: studio.upload.isPending,
    uploadReferences: (files) => {
      setEditorMode('normal')
      studio.uploadReferences(files)
    },
    canAddReference: studio.canAddReference,
  })
  const newCreation = () => {
    if (studio.busy) return
    studio.newCreation()
    openEditor()
  }
  const showHistory = () => {
    setHistoryOpen(true)
    historyRef.current?.scrollIntoView({ block: 'start' })
    historyRef.current?.focus({ preventScroll: true })
  }

  return (
    <>
      <header className='sticky top-0 z-10 flex min-h-14 shrink-0 items-center justify-between gap-2 border-b border-(--studio-subtle-border) bg-(--studio-surface)/85 px-3 py-2 backdrop-blur-md md:gap-4 md:px-6'>
        <div className='flex min-w-0 items-center gap-2.5 md:gap-5 max-md:[&>a]:px-0 max-md:[&>a>span]:hidden'>
          <SystemBrand variant='inline' />
          <span
            className='bg-border h-6 w-px shrink-0 md:h-7'
            aria-hidden='true'
          />
          <div className='flex min-w-0 items-baseline gap-2'>
            <h1 className='truncate text-sm font-semibold md:text-[15px]'>
              {t('Image creation')}
            </h1>
            <p className='text-muted-foreground hidden text-[11px] md:block'>
              {t('Image studio')}
            </p>
          </div>
        </div>
        <div className='flex min-w-0 items-center gap-1 md:gap-2 max-md:[&>a]:px-2 max-md:[&>a]:text-[11px]'>
          <Button
            type='button'
            variant='ghost'
            size='sm'
            aria-label={t('New creation')}
            disabled={studio.busy}
            onClick={newCreation}
          >
            <HugeiconsIcon icon={Add01Icon} aria-hidden='true' />
            <span className='hidden md:inline'>{t('New creation')}</span>
          </Button>
          <Button
            type='button'
            variant='ghost'
            size='icon'
            className='min-[75rem]:hidden'
            aria-label={t('Generation records')}
            aria-controls='studio-history'
            onClick={showHistory}
          >
            <HugeiconsIcon icon={Clock01Icon} aria-hidden='true' />
          </Button>
          <Button
            type='button'
            variant='outline'
            size='sm'
            className='rounded-full max-[75rem]:hidden'
            aria-label={historyLabel}
            title={historyLabel}
            aria-controls='studio-history'
            aria-expanded={historyOpen}
            onClick={() => setHistoryOpen(!historyOpen)}
          >
            <HugeiconsIcon
              icon={historyOpen ? PanelRightCloseIcon : PanelRightOpenIcon}
              aria-hidden='true'
            />
            {historyLabel}
          </Button>
          <Button
            variant='outline'
            className='rounded-full'
            nativeButton={false}
            role='link'
            render={
              <Link
                to={props.authenticated ? '/dashboard' : '/sign-in'}
                search={
                  props.authenticated
                    ? undefined
                    : { redirect: '/image-studio' }
                }
              />
            }
          >
            {props.authenticated ? t('Back to Dashboard') : t('Sign in')}
          </Button>
        </div>
      </header>
      <div className='flex min-h-0 flex-1 flex-col overflow-y-auto min-[75rem]:flex-row min-[75rem]:overflow-hidden'>
        <section
          aria-label={t('Creative canvas')}
          className='flex min-h-[calc(100svh-56px)] min-w-0 flex-none flex-col p-3 md:p-5 min-[75rem]:min-h-0 min-[75rem]:flex-1'
        >
          <div className='studio-stage relative flex min-h-[calc(100svh-96px)] min-w-0 flex-1 flex-col overflow-hidden rounded-[28px] border border-(--studio-subtle-border) min-[75rem]:min-h-0 min-[75rem]:pb-10'>
            <StudioParticles />
            <div className='flex min-h-[36svh] min-w-0 flex-1 flex-col min-[75rem]:min-h-0'>
              {views.canvas}
            </div>
            <form
              noValidate
              data-dragging={dragging}
              data-has-prompt={Boolean(studio.prompt.trim())}
              onDragOver={(event) => {
                if (
                  !event.dataTransfer.types.includes('Files') &&
                  !event.dataTransfer.types.includes('text/uri-list')
                ) {
                  return
                }
                event.preventDefault()
                if (!studio.busy && props.authenticated) setDragging(true)
              }}
              onDragLeave={(event) => {
                if (
                  !(event.relatedTarget instanceof Node) ||
                  !event.currentTarget.contains(event.relatedTarget)
                ) {
                  setDragging(false)
                }
              }}
              onDrop={(event) => {
                event.preventDefault()
                setDragging(false)
                setEditorMode('normal')
                const files = [...event.dataTransfer.files]
                const url = event.dataTransfer
                  .getData('text/uri-list')
                  .split('\n')
                  .find((line) => /^https?:\/\//.test(line))
                if (files.length) studio.uploadReferences(files)
                else if (url) studio.uploadReferences([url.trim()])
              }}
              onSubmit={studio.submit}
              onChange={() => {
                studio.requestKey.current = ''
              }}
              className='studio-composer mx-auto mb-3 flex max-h-[60svh] w-[calc(100%-24px)] max-w-[820px] shrink-0 flex-col gap-2 overflow-y-auto rounded-3xl border border-(--studio-panel-border) bg-(--studio-panel)/95 p-3 transition-[border-color,box-shadow] focus-within:border-(--studio-accent)/50 data-[dragging=true]:border-dashed data-[dragging=true]:border-(--studio-accent) data-[dragging=true]:bg-(--studio-accent-soft) md:px-4 md:py-3'
              aria-label={t('Image prompt editor')}
            >
              <div className='flex min-w-0 items-center justify-between gap-3'>
                <div className='flex min-w-0 items-center gap-2 rounded-full border border-(--studio-panel-border) bg-(--studio-surface) py-1 ps-1 pe-3 text-xs'>
                  <span className='grid size-6 shrink-0 place-items-center rounded-full bg-(--studio-accent-soft) text-(--studio-accent)'>
                    <HugeiconsIcon
                      icon={Image01Icon}
                      className='size-3.5'
                      aria-hidden='true'
                    />
                  </span>
                  <span
                    className='truncate text-(--studio-soft-text)'
                    title={studio.model}
                  >
                    {props.authenticated
                      ? studio.model || t('No image models available')
                      : t('Sign in to choose a model')}
                  </span>
                </div>
                <div className='flex items-center gap-1'>
                  <StudioModelSettings
                    keys={props.keys}
                    userGroup={studio.userGroup}
                    image={{ tokenId: studio.tokenId, model: studio.model }}
                    text={{ ...studio.text, model: studio.textModel }}
                    imageModels={studio.models}
                    textModels={studio.textModels}
                    disabled={studio.busy || !props.authenticated}
                    onImageChange={(selection) => {
                      studio.requestKey.current = ''
                      studio.form.setValue('tokenId', String(selection.tokenId))
                      studio.form.setValue('model', selection.model)
                    }}
                    onTextChange={studio.setText}
                  />
                  <Button
                    type='button'
                    variant='ghost'
                    size='icon-sm'
                    aria-label={
                      editorMode === 'expanded'
                        ? t('Restore prompt editor')
                        : t('Expand prompt editor')
                    }
                    aria-pressed={editorMode === 'expanded'}
                    onClick={() =>
                      setEditorMode(
                        editorMode === 'expanded' ? 'normal' : 'expanded'
                      )
                    }
                  >
                    <HugeiconsIcon
                      icon={
                        editorMode === 'expanded'
                          ? MinimizeScreenIcon
                          : FullScreenIcon
                      }
                      aria-hidden='true'
                    />
                  </Button>
                  <Button
                    type='button'
                    variant='ghost'
                    size='icon-sm'
                    aria-label={
                      editorMode === 'collapsed'
                        ? t('Show prompt editor')
                        : t('Collapse prompt editor')
                    }
                    aria-expanded={editorMode !== 'collapsed'}
                    aria-controls='studio-editor-body'
                    onClick={() =>
                      setEditorMode(
                        editorMode === 'collapsed' ? 'normal' : 'collapsed'
                      )
                    }
                  >
                    <HugeiconsIcon
                      icon={
                        editorMode === 'collapsed'
                          ? ArrowUp01Icon
                          : ArrowDown01Icon
                      }
                      aria-hidden='true'
                    />
                  </Button>
                </div>
              </div>
              {editorMode === 'collapsed' && (
                <p className='text-muted-foreground truncate text-xs'>
                  {studio.prompt || t('Describe your image')}
                </p>
              )}
              <div
                id='studio-editor-body'
                hidden={editorMode === 'collapsed'}
                className='flex flex-col gap-2'
              >
                {!props.settings.enabled && (
                  <p
                    role='status'
                    className='text-muted-foreground text-[11px] leading-[1.6]'
                  >
                    {t('Image generation is disabled')}
                  </p>
                )}
                {props.authenticated && !props.keys.length && (
                  <p
                    role='status'
                    className='text-muted-foreground text-[11px] leading-[1.6]'
                  >
                    {t(
                      'No API keys available. Create your first API key to get started.'
                    )}{' '}
                    <Link
                      to='/keys'
                      className='text-primary underline underline-offset-4'
                    >
                      {t('API Keys')}
                    </Link>
                  </p>
                )}
                {studio.models.isError && (
                  <p role='alert' className='text-destructive text-xs'>
                    {t('Failed to load enabled models')}
                  </p>
                )}
                <Field
                  data-disabled={studio.busy}
                  data-invalid={Boolean(studio.form.formState.errors.prompt)}
                  className='gap-2'
                >
                  <FieldLabel className='sr-only' htmlFor='studio-prompt'>
                    {t('Describe your image')}
                  </FieldLabel>
                  <Textarea
                    id='studio-prompt'
                    rows={1}
                    className={cn(
                      'studio-prompt-input resize-none rounded-none border-0 px-1 py-1 text-sm leading-[1.75] focus-visible:ring-0',
                      editorMode === 'expanded' && 'studio-prompt-expanded'
                    )}
                    placeholder={t('Describe your image')}
                    {...studio.form.register('prompt')}
                    disabled={studio.busy}
                    aria-invalid={Boolean(studio.form.formState.errors.prompt)}
                    aria-describedby='studio-prompt-help'
                    maxLength={studio.maxCharacters}
                    onKeyDown={(event) => {
                      if (
                        (event.metaKey || event.ctrlKey) &&
                        event.key === 'Enter' &&
                        !event.nativeEvent.isComposing
                      ) {
                        event.preventDefault()
                        void studio.submit()
                      }
                    }}
                  />
                </Field>
                {studio.upload.isError && (
                  <p role='alert' className='text-destructive text-xs'>
                    {t(studioErrorMessage(studio.upload.error))}
                  </p>
                )}
                {studio.optimize.isError && (
                  <p role='alert' className='text-destructive text-xs'>
                    {studioErrorMessage(studio.optimize.error)}
                  </p>
                )}
                {studio.runLimitReached && (
                  <p role='status' className='text-muted-foreground text-xs'>
                    {t(
                      'Up to {{count}} images can be generating at once. You can create more when one finishes.',
                      { count: props.settings.max_running_runs }
                    )}
                  </p>
                )}
                {studio.generate.isError && (
                  <p role='alert' className='text-destructive text-xs'>
                    {t(studioErrorMessage(studio.generate.error))}
                  </p>
                )}
                {studio.optimize.isSuccess &&
                  studio.prompt === studio.optimize.data.suggested_prompt && (
                    <div className='text-muted-foreground flex items-center gap-2 text-xs'>
                      <span role='status'>{t('Prompt optimized')}</span>
                      <Button
                        type='button'
                        size='xs'
                        variant='ghost'
                        disabled={studio.busy}
                        onClick={() => {
                          if (studio.optimize.variables !== undefined) {
                            studio.changePrompt(studio.optimize.variables)
                          }
                          studio.optimize.reset()
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
                <div className='studio-composer-actions flex flex-wrap items-end justify-between gap-2 md:gap-3'>
                  <div className='flex min-w-0 flex-1 flex-wrap items-center gap-1.5 md:gap-2'>
                    <Label
                      htmlFor='studio-aspect-ratio'
                      className='text-muted-foreground shrink-0 text-xs'
                    >
                      {t('Aspect ratio')}
                    </Label>
                    <NativeSelect
                      id='studio-aspect-ratio'
                      value={studio.size}
                      className='h-8 w-44 [&>select]:rounded-full [&>select]:text-xs'
                      disabled={studio.busy || !props.authenticated}
                      onChange={(event) => {
                        studio.requestKey.current = ''
                        studio.setSize(event.target.value as StudioImageSize)
                      }}
                    >
                      <NativeSelectOption value='auto'>
                        {t('Auto (from prompt)')}
                      </NativeSelectOption>
                      {Object.entries(studioAspectRatios).map(
                        ([ratio, size]) => (
                          <NativeSelectOption key={ratio} value={size}>
                            {ratio}
                          </NativeSelectOption>
                        )
                      )}
                    </NativeSelect>
                    <StudioReferenceImages
                      references={studio.references}
                      uploading={studio.upload.isPending}
                      disabled={!studio.editable}
                      maxImages={props.settings.max_reference_images}
                      onUpload={studio.uploadReferences}
                      onRemove={studio.removeReference}
                    />
                    <Button
                      type='button'
                      size='sm'
                      variant='ghost'
                      disabled={
                        !studio.editable ||
                        !studio.textModel ||
                        studio.textModels.isFetching ||
                        !studio.prompt.trim() ||
                        studio.prompt.length > studio.maxCharacters
                      }
                      className='rounded-full'
                      onClick={() => studio.optimize.mutate(studio.prompt)}
                    >
                      <HugeiconsIcon icon={SparklesIcon} aria-hidden='true' />
                      {studio.optimize.isPending
                        ? t('Optimizing...')
                        : t('Optimize prompt')}
                    </Button>
                  </div>
                  {props.authenticated ? (
                    <Button
                      type='submit'
                      size='lg'
                      className='studio-generate min-h-10 min-w-[110px] rounded-full px-5 font-semibold text-zinc-950 md:min-w-[136px]'
                      disabled={!studio.canGenerate}
                    >
                      {studio.generating ? (
                        <Spinner />
                      ) : (
                        <HugeiconsIcon icon={SparklesIcon} aria-hidden='true' />
                      )}
                      {studio.generating
                        ? t('Generating...')
                        : t('Create images')}
                    </Button>
                  ) : (
                    <Button
                      size='lg'
                      className='studio-generate min-h-10 rounded-full px-5 font-semibold text-zinc-950'
                      nativeButton={false}
                      role='link'
                      render={
                        <Link
                          to='/sign-in'
                          search={{ redirect: '/image-studio' }}
                        />
                      }
                    >
                      {t('Sign in to create images')}
                    </Button>
                  )}
                </div>
                <p
                  className='text-muted-foreground px-1 text-[11px] leading-relaxed'
                  aria-live='polite'
                >
                  {dragging
                    ? t('Release to add reference images')
                    : t(
                        'Drag local images or history thumbnails into the prompt to add references.'
                      )}
                </p>
                <div className='studio-composer-help'>
                  <div
                    id='studio-prompt-help'
                    className='text-muted-foreground flex min-w-0 justify-between gap-3 border-t border-(--studio-subtle-border) pt-2 text-[11px]'
                  >
                    <span className='truncate'>
                      {t('Text model')}:{' '}
                      {props.authenticated
                        ? studio.textModel || t('No models available')
                        : t('Sign in to choose a model')}
                      <span className='hidden lg:inline'>
                        {' · '}
                        {t(
                          'Drop images here. References expire automatically.'
                        )}
                      </span>
                    </span>
                    <span className='shrink-0 tabular-nums'>
                      {studio.prompt.length} / {studio.maxCharacters}
                      <kbd className='ml-3 hidden rounded border border-(--studio-panel-border) bg-(--studio-surface) px-1.5 py-px font-[inherit] md:inline'>
                        ⌘ / Ctrl ↵
                      </kbd>
                    </span>
                  </div>
                </div>
              </div>
            </form>
            <p className='text-muted-foreground/70 mb-3 shrink-0 px-4 text-center text-[11px] leading-relaxed'>
              {t(
                'Generated images expire automatically. Download them in time.'
              )}
            </p>
          </div>
        </section>
        <div
          id='studio-history'
          ref={historyRef}
          tabIndex={-1}
          className={cn(
            'flex min-h-[280px] min-w-0 flex-none border-t border-(--studio-subtle-border) bg-(--studio-surface) -outline-offset-3 min-[75rem]:min-h-0 min-[75rem]:w-[300px] min-[75rem]:border-t-0 min-[75rem]:border-l',
            !historyOpen && 'min-[75rem]:hidden'
          )}
        >
          {views.history}
        </div>
      </div>
    </>
  )
}
