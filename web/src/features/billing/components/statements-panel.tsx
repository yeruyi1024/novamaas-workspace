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
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import { billingTimestamp, getStatements } from '../api'
import { statementStatusKeys } from '../constants'
import { MonthlyPreviewCard } from './monthly-preview-card'
import { StatementDetail } from './statement-detail'

export function StatementsPanel(props: {
  userId: number
  currentUserId: number
  admin: boolean
  onConfigureIdentity?: () => void
  onSelectDay?: (date: string) => void
}) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const [selected, setSelected] = useState<string | null>(null)
  const [status, setStatus] = useState<string | null>('all')
  const [cursors, setCursors] = useState<{ at: number; id: string }[]>([])
  const cursor = cursors.at(-1)
  const list = useQuery({
    queryKey: ['billing', 'statements', props.userId, cursor, status],
    queryFn: () => getStatements(props.userId, cursor, status ?? undefined),
    refetchInterval: (query) =>
      query.state.data?.some((item) => item.status === 'preparing')
        ? 3000
        : false,
  })
  const statusOptions = [
    { value: 'all', label: t('All statuses') },
    ...Object.entries(statementStatusKeys).map(([value, label]) => ({
      value,
      label: t(label),
    })),
  ]
  return (
    <div className='flex flex-col gap-5'>
      {props.admin && (
        <MonthlyPreviewCard
          userId={props.userId}
          onViewStatement={setSelected}
          onConfigureIdentity={props.onConfigureIdentity}
          onSelectDay={props.onSelectDay}
        />
      )}
      {!props.admin && (
        <p className='text-muted-foreground text-sm'>
          {t(
            'Statements appear here after an administrator issues them. Open a statement to review, confirm, and download the PDF.'
          )}
        </p>
      )}
      <Card>
        <CardHeader className='flex flex-row flex-wrap items-center justify-between gap-3'>
          <CardTitle>{t('Billing statements')}</CardTitle>
          <div className='flex gap-2'>
            <Select
              items={statusOptions}
              value={status}
              onValueChange={(value) => {
                setStatus(value)
                setCursors([])
              }}
            >
              <SelectTrigger aria-label={t('Status')}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {statusOptions.map((item) => (
                    <SelectItem key={item.value} value={item.value}>
                      {item.label}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
            <Button
              variant='outline'
              onClick={() =>
                void client.invalidateQueries({ queryKey: ['billing'] })
              }
            >
              {t('Refresh')}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {list.isPending && <p>{t('Loading...')}</p>}
          {list.isError && <p role='alert'>{t('Failed to load data')}</p>}
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Month')}</TableHead>
                <TableHead>{t('Version')}</TableHead>
                <TableHead>{t('Status')}</TableHead>
                <TableHead>{t('Confirmed at')}</TableHead>
                <TableHead>{t('Actions')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {list.data?.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>{item.month}</TableCell>
                  <TableCell>{item.revision}</TableCell>
                  <TableCell>{t(statementStatusKeys[item.status])}</TableCell>
                  <TableCell>{billingTimestamp(item.confirmed_at)}</TableCell>
                  <TableCell>
                    <Button
                      variant='outline'
                      size='sm'
                      onClick={() => setSelected(item.id)}
                    >
                      {t('View')}
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {list.data?.length === 0 && (
            <p className='text-muted-foreground py-6 text-center'>
              {t('No billing statements yet.')}
            </p>
          )}
        </CardContent>
      </Card>
      <div className='flex justify-end gap-2'>
        <Button
          variant='outline'
          disabled={cursors.length === 0}
          onClick={() => setCursors(cursors.slice(0, -1))}
        >
          {t('Previous')}
        </Button>
        <Button
          variant='outline'
          disabled={list.data?.length !== 100}
          onClick={() => {
            const last = list.data?.at(-1)
            if (last) {
              setCursors([...cursors, { at: last.created_at, id: last.id }])
            }
          }}
        >
          {t('Next')}
        </Button>
      </div>
      <StatementDetail
        id={selected}
        onClose={() => setSelected(null)}
        admin={props.admin}
        currentUserId={props.currentUserId}
      />
    </div>
  )
}
