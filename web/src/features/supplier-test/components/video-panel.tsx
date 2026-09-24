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
import {
  Download,
  ExternalLink,
  Network,
  Play,
  RefreshCw,
  Send,
  Upload,
  Video,
  X,
  Zap,
} from 'lucide-react'
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type ChangeEvent,
  type Dispatch,
  type SetStateAction,
} from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { TabsContent } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'

import { querySupplierVideoTask, type QueryVideoTaskResult } from '../api'
import {
  DEFAULT_VIDEO_PROMPT,
  ENDPOINT_PATH_PRESETS,
  VIDEO_RATIOS,
  VIDEO_RESOLUTIONS,
  VIDEO_ROLES,
} from '../constants'
import type { CheckResult, VideoForm, VideoMetrics } from '../types'
import {
  buildVideoRequestPayload,
  parseVideoPayloadToForm,
} from '../video-json'
import { CheckTable } from './check-table'
import { RawJsonDialog } from './raw-json-dialog'

export function VideoPanel(props: {
  video: VideoForm
  busy: boolean
  model?: string
  baseUrl?: string
  apiKey?: string
  videoChecks: CheckResult[]
  videoMetrics: VideoMetrics | null
  onVideoChange: Dispatch<SetStateAction<VideoForm>>
  onExportPdf?: () => void
  onSendRawJson?: (rawJson: string) => void
  onModelChange?: (model: string) => void
  onRunCheck?: (checkId?: string) => void
  onManualQueryResult?: (res: QueryVideoTaskResult) => void
}) {
  const { t } = useTranslation()
  const firstFileInputRef = useRef<HTMLInputElement>(null)
  const lastFileInputRef = useRef<HTMLInputElement>(null)

  const [activeJsonTab, setActiveJsonTab] = useState<
    'request' | 'poll' | 'submit'
  >('request')
  const [manualPollJson, setManualPollJson] = useState<string>('')
  const [isManualQuerying, setIsManualQuerying] = useState(false)

  const busy = props.busy
  useEffect(() => {
    if (busy) {
      setManualPollJson('')
    }
  }, [busy])

  // Real-time request JSON preview based on current form
  const previewPayload = buildVideoRequestPayload(
    props.model || '',
    props.video
  )
  const previewRequestJson = previewPayload
    ? JSON.stringify(previewPayload, null, 2)
    : ''
  const onVideoChange = props.onVideoChange
  const handleEffectiveJsonChange = useCallback(
    (next: string) => {
      onVideoChange((current) =>
        (current.rawPayload ?? '') === next
          ? current
          : { ...current, rawPayload: next }
      )
    },
    [onVideoChange]
  )

  // Effective latest Task ID
  const effectiveTaskId =
    props.video.taskId?.trim() || props.videoMetrics?.task_id?.trim() || ''

  // Auto-sync Task ID from video metrics when new task_id is returned
  const lastMetricsTaskIdRef = useRef<string>('')
  const videoMetricsTaskId = props.videoMetrics?.task_id
  useEffect(() => {
    const newTaskId = videoMetricsTaskId?.trim()
    if (newTaskId && newTaskId !== lastMetricsTaskIdRef.current) {
      lastMetricsTaskIdRef.current = newTaskId
      onVideoChange((cur) => ({
        ...cur,
        taskId: newTaskId,
      }))
    }
  }, [videoMetricsTaskId, onVideoChange])

  const handleApplyJsonToForm = (rawJson: string) => {
    const res = parseVideoPayloadToForm(rawJson)
    if (!res.success) return
    if (res.model && props.onModelChange) {
      props.onModelChange(res.model)
    }
    if (res.videoPatch) {
      props.onVideoChange((cur) => ({
        ...cur,
        ...res.videoPatch,
      }))
    }
  }

  const handleFirstFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    if (!file.type.startsWith('image/')) {
      toast.error(t('Please select an image file'))
      return
    }
    const reader = new FileReader()
    reader.addEventListener('load', (event) => {
      const result = event.target?.result
      if (typeof result === 'string') {
        props.onVideoChange((current) => ({
          ...current,
          base64Data: result,
        }))
        toast.success(
          t('Image loaded as Base64 ({{size}} KB)', {
            size: Math.round(file.size / 1024),
          })
        )
      }
    })
    reader.addEventListener('error', () => {
      toast.error(t('Failed to read image file'))
    })
    reader.readAsDataURL(file)
  }

  const handleLastFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    if (!file.type.startsWith('image/')) {
      toast.error(t('Please select an image file'))
      return
    }
    const reader = new FileReader()
    reader.addEventListener('load', (event) => {
      const result = event.target?.result
      if (typeof result === 'string') {
        props.onVideoChange((current) => ({
          ...current,
          lastFrameBase64: result,
        }))
        toast.success(
          t('End frame image loaded ({{size}} KB)', {
            size: Math.round(file.size / 1024),
          })
        )
      }
    })
    reader.addEventListener('error', () => {
      toast.error(t('Failed to read image file'))
    })
    reader.readAsDataURL(file)
  }

  // Active Manual Query
  const handleManualQuery = async () => {
    if (!effectiveTaskId) {
      toast.error(t('Please submit task first or provide a Task ID'))
      return
    }
    if (!props.baseUrl?.trim()) {
      toast.error(t('Enter a base URL first'))
      return
    }
    setIsManualQuerying(true)
    try {
      const res = await querySupplierVideoTask({
        base_url: props.baseUrl.trim(),
        api_key: props.apiKey || '',
        task_id: effectiveTaskId,
        custom_path: props.video.customPath,
      })
      if (res.raw_response) {
        try {
          setManualPollJson(
            JSON.stringify(JSON.parse(res.raw_response), null, 2)
          )
        } catch {
          setManualPollJson(res.raw_response)
        }
      } else {
        setManualPollJson(JSON.stringify(res, null, 2))
      }
      setActiveJsonTab('poll')
      if (!res.success) {
        toast.error(res.message || t('Failed to query task'))
      } else if (res.status) {
        props.onManualQueryResult?.(res)
        toast.success(t('Task status: {{status}}', { status: res.status }))
      } else if (res.message) {
        toast.error(res.message)
      }
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : t('Failed to query task')
      )
    } finally {
      setIsManualQuerying(false)
    }
  }

  const handleRunSingleCheck = (checkId: string) => {
    if (
      (checkId === 'video_poll' || checkId === 'video_result') &&
      !effectiveTaskId
    ) {
      toast.error(t('Please submit task first or provide a Task ID'))
      return
    }
    props.onRunCheck?.(checkId)
  }

  return (
    <TabsContent value='video' className='mt-0 space-y-3'>
      {/* 2-Column Responsive Layout */}
      <div className='grid grid-cols-1 items-start gap-4 lg:grid-cols-12'>
        {/* Left Column: Form & Actions */}
        <div className='space-y-3 lg:col-span-7'>
          {/* Prompt Input */}
          <div className='space-y-1.5'>
            <div className='flex items-center justify-between'>
              <Label htmlFor='video-prompt' className='text-xs font-medium'>
                {t('Generation Prompt')}
              </Label>
              <Button
                type='button'
                variant='ghost'
                size='sm'
                className='text-muted-foreground h-5 px-1.5 text-xs'
                onClick={() =>
                  props.onVideoChange((c) => ({
                    ...c,
                    prompt: DEFAULT_VIDEO_PROMPT,
                  }))
                }
              >
                {t('Reset Prompt')}
              </Button>
            </div>
            <Textarea
              id='video-prompt'
              rows={2}
              value={props.video.prompt}
              disabled={props.busy}
              placeholder={t('Describe video scene, lighting, style...')}
              onChange={(e) =>
                props.onVideoChange((c) => ({
                  ...c,
                  prompt: e.target.value,
                }))
              }
              className='text-xs'
            />
          </div>

          {/* Endpoint Path & Route Compatibility */}
          <Card className='p-2.5'>
            <div className='mb-2 flex flex-wrap items-center justify-between gap-2'>
              <div className='flex items-center gap-1.5 text-xs font-semibold'>
                <Network className='text-primary size-3.5' />
                <span>{t('Endpoint Path & Route Compatibility')}</span>
              </div>
              {props.video.customPath.trim() ? (
                <Badge variant='secondary' className='font-mono text-[10px]'>
                  {props.video.customPath}
                </Badge>
              ) : (
                <Badge variant='outline' className='text-[10px]'>
                  {t('Auto Detect / Standard')}
                </Badge>
              )}
            </div>
            <div className='grid gap-2 sm:grid-cols-2'>
              <div className='space-y-1'>
                <Label className='text-muted-foreground text-[11px]'>
                  {t('Quick Path Preset')}
                </Label>
                <Select
                  value={props.video.customPath || '__auto__'}
                  disabled={props.busy}
                  onValueChange={(val) => {
                    props.onVideoChange((c) => ({
                      ...c,
                      customPath: val === '__auto__' ? '' : val || '',
                    }))
                  }}
                >
                  <SelectTrigger className='h-7 w-full text-xs'>
                    <SelectValue placeholder={t('Select path preset')} />
                  </SelectTrigger>
                  <SelectContent>
                    {ENDPOINT_PATH_PRESETS.map((preset) => (
                      <SelectItem
                        key={preset.value || '__auto__'}
                        value={preset.value || '__auto__'}
                        className='text-xs'
                      >
                        {preset.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className='space-y-1'>
                <Label className='text-muted-foreground text-[11px]'>
                  {t('Custom Path String')}
                </Label>
                <Input
                  placeholder='/api/v3/contents/generations/tasks'
                  value={props.video.customPath}
                  disabled={props.busy}
                  onChange={(e) =>
                    props.onVideoChange((c) => ({
                      ...c,
                      customPath: e.target.value,
                    }))
                  }
                  className='h-7 font-mono text-xs'
                />
              </div>
            </div>
            <p className='text-muted-foreground mt-1.5 text-[10px] leading-tight'>
              {t(
                'Supports custom vendor path isolation. You can specify a path override here, or directly fill in a full URL (with /contents/generations/tasks) in Base URL above. Both are automatically normalized and supported.'
              )}
            </p>
          </Card>

          {/* Compact First Frame / Reference Image Card */}
          <Card className='p-2.5'>
            <div className='flex flex-wrap items-center justify-between gap-2'>
              <div className='flex items-center space-x-2'>
                <Checkbox
                  id='toggle-has-image'
                  checked={props.video.hasImage}
                  disabled={props.busy}
                  onCheckedChange={(checked) =>
                    props.onVideoChange((cur) => ({
                      ...cur,
                      hasImage: Boolean(checked),
                    }))
                  }
                />
                <Label
                  htmlFor='toggle-has-image'
                  className='cursor-pointer text-xs font-semibold'
                >
                  {t('First Frame / Reference Image')}
                </Label>
              </div>

              {props.video.hasImage && (
                <div className='flex items-center gap-2'>
                  <Select
                    value={props.video.uploadMode}
                    disabled={props.busy}
                    onValueChange={(val) => {
                      if (val === 'url' || val === 'base64') {
                        props.onVideoChange((c) => ({ ...c, uploadMode: val }))
                      }
                    }}
                  >
                    <SelectTrigger className='h-6 w-24 text-[11px]'>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value='url'>{t('URL Mode')}</SelectItem>
                      <SelectItem value='base64'>{t('Base64 Mode')}</SelectItem>
                    </SelectContent>
                  </Select>

                  <Select
                    value={props.video.role}
                    disabled={props.busy}
                    onValueChange={(val) => {
                      if (val) {
                        props.onVideoChange((c) => ({ ...c, role: val }))
                      }
                    }}
                  >
                    <SelectTrigger className='h-6 w-28 text-[11px]'>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {VIDEO_ROLES.map((role) => (
                        <SelectItem key={role.value} value={role.value}>
                          {role.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              )}
            </div>

            {props.video.hasImage && (
              <div className='mt-2 space-y-1.5 border-t pt-2'>
                {props.video.uploadMode === 'url' ? (
                  <Input
                    placeholder='https://.../sample.jpg'
                    value={props.video.imageUrl}
                    disabled={props.busy}
                    onChange={(e) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        imageUrl: e.target.value,
                      }))
                    }
                    className='h-7 text-xs'
                  />
                ) : (
                  <div className='flex items-center gap-2'>
                    <input
                      type='file'
                      ref={firstFileInputRef}
                      accept='image/*'
                      className='hidden'
                      onChange={handleFirstFileChange}
                    />
                    <Button
                      type='button'
                      variant='outline'
                      size='sm'
                      className='h-7 gap-1 text-xs'
                      onClick={() => firstFileInputRef.current?.click()}
                    >
                      <Upload className='size-3' />
                      {t('Select Local Image')}
                    </Button>
                    <span className='text-muted-foreground truncate text-[11px]'>
                      {props.video.base64Data
                        ? t('Image loaded')
                        : t('No image selected')}
                    </span>
                  </div>
                )}

                {/* Inline Compact Preview */}
                {(props.video.uploadMode === 'url'
                  ? props.video.imageUrl
                  : props.video.base64Data) && (
                  <div className='bg-muted/30 flex items-center gap-2 rounded border p-1'>
                    <img
                      src={
                        props.video.uploadMode === 'url'
                          ? props.video.imageUrl
                          : props.video.base64Data
                      }
                      alt='Preview'
                      className='size-7 rounded object-cover'
                      onError={(e) => {
                        ;(e.target as HTMLElement).style.display = 'none'
                      }}
                    />
                    <span className='text-muted-foreground flex-1 truncate font-mono text-[11px]'>
                      {props.video.uploadMode === 'url'
                        ? props.video.imageUrl
                        : 'data:image/...;base64'}
                    </span>
                    <Button
                      type='button'
                      variant='ghost'
                      size='sm'
                      className='text-muted-foreground hover:text-destructive h-5 w-5 p-0'
                      onClick={() =>
                        props.onVideoChange((c) => ({
                          ...c,
                          imageUrl: '',
                          base64Data: '',
                        }))
                      }
                    >
                      <X className='size-3' />
                    </Button>
                  </div>
                )}
              </div>
            )}
          </Card>

          {/* Compact End Frame Card */}
          <Card className='p-2.5'>
            <div className='flex flex-wrap items-center justify-between gap-2'>
              <div className='flex items-center space-x-2'>
                <Checkbox
                  id='toggle-has-last-frame'
                  checked={props.video.hasLastFrame}
                  disabled={props.busy}
                  onCheckedChange={(checked) =>
                    props.onVideoChange((cur) => ({
                      ...cur,
                      hasLastFrame: Boolean(checked),
                    }))
                  }
                />
                <Label
                  htmlFor='toggle-has-last-frame'
                  className='cursor-pointer text-xs font-semibold'
                >
                  {t('End Frame (Dual-frame transition)')}
                </Label>
              </div>

              {props.video.hasLastFrame && (
                <Select
                  value={props.video.lastFrameMode}
                  disabled={props.busy}
                  onValueChange={(val) => {
                    if (val === 'url' || val === 'base64') {
                      props.onVideoChange((c) => ({ ...c, lastFrameMode: val }))
                    }
                  }}
                >
                  <SelectTrigger className='h-6 w-24 text-[11px]'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value='url'>{t('URL Mode')}</SelectItem>
                    <SelectItem value='base64'>{t('Base64 Mode')}</SelectItem>
                  </SelectContent>
                </Select>
              )}
            </div>

            {props.video.hasLastFrame && (
              <div className='mt-2 space-y-1.5 border-t pt-2'>
                {props.video.lastFrameMode === 'url' ? (
                  <Input
                    placeholder='https://.../end-frame.jpg'
                    value={props.video.lastFrameUrl}
                    disabled={props.busy}
                    onChange={(e) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        lastFrameUrl: e.target.value,
                      }))
                    }
                    className='h-7 text-xs'
                  />
                ) : (
                  <div className='flex items-center gap-2'>
                    <input
                      type='file'
                      ref={lastFileInputRef}
                      accept='image/*'
                      className='hidden'
                      onChange={handleLastFileChange}
                    />
                    <Button
                      type='button'
                      variant='outline'
                      size='sm'
                      className='h-7 gap-1 text-xs'
                      onClick={() => lastFileInputRef.current?.click()}
                    >
                      <Upload className='size-3' />
                      {t('Select End Frame')}
                    </Button>
                    <span className='text-muted-foreground truncate text-[11px]'>
                      {props.video.lastFrameBase64
                        ? t('Image loaded')
                        : t('No image selected')}
                    </span>
                  </div>
                )}
              </div>
            )}
          </Card>

          {/* Compact Advanced Parameters Grid */}
          <Card className='p-2.5'>
            <div className='mb-2 flex items-center justify-between'>
              <span className='text-xs font-semibold'>
                {t('Advanced Parameters (Sent only when checked)')}
              </span>
              <span className='text-muted-foreground text-[10px]'>
                {t('Unchecked fields use upstream model defaults')}
              </span>
            </div>

            <div className='grid grid-cols-2 gap-2 sm:grid-cols-4'>
              {/* Resolution */}
              <div className='bg-muted/10 space-y-1 rounded border p-1.5'>
                <div className='flex items-center space-x-1.5'>
                  <Checkbox
                    id='opt-resolution'
                    checked={props.video.hasResolution}
                    disabled={props.busy}
                    onCheckedChange={(checked) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        hasResolution: Boolean(checked),
                      }))
                    }
                  />
                  <Label
                    htmlFor='opt-resolution'
                    className='cursor-pointer text-[11px] font-medium'
                  >
                    {t('Resolution')}
                  </Label>
                </div>
                {props.video.hasResolution && (
                  <Select
                    value={props.video.resolution}
                    disabled={props.busy}
                    onValueChange={(val) => {
                      if (val) {
                        props.onVideoChange((c) => ({ ...c, resolution: val }))
                      }
                    }}
                  >
                    <SelectTrigger className='h-6 text-[11px]'>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {VIDEO_RESOLUTIONS.map((res) => (
                        <SelectItem key={res} value={res}>
                          {res}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </div>

              {/* Ratio */}
              <div className='bg-muted/10 space-y-1 rounded border p-1.5'>
                <div className='flex items-center space-x-1.5'>
                  <Checkbox
                    id='opt-ratio'
                    checked={props.video.hasRatio}
                    disabled={props.busy}
                    onCheckedChange={(checked) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        hasRatio: Boolean(checked),
                      }))
                    }
                  />
                  <Label
                    htmlFor='opt-ratio'
                    className='cursor-pointer text-[11px] font-medium'
                  >
                    {t('Ratio')}
                  </Label>
                </div>
                {props.video.hasRatio && (
                  <Select
                    value={props.video.ratio}
                    disabled={props.busy}
                    onValueChange={(val) => {
                      if (val) {
                        props.onVideoChange((c) => ({ ...c, ratio: val }))
                      }
                    }}
                  >
                    <SelectTrigger className='h-6 text-[11px]'>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {VIDEO_RATIOS.map((ratio) => (
                        <SelectItem key={ratio} value={ratio}>
                          {ratio}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </div>

              {/* Duration */}
              <div className='bg-muted/10 space-y-1 rounded border p-1.5'>
                <div className='flex items-center space-x-1.5'>
                  <Checkbox
                    id='opt-duration'
                    checked={props.video.hasDuration}
                    disabled={props.busy}
                    onCheckedChange={(checked) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        hasDuration: Boolean(checked),
                      }))
                    }
                  />
                  <Label
                    htmlFor='opt-duration'
                    className='cursor-pointer text-[11px] font-medium'
                  >
                    {t('Duration (s)')}
                  </Label>
                </div>
                {props.video.hasDuration && (
                  <Input
                    type='number'
                    min={1}
                    max={60}
                    value={props.video.duration}
                    disabled={props.busy}
                    onChange={(e) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        duration: Number.parseInt(e.target.value, 10) || 5,
                      }))
                    }
                    className='h-6 text-[11px]'
                  />
                )}
              </div>

              {/* Watermark */}
              <div className='bg-muted/10 space-y-1 rounded border p-1.5'>
                <div className='flex items-center space-x-1.5'>
                  <Checkbox
                    id='opt-watermark'
                    checked={props.video.hasWatermark}
                    disabled={props.busy}
                    onCheckedChange={(checked) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        hasWatermark: Boolean(checked),
                      }))
                    }
                  />
                  <Label
                    htmlFor='opt-watermark'
                    className='cursor-pointer text-[11px] font-medium'
                  >
                    {t('Watermark')}
                  </Label>
                </div>
                {props.video.hasWatermark && (
                  <div className='flex items-center justify-between pt-0.5'>
                    <span className='text-muted-foreground text-[10px]'>
                      {props.video.watermark ? t('Enabled') : t('Disabled')}
                    </span>
                    <Switch
                      checked={props.video.watermark}
                      disabled={props.busy}
                      onCheckedChange={(checked) =>
                        props.onVideoChange((c) => ({
                          ...c,
                          watermark: Boolean(checked),
                        }))
                      }
                      className='scale-75'
                    />
                  </div>
                )}
              </div>

              {/* Seed */}
              <div className='bg-muted/10 space-y-1 rounded border p-1.5'>
                <div className='flex items-center space-x-1.5'>
                  <Checkbox
                    id='opt-seed'
                    checked={props.video.hasSeed}
                    disabled={props.busy}
                    onCheckedChange={(checked) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        hasSeed: Boolean(checked),
                      }))
                    }
                  />
                  <Label
                    htmlFor='opt-seed'
                    className='cursor-pointer text-[11px] font-medium'
                  >
                    {t('Seed')}
                  </Label>
                </div>
                {props.video.hasSeed && (
                  <Input
                    placeholder='123456'
                    value={props.video.seed}
                    disabled={props.busy}
                    onChange={(e) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        seed: e.target.value,
                      }))
                    }
                    className='h-6 text-[11px]'
                  />
                )}
              </div>

              {/* Audio Generation */}
              <div className='bg-muted/10 space-y-1 rounded border p-1.5'>
                <div className='flex items-center space-x-1.5'>
                  <Checkbox
                    id='opt-audio'
                    checked={props.video.hasGenerateAudio}
                    disabled={props.busy}
                    onCheckedChange={(checked) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        hasGenerateAudio: Boolean(checked),
                      }))
                    }
                  />
                  <Label
                    htmlFor='opt-audio'
                    className='cursor-pointer text-[11px] font-medium'
                  >
                    {t('Generate Audio')}
                  </Label>
                </div>
                {props.video.hasGenerateAudio && (
                  <div className='flex items-center justify-between pt-0.5'>
                    <span className='text-muted-foreground text-[10px]'>
                      {props.video.generateAudio ? t('Enabled') : t('Disabled')}
                    </span>
                    <Switch
                      checked={props.video.generateAudio}
                      disabled={props.busy}
                      onCheckedChange={(checked) =>
                        props.onVideoChange((c) => ({
                          ...c,
                          generateAudio: Boolean(checked),
                        }))
                      }
                      className='scale-75'
                    />
                  </div>
                )}
              </div>

              {/* Return Last Frame */}
              <div className='bg-muted/10 space-y-1 rounded border p-1.5'>
                <div className='flex items-center space-x-1.5'>
                  <Checkbox
                    id='opt-return-last'
                    checked={props.video.hasReturnLastFrame}
                    disabled={props.busy}
                    onCheckedChange={(checked) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        hasReturnLastFrame: Boolean(checked),
                      }))
                    }
                  />
                  <Label
                    htmlFor='opt-return-last'
                    className='cursor-pointer text-[11px] font-medium'
                  >
                    {t('Return Last Frame')}
                  </Label>
                </div>
                {props.video.hasReturnLastFrame && (
                  <div className='flex items-center justify-between pt-0.5'>
                    <span className='text-muted-foreground text-[10px]'>
                      {props.video.returnLastFrame
                        ? t('Enabled')
                        : t('Disabled')}
                    </span>
                    <Switch
                      checked={props.video.returnLastFrame}
                      disabled={props.busy}
                      onCheckedChange={(checked) =>
                        props.onVideoChange((c) => ({
                          ...c,
                          returnLastFrame: Boolean(checked),
                        }))
                      }
                      className='scale-75'
                    />
                  </div>
                )}
              </div>

              {/* Custom Extra JSON */}
              <div className='bg-muted/10 col-span-2 space-y-1 rounded border p-1.5 sm:col-span-4'>
                <div className='flex items-center space-x-1.5'>
                  <Checkbox
                    id='opt-custom-json'
                    checked={props.video.hasCustomJson}
                    disabled={props.busy}
                    onCheckedChange={(checked) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        hasCustomJson: Boolean(checked),
                      }))
                    }
                  />
                  <Label
                    htmlFor='opt-custom-json'
                    className='cursor-pointer text-[11px] font-medium'
                  >
                    {t('Custom Extra Parameters (Merged into top-level JSON)')}
                  </Label>
                </div>
                {props.video.hasCustomJson && (
                  <Textarea
                    rows={1}
                    placeholder='{"camera_motion": "pan_left"}'
                    value={props.video.customJson}
                    disabled={props.busy}
                    onChange={(e) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        customJson: e.target.value,
                      }))
                    }
                    className='font-mono text-xs'
                  />
                )}
              </div>
            </div>
          </Card>

          {/* Task Control & Manual Query Toolbar */}
          <Card className='border-primary/20 bg-primary/5 p-2.5'>
            <div className='space-y-2'>
              <div className='flex flex-wrap items-center justify-between gap-2'>
                <div className='flex flex-1 items-center gap-2'>
                  <Label className='text-muted-foreground shrink-0 text-xs font-medium'>
                    {t('Task ID')}:
                  </Label>
                  <Input
                    placeholder={t('Enter Task ID or generate by submitting')}
                    value={props.video.taskId ?? ''}
                    disabled={props.busy || isManualQuerying}
                    onChange={(e) =>
                      props.onVideoChange((c) => ({
                        ...c,
                        taskId: e.target.value,
                      }))
                    }
                    className='h-7 font-mono text-xs'
                  />
                </div>

                <Button
                  type='button'
                  variant='secondary'
                  size='sm'
                  disabled={props.busy || isManualQuerying || !effectiveTaskId}
                  className='h-7 gap-1 text-xs'
                  onClick={handleManualQuery}
                >
                  <Zap className='size-3 text-amber-500' />
                  {isManualQuerying ? t('Querying...') : t('Query Status Now')}
                </Button>
              </div>

              {/* Execution Actions */}
              <div className='flex flex-wrap items-center justify-between gap-2 pt-1'>
                <div className='flex items-center gap-2'>
                  <Button
                    type='button'
                    variant='outline'
                    size='sm'
                    disabled={props.busy}
                    onClick={() => handleRunSingleCheck('video_submit')}
                    className='h-7 text-xs'
                  >
                    <Send className='mr-1 size-3' />
                    {t('Submit Only')}
                  </Button>

                  <Button
                    type='button'
                    variant='outline'
                    size='sm'
                    disabled={props.busy || !effectiveTaskId}
                    onClick={() => handleRunSingleCheck('video_poll')}
                    className='h-7 text-xs'
                  >
                    <RefreshCw className='mr-1 size-3' />
                    {t('Poll Only')}
                  </Button>
                </div>

                <div className='flex items-center gap-2'>
                  <Button
                    type='button'
                    size='sm'
                    disabled={props.busy}
                    onClick={() => props.onRunCheck?.()}
                    className='h-7 gap-1 text-xs font-semibold'
                  >
                    <Play className='size-3 fill-current' />
                    {t('Start Video Test (Submit & Poll)')}
                  </Button>
                </div>
              </div>
            </div>
          </Card>

          {/* Step Checks Table */}
          <div className='space-y-1.5'>
            <p className='text-xs font-medium'>
              {t('Task Execution Pipeline')}
            </p>
            <CheckTable
              checks={props.videoChecks}
              busy={props.busy}
              onRun={(id) => handleRunSingleCheck(id)}
            />
          </div>

          {/* Generated Video Output Player */}
          {props.videoMetrics?.video_url && (
            <Card className='border-primary/20 p-3'>
              <div className='flex items-center justify-between pb-2'>
                <div className='flex items-center gap-1.5'>
                  <Video className='text-primary size-4' />
                  <span className='text-xs font-semibold'>
                    {t('Generated Video Output')}
                  </span>
                </div>
                <div className='flex items-center gap-2'>
                  {props.onExportPdf && (
                    <Button
                      variant='outline'
                      size='sm'
                      className='h-6 gap-1 text-[11px]'
                      onClick={props.onExportPdf}
                    >
                      <Download className='size-3' />
                      {t('Export PDF')}
                    </Button>
                  )}
                  <a
                    href={props.videoMetrics.video_url}
                    target='_blank'
                    rel='noreferrer'
                    className='text-primary flex items-center gap-1 text-xs hover:underline'
                  >
                    {t('Open URL')}
                    <ExternalLink className='size-3' />
                  </a>
                </div>
              </div>
              <div className='flex justify-center overflow-hidden rounded bg-black'>
                <video
                  src={props.videoMetrics.video_url}
                  controls
                  autoPlay
                  playsInline
                  className='max-h-80 w-auto max-w-full'
                />
              </div>
            </Card>
          )}
        </div>

        <div className='lg:sticky lg:top-4 lg:col-span-5'>
          <RawJsonDialog
            videoMetrics={props.videoMetrics}
            previewRequestJson={previewRequestJson}
            pollJsonOverride={manualPollJson}
            busy={props.busy}
            activeTab={activeJsonTab}
            onActiveTabChange={setActiveJsonTab}
            onApplyToForm={handleApplyJsonToForm}
            onSendRawJson={props.onSendRawJson}
            onEffectiveJsonChange={handleEffectiveJsonChange}
          />
        </div>
      </div>
    </TabsContent>
  )
}
