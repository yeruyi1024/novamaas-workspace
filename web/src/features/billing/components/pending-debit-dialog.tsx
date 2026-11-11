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

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
  DialogHeader,
} from '@/components/ui/dialog'
import { handleServerError } from '@/lib/handle-server-error'

import { reviewedDebits, type ReviewedDebits } from '../api'

export function PendingDebitDialog({
  account,
  month,
}: {
  account: number
  month: string
}) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const [open, setOpen] = useState(false)
  const key = ['billing', 'reviewed-debits', account, month]
  const preview = useQuery({
    queryKey: key,
    queryFn: () => reviewedDebits(account, month),
    enabled: open,
    retry: false,
    refetchOnWindowFocus: false,
  })
  const apply = useMutation({
    mutationFn: (data: ReviewedDebits) => reviewedDebits(account, month, data),
    onSuccess: (result) => {
      client.setQueryData(key, result)
      void client.invalidateQueries({ queryKey: ['billing'] })
    },
    onError: (error) => {
      handleServerError(error)
      void preview.refetch()
    },
  })
  const data = preview.isError ? undefined : preview.data
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
        <DialogContent className='max-h-[85dvh] overflow-y-auto sm:max-w-xl'>
          <DialogHeader>
            <DialogTitle>{t('Collect outstanding charges')}</DialogTitle>
            <DialogDescription>
              {month}
              {data && ` · ${data.username} (#${data.user_id})`}
            </DialogDescription>
          </DialogHeader>
          {preview.isFetching && <p>{t('Loading...')}</p>}
          {preview.isError && (
            <p role='alert'>
              {t(
                'Wallet debit unavailable. Recheck the ledger, balance and historical import period.'
              )}
            </p>
          )}
          {data?.rows.map((row) => (
            <div key={row.request_id} className='border-t pt-2 text-sm'>
              <p>
                {row.supplier ? t('Supplier statement') : t('Usage log')} · CNY{' '}
                {row.amount}
              </p>
              <p className='break-all'>
                {t('Request ID')}: {row.request_id}
              </p>
              {row.entry_id && (
                <p>
                  {t('Wallet debit recorded (entry {{id}}).', {
                    id: row.entry_id,
                  })}
                </p>
              )}
            </div>
          ))}
          {data && data.rows.some((row) => row.status === 'ready') && (
            <>
              <p className='font-semibold'>
                {t('Total')}: CNY {data.total}
              </p>
              <Button
                disabled={preview.isFetching || apply.isPending}
                onClick={() => apply.mutate(data)}
              >
                {t('Confirm wallet debit')}
              </Button>
            </>
          )}
        </DialogContent>
      </Dialog>
    </>
  )
}
