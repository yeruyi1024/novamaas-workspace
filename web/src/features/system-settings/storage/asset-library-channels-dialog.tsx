import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { cn } from '@/lib/utils'

import type { AssetChannelConfig, AssetLibraryConnection } from './types'

export function AssetLibraryChannelsDialog(props: {
  library: AssetLibraryConnection | null
  channels: AssetChannelConfig[]
  saving: boolean
  error?: string
  onOpenChange: (open: boolean) => void
  onReload: () => void
  onSave: (channelIDs: number[]) => void
}) {
  const { t } = useTranslation()
  const [selected, setSelected] = useState<number[]>([])
  useEffect(() => {
    setSelected(props.library?.channels.map((channel) => channel.id) ?? [])
  }, [props.library])

  return (
    <Dialog open={props.library !== null} onOpenChange={props.onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>{t('Manage channels')}</DialogTitle>
          <DialogDescription>{props.library?.name} · {t('Select channels to link. Clear a selection to unlink without changing credentials or existing assets.')}</DialogDescription>
        </DialogHeader>
        <p className='text-muted-foreground text-xs'>{t('Changing a linked channel API key does not unlink it. Unlink the channel here before pointing it at a different upstream account.')}</p>
        {props.error && <Alert variant='destructive'><AlertDescription>{props.error}</AlertDescription></Alert>}
        <fieldset className='max-h-[52dvh] space-y-2 overflow-y-auto overscroll-contain rounded-lg border p-3'>
          <legend className='px-1 text-sm font-medium'>{t('Linked channels')}</legend>
          {props.channels.length === 0 && <p className='text-muted-foreground text-sm'>{t('Create a DoubaoVideo or Volc Native channel first.')}</p>}
          {props.channels.map((channel) => {
            const occupied = (channel.library_id ?? 0) > 0 && channel.library_id !== props.library?.id
            return (
              <div key={channel.channel_id} className={cn('flex items-start gap-2 rounded-md p-2 text-sm',occupied && 'opacity-50')}>
                <Checkbox
                  id={`asset-account-channel-${channel.channel_id}`}
                  disabled={occupied || props.saving}
                  checked={selected.includes(channel.channel_id)}
                  onCheckedChange={(checked) => setSelected((current) => checked ? [...current, channel.channel_id] : current.filter((id) => id !== channel.channel_id))}
                />
                <label htmlFor={`asset-account-channel-${channel.channel_id}`} className='min-w-0 break-words'>
                  {channel.channel_name} · ID {channel.channel_id}
                  {occupied && <span className='text-muted-foreground block text-xs'>{t('Currently linked to {{library}}', {library: channel.library_name})}</span>}
                </label>
              </div>
            )
          })}
        </fieldset>
        <DialogFooter>
          {props.error && <Button variant='outline' disabled={props.saving} onClick={props.onReload}>{t('Reload')}</Button>}
          <Button variant='outline' disabled={props.saving} onClick={() => props.onOpenChange(false)}>{t('Cancel')}</Button>
          <Button disabled={props.saving} onClick={() => props.onSave(selected)}>{t('Save')}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
