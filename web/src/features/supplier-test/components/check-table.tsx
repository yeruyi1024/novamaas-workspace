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

import { StatusBadge } from '@/components/status-badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import { statusLabel, statusVariant } from '../formatters'
import type { CheckResult } from '../types'

export function CheckTable(props: {
  checks: CheckResult[]
  busy: boolean
  onRun?: (id: string) => void
}) {
  const { t } = useTranslation()
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{t('Check')}</TableHead>
          <TableHead>{t('Result')}</TableHead>
          <TableHead>{t('Detail')}</TableHead>
          {props.onRun ? <TableHead className='w-24'>{t('Action')}</TableHead> : null}
        </TableRow>
      </TableHeader>
      <TableBody>
        {props.checks.map((check) => (
          <TableRow key={check.id}>
            <TableCell>
              <div>{t(check.title)}</div>
              {check.hintKey ? (
                <p className='text-muted-foreground mt-0.5 max-w-prose text-xs leading-snug'>
                  {t(check.hintKey)}
                </p>
              ) : null}
            </TableCell>
            <TableCell>
              <StatusBadge
                variant={statusVariant(check.status)}
                copyable={false}
                label={t(statusLabel(check.status))}
              />
            </TableCell>
            <TableCell className='text-muted-foreground max-w-xl whitespace-pre-wrap'>
              {check.message ?? ''}
            </TableCell>
            {props.onRun ? (
              <TableCell>
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  disabled={props.busy}
                  onClick={() => props.onRun?.(check.id)}
                >
                  {t('Run this check')}
                </Button>
              </TableCell>
            ) : null}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
