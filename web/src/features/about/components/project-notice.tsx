import { useTranslation } from 'react-i18next'

export function ProjectNotice() {
  const { t } = useTranslation()
  return (
    <section
      className='border-border/60 bg-muted/20 mx-auto mb-8 max-w-4xl rounded-2xl border px-6 py-6 text-sm'
      aria-labelledby='project-notice-title'
    >
      <h2 id='project-notice-title' className='mb-3 font-semibold'>
        {t('Open-source information')}
      </h2>
      <p className='text-muted-foreground leading-relaxed'>
        {t('This platform is a localized fork of New API v1.0.0-rc.26.')}
      </p>
      <p className='text-muted-foreground mt-2 leading-relaxed'>
        {t('Frontend design and development by New API contributors.')}
      </p>
      <div className='mt-4 flex flex-wrap gap-x-5 gap-y-2'>
        <a
          className='text-primary underline underline-offset-4'
          href='https://github.com/QuantumNous/new-api'
          target='_blank'
          rel='noopener noreferrer'
        >
          New API · QuantumNous
        </a>
        <a
          className='text-primary underline underline-offset-4'
          href='https://github.com/yeruyi1024/novamaas-workspace/blob/main/LICENSE'
          target='_blank'
          rel='noopener noreferrer'
        >
          {t('AGPL v3.0 License')}
        </a>
        <a
          className='text-primary underline underline-offset-4'
          href='https://github.com/yeruyi1024/novamaas-workspace'
          target='_blank'
          rel='noopener noreferrer'
        >
          {t('Source Code')}
        </a>
      </div>
    </section>
  )
}
