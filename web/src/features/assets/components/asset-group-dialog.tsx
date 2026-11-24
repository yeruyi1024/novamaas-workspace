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
import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import * as z from 'zod'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'
import { Textarea } from '@/components/ui/textarea'

export type AssetGroupFormValues = { name: string; description: string }

export function AssetGroupDialog(props: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (values: AssetGroupFormValues) => void
  pending: boolean
  /** Editing an existing group pre-fills its values. */
  initialValues?: AssetGroupFormValues
}) {
  const { t } = useTranslation()
  const schema = z.object({
    name: z.string().trim().min(1, t('Group name is required')).max(64),
    description: z.string().trim().max(300),
  })
  const form = useForm<AssetGroupFormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', description: '' },
  })
  const initialName = props.initialValues?.name ?? ''
  const initialDescription = props.initialValues?.description ?? ''
  useEffect(() => {
    form.reset({ name: initialName, description: initialDescription })
  }, [form, props.open, initialName, initialDescription])
  const editing = props.initialValues !== undefined

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {editing ? t('Rename group') : t('Create asset group')}
          </DialogTitle>
          <DialogDescription>
            {t('Use groups to organize assets by product, campaign, or team.')}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(props.onSubmit)}
            className='flex flex-col gap-4'
          >
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Name')}</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='description'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Description')}</FormLabel>
                  <FormControl>
                    <Textarea {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <Button type='submit' disabled={props.pending}>
                {props.pending && <Spinner data-icon='inline-start' />}
                {editing ? t('Save') : t('Create')}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
