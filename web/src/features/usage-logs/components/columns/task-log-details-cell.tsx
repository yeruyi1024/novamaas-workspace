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
import { Braces, Download, Info } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'

import { TASK_ACTIONS, TASK_STATUS } from '../../constants'
import {
  downloadTaskVideo,
  canGetTaskInformation,
  getTaskInformation,
  getTaskRequestBody,
  getTaskVideoContentInfo,
} from '../../task-content-api'
import type { TaskLog } from '../../types'
import { FailReasonDialog } from '../dialogs/fail-reason-dialog'
import { RequestBodyDialog } from '../dialogs/request-body-dialog'
import { TaskInformationDialog } from '../dialogs/task-information-dialog'

const VIDEO_ACTIONS = new Set<string>([
  TASK_ACTIONS.GENERATE,
  TASK_ACTIONS.TEXT_GENERATE,
  TASK_ACTIONS.FIRST_TAIL_GENERATE,
  TASK_ACTIONS.REFERENCE_GENERATE,
  TASK_ACTIONS.REMIX_GENERATE,
])

function formatJson(body: unknown): string {
  if (typeof body === 'string') {
    try {
      return JSON.stringify(JSON.parse(body), null, 2)
    } catch {
      return body
    }
  }
  return JSON.stringify(body, null, 2) ?? 'null'
}

function videoExtension(contentType: string): string {
  if (contentType.includes('webm')) return 'webm'
  if (contentType.includes('quicktime')) return 'mov'
  if (contentType.includes('matroska')) return 'mkv'
  return 'mp4'
}

export function TaskLogDetailsCell({
  isAdmin,
  log,
}: {
  isAdmin: boolean
  log: TaskLog
}) {
  const { t } = useTranslation()
  const [failReasonOpen, setFailReasonOpen] = useState(false)
  const [requestBodyOpen, setRequestBodyOpen] = useState(false)
  const [requestBodyJson, setRequestBodyJson] = useState<string | null>(null)
  const [requestBodyLoading, setRequestBodyLoading] = useState(false)
  const [requestBodyError, setRequestBodyError] = useState(false)
  const [videoDownloading, setVideoDownloading] = useState(false)
  const [taskInformationOpen, setTaskInformationOpen] = useState(false)
  const [taskInformationJson, setTaskInformationJson] = useState<string | null>(
    null
  )
  const [taskInformationLoading, setTaskInformationLoading] = useState(false)
  const [taskInformationError, setTaskInformationError] = useState(false)

  const canDownloadVideo =
    log.status === TASK_STATUS.SUCCESS &&
    VIDEO_ACTIONS.has(log.action) &&
    Boolean(log.result_url?.trim())

  const handleRequestBody = async () => {
    setRequestBodyOpen(true)
    if (requestBodyJson !== null || requestBodyLoading) return

    setRequestBodyLoading(true)
    setRequestBodyError(false)
    try {
      const response = await getTaskRequestBody(log.task_id)
      if (!response.success || response.data === undefined) {
        throw new Error(response.message || 'request body unavailable')
      }
      setRequestBodyJson(formatJson(response.data))
    } catch {
      setRequestBodyError(true)
    } finally {
      setRequestBodyLoading(false)
    }
  }

  const handleTaskInformation = async () => {
    setTaskInformationOpen(true)
    if (taskInformationLoading) return

    setTaskInformationLoading(true)
    setTaskInformationError(false)
    setTaskInformationJson(null)
    try {
      const response = await getTaskInformation(log.task_id, log.platform)
      setTaskInformationJson(formatJson(response))
    } catch {
      setTaskInformationError(true)
    } finally {
      setTaskInformationLoading(false)
    }
  }

  const handleVideoDownload = async () => {
    if (videoDownloading) return
    setVideoDownloading(true)
    try {
      const contentInfo = await getTaskVideoContentInfo(log.task_id)
      if (!contentInfo.success || !contentInfo.data) {
        throw new Error(contentInfo.message || 'video content unavailable')
      }
      if (contentInfo.data.delivery_mode === 'redirect') {
        if (!contentInfo.data.url) {
          throw new Error('video redirect URL unavailable')
        }
        const anchor = document.createElement('a')
        anchor.href = contentInfo.data.url
        anchor.download = `${log.task_id}.mp4`
        anchor.target = '_blank'
        anchor.rel = 'noopener'
        document.body.append(anchor)
        anchor.click()
        anchor.remove()
        return
      }

      const response = await downloadTaskVideo(log.task_id)
      const url = URL.createObjectURL(response.data)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = `${log.task_id}.${videoExtension(response.headers['content-type'])}`
      document.body.append(anchor)
      anchor.click()
      anchor.remove()
      URL.revokeObjectURL(url)
    } catch {
      toast.error(t('Failed to download video'))
    } finally {
      setVideoDownloading(false)
    }
  }

  const hasDetails =
    canDownloadVideo ||
    (isAdmin && log.request_body_available) ||
    canGetTaskInformation(log.platform) ||
    Boolean(log.fail_reason)
  if (!hasDetails) {
    return <span className='text-muted-foreground/60 text-xs'>-</span>
  }

  return (
    <div className='flex max-w-[220px] flex-col items-start gap-1'>
      {canDownloadVideo ? (
        <Button
          type='button'
          variant='ghost'
          size='xs'
          disabled={videoDownloading}
          onClick={handleVideoDownload}
        >
          {videoDownloading ? (
            <Spinner data-icon='inline-start' />
          ) : (
            <Download data-icon='inline-start' />
          )}
          {videoDownloading ? t('Downloading...') : t('Download video')}
        </Button>
      ) : null}
      {isAdmin && log.request_body_available ? (
        <Button
          type='button'
          variant='ghost'
          size='xs'
          onClick={handleRequestBody}
        >
          <Braces data-icon='inline-start' />
          {t('View request body')}
        </Button>
      ) : null}
      {canGetTaskInformation(log.platform) ? (
        <Button
          type='button'
          variant='ghost'
          size='xs'
          disabled={taskInformationLoading}
          onClick={handleTaskInformation}
        >
          {taskInformationLoading ? (
            <Spinner data-icon='inline-start' />
          ) : (
            <Info data-icon='inline-start' />
          )}
          {t('View information')}
        </Button>
      ) : null}
      {log.fail_reason ? (
        <button
          type='button'
          className='group flex max-w-full items-center text-left text-xs'
          onClick={() => setFailReasonOpen(true)}
          title={t('Click to view full error message')}
        >
          <span className='text-destructive truncate leading-snug group-hover:underline'>
            {log.fail_reason}
          </span>
        </button>
      ) : null}
      <FailReasonDialog
        failReason={log.fail_reason || ''}
        open={failReasonOpen}
        onOpenChange={setFailReasonOpen}
      />
      <RequestBodyDialog
        error={requestBodyError}
        json={requestBodyJson}
        loading={requestBodyLoading}
        open={requestBodyOpen}
        onOpenChange={setRequestBodyOpen}
      />
      <TaskInformationDialog
        error={taskInformationError}
        json={taskInformationJson}
        loading={taskInformationLoading}
        open={taskInformationOpen}
        onOpenChange={setTaskInformationOpen}
      />
    </div>
  )
}
