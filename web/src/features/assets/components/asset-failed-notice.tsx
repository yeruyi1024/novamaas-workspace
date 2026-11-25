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
import { CircleAlert, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'

/**
 * Surfaces assets that can no longer be used (rejected by review or failed to
 * synchronize) in the current scope, with a shortcut to view or clear them.
 */
export function AssetFailedNotice(props: {
  count: number
  viewing: boolean
  pending: boolean
  onView: () => void
  onClear: () => void
}) {
  const { t } = useTranslation()

  return (
    <div
      role='status'
      className='border-destructive/25 bg-destructive/5 flex flex-wrap items-center gap-x-3 gap-y-2 rounded-xl border px-3 py-2 text-xs'
    >
      <CircleAlert
        className='text-destructive size-4 shrink-0'
        strokeWidth={1.75}
        aria-hidden='true'
      />
      <p className='min-w-0 flex-1'>
        {t(
          '{{count}} failed assets cannot be used (rejected by review or failed to synchronize).',
          { count: props.count }
        )}
      </p>
      <div className='flex shrink-0 gap-1.5'>
        {!props.viewing && (
          <Button size='xs' variant='ghost' onClick={props.onView}>
            {t('View')}
          </Button>
        )}
        <Button
          size='xs'
          variant='outline'
          className='text-destructive hover:bg-destructive/10 hover:text-destructive border-destructive/30'
          disabled={props.pending}
          onClick={props.onClear}
        >
          <Trash2 className='size-3' strokeWidth={1.75} aria-hidden='true' />
          {t('Clear failed assets')}
        </Button>
      </div>
    </div>
  )
}
