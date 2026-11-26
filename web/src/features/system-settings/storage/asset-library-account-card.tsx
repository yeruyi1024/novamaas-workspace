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
import { AlertCircleIcon, CloudServerIcon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardFooter } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { cn } from '@/lib/utils'

import { canUpload, pausedUploadCount } from './storage-utils'
import {
  ASSET_PROTOCOL_YOUFANG_REST,
  type AssetLibraryConnection,
} from './types'

export function AssetLibraryAccountCard(props: {
  library: AssetLibraryConnection
  onConfigure: (library: AssetLibraryConnection) => void
  onManageChannels: (library: AssetLibraryConnection) => void
  onViewLogs: (libraryId: number) => void
  onSyncMissing: (library: AssetLibraryConnection) => void
  onRebuild?: (library: AssetLibraryConnection) => void
  onClearFailed: (library: AssetLibraryConnection) => void
  onDelete: (library: AssetLibraryConnection) => void
}) {
  const { t } = useTranslation()
  const library = props.library
  const summary = library.summary
  const deleting = Boolean(library.delete_requested_at)
  const credentialsUnavailable = library.credential_status === 'unavailable'
  const canSync = !deleting && !credentialsUnavailable && canUpload(library)
  const canFill = canSync && !library.filling
  const inFlight = summary.pending + summary.processing
  const pausedUploads = pausedUploadCount(library)
  const percent =
    summary.total > 0 ? Math.round((summary.active / summary.total) * 100) : 0

  let statusLabel = library.enabled ? t('Enabled') : t('Disabled')
  let statusDot = library.enabled ? 'bg-emerald-500' : 'bg-muted-foreground/50'
  if (deleting) {
    statusLabel = t('Deleting...')
    statusDot = 'bg-amber-500'
  }
  const details = [
    {
      label: t('Protocol'),
      value:
        library.protocol === ASSET_PROTOCOL_YOUFANG_REST
          ? t('YooFang REST')
          : t('Volcengine Action compatible'),
    },
    {
      label: t('Authentication'),
      value:
        library.auth_type === 'ak_sk'
          ? t('AK/SK signature')
          : t('Bearer API key'),
    },
    { label: t('Rate limit'), value: `${library.qpm} QPM` },
    { label: t('Base URL'), value: library.base_url, mono: true },
  ]

  return (
    <Card
      size='sm'
      role='region'
      aria-label={library.name}
      className={cn(
        'h-full w-full gap-0 py-0',
        !library.enabled && 'opacity-90'
      )}
    >
      <header className='flex items-start gap-3 px-4 pt-4 pb-3'>
        <div className='bg-primary/10 text-primary flex size-9 shrink-0 items-center justify-center rounded-lg'>
          <HugeiconsIcon
            icon={CloudServerIcon}
            className='size-5'
            aria-hidden='true'
          />
        </div>
        <div className='min-w-0 flex-1'>
          <h3 className='truncate font-semibold' title={library.name}>
            {library.name}
          </h3>
          <p className='text-muted-foreground truncate text-xs'>
            <span title={library.provider_name || library.name}>
              {library.provider_name || library.name}
            </span>
            <span className='tabular-nums'> · ID {library.id}</span>
          </p>
        </div>
        <span className='bg-muted/60 inline-flex shrink-0 items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium'>
          <span
            className={cn('size-1.5 rounded-full', statusDot)}
            aria-hidden='true'
          />
          {statusLabel}
        </span>
      </header>

      {credentialsUnavailable && (
        <p
          role='alert'
          className='bg-destructive/10 text-destructive mx-4 mb-3 flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-xs'
        >
          <HugeiconsIcon
            icon={AlertCircleIcon}
            className='size-3.5 shrink-0'
            aria-hidden='true'
          />
          {t('Credentials unavailable')}
        </p>
      )}

      <section
        aria-label={t('Synchronized')}
        className='bg-muted/30 mx-4 rounded-lg px-3 py-2.5'
      >
        <div className='flex items-baseline justify-between gap-2 text-xs'>
          <span className='text-muted-foreground'>{t('Synchronized')}</span>
          <span className='tabular-nums'>
            <strong className='text-sm font-semibold'>{summary.active}</strong>
            <span className='text-muted-foreground'> / {summary.total}</span>
          </span>
        </div>
        <Progress
          value={percent}
          aria-label={t('Synchronized')}
          className='mt-2 [&_[data-slot=progress-track]]:h-1.5'
        />
        <div className='mt-2 flex flex-wrap items-center gap-1.5 text-xs tabular-nums'>
          {library.filling && (
            <Badge variant='secondary'>{t('Queueing assets...')}</Badge>
          )}
          {inFlight > 0 && (
            <Badge variant='secondary'>
              {t('Synchronizing')} · {inFlight}
            </Badge>
          )}
          {summary.failed > 0 && (
            <Badge variant='destructive'>
              {t('Failed')} · {summary.failed}
            </Badge>
          )}
          {summary.failed + pausedUploads > 0 && (
            <Button
              size='xs'
              variant='ghost'
              className='text-destructive hover:bg-destructive/10 hover:text-destructive'
              disabled={deleting}
              onClick={() => props.onClearFailed(library)}
            >
              {t('Clear failures')}
            </Button>
          )}
          {library.missing > 0 && (
            <Badge variant='warning'>
              {t('Missing assets')} · {library.missing}
            </Badge>
          )}
          <div className='ml-auto flex gap-1'>
            <Button
              size='xs'
              variant='ghost'
              disabled={
                !canFill || (library.missing === 0 && summary.failed === 0)
              }
              onClick={() => props.onSyncMissing(library)}
            >
              {t('Sync missing assets')}
            </Button>
            {props.onRebuild && (
              <Button
                size='xs'
                variant='ghost'
                disabled={!canFill}
                onClick={() => props.onRebuild?.(library)}
              >
                {t('Rebuild sync')}
              </Button>
            )}
          </div>
        </div>
        {library.fill_error && (
          <p className='text-destructive mt-2 text-xs break-words'>
            {t('Background synchronization failed: {{error}}', {
              error: library.fill_error,
            })}
          </p>
        )}
      </section>

      <dl className='grid grid-cols-[auto_minmax(0,1fr)] gap-x-4 gap-y-1.5 px-4 pt-3 text-xs'>
        {details.map((detail) => (
          <div key={detail.label} className='contents'>
            <dt className='text-muted-foreground'>{detail.label}</dt>
            <dd
              className={cn('truncate', detail.mono && 'font-mono')}
              title={detail.value}
            >
              {detail.value}
            </dd>
          </div>
        ))}
      </dl>

      <div
        role='group'
        aria-label={t('Linked channels')}
        className='flex-1 px-4 pt-3 pb-4'
      >
        <p className='text-muted-foreground mb-1.5 flex items-center gap-1.5 text-xs'>
          {t('Linked channels')}
          <span className='bg-muted text-foreground rounded px-1.5 tabular-nums'>
            {library.channels.length}
          </span>
        </p>
        {library.channels.length > 0 ? (
          <ul
            aria-label={t('Linked channels')}
            tabIndex={0}
            className='focus-visible:ring-ring flex max-h-32 flex-col gap-1 overflow-y-auto overscroll-contain rounded-md focus-visible:ring-2 focus-visible:outline-none'
          >
            {library.channels.map((channel) => (
              <li
                key={channel.id}
                className='border-border/60 flex min-w-0 shrink-0 items-center gap-2 rounded-md border px-2 py-1 text-xs'
              >
                <span
                  aria-hidden='true'
                  className={cn(
                    'size-1.5 shrink-0 rounded-full',
                    channel.enabled
                      ? 'bg-emerald-500'
                      : 'bg-muted-foreground/40'
                  )}
                />
                <span
                  className='min-w-0 flex-1 truncate'
                  title={`${channel.name} · ID ${channel.id}`}
                >
                  {channel.name} · ID {channel.id}
                </span>
                {!channel.enabled && (
                  <Badge variant='secondary'>{t('Disabled')}</Badge>
                )}
              </li>
            ))}
          </ul>
        ) : (
          <p className='text-muted-foreground rounded-md border border-dashed px-2.5 py-2 text-xs'>
            {t(
              'Historical assets and logs are retained. Link channels to use this library.'
            )}
          </p>
        )}
      </div>

      <CardFooter className='mt-auto flex-wrap gap-1.5 border-t py-2.5'>
        <Button
          size='sm'
          variant='outline'
          disabled={deleting}
          onClick={() => props.onManageChannels(library)}
        >
          {t('Manage channels')}
        </Button>
        <Button
          size='sm'
          variant='ghost'
          onClick={() => props.onViewLogs(library.id)}
        >
          {t('View logs')}
        </Button>
        <div className='ml-auto flex gap-1.5'>
          {library.channels.length === 0 && (
            <Button
              size='sm'
              variant='destructive'
              disabled={deleting && summary.failed === 0}
              onClick={() => props.onDelete(library)}
            >
              {t('Delete account and copies')}
            </Button>
          )}
          <Button
            size='sm'
            disabled={deleting}
            onClick={() => props.onConfigure(library)}
          >
            {t('Configure')}
          </Button>
        </div>
      </CardFooter>
    </Card>
  )
}
