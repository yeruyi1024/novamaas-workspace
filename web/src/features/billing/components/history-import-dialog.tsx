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
import { toast } from 'sonner'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { Badge } from '@/components/ui/badge'
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
import { Textarea } from '@/components/ui/textarea'
import { handleServerError } from '@/lib/handle-server-error'

import {
  billingTimestamp,
  confirmBillingHistoryImport,
  downloadBillingHistorySource,
  getBillingHistoryImports,
  reviewBillingHistory,
} from '../api'
import type { BillingHistoryReview } from '../types'
import { BillingRows } from './billing-rows'

export function HistoryImportDialog(props: {
  userId: number
  month: string
  profileId: number | null
  onViewStatement: (id: string) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const review = useQuery({
    queryKey: ['billing', 'history-review', props.userId, props.month],
    queryFn: () => reviewBillingHistory(props.userId, props.month),
    enabled: open,
    refetchOnWindowFocus: false,
  })
  const imports = useQuery({
    queryKey: ['billing', 'history-imports', props.userId, props.month],
    queryFn: () => getBillingHistoryImports(props.userId, props.month),
    enabled: open,
  })
  const download = useMutation({
    mutationFn: (id: string) => downloadBillingHistorySource(props.userId, id),
    onError: handleServerError,
  })
  return (
    <>
      <Button
        variant='outline'
        className='self-start'
        onClick={() => setOpen(true)}
      >
        {t('Review historical import')}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className='max-h-[90dvh] overflow-y-auto sm:max-w-4xl'>
          <DialogHeader>
            <DialogTitle>{t('Historical accounting review')}</DialogTitle>
            <DialogDescription>
              {t(
                'Review does not import data. Only an explicit confirmation archives the source records and adds historical ledger entries, without changing wallet balances.'
              )}
            </DialogDescription>
          </DialogHeader>
          {review.isFetching && <p>{t('Loading...')}</p>}
          {review.isError && <p role='alert'>{t('Failed to load data')}</p>}
          <Button
            variant='outline'
            className='self-start'
            disabled={review.isFetching}
            onClick={() => {
              void review.refetch()
              void imports.refetch()
            }}
          >
            {t('Refresh review')}
          </Button>
          {review.data && !review.isError && (
            <HistoryReviewContent
              key={`${review.data.user_id}:${review.data.month}:${review.data.source_sha256}`}
              review={review.data}
              profileId={props.profileId}
              refreshing={review.isFetching}
              onClose={() => setOpen(false)}
              onViewStatement={props.onViewStatement}
            />
          )}
          {imports.isError && <p role='alert'>{t('Failed to load data')}</p>}
          {Boolean(imports.data?.length) && (
            <section className='flex flex-col gap-3 border-t pt-4'>
              <h3 className='font-medium'>{t('Historical import records')}</h3>
              {imports.data?.map((batch) => (
                <div key={batch.id} className='flex flex-col gap-2 text-sm'>
                  <p>
                    {batch.month} · {t('Records')}: {batch.records} ·{' '}
                    {billingTimestamp(batch.confirmed_at)}
                  </p>
                  <p>
                    {t('Operator')}:{' '}
                    {batch.confirmed_by_username || t('Unknown')} (#
                    {batch.confirmed_by}) · {batch.note}
                  </p>
                  <p className='font-mono text-xs break-all'>
                    SHA-256: {batch.source_sha256}
                  </p>
                  <Button
                    variant='outline'
                    className='self-start'
                    disabled={download.isPending}
                    onClick={() => download.mutate(batch.id)}
                  >
                    {t('Download import evidence')}
                  </Button>
                </div>
              ))}
            </section>
          )}
        </DialogContent>
      </Dialog>
    </>
  )
}

function HistoryReviewContent(props: {
  review: BillingHistoryReview
  profileId: number | null
  refreshing: boolean
  onClose: () => void
  onViewStatement: (id: string) => void
}) {
  const { t } = useTranslation()
  const fieldId = useId()
  const client = useQueryClient()
  const [requestId] = useState(() => crypto.randomUUID().replaceAll('-', ''))
  const [acknowledged, setAcknowledged] = useState(false)
  const [note, setNote] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const review = props.review
  const customerLabel = `${review.customer.username} (#${review.user_id})`
  const labels: Record<string, string> = {
    month_closed: t('Month closed plus 24 hours'),
    identity_complete: t('Company title and tax ID complete'),
    history_within_limit: t(
      'Historical records exist and fit the online import limit'
    ),
    history_empty_ledger: t(
      'No existing formal entries or summaries in this month'
    ),
    operations_settled: t('All reservations settled'),
    no_active_statement: t('No active statement for this month'),
    history_start_safe: t('Accounting coverage can be preserved safely'),
  }
  const apply = useMutation({
    mutationFn: () =>
      confirmBillingHistoryImport(review.user_id, {
        id: requestId,
        month: review.month,
        storage_profile_id: props.profileId ?? 0,
        source_sha256: review.source_sha256,
        acknowledged,
        note: note.trim(),
      }),
    onSuccess: () => {
      toast.success(
        t('Historical records imported. Wallet balances are unchanged.')
      )
      void client.invalidateQueries({ queryKey: ['billing'] })
      props.onClose()
    },
    onError: handleServerError,
  })
  const disabled =
    !review.ready ||
    !props.profileId ||
    !acknowledged ||
    !note.trim() ||
    props.refreshing ||
    apply.isPending
  return (
    <div className='flex flex-col gap-4'>
      <p className='text-sm break-words'>
        {t('Statement customer')}: {customerLabel} · {t('Month')}:{' '}
        {review.month}
      </p>
      <p className='text-sm'>
        {t('Records')}:{' '}
        {review.source_count > review.record_limit
          ? `>${review.record_limit}`
          : review.source_count}{' '}
        / {t('Online import limit')}: {review.record_limit}
      </p>
      <ul className='grid gap-2 text-sm sm:grid-cols-2'>
        {review.checks.map((check) => (
          <li key={check.code} className='flex items-start gap-2'>
            <Badge variant={check.passed ? 'secondary' : 'outline'}>
              {check.passed ? t('Passed') : t('Not ready')}
            </Badge>
            <span>{labels[check.code] ?? check.code}</span>
          </li>
        ))}
      </ul>
      {!review.ready && (
        <p role='alert' className='text-sm'>
          {t(
            'Import is blocked. Void active drafts explicitly before retrying. Mixed-ledger or oversized months require a separately reviewed migration; no records will be truncated or overwritten.'
          )}
        </p>
      )}
      {review.existing_statement && (
        <Button
          variant='outline'
          className='self-start'
          onClick={() => {
            props.onClose()
            props.onViewStatement(review.existing_statement)
          }}
        >
          {t('View existing statement')}
        </Button>
      )}
      {!props.profileId && (
        <p role='alert' className='text-sm'>
          {t('Select archive storage on the monthly preview before importing.')}
        </p>
      )}
      {review.new_start_at !== review.old_start_at && (
        <p role='alert' className='text-sm'>
          {t('Import will update the accounting start after confirmation')}:{' '}
          {billingTimestamp(review.old_start_at)} →{' '}
          {billingTimestamp(review.new_start_at)}
        </p>
      )}
      {review.snapshot && (
        <BillingRows
          rows={review.snapshot.days}
          total={review.snapshot.total}
          symbol={review.snapshot.currency.symbol}
          roundingDifference={review.snapshot.rounding_difference}
        />
      )}
      <p className='text-muted-foreground text-sm'>
        {t(
          'Historical entries keep the original recorded quota and log timestamps. This does not reconstruct historical wallet balances or certify that old logs are complete. Verify consumption, refunds, and administrator funding records before confirming.'
        )}
      </p>
      <p className='font-mono text-xs break-all'>
        SHA-256: {review.source_sha256 || '—'}
      </p>
      <Field>
        <FieldLabel htmlFor={`${fieldId}-note`}>
          {t('Verification basis')}
        </FieldLabel>
        <Textarea
          id={`${fieldId}-note`}
          value={note}
          onChange={(event) => setNote(event.target.value)}
          maxLength={2000}
          disabled={apply.isPending}
        />
      </Field>
      <Field orientation='horizontal'>
        <Checkbox
          id={`${fieldId}-ack`}
          checked={acknowledged}
          onCheckedChange={(value) => setAcknowledged(value === true)}
          disabled={apply.isPending}
        />
        <FieldLabel htmlFor={`${fieldId}-ack`}>
          {t(
            'I have verified the source records and authorize this historical import without charging the customer again.'
          )}
        </FieldLabel>
      </Field>
      <Button
        className='self-start'
        disabled={disabled}
        onClick={() => setConfirmOpen(true)}
      >
        {t('Confirm historical import')}
      </Button>
      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title={t('Confirm historical import')}
        desc={`${customerLabel} · ${review.month} · ${review.snapshot?.currency.code ?? ''} ${review.snapshot?.total.amount ?? ''}. ${t('This import is permanent and will not change the wallet balance. Continue only after verifying the source records.')}`}
        confirmText={t('Import verified records')}
        disabled={disabled}
        isLoading={apply.isPending}
        handleConfirm={() => apply.mutate()}
      />
    </div>
  )
}
