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
import { lazy, Suspense } from 'react'
import { useTranslation } from 'react-i18next'

import { Dialog } from '@/components/dialog'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'

const RequestJsonViewer = lazy(() => import('./request-json-viewer'))

interface TaskInformationDialogProps {
  error: boolean
  json: string | null
  loading: boolean
  onOpenChange: (open: boolean) => void
  open: boolean
}

export function TaskInformationDialog(props: TaskInformationDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog
      open={props.open}
      onOpenChange={props.onOpenChange}
      title={t('Task Information')}
      description={t(
        'Latest task information returned by the task query endpoint.'
      )}
      contentClassName='sm:max-w-3xl'
      contentHeight='min(70vh, 42rem)'
    >
      {props.loading ? (
        <div className='text-muted-foreground flex min-h-40 items-center justify-center gap-2 text-sm'>
          <Spinner />
          <span>{t('Loading task information...')}</span>
        </div>
      ) : null}
      {!props.loading && props.error ? (
        <Alert variant='destructive'>
          <AlertCircle />
          <AlertTitle>{t('Task information unavailable')}</AlertTitle>
          <AlertDescription>{t('Request failed')}</AlertDescription>
        </Alert>
      ) : null}
      {!props.loading && !props.error && props.json !== null ? (
        <Suspense
          fallback={
            <div className='flex min-h-40 items-center justify-center'>
              <Spinner />
            </div>
          }
        >
          <RequestJsonViewer
            filename='task-information.json'
            json={props.json}
          />
        </Suspense>
      ) : null}
    </Dialog>
  )
}
