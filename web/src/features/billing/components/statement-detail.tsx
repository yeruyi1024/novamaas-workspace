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
import { useId, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Field, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { handleServerError } from '@/lib/handle-server-error'

import {
  billingTimestamp,
  downloadStatement,
  getStatement,
  statementAction,
} from '../api'
import { statementStatusKeys, statementActionKeys } from '../constants'
import type { BillingSnapshot, StatementDetail as Detail } from '../types'
import { BillingRows } from './billing-rows'

export function StatementDetail(props: {
  id: string | null
  onClose: () => void
  admin: boolean
  currentUserId: number
}) {
  const { t } = useTranslation()
  const query = useQuery({
    queryKey: ['billing', 'statement', props.id],
    queryFn: () => getStatement(props.id ?? ''),
    enabled: Boolean(props.id),
    refetchInterval: (state) =>
      state.state.data?.statement.status === 'preparing' ? 3000 : false,
  })
  return (
    <Dialog
      open={Boolean(props.id)}
      onOpenChange={(open) => !open && props.onClose()}
    >
      <DialogContent className='max-h-[90dvh] overflow-y-auto sm:max-w-4xl'>
        <DialogHeader>
          <DialogTitle>{t('Billing statement')}</DialogTitle>
          <DialogDescription>
            {t(
              'Review the frozen billing identity and daily amounts before confirming.'
            )}
          </DialogDescription>
        </DialogHeader>
        {query.isPending && <p>{t('Loading...')}</p>}
        {query.isError && <p role='alert'>{t('Failed to load data')}</p>}
        {query.data && (
          <StatementContent
            key={`${query.data.statement.id}:${query.data.statement.status}`}
            detail={query.data}
            admin={props.admin}
            currentUserId={props.currentUserId}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}
function StatementContent(props: {
  detail: Detail
  admin: boolean
  currentUserId: number
}) {
  const { t } = useTranslation()
  const id = useId()
  const client = useQueryClient()
  const [acknowledged, setAcknowledged] = useState(false)
  const [note, setNote] = useState('')
  const [intent, setIntent] = useState('')
  const [part, setPart] = useState(1)
  const statement = props.detail.statement
  const snapshot = JSON.parse(statement.snapshot) as BillingSnapshot
  const customerName =
    snapshot.username || props.detail.customer?.username || t('Unknown')
  const customerLabel = `${customerName} (#${statement.user_id})`
  const sourceBlocked = Boolean(props.detail.source_warning)
  const mutate = useMutation({
    mutationFn: (action: string) =>
      statementAction(statement.id, action, statement.manifest_sha256, note),
    onSuccess: () => {
      setIntent('')
      void client.invalidateQueries({ queryKey: ['billing'] })
    },
    onError: handleServerError,
  })
  const download = useMutation({
    mutationFn: (file: { kind: string; ordinal?: number }) =>
      downloadStatement(statement.id, file.kind, file.ordinal),
    onError: handleServerError,
  })
  const canConfirm =
    statement.user_id === props.currentUserId &&
    (statement.status === 'issued' || statement.status === 'disputed')
  return (
    <div className='flex flex-col gap-5'>
      <div className='grid gap-2 rounded-lg border p-4 text-sm sm:grid-cols-2'>
        <p className='break-words sm:col-span-2'>
          {t('Statement customer')}: {customerLabel}
        </p>
        <p>
          {t('Month')}: {statement.month} / {t('Version')} {statement.revision}
        </p>
        <p>
          {t('Status')}: {t(statementStatusKeys[statement.status])}
        </p>
        <p>
          {t('Company title')}: {snapshot.company_title}
        </p>
        <p>
          {t('Tax ID')}: {snapshot.tax_id}
        </p>
        <p>
          {t('Issued at')}: {billingTimestamp(statement.issued_at)}
        </p>
        <p>
          {t('Confirmed at')}: {billingTimestamp(statement.confirmed_at)}
        </p>
        <p>
          {t('Confirmation due')}: {billingTimestamp(statement.due_at)}
        </p>
        <p>{snapshot.currency.code} / Asia/Shanghai</p>
      </div>
      {sourceBlocked && (
        <p
          role='alert'
          className='border-destructive text-destructive rounded-lg border p-4 text-sm'
        >
          {t(
            'Historical usage exists but this statement has no formal entries. Issuing and confirmation are blocked. An administrator must void the draft and reconcile historical data before creating a new version.'
          )}
        </p>
      )}
      {statement.last_error && (
        <p role='alert' className='text-destructive'>
          {t('Archive preparation failed. Check server logs and retry.')}
        </p>
      )}
      <BillingRows
        rows={snapshot.days}
        total={snapshot.total}
        symbol={snapshot.currency.symbol}
        roundingDifference={snapshot.rounding_difference}
      />
      <p className='text-muted-foreground text-sm'>
        {t(
          'Amounts use six decimals. Funding and temporary reservations are not consumption. This statement is not an invoice or bank receipt.'
        )}
      </p>
      <div className='flex flex-wrap gap-2'>
        {props.detail.artifacts
          .filter((item) => item.kind === 'pdf' || item.kind === 'manifest')
          .map((item) => (
            <Button
              key={item.id}
              variant='outline'
              disabled={download.isPending}
              onClick={() => download.mutate({ kind: item.kind })}
            >
              {item.kind === 'pdf' ? t('Download PDF') : t('Download manifest')}
            </Button>
          ))}
        {statement.status === 'confirmed' && (
          <Button
            variant='outline'
            disabled={download.isPending}
            onClick={() => download.mutate({ kind: 'receipt' })}
          >
            {t('Download confirmation receipt')}
          </Button>
        )}
        {props.detail.artifacts.some((item) => item.kind === 'details') && (
          <details className='w-full'>
            <summary className='cursor-pointer py-2'>
              {t('Archived detail files')}
            </summary>
            <div className='flex max-h-48 flex-wrap gap-2 overflow-auto'>
              {props.detail.artifacts
                .filter((item) => item.kind === 'details')
                .map((item) => (
                  <Button
                    key={item.id}
                    size='sm'
                    variant='outline'
                    disabled={download.isPending}
                    onClick={() =>
                      download.mutate({
                        kind: 'details',
                        ordinal: item.ordinal,
                      })
                    }
                  >
                    {t('Part')} {item.ordinal + 1} ({item.rows})
                  </Button>
                ))}
            </div>
          </details>
        )}
      </div>
      {(props.detail.detail_count ?? 0) > 200 && (
        <Field>
          <FieldLabel htmlFor={`${id}-part`}>
            {t('Part')} (1 - {props.detail.detail_count})
          </FieldLabel>
          <div className='flex gap-2'>
            <Input
              id={`${id}-part`}
              type='number'
              min={1}
              max={props.detail.detail_count}
              value={part}
              onChange={(event) => setPart(Number(event.target.value))}
            />
            <Button
              variant='outline'
              disabled={
                download.isPending ||
                !Number.isInteger(part) ||
                part < 1 ||
                part > (props.detail.detail_count ?? 0)
              }
              onClick={() =>
                download.mutate({ kind: 'details', ordinal: part - 1 })
              }
            >
              {t('Download')}
            </Button>
          </div>
        </Field>
      )}
      {statement.manifest_sha256 && (
        <p className='text-muted-foreground font-mono text-xs break-all'>
          SHA-256: {statement.manifest_sha256}
        </p>
      )}
      {canConfirm && (
        <Field orientation='horizontal'>
          <Checkbox
            id={id}
            checked={acknowledged}
            onCheckedChange={(value) => setAcknowledged(value === true)}
          />
          <FieldLabel htmlFor={id}>
            {t('I have checked this statement and agree to confirm it.')}
          </FieldLabel>
        </Field>
      )}
      {(canConfirm || props.admin) &&
        statement.status !== 'confirmed' &&
        statement.status !== 'void' && (
          <Field>
            <FieldLabel htmlFor={`${id}-note`}>
              {t('Reason / reply')}
            </FieldLabel>
            <Textarea
              id={`${id}-note`}
              value={note}
              onChange={(event) => setNote(event.target.value)}
              maxLength={2000}
            />
          </Field>
        )}
      <div className='flex flex-wrap gap-2'>
        {canConfirm && (
          <Button
            disabled={!acknowledged || mutate.isPending || sourceBlocked}
            onClick={() => setIntent('confirm')}
          >
            {t('Confirm statement')}
          </Button>
        )}
        {canConfirm && statement.status === 'issued' && (
          <Button
            variant='outline'
            disabled={!note.trim() || mutate.isPending}
            onClick={() => setIntent('dispute')}
          >
            {t('Raise a dispute')}
          </Button>
        )}
        {props.admin && statement.status === 'draft' && (
          <Button
            disabled={sourceBlocked || mutate.isPending}
            onClick={() => setIntent('issue')}
          >
            {t('Issue to customer')}
          </Button>
        )}
        {props.admin && statement.status === 'disputed' && (
          <Button
            variant='outline'
            disabled={!note.trim() || mutate.isPending}
            onClick={() => mutate.mutate('reply')}
          >
            {t('Reply')}
          </Button>
        )}
        {props.admin && statement.status === 'failed' && (
          <Button variant='outline' onClick={() => mutate.mutate('retry')}>
            {t('Retry')}
          </Button>
        )}
        {props.admin &&
          statement.status !== 'confirmed' &&
          statement.status !== 'void' && (
            <Button
              variant='destructive'
              disabled={!note.trim()}
              onClick={() => setIntent('void')}
            >
              {t('Void statement')}
            </Button>
          )}
      </div>
      <div className='flex flex-col gap-2'>
        <h3 className='font-medium'>{t('Statement history')}</h3>
        {props.detail.events.map((event) => (
          <p key={event.id} className='text-muted-foreground text-sm'>
            {billingTimestamp(event.created_at)} · {t('Operator')}:{' '}
            {event.actor_username || t('Unknown')} (#{event.actor_id}) ·{' '}
            {t(statementActionKeys[event.action] ?? 'Unknown')} {event.note}
          </p>
        ))}
      </div>
      <ConfirmDialog
        open={Boolean(intent)}
        onOpenChange={(open) => !open && setIntent('')}
        title={t('Confirm action')}
        desc={`${t('Statement customer')}: ${customerLabel}. ${t('This action will be recorded in the statement history. Confirmed statements cannot be overwritten.')}`}
        confirmText={t('Confirm')}
        isLoading={mutate.isPending}
        destructive={intent === 'void'}
        handleConfirm={() => mutate.mutate(intent)}
      />
    </div>
  )
}
