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
import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { ArrowUpRight, Layers } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { PublicLayout } from '@/components/layout'
import { RichContent } from '@/components/rich-content'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'
import { isHttpUrl, isLikelyHtml } from '@/lib/content-format'

import { getAboutContent } from './api'
import { ProjectNotice } from './components/project-notice'

function EmptyAboutState() {
  const { t } = useTranslation()
  const { systemName } = useSystemConfig()

  return (
    <section className='mx-auto flex min-h-[70vh] max-w-4xl flex-col items-center justify-center gap-6 px-4 py-16 text-center'>
      <div className='maas-feature-icon border-border/50 flex size-16 items-center justify-center rounded-2xl border'>
        <Layers aria-hidden className='maas-accent size-8' strokeWidth={1.5} />
      </div>
      <p className='maas-accent text-xs font-semibold tracking-widest uppercase'>
        {t('Model as a Service')}
      </p>
      <h1 className='maas-gradient-text text-4xl font-bold tracking-tight break-words sm:text-5xl'>
        {systemName}
      </h1>
      <p className='max-w-2xl text-lg font-medium'>
        {t('Unified model services for AI productivity')}
      </p>
      <p className='text-muted-foreground max-w-2xl text-sm leading-7'>
        {t(
          'Connect models, applications and people through one platform. Manage API access, usage and costs in one workspace.'
        )}
      </p>
      <div className='flex flex-wrap justify-center gap-3'>
        <Button render={<Link to='/pricing' />}>{t('Model Square')}</Button>
        <Button
          variant='outline'
          render={
            <a
              href='https://ai.shilijia.xyz/'
              target='_blank'
              rel='noopener noreferrer'
            />
          }
        >
          {t('Explore Shilijia AI')}
          <ArrowUpRight aria-hidden className='size-4' />
        </Button>
      </div>
      <a
        className='text-muted-foreground hover:text-foreground mt-4 text-xs underline underline-offset-4'
        href='https://github.com/yeruyi1024/novamaas-workspace'
        target='_blank'
        rel='noopener noreferrer'
      >
        {t('Source Code')}
      </a>
    </section>
  )
}

export function About() {
  const { t } = useTranslation()
  const { data, isLoading } = useQuery({
    queryKey: ['about-content'],
    queryFn: getAboutContent,
  })

  const rawContent = data?.data?.trim() ?? ''
  const hasContent = rawContent.length > 0
  const isUrl = hasContent && isHttpUrl(rawContent)
  const contentIsHtml = hasContent && isLikelyHtml(rawContent)

  if (isLoading) {
    return (
      <PublicLayout>
        <div className='mx-auto flex max-w-4xl flex-col gap-4 py-12'>
          <Skeleton className='h-8 w-[45%]' />
          <Skeleton className='h-4 w-full' />
          <Skeleton className='h-4 w-[90%]' />
          <Skeleton className='h-4 w-[80%]' />
        </div>
        <ProjectNotice />
      </PublicLayout>
    )
  }

  if (!hasContent) {
    return (
      <PublicLayout appearance='maas'>
        <EmptyAboutState />
        <ProjectNotice />
      </PublicLayout>
    )
  }

  if (isUrl) {
    return (
      <PublicLayout showMainContainer={false}>
        <iframe
          src={rawContent}
          className='h-[calc(100vh-3.5rem)] w-full border-0'
          title={t('About')}
          sandbox='allow-forms allow-popups allow-popups-to-escape-sandbox allow-scripts'
        />
        <ProjectNotice />
      </PublicLayout>
    )
  }

  if (contentIsHtml) {
    return (
      <PublicLayout showMainContainer={false}>
        <RichContent
          mode='html'
          htmlVariant='isolated'
          content={rawContent}
          className='prose-neutral dark:prose-invert max-w-none'
        />
        <ProjectNotice />
      </PublicLayout>
    )
  }

  return (
    <PublicLayout>
      <div className='mx-auto max-w-6xl px-4 py-8'>
        <RichContent
          mode='markdown'
          content={rawContent}
          className='prose-neutral dark:prose-invert max-w-none'
        />
      </div>
      <ProjectNotice />
    </PublicLayout>
  )
}
