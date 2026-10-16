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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
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
import {
  billingPreviousMonth,
  billingTimestamp,
  billingToday,
} from '@/features/billing/api'
import { statementStatusKeys } from '@/features/billing/constants'
import { handleServerError } from '@/lib/handle-server-error'

import {
  getUpstreamReport,
  getUpstreamStatements,
  getUpstreamStorageProfiles,
  prepareUpstreamStatement,
} from '../api'
import { ReportSummary } from './report-summary'
import { StatementDetail } from './statement-detail'

export function StatementsPanel(props: { groupId: number }) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const [month, setMonth] = useState(billingPreviousMonth)
  const [profileId, setProfileId] = useState<number | null>(null)
  const [status, setStatus] = useState<string | null>('all')
  const [cursors, setCursors] = useState<{ at: number; id: string }[]>([])
  const [selected, setSelected] = useState<string | null>(null)
  const cursor = cursors.at(-1)
  const [year, monthNumber] = month.split('-').map(Number)
  const earliestDraftDate = /^\d{4}-\d{2}$/.test(month)
    ? new Date(Date.UTC(year, monthNumber, 2)).toISOString().slice(0, 10)
    : ''
  const closedMonth = Boolean(
    earliestDraftDate && billingToday() >= earliestDraftDate
  )
  const profiles = useQuery({
    queryKey: ['upstream-billing', 'storage-profiles'],
    queryFn: getUpstreamStorageProfiles,
  })
  const preview = useQuery({
    queryKey: ['upstream-billing', 'report', props.groupId, { month }],
    queryFn: () => getUpstreamReport(props.groupId, { month }),
    enabled: closedMonth,
  })
  const statements = useQuery({
    queryKey: ['upstream-billing', 'statements', props.groupId, status, cursor],
    queryFn: () =>
      getUpstreamStatements(props.groupId, status || 'all', cursor),
    refetchInterval: (query) =>
      query.state.data?.some((item) => item.status === 'preparing')
        ? 3000
        : false,
  })
  const create = useMutation({
    mutationFn: () =>
      prepareUpstreamStatement(props.groupId, month, profileId || 0),
    onSuccess: (statement) => {
      setSelected(statement.id)
      void client.invalidateQueries({ queryKey: ['upstream-billing'] })
    },
    onError: handleServerError,
  })
  const profileOptions = [
    { value: null, label: t('Select storage profile') },
    ...(profiles.data || []).map((item) => ({
      value: item.id,
      label: item.name,
    })),
  ]
  const statusOptions = [
    { value: 'all', label: t('All statuses') },
    ...(['preparing', 'draft', 'confirmed', 'void', 'failed'] as const).map(
      (value) => ({ value, label: t(statementStatusKeys[value]) })
    ),
  ]
  return (
    <div className='flex flex-col gap-5'>
      <Card>
        <CardHeader>
          <CardTitle>{t('Monthly reconciliation')}</CardTitle>
          <CardDescription>
            {t(
              'Preview a closed month, create a draft, then internally confirm and archive its PDF, Excel and usage evidence.'
            )}
          </CardDescription>
        </CardHeader>
        <CardContent className='flex flex-col gap-4'>
          <FieldGroup className='grid gap-4 sm:grid-cols-3'>
            <Field>
              <FieldLabel htmlFor='upstream-draft-month'>
                {t('Statement month')}
              </FieldLabel>
              <Input
                id='upstream-draft-month'
                type='month'
                min='1970-02'
                max={billingPreviousMonth()}
                value={month}
                onChange={(event) => setMonth(event.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor='upstream-storage'>
                {t('Archive storage')}
              </FieldLabel>
              <Select
                items={profileOptions}
                value={profileId}
                onValueChange={setProfileId}
              >
                <SelectTrigger id='upstream-storage' className='w-full'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {profileOptions.map((item) => (
                      <SelectItem
                        key={item.value || 'empty'}
                        value={item.value}
                      >
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
            <div className='flex items-end'>
              <Button
                variant='outline'
                disabled={!closedMonth || preview.isFetching}
                onClick={() => {
                  void preview.refetch()
                  void profiles.refetch()
                }}
              >
                {t('Preview')}
              </Button>
            </div>
          </FieldGroup>
          {!closedMonth && (
            <p role='alert'>
              {t(
                'Select a month that has been closed for at least 24 hours to create an archived draft.'
              )}
            </p>
          )}
          {profiles.isError && (
            <p role='alert'>
              {t('Failed to load archive storage. Refresh to retry.')}
            </p>
          )}
          {profiles.data?.length === 0 && (
            <p role='alert'>
              {t('Configure an enabled Aliyun OSS storage profile first.')}
            </p>
          )}
          {preview.isFetching && <p role='status'>{t('Loading...')}</p>}
          {preview.isError && <p role='alert'>{t('Failed to load data')}</p>}
          {preview.data && closedMonth && !preview.isError && (
            <ReportSummary report={preview.data} />
          )}
          <Button
            className='self-start'
            disabled={
              !closedMonth ||
              !profileId ||
              !profiles.data?.some((item) => item.id === profileId) ||
              !preview.data ||
              preview.isError ||
              preview.isFetching ||
              create.isPending
            }
            onClick={() => create.mutate()}
          >
            {create.isPending ? t('Preparing') : t('Create archived draft')}
          </Button>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className='gap-4'>
          <CardTitle>{t('Upstream statements')}</CardTitle>
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
        </CardHeader>
        <CardContent className='flex flex-col gap-4'>
          {statements.isPending && <p role='status'>{t('Loading...')}</p>}
          {statements.isError && <p role='alert'>{t('Failed to load data')}</p>}
          {statements.data && !statements.isError && (
            <Table aria-label={t('Upstream statements')}>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('Statement month')}</TableHead>
                  <TableHead>{t('Version')}</TableHead>
                  <TableHead>{t('Status')}</TableHead>
                  <TableHead>{t('Created at')}</TableHead>
                  <TableHead>{t('Confirmed at')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {statements.data.map((statement) => (
                  <TableRow key={statement.id}>
                    <TableCell>
                      <Button
                        variant='link'
                        className='h-auto p-0'
                        onClick={() => setSelected(statement.id)}
                      >
                        {statement.month}
                      </Button>
                    </TableCell>
                    <TableCell>{statement.revision}</TableCell>
                    <TableCell>
                      <Badge variant='outline'>
                        {t(statementStatusKeys[statement.status])}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {billingTimestamp(statement.created_at)}
                    </TableCell>
                    <TableCell>
                      {billingTimestamp(statement.confirmed_at || 0)}
                    </TableCell>
                  </TableRow>
                ))}
                {statements.data.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={5}>{t('No records')}</TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          )}
          <div className='flex flex-wrap justify-end gap-2'>
            <Button
              variant='outline'
              disabled={statements.isFetching}
              onClick={() => void statements.refetch()}
            >
              {t('Refresh')}
            </Button>
            <Button
              variant='outline'
              disabled={cursors.length === 0 || statements.isFetching}
              onClick={() => setCursors((values) => values.slice(0, -1))}
            >
              {t('Previous')}
            </Button>
            <Button
              variant='outline'
              disabled={
                statements.data?.length !== 50 ||
                statements.isFetching ||
                statements.isError
              }
              onClick={() => {
                const last = statements.data?.at(-1)
                if (last) {
                  setCursors((values) => [
                    ...values,
                    { at: last.created_at, id: last.id },
                  ])
                }
              }}
            >
              {t('Next')}
            </Button>
          </div>
        </CardContent>
      </Card>
      {selected && (
        <StatementDetail
          key={selected}
          id={selected}
          onClose={() => setSelected(null)}
        />
      )}
    </div>
  )
}
