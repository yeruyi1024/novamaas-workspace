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

import type { CostCoverage } from '../lib/cost-coverage'

export function CostAmount(props: {
  amount: string
  symbol: string
  coverage: CostCoverage
  align?: 'start' | 'end'
}) {
  const { t } = useTranslation()
  return (
    <span
      className={
        props.align === 'start'
          ? 'inline-flex max-w-40 flex-col items-start text-left'
          : 'inline-flex max-w-40 flex-col items-end text-right'
      }
    >
      <span>
        {props.coverage === 'missing' || props.coverage === 'excluded'
          ? t('Unknown')
          : `${props.symbol} ${props.amount}`}
      </span>
      {props.coverage === 'partial' && (
        <span className='text-muted-foreground text-xs leading-tight whitespace-normal'>
          {t('Known cost subtotal (incomplete)')}
        </span>
      )}
      {props.coverage === 'missing' && (
        <span className='text-muted-foreground text-xs leading-tight'>
          {t('Missing cost')}
        </span>
      )}
      {props.coverage === 'excluded' && (
        <span className='text-muted-foreground text-xs leading-tight whitespace-normal'>
          {t('Not included in cost statistics')}
        </span>
      )}
    </span>
  )
}
