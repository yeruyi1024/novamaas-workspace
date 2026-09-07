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

interface RequestBodyDialogProps {
  error: boolean
  json: string | null
  loading: boolean
  onOpenChange: (open: boolean) => void
  open: boolean
}

export function RequestBodyDialog({
  error,
  json,
  loading,
  onOpenChange,
  open,
}: RequestBodyDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={t('Request Body')}
      description={t(
        'Formatted JSON submitted by the user for this video task.'
      )}
      contentClassName='sm:max-w-3xl'
      contentHeight='min(70vh, 42rem)'
    >
      {loading ? (
        <div className='text-muted-foreground flex min-h-40 items-center justify-center gap-2 text-sm'>
          <Spinner />
          <span>{t('Loading request body...')}</span>
        </div>
      ) : null}
      {!loading && error ? (
        <Alert variant='destructive'>
          <AlertCircle />
          <AlertTitle>{t('Request body unavailable')}</AlertTitle>
          <AlertDescription>{t('Request failed')}</AlertDescription>
        </Alert>
      ) : null}
      {!loading && !error && json !== null ? (
        <Suspense
          fallback={
            <div className='flex min-h-40 items-center justify-center'>
              <Spinner />
            </div>
          }
        >
          <RequestJsonViewer json={json} />
        </Suspense>
      ) : null}
    </Dialog>
  )
}
