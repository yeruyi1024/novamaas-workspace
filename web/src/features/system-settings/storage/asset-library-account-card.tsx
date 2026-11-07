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
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

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
  onDelete: (library: AssetLibraryConnection) => void
}) {
  const { t } = useTranslation()
  const providerName = props.library.provider_name || props.library.name
  const protocolLabel =
    props.library.protocol === ASSET_PROTOCOL_YOUFANG_REST
      ? t('YooFang REST')
      : t('Volcengine Action compatible')
  const authLabel =
    props.library.auth_type === 'ak_sk'
      ? t('AK/SK signature')
      : t('Bearer API key')
  const connectionLabel = `${protocolLabel} · ${authLabel} · ${props.library.qpm} QPM`
  const deleting = Boolean(props.library.delete_requested_at)
  let statusLabel = props.library.enabled ? t('Enabled') : t('Disabled')
  if (deleting) statusLabel = t('Deleting...')

  return (
    <Card
      size='sm'
      role='region'
      aria-label={props.library.name}
      className='aspect-square min-h-0 w-full max-w-md'
    >
      <CardHeader className='shrink-0'>
        <div className='flex min-w-0 items-center justify-between gap-2'>
          <CardTitle className='min-w-0 flex-1 truncate'>
            <h3 title={props.library.name}>{props.library.name}</h3>
          </CardTitle>
          <Badge variant={props.library.enabled ? 'default' : 'secondary'}>
            {statusLabel}
          </Badge>
        </div>
        <div className='flex min-w-0 flex-wrap items-center gap-1.5'>
          <Badge variant='outline' className='max-w-[50%] min-w-0'>
            <span className='truncate' title={providerName}>
              {providerName}
            </span>
          </Badge>
          <Badge variant='outline'>ID {props.library.id}</Badge>
          <Badge variant='outline'>
            {t('{{count}} linked channels', {
              count: props.library.channels.length,
            })}
          </Badge>
          {props.library.channels.length === 0 && (
            <Badge variant='secondary'>
              {t('Unlinked / historical library')}
            </Badge>
          )}
        </div>
        <CardDescription className='truncate' title={connectionLabel}>
          {connectionLabel}
        </CardDescription>
        <p
          className='text-muted-foreground truncate font-mono text-xs'
          title={props.library.base_url}
        >
          {props.library.base_url}
        </p>
      </CardHeader>
      <CardContent
        role='group'
        aria-label={t('Linked channels')}
        className='flex min-h-0 flex-1 flex-col gap-3 overflow-hidden'
      >
        {props.library.channels.length > 0 ? (
          <ul
            aria-label={t('Linked channels')}
            tabIndex={0}
            className='focus-visible:ring-ring flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto overscroll-contain rounded-md focus-visible:ring-2 focus-visible:outline-none'
          >
            {props.library.channels.map((channel) => (
              <li
                key={channel.id}
                className='bg-muted/40 flex min-w-0 shrink-0 items-center gap-2 rounded-md px-2 py-1 text-xs'
              >
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
          <p className='text-muted-foreground min-h-0 flex-1 overflow-y-auto text-xs'>
            {t(
              'Historical assets and logs are retained. Link channels to use this library.'
            )}
          </p>
        )}
        <div className='flex shrink-0 flex-wrap items-center gap-x-3 gap-y-1 text-xs tabular-nums'>
          <span>
            {t('Synchronized')}{' '}
            <strong>
              {props.library.summary.active} / {props.library.summary.total}
            </strong>
          </span>
          {props.library.summary.pending + props.library.summary.processing >
            0 && (
            <Badge variant='secondary'>
              {t('Synchronizing')} ·{' '}
              {props.library.summary.pending + props.library.summary.processing}
            </Badge>
          )}
          {props.library.summary.failed > 0 && (
            <Badge variant='destructive'>
              {t('Failed')} · {props.library.summary.failed}
            </Badge>
          )}
          {props.library.credential_status === 'unavailable' && (
            <Badge variant='destructive'>{t('Credentials unavailable')}</Badge>
          )}
          {props.library.missing > 0 && (
            <Badge variant='outline'>
              {t('Missing assets')} · {props.library.missing}
            </Badge>
          )}
        </div>
      </CardContent>
      <CardFooter className='shrink-0 flex-wrap justify-end gap-1.5'>
        <Button
          size='xs'
          variant='outline'
          disabled={deleting}
          onClick={() => props.onManageChannels(props.library)}
        >
          {t('Manage channels')}
        </Button>
        <Button
          size='xs'
          variant='outline'
          onClick={() => props.onViewLogs(props.library.id)}
        >
          {t('View logs')}
        </Button>
        <Button
          size='xs'
          variant='outline'
          disabled={
            !props.library.enabled ||
            props.library.credential_status === 'unavailable' ||
            !props.library.channels.some((channel) => channel.enabled) ||
            props.library.missing === 0
          }
          onClick={() => props.onSyncMissing(props.library)}
        >
          {t('Sync missing assets')}
        </Button>
        <Button
          size='xs'
          disabled={deleting}
          onClick={() => props.onConfigure(props.library)}
        >
          {t('Configure')}
        </Button>
        {props.library.channels.length === 0 && (
          <Button
            size='xs'
            variant='destructive'
            disabled={deleting && props.library.summary.failed === 0}
            onClick={() => props.onDelete(props.library)}
          >
            {t('Delete account and copies')}
          </Button>
        )}
      </CardFooter>
    </Card>
  )
}
