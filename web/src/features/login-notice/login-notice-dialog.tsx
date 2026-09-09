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
import { useMutation, useQuery } from '@tanstack/react-query'
import { AlertTriangle, Bell, ShieldAlert } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { RichContent } from '@/components/rich-content'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Spinner } from '@/components/ui/spinner'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useAuthStore } from '@/stores/auth-store'

import { acknowledgeLoginNotice, getLoginNotice } from './api'
import { createDeviceFingerprint } from './device-fingerprint'

export function LoginNoticeDialog() {
  const { t } = useTranslation()
  const sessionID = useAuthStore((state) => state.auth.session?.sid)
  const [acknowledgedSessionID, setAcknowledgedSessionID] = useState<string>()
  const noticeQuery = useQuery({
    queryKey: ['login-notice', sessionID],
    queryFn: getLoginNotice,
    enabled: Boolean(sessionID),
    staleTime: Infinity,
    retry: 1,
  })
  const acknowledgement = useMutation({
    mutationFn: async () => {
      const fingerprint = await createDeviceFingerprint()
      await acknowledgeLoginNotice(fingerprint)
    },
    onSuccess: () => setAcknowledgedSessionID(sessionID),
    onError: () => toast.error(t('Failed to record acknowledgement')),
  })

  const notice = noticeQuery.data
  const open = Boolean(
    sessionID &&
    notice &&
    !notice.acknowledged &&
    acknowledgedSessionID !== sessionID
  )
  if (!notice) return null

  const periods = [
    { label: t('Today'), stats: notice.statistics.today },
    { label: t('Last 7 days'), stats: notice.statistics.seven_days },
    { label: t('Last 30 days'), stats: notice.statistics.thirty_days },
  ]

  return (
    <AlertDialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen && notice.requires_acknowledgement) return
        if (!nextOpen) setAcknowledgedSessionID(sessionID)
      }}
    >
      <AlertDialogContent className='max-h-[min(90svh,760px)] max-w-[calc(100%-2rem)] sm:max-w-2xl'>
        <AlertDialogHeader className='place-items-start text-left'>
          <AlertDialogTitle>{t('Login information')}</AlertDialogTitle>
          <AlertDialogDescription>
            {t('Review system announcements and your recent video activity.')}{' '}
            {t(
              'Confirming this notice records a privacy-preserving device fingerprint for audit purposes.'
            )}
          </AlertDialogDescription>
        </AlertDialogHeader>

        <ScrollArea className='max-h-[min(65svh,560px)] pr-3'>
          <div className='space-y-4'>
            <section aria-labelledby='login-notice-announcements'>
              <h3
                id='login-notice-announcements'
                className='mb-2 flex items-center gap-2 text-sm font-medium'
              >
                <Bell className='size-4' />
                {t('System announcements')}
              </h3>
              {notice.announcements.length > 0 ? (
                <div className='space-y-2'>
                  {notice.announcements.map((announcement, index) => (
                    <Card
                      key={announcement.id ?? `announcement-${index}`}
                      size='sm'
                    >
                      <CardHeader>
                        <CardTitle className='flex items-center justify-between gap-2'>
                          <span>{t('Announcement')}</span>
                          {announcement.publishDate ? (
                            <Badge variant='outline'>
                              {announcement.publishDate}
                            </Badge>
                          ) : null}
                        </CardTitle>
                      </CardHeader>
                      <CardContent className='space-y-2'>
                        <RichContent breaks content={announcement.content} />
                        {announcement.extra ? (
                          <RichContent
                            breaks
                            content={announcement.extra}
                            className='text-muted-foreground'
                          />
                        ) : null}
                      </CardContent>
                    </Card>
                  ))}
                </div>
              ) : (
                <Empty className='border py-5'>
                  <EmptyHeader>
                    <EmptyMedia variant='icon'>
                      <Bell />
                    </EmptyMedia>
                    <EmptyTitle>
                      {t('No announcements at this time')}
                    </EmptyTitle>
                    <EmptyDescription>
                      {t(
                        'There are no new system announcements for this login.'
                      )}
                    </EmptyDescription>
                  </EmptyHeader>
                </Empty>
              )}
            </section>

            <section aria-labelledby='login-notice-activity'>
              <h3
                id='login-notice-activity'
                className='mb-2 flex items-center gap-2 text-sm font-medium'
              >
                <ShieldAlert className='size-4' />
                {t('Video generation and violation statistics')}
              </h3>
              <div className='overflow-hidden rounded-lg border'>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('Period')}</TableHead>
                      <TableHead className='text-right'>
                        {t('Video generations')}
                      </TableHead>
                      <TableHead className='text-right'>
                        {t('Violations')}
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {periods.map((period) => (
                      <TableRow key={period.label}>
                        <TableCell>{period.label}</TableCell>
                        <TableCell className='text-right tabular-nums'>
                          {period.stats.generated}
                        </TableCell>
                        <TableCell className='text-right tabular-nums'>
                          <span
                            className={
                              period.stats.violations > 0
                                ? 'text-destructive font-medium'
                                : undefined
                            }
                          >
                            {period.stats.violations}
                          </span>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </section>

            {notice.requires_acknowledgement ? (
              <Alert variant='destructive'>
                <AlertTriangle />
                <AlertTitle>{t('Acknowledgement required')}</AlertTitle>
                <AlertDescription>
                  {t(
                    'A violation was recorded in the last 7 days. You must acknowledge this notice before closing it.'
                  )}
                </AlertDescription>
              </Alert>
            ) : null}
          </div>
        </ScrollArea>

        <AlertDialogFooter>
          <AlertDialogAction
            type='button'
            disabled={acknowledgement.isPending}
            onClick={() => acknowledgement.mutate()}
          >
            {acknowledgement.isPending ? (
              <Spinner data-icon='inline-start' />
            ) : null}
            {notice.requires_acknowledgement ? t('I acknowledge') : t('Got it')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
