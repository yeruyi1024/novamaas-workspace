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
  FolderPlus,
  KeyRound,
  Layers,
  Upload,
  UserX,
  Webhook,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { useWebhookCapabilities } from '@/features/webhooks/hooks/use-webhook-capabilities'

/**
 * Page actions: expose developer and administrator tools directly in the header
 * toolbar so AK/SK access, webhooks, group management, and bulk admin actions
 * stay one click away instead of hidden behind an overflow menu.
 */
export function AssetLibraryActions(props: {
  isAdmin: boolean
  canUpload: boolean
  onUpload: () => void
  onCreateGroup: () => void
  onOpenAccessKeys: () => void
  onOpenWebhooks: () => void
  onCopyGroups: () => void
  onPurgeUser: () => void
}) {
  const { t } = useTranslation()
  const capabilities = useWebhookCapabilities()
  const webhooksAvailable =
    !capabilities.error &&
    Boolean(
      capabilities.data?.asset_library_enabled ||
      capabilities.data?.manual_enabled
    )

  return (
    <div className='flex flex-wrap items-center justify-end gap-2'>
      <Button size='sm' variant='outline' onClick={props.onOpenAccessKeys}>
        <KeyRound className='size-3.5' strokeWidth={1.75} aria-hidden='true' />
        {t('AK/SK access')}
      </Button>
      {webhooksAvailable && (
        <Button size='sm' variant='outline' onClick={props.onOpenWebhooks}>
          <Webhook className='size-3.5' strokeWidth={1.75} aria-hidden='true' />
          {t('Webhooks')}
        </Button>
      )}
      {props.isAdmin && (
        <>
          <Button size='sm' variant='outline' onClick={props.onCopyGroups}>
            <Layers
              className='size-3.5'
              strokeWidth={1.75}
              aria-hidden='true'
            />
            {t('Copy asset groups')}
          </Button>
          <Button
            size='sm'
            variant='outline'
            className='text-destructive hover:bg-destructive/10 hover:text-destructive border-destructive/30'
            onClick={props.onPurgeUser}
          >
            <UserX className='size-3.5' strokeWidth={1.75} aria-hidden='true' />
            {t('Clear user assets')}
          </Button>
        </>
      )}
      <Button size='sm' variant='outline' onClick={props.onCreateGroup}>
        <FolderPlus
          className='size-3.5'
          strokeWidth={1.75}
          aria-hidden='true'
        />
        {t('New group')}
      </Button>
      <Tooltip>
        <TooltipTrigger render={<span className='inline-flex' />}>
          <Button
            size='sm'
            disabled={!props.canUpload}
            onClick={props.onUpload}
          >
            <Upload
              className='size-3.5'
              strokeWidth={1.75}
              aria-hidden='true'
            />
            {t('Upload asset')}
          </Button>
        </TooltipTrigger>
        {!props.canUpload && (
          <TooltipContent>
            {t('Create an asset group before uploading')}
          </TooltipContent>
        )}
      </Tooltip>
    </div>
  )
}
