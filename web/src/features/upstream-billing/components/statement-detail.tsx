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
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { billingTimestamp } from '@/features/billing/api'
import {
  statementActionKeys,
  statementDownloadKeys,
  statementStatusKeys,
} from '@/features/billing/constants'
import { handleServerError } from '@/lib/handle-server-error'
import { useAuthStore } from '@/stores/auth-store'

import {
  downloadUpstreamStatement,
  getUpstreamStatement,
  upstreamStatementAction,
} from '../api'
import { canManageUpstreamBilling } from '../permissions'
import type { UpstreamStatementDetail as Detail } from '../types'
import { ChannelUserCosts } from './channel-user-costs'
import { CostRows } from './cost-rows'
import { ReportSummary } from './report-summary'
import { StatementAttachments } from './statement-attachments'

const upstreamStatementActionKeys: Record<string, string> = {
  ...statementActionKeys,
  create: 'Create archived draft',
  attachment_upload: 'Upload attachment',
  attachment_ready: 'Attachment uploaded',
  attachment_failed: 'Attachment upload failed',
  attachment_expired: 'Attachment upload expired',
}

export function StatementDetail(props: { id: string; onClose: () => void }) {
  const { t } = useTranslation()
  const query = useQuery({
    queryKey: ['upstream-billing', 'statement', props.id],
    queryFn: () => getUpstreamStatement(props.id),
    refetchInterval: (state) =>
      state.state.data?.statement.status === 'preparing' ||
      (state.state.data?.pending_uploads ?? 0) > 0
        ? 3000
        : false,
  })
  return (
    <Dialog open onOpenChange={(open) => !open && props.onClose()}>
      <DialogContent className='max-h-[90dvh] overflow-y-auto sm:max-w-6xl'>
        <DialogHeader>
          <DialogTitle>{t('Upstream statement')}</DialogTitle>
          <DialogDescription>
            {t(
              'Review the frozen channels, costs and usage before internally confirming this archive.'
            )}
          </DialogDescription>
        </DialogHeader>
        {query.isPending && <p role='status'>{t('Loading...')}</p>}
        {query.isError && (
          <p role='alert'>
            {t('Failed to load data')}{' '}
            <Button variant='link' onClick={() => void query.refetch()}>
              {t('Retry')}
            </Button>
          </p>
        )}
        {query.data && !query.isError && (
          <StatementContent
            key={`${props.id}:${query.data.statement.status}`}
            detail={query.data}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}

function StatementContent(props: { detail: Detail }) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const session = useAuthStore((state) => state.auth.session)
  const canManage = useAuthStore((state) =>
    canManageUpstreamBilling(state.auth.user)
  )
  const [acknowledgedDigest, setAcknowledgedDigest] = useState<string | null>(
    null
  )
  const [uploading, setUploading] = useState(false)
  const [reason, setReason] = useState('')
  const [intent, setIntent] = useState<{
    action: 'confirm' | 'void'
    manifest: string
    evidence: string
  } | null>(null)
  const [part, setPart] = useState(1)
  const statement = props.detail.statement
  const snapshot = props.detail.snapshot
  const reviewDigest = `${statement.manifest_sha256}:${props.detail.evidence_sha256}`
  const acknowledged = acknowledgedDigest === reviewDigest
  const canConfirm =
    canManage && statement.status === 'draft' && Boolean(session?.sid)
  const canVoid = canManage && ['draft', 'failed'].includes(statement.status)
  const voidBlocked = props.detail.pending_uploads > 0 || uploading
  const reviewChanged =
    intent?.action === 'confirm' &&
    (intent.manifest !== statement.manifest_sha256 ||
      intent.evidence !== props.detail.evidence_sha256)
  const canAcknowledge =
    canConfirm &&
    Boolean(statement.manifest_sha256) &&
    /^[a-f0-9]{64}$/i.test(props.detail.evidence_sha256) &&
    props.detail.pending_uploads === 0 &&
    !uploading &&
    Boolean(snapshot)
  const mutate = useMutation({
    mutationFn: (input: Parameters<typeof upstreamStatementAction>[1]) =>
      upstreamStatementAction(statement.id, input),
    onSuccess: () => {
      setIntent(null)
      setAcknowledgedDigest(null)
      void client.invalidateQueries({ queryKey: ['upstream-billing'] })
    },
    onError: (error) => {
      handleServerError(error)
      setIntent(null)
      setAcknowledgedDigest(null)
      void client.invalidateQueries({
        queryKey: ['upstream-billing', 'statement', statement.id],
      })
    },
  })
  const download = useMutation({
    mutationFn: (file: { kind: string; ordinal?: number }) =>
      downloadUpstreamStatement(statement.id, file.kind, file.ordinal),
    onError: handleServerError,
  })
  const files = props.detail.artifacts.filter((file) =>
    ['pdf', 'xlsx', 'snapshot', 'manifest'].includes(file.kind)
  )
  const validPart =
    Number.isInteger(part) && part >= 1 && part <= props.detail.detail_count
  return (
    <div className='flex flex-col gap-5'>
      <div className='grid gap-2 rounded-lg border p-4 text-sm sm:grid-cols-2'>
        <p className='break-words'>
          {t('Upstream name')}:{' '}
          {snapshot?.group.name || `#${statement.group_id}`}
        </p>
        <p>
          {t('Statement month')}: {statement.month} · {t('Version')}{' '}
          {statement.revision}
        </p>
        <p>
          {t('Status')}:{' '}
          <Badge variant='outline'>
            {t(statementStatusKeys[statement.status])}
          </Badge>
        </p>
        <p>
          {t('Created at')}: {billingTimestamp(statement.created_at)}
        </p>
        <p>
          {t('Confirmed at')}: {billingTimestamp(statement.confirmed_at || 0)}
        </p>
        <p>{snapshot?.currency.code || '—'} · Asia/Shanghai</p>
      </div>
      {statement.status === 'preparing' && (
        <p role='status'>{t('Preparing')}</p>
      )}
      {statement.last_error && (
        <p role='alert'>
          {t('Archive preparation failed. Check server logs and retry.')}
        </p>
      )}
      {snapshot && (
        <>
          <ReportSummary report={snapshot} />
          <h3 className='font-medium'>{t('Daily upstream costs')}</h3>
          <CostRows
            mode='day'
            scope={snapshot}
            rows={snapshot.days}
            total={snapshot.total}
            symbol={snapshot.currency.symbol}
          />
          <h3 className='font-medium'>{t('Upstream model usage')}</h3>
          <CostRows
            mode='model'
            scope={snapshot}
            rows={snapshot.models}
            symbol={snapshot.currency.symbol}
          />
          <h3 className='font-medium'>
            {t('User cost allocation by channel')}
          </h3>
          <ChannelUserCosts
            rows={snapshot.channel_users}
            schemaVersion={snapshot.schema_version}
            symbol={snapshot.currency.symbol}
            frozen
          />
        </>
      )}
      <div className='flex flex-wrap gap-2'>
        {files.map((file) => (
          <Button
            key={`${file.kind}:${file.ordinal}`}
            variant='outline'
            disabled={download.isPending}
            onClick={() =>
              download.mutate({ kind: file.kind, ordinal: file.ordinal })
            }
          >
            {file.kind === 'snapshot'
              ? t('Download snapshot')
              : t(statementDownloadKeys[file.kind] || 'Download')}
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
      </div>
      {props.detail.detail_count > 0 && (
        <Field>
          <FieldLabel htmlFor='upstream-detail-part'>
            {t('Archived detail files')} ({props.detail.detail_count})
          </FieldLabel>
          <div className='flex gap-2'>
            <Input
              id='upstream-detail-part'
              type='number'
              min={1}
              max={props.detail.detail_count}
              value={part}
              onChange={(event) => setPart(Number(event.target.value))}
            />
            <Button
              variant='outline'
              disabled={!validPart || download.isPending}
              onClick={() =>
                validPart &&
                download.mutate({
                  kind: 'details',
                  ordinal: part - 1,
                })
              }
            >
              {t('Download')}
            </Button>
          </div>
        </Field>
      )}
      <StatementAttachments
        detail={props.detail}
        canUpload={
          canManage && statement.status === 'draft' && !mutate.isPending
        }
        onUploadingChange={setUploading}
        onReviewReset={() => setAcknowledgedDigest(null)}
      />
      {statement.manifest_sha256 && (
        <p className='text-muted-foreground font-mono text-xs break-all'>
          SHA-256: {statement.manifest_sha256}
        </p>
      )}
      {canManage && statement.status === 'draft' && !session?.sid && (
        <p role='alert'>
          {t('Sign in with a browser session to confirm this statement.')}
        </p>
      )}
      {canManage && statement.status === 'draft' && (
        <Field orientation='horizontal' data-disabled={!canAcknowledge}>
          <Checkbox
            id='upstream-acknowledged'
            checked={acknowledged}
            disabled={!canAcknowledge || mutate.isPending}
            onCheckedChange={(checked) =>
              setAcknowledgedDigest(checked === true ? reviewDigest : null)
            }
          />
          <FieldLabel htmlFor='upstream-acknowledged'>
            {t('I have checked this statement and agree to confirm it.')}
          </FieldLabel>
        </Field>
      )}
      {canVoid && (
        <Field>
          <FieldLabel htmlFor='upstream-statement-reason'>
            {t('Review note')}
          </FieldLabel>
          <Textarea
            id='upstream-statement-reason'
            value={reason}
            maxLength={2000}
            placeholder={t('Optional when confirming; required when voiding.')}
            disabled={mutate.isPending}
            onChange={(event) => setReason(event.target.value)}
          />
        </Field>
      )}
      <div className='flex flex-wrap gap-2'>
        {canManage && statement.status === 'draft' && (
          <Button
            disabled={!canAcknowledge || !acknowledged || mutate.isPending}
            onClick={() =>
              setIntent({
                action: 'confirm',
                manifest: statement.manifest_sha256,
                evidence: props.detail.evidence_sha256,
              })
            }
          >
            {t('Confirm statement')}
          </Button>
        )}
        {statement.status === 'failed' && (
          <Button
            variant='outline'
            disabled={mutate.isPending}
            onClick={() =>
              mutate.mutate({
                action: 'retry',
                manifest_sha256: statement.manifest_sha256,
                reason: '',
                acknowledged: false,
              })
            }
          >
            {t('Retry')}
          </Button>
        )}
        {canVoid && (
          <Button
            variant='destructive'
            disabled={!reason.trim() || mutate.isPending || voidBlocked}
            onClick={() =>
              setIntent({
                action: 'void',
                manifest: statement.manifest_sha256,
                evidence: props.detail.evidence_sha256,
              })
            }
          >
            {t('Void statement')}
          </Button>
        )}
      </div>
      <section
        className='flex flex-col gap-2'
        aria-label={t('Statement history')}
      >
        <h3 className='font-medium'>{t('Statement history')}</h3>
        {props.detail.events.map((event) => (
          <p
            key={event.id}
            className='text-muted-foreground text-sm break-words'
          >
            {billingTimestamp(event.created_at)} · {t('Operator')}:{' '}
            {event.actor_username || t('Unknown')} (#{event.actor_id}) ·{' '}
            {t(upstreamStatementActionKeys[event.action] || 'Unknown')}{' '}
            {event.note}
          </p>
        ))}
        {props.detail.events.length === 0 && (
          <p className='text-muted-foreground text-sm'>{t('No records')}</p>
        )}
      </section>
      <ConfirmDialog
        open={Boolean(intent)}
        onOpenChange={(open) => !open && setIntent(null)}
        title={t('Confirm action')}
        desc={
          intent?.action === 'void'
            ? t(
                'Void this unconfirmed upstream statement? The reason will be kept in the audit history.'
              )
            : t(
                'Confirm and archive this frozen upstream statement? Confirmed statements cannot be changed.'
              )
        }
        confirmText={t('Confirm')}
        destructive={intent?.action === 'void'}
        isLoading={mutate.isPending}
        disabled={
          intent?.action === 'confirm'
            ? !canAcknowledge || !acknowledged || reviewChanged
            : !canVoid || !reason.trim() || voidBlocked
        }
        handleConfirm={() => {
          if (
            !intent ||
            (intent.action === 'void' && (!canVoid || voidBlocked)) ||
            (intent.action === 'confirm' &&
              (!canAcknowledge || !acknowledged || reviewChanged))
          ) {
            return
          }
          mutate.mutate({
            action: intent.action,
            manifest_sha256: intent.manifest,
            ...(intent.action === 'confirm'
              ? { evidence_sha256: intent.evidence }
              : {}),
            reason: reason.trim(),
            acknowledged: intent.action === 'confirm' && acknowledged,
          })
        }}
      >
        {intent?.action === 'confirm' &&
          snapshot &&
          !snapshot.coverage.complete && (
            <p className='text-sm text-amber-900 dark:text-amber-200'>
              {t(
                'Reconciliation warnings remain in this archive after confirmation.'
              )}
            </p>
          )}
        {reviewChanged && (
          <p role='alert' className='text-destructive text-sm'>
            {t(
              'Supplier evidence changed. Close this dialog and review the updated attachments before confirming again.'
            )}
          </p>
        )}
      </ConfirmDialog>
    </div>
  )
}
