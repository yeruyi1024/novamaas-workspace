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
import { AlertCircle } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Dialog } from '@/components/dialog'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'

import type { TaskRequestSnapshots } from '../../task-content-api'
import { RequestSnapshotsViewer } from './request-snapshots-viewer'

interface RequestBodyDialogProps {
  error: boolean
  loading: boolean
  onOpenChange: (open: boolean) => void
  open: boolean
  snapshots: TaskRequestSnapshots | null
}

export function RequestBodyDialog(props: RequestBodyDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog
      open={props.open}
      onOpenChange={props.onOpenChange}
      title={t('Request Snapshots')}
      description={t(
        'Original client request and actual upstream request snapshots for this video task.'
      )}
      contentClassName='sm:max-w-3xl'
      contentHeight='min(70vh, 42rem)'
    >
      {props.loading ? (
        <div className='text-muted-foreground flex min-h-40 items-center justify-center gap-2 text-sm'>
          <Spinner />
          <span>{t('Loading request body...')}</span>
        </div>
      ) : null}
      {!props.loading && props.error ? (
        <Alert variant='destructive'>
          <AlertCircle />
          <AlertTitle>{t('Request body unavailable')}</AlertTitle>
          <AlertDescription>{t('Request failed')}</AlertDescription>
        </Alert>
      ) : null}
      {!props.loading && !props.error && props.snapshots !== null ? (
        <RequestSnapshotsViewer
          original={props.snapshots.original}
          upstream={props.snapshots.upstream}
        />
      ) : null}
    </Dialog>
  )
}
