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
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'

/** Destructive confirmation shared by single and batch asset deletion. */
export function AssetConfirmDialog(props: {
  open: boolean
  title: string
  description: string
  pending: boolean
  /** Optional confirmation checkbox text for high-impact batch deletes. */
  acknowledgeLabel?: string
  checkboxId?: string
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}) {
  const { t } = useTranslation()
  const [acknowledged, setAcknowledged] = useState(false)

  useEffect(() => {
    if (!props.open) setAcknowledged(false)
  }, [props.open, props.title, props.acknowledgeLabel])

  const requiresAck = Boolean(props.acknowledgeLabel)
  const checkboxId = props.checkboxId ?? 'asset-confirm-ack'

  return (
    <AlertDialog
      open={props.open}
      onOpenChange={(open) => {
        if (props.pending) return
        if (!open) setAcknowledged(false)
        props.onOpenChange(open)
      }}
    >
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{props.title}</AlertDialogTitle>
          <AlertDialogDescription>{props.description}</AlertDialogDescription>
        </AlertDialogHeader>
        {requiresAck && (
          <div className='border-destructive/30 bg-destructive/5 flex items-start gap-2.5 rounded-lg border p-3'>
            <Checkbox
              id={checkboxId}
              checked={acknowledged}
              onCheckedChange={(checked) => setAcknowledged(checked === true)}
            />
            <Label
              htmlFor={checkboxId}
              className='text-sm leading-snug font-normal'
            >
              {props.acknowledgeLabel}
            </Label>
          </div>
        )}
        <AlertDialogFooter>
          <AlertDialogCancel disabled={props.pending}>
            {t('Cancel')}
          </AlertDialogCancel>
          <AlertDialogAction
            variant='destructive'
            disabled={props.pending || (requiresAck && !acknowledged)}
            onClick={(event) => {
              event.preventDefault()
              if (props.pending) return
              props.onConfirm()
            }}
          >
            {props.pending ? t('Deleting...') : t('Delete')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
