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

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import {
  displayMeasured,
  displayThreshold,
  isInformationalRow,
  overallLabel,
  VERDICT_LABEL,
  type Assessment,
} from '../baselines'
import { verdictClass } from '../formatters'

export function AssessmentTable(props: { assessment: Assessment; title?: string }) {
  const { t } = useTranslation()
  const rows = props.assessment.rows.filter((row) => !isInformationalRow(row))
  if (rows.length === 0) return null
  return (
    <div className='mt-4 space-y-2 overflow-x-auto'>
      {props.title ? (
        <p className='text-sm font-medium'>{props.title}</p>
      ) : null}
      <div className={verdictClass(props.assessment.overall)}>
        {t(overallLabel(props.assessment.overall))}
      </div>
      <Table className='min-w-[640px]'>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Metric')}</TableHead>
            <TableHead>{t('Measured')}</TableHead>
            <TableHead>{t('Threshold')}</TableHead>
            <TableHead>{t('Verdict')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.id}>
              <TableCell className='whitespace-nowrap'>{t(row.label)}</TableCell>
              <TableCell className='font-medium whitespace-nowrap'>
                {displayMeasured(row, t)}
              </TableCell>
              <TableCell className='text-muted-foreground'>
                {displayThreshold(row, t)}
              </TableCell>
              <TableCell className={verdictClass(row.verdict)}>
                {t(VERDICT_LABEL[row.verdict])}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
