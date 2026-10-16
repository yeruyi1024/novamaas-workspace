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
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Field, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { billingTimestamp } from '@/features/billing/api'
import { handleServerError } from '@/lib/handle-server-error'

import { downloadUpstreamAttachment, uploadUpstreamAttachment } from '../api'
import type { UpstreamStatementDetail } from '../types'

export function StatementAttachments(props: {
  detail: UpstreamStatementDetail
  canUpload: boolean
  onUploadingChange: (uploading: boolean) => void
  onReviewReset: () => void
}) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const [file, setFile] = useState<File | null>(null)
  const [error, setError] = useState('')
  const [inputVersion, setInputVersion] = useState(0)
  const detail = props.detail
  const upload = useMutation({
    mutationFn: (attachment: File) =>
      uploadUpstreamAttachment(detail.statement.id, attachment),
    onMutate: () => {
      props.onUploadingChange(true)
      props.onReviewReset()
    },
    onSuccess: async () => {
      setFile(null)
      setInputVersion((value) => value + 1)
      await client.invalidateQueries({
        queryKey: ['upstream-billing', 'statement', detail.statement.id],
      })
    },
    onError: handleServerError,
    onSettled: () => props.onUploadingChange(false),
  })
  const download = useMutation({
    mutationFn: (attachment: UpstreamStatementDetail['attachments'][number]) =>
      downloadUpstreamAttachment(
        detail.statement.id,
        attachment.id,
        attachment.original_filename
      ),
    onError: handleServerError,
  })
  const limitReached = detail.attachments.length + detail.pending_uploads >= 20
  return (
    <section
      className='flex flex-col gap-3'
      aria-label={t('Supplier statement attachments')}
    >
      <h3 className='font-medium'>{t('Supplier statement attachments')}</h3>
      <p className='text-muted-foreground text-sm'>
        {t(
          'Optional supplier evidence is included in the review digest. Confirmed statements and their attachments are read-only.'
        )}
      </p>
      {detail.attachments.length === 0 && (
        <p className='text-muted-foreground text-sm'>
          {t('No supplier attachments')}
        </p>
      )}
      {detail.attachments.map((attachment) => (
        <div
          key={attachment.id}
          className='flex flex-col gap-2 rounded-lg border p-3 text-sm'
        >
          <div className='flex flex-wrap items-center justify-between gap-2'>
            <span className='min-w-0 font-medium break-all'>
              {attachment.original_filename}
            </span>
            <Button
              variant='outline'
              size='sm'
              disabled={download.isPending}
              onClick={() => download.mutate(attachment)}
            >
              {t('Download attachment')}
            </Button>
          </div>
          <p className='text-muted-foreground'>
            {attachment.content_type} · {t('Size in bytes')}:{' '}
            {attachment.size.toLocaleString()}
          </p>
          <p className='text-muted-foreground'>
            {t('Uploaded by')}: {attachment.uploaded_username || t('Unknown')}{' '}
            (#{attachment.uploaded_by}) · {t('Uploaded at')}:{' '}
            {billingTimestamp(attachment.uploaded_at)}
          </p>
          <p className='text-muted-foreground font-mono text-xs break-all'>
            SHA-256: {attachment.sha256}
          </p>
        </div>
      ))}
      {detail.pending_uploads > 0 && (
        <p role='status' className='text-muted-foreground text-sm'>
          {t(
            'Attachment uploads are in progress. Wait for them to finish before confirming.'
          )}
        </p>
      )}
      {props.canUpload && (
        <>
          <Field>
            <FieldLabel htmlFor='upstream-supplier-file'>
              {t('Choose supplier attachment')}
            </FieldLabel>
            <Input
              key={inputVersion}
              id='upstream-supplier-file'
              type='file'
              accept='.pdf,.xlsx,.csv,.png,.jpg,.jpeg,application/pdf,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,text/csv,image/png,image/jpeg'
              disabled={upload.isPending || limitReached}
              onChange={(event) => {
                const selected = event.target.files?.[0] || null
                setFile(null)
                setError('')
                if (!selected) return
                if (selected.size > 20 * 1024 * 1024) {
                  setError(t('Each attachment must be 20 MiB or smaller.'))
                  return
                }
                if (!/\.(pdf|xlsx|csv|png|jpe?g)$/i.test(selected.name)) {
                  setError(
                    t('Choose a PDF, XLSX, CSV, PNG or JPEG attachment.')
                  )
                  return
                }
                setFile(selected)
              }}
            />
          </Field>
          <p className='text-muted-foreground text-sm'>
            {t(
              'Up to 20 attachments, 20 MiB each. PDF, XLSX, CSV, PNG and JPEG are supported.'
            )}
          </p>
          {limitReached && (
            <p role='alert'>
              {t('A statement can have at most 20 supplier attachments.')}
            </p>
          )}
          {error && <p role='alert'>{error}</p>}
          <Button
            className='self-start'
            variant='outline'
            disabled={!file || upload.isPending || limitReached}
            onClick={() => file && upload.mutate(file)}
          >
            {upload.isPending ? t('Uploading...') : t('Upload attachment')}
          </Button>
        </>
      )}
      <p className='text-muted-foreground font-mono text-xs break-all'>
        {t('Evidence SHA-256')}: {detail.evidence_sha256 || '—'}
      </p>
    </section>
  )
}
