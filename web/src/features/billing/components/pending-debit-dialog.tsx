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
import { useQueryClient, useMutation, useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
  DialogHeader,
} from '@/components/ui/dialog'
import { handleServerError } from '@/lib/handle-server-error'

import { collectReviewedDebit, reviewDebits, type ReviewedDebit } from '../api'

type Props = { account: number; month: string }
export function PendingDebitDialog(props: Props) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  return (
    <>
      <Button
        className='self-start'
        variant='outline'
        onClick={() => setOpen(true)}
      >
        {t('Collect outstanding charges')}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        {open && <ReviewContent {...props} />}
      </Dialog>
    </>
  )
}
function ReviewContent({ account, month }: Props) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const preview = useQuery({
    queryKey: ['billing', 'reviewed-debits', account, month],
    queryFn: () => reviewDebits({ account, month }),
    retry: false,
    refetchOnWindowFocus: false,
  })
  const [receipts, setReceipts] = useState<Record<string, number>>({})
  const apply = useMutation({
    mutationFn: (row: ReviewedDebit) =>
      collectReviewedDebit({ account, month }, row),
    onSuccess: (row) => {
      const entryId = row.entry_id
      if (entryId) {
        setReceipts((prev) => ({ ...prev, [row.request_id]: entryId }))
      }
      toast.success(
        t('Wallet debit recorded (entry {{id}}).', { id: row.entry_id })
      )
      void client.invalidateQueries({ queryKey: ['billing'] })
    },
    onError: handleServerError,
  })
  return (
    <DialogContent className='max-h-[85dvh] overflow-y-auto sm:max-w-xl'>
      <DialogHeader>
        <DialogTitle>{t('Collect outstanding charges')}</DialogTitle>
        <DialogDescription>
          {month} ·{' '}
          {t('Confirm this request has not been collected separately.')}
        </DialogDescription>
      </DialogHeader>
      {preview.isPending && <p>{t('Loading...')}</p>}
      {preview.isError && (
        <p role='alert'>
          {t('Check the evidence file, account, month and billing currency.')}
        </p>
      )}
      {preview.data?.map((row) => {
        const entry = receipts[row.request_id] ?? row.entry_id
        return (
          <section
            key={row.request_id}
            className='space-y-2 border-t pt-3 text-sm'
          >
            <p>
              {row.username} (#{row.user_id}) · {month}
            </p>
            <p className='break-all'>
              {t('Request ID')}: {row.request_id}
            </p>
            <p>
              {row.source === 'usage_log'
                ? t('Usage log')
                : t('Supplier statement')}
            </p>
            {row.upstream_request_id && (
              <p className='break-all'>
                {t('Upstream Request ID')}: {row.upstream_request_id}
              </p>
            )}
            <p className='font-semibold'>
              {t('Wallet debit')}: CNY {row.amount}
            </p>
            {entry && (
              <p>{t('Wallet debit recorded (entry {{id}}).', { id: entry })}</p>
            )}
            {row.status === 'blocked' && (
              <p>
                {t(
                  'Wallet debit unavailable. Recheck the ledger, balance and historical import period.'
                )}
              </p>
            )}
            {!entry && row.status === 'ready' && (
              <Button
                disabled={apply.isPending}
                onClick={() => apply.mutate(row)}
              >
                {t('Confirm wallet debit')}
              </Button>
            )}
          </section>
        )
      })}
    </DialogContent>
  )
}
