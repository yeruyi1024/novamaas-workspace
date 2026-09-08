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
import { toast } from 'sonner'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'

import {
  getCurrentRequestBodyArchiveTask,
  getRequestBodyArchiveTask,
  startRequestBodyArchiveTask,
} from '../api'
import { SettingsControlGroup } from '../components/settings-form-layout'
import type { RequestBodyArchiveTask } from '../types'

function isActive(task: RequestBodyArchiveTask | null) {
  return task?.status === 'pending' || task?.status === 'running'
}

export function RequestBodyArchiveControl() {
  const { t } = useTranslation()
  const [task, setTask] = useState<RequestBodyArchiveTask | null>(null)
  const [isStarting, setIsStarting] = useState(false)
  const [confirmOpen, setConfirmOpen] = useState(false)

  useEffect(() => {
    let cancelled = false
    getCurrentRequestBodyArchiveTask()
      .then((response) => {
        if (!cancelled && response.success && response.data) {
          setTask(response.data)
        }
      })
      .catch(() => undefined)
    return () => {
      cancelled = true
    }
  }, [])

  const taskId = task?.task_id
  const active = isActive(task)
  useEffect(() => {
    if (!taskId || !active) return

    let cancelled = false
    const interval = window.setInterval(async () => {
      try {
        const response = await getRequestBodyArchiveTask(taskId)
        if (cancelled || !response.success || !response.data) return
        setTask(response.data)
        if (response.data.status === 'succeeded') {
          toast.success(
            t('{{count}} request bodies archived and hot rows cleaned.', {
              count: response.data.result?.archived_count ?? 0,
            })
          )
        } else if (response.data.status === 'failed') {
          toast.error(
            response.data.error || t('Failed to archive request bodies')
          )
        }
      } catch {
        // Keep polling; the task lease and state are persisted by the backend.
      }
    }, 1000)
    return () => {
      cancelled = true
      window.clearInterval(interval)
    }
  }, [active, taskId, t])

  const handleStart = async () => {
    setConfirmOpen(false)
    setIsStarting(true)
    try {
      const response = await startRequestBodyArchiveTask()
      if (!response.success || !response.data) {
        throw new Error(
          response.message || t('Failed to archive request bodies')
        )
      }
      setTask(response.data)
      toast.success(t('Request body archive task started.'))
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : t('Failed to archive request bodies')
      toast.error(message)
    } finally {
      setIsStarting(false)
    }
  }

  const progress = Math.min(100, Math.max(0, task?.state?.progress ?? 0))
  const processed = task?.state?.processed ?? 0
  const total = task?.state?.total ?? 0

  return (
    <SettingsControlGroup className='space-y-3'>
      <div>
        <h4 className='text-sm font-medium'>{t('Archive request bodies')}</h4>
        <p className='text-muted-foreground text-sm'>
          {t(
            'Scan historical task and usage logs, preserve request bodies in the audit archive, and remove duplicate payloads from hot rows.'
          )}
        </p>
      </div>
      <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <AlertDialogTrigger
          render={
            <Button
              type='button'
              variant='outline'
              disabled={isStarting || active}
            />
          }
        >
          {isStarting || active
            ? t('Archiving...')
            : t('Scan and archive history')}
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t('Archive historical request bodies?')}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t(
                'The scan runs in batches and can be resumed safely. Request bodies remain available to administrators, while duplicate copies are removed from task and usage log rows.'
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('Cancel')}</AlertDialogCancel>
            <AlertDialogAction onClick={handleStart}>
              {t('Start archive')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      {task ? (
        <div className='rounded-md border p-3'>
          <div className='mb-2 flex items-center justify-between gap-3 text-sm'>
            <span className='font-medium'>
              {t('Request body archive progress')}
            </span>
            <span className='text-muted-foreground tabular-nums'>
              {progress}%
            </span>
          </div>
          <Progress value={progress} />
          <div className='text-muted-foreground mt-2 text-xs'>
            {t('{{processed}} of {{total}} rows scanned.', {
              processed,
              total,
            })}
          </div>
          {task.status === 'failed' && task.error ? (
            <div className='text-destructive mt-2 text-xs'>{task.error}</div>
          ) : null}
        </div>
      ) : null}
    </SettingsControlGroup>
  )
}
