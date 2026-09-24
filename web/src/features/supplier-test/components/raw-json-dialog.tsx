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
  Check,
  Copy,
  Edit3,
  Eye,
  Play,
  RotateCcw,
  Sparkles,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { copyToClipboard } from '@/lib/copy-to-clipboard'

import type { VideoMetrics } from '../types'
import { parseVideoPayloadToForm } from '../video-json'

function formatJSON(raw?: string): string {
  if (!raw || !raw.trim()) return ''
  try {
    const parsed = JSON.parse(raw)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return raw
  }
}

function JsonCodeViewer({ code, label }: { code: string; label: string }) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    if (!code) return
    const ok = await copyToClipboard(code)
    if (!ok) {
      toast.error(t('Failed to copy'))
      return
    }
    setCopied(true)
    toast.success(t('JSON copied to clipboard'))
    setTimeout(() => setCopied(false), 2000)
  }

  if (!code) {
    return (
      <div className='bg-muted/40 text-muted-foreground flex h-60 items-center justify-center rounded-lg border border-dashed text-sm'>
        {t('No data available for {{label}} yet', { label })}
      </div>
    )
  }

  return (
    <div className='relative'>
      <div className='absolute top-2 right-2 z-10'>
        <Button
          variant='secondary'
          size='xs'
          className='h-7 gap-1 px-2 text-xs'
          onClick={handleCopy}
        >
          {copied ? (
            <Check className='size-3.5' />
          ) : (
            <Copy className='size-3.5' />
          )}
          {copied ? t('Copied') : t('Copy JSON')}
        </Button>
      </div>
      <pre className='bg-muted/60 max-h-[min(70vh,640px)] overflow-auto rounded-lg border p-3.5 pt-8 font-mono text-xs leading-relaxed break-all whitespace-pre-wrap select-text'>
        {code}
      </pre>
    </div>
  )
}

export function RawJsonDialog(props: {
  videoMetrics?: VideoMetrics | null
  previewRequestJson?: string
  pollJsonOverride?: string
  busy?: boolean
  activeTab?: 'request' | 'submit' | 'poll'
  onActiveTabChange?: (tab: 'request' | 'submit' | 'poll') => void
  onSendRawJson?: (rawJson: string) => void
  onApplyToForm?: (rawJson: string) => void
  onEffectiveJsonChange?: (json: string) => void
}) {
  const { t } = useTranslation()
  const [uncontrolledTab, setUncontrolledTab] = useState<
    'request' | 'submit' | 'poll'
  >('request')
  const activeTab = props.activeTab ?? uncontrolledTab
  const setActiveTab = (tab: 'request' | 'submit' | 'poll') => {
    setUncontrolledTab(tab)
    props.onActiveTabChange?.(tab)
  }
  const [editMode, setEditMode] = useState(false)
  const [customText, setCustomText] = useState('')

  const displayRequestJSON = props.previewRequestJson || ''

  const submitJSON = formatJSON(props.videoMetrics?.raw_submit_response_json)
  const pollJSON = formatJSON(
    props.pollJsonOverride || props.videoMetrics?.raw_poll_response_json
  )
  const onEffectiveJsonChange = props.onEffectiveJsonChange
  useEffect(() => {
    if (!editMode) setCustomText(displayRequestJSON)
  }, [editMode, displayRequestJSON])
  useEffect(() => {
    onEffectiveJsonChange?.(editMode ? customText.trim() : '')
  }, [editMode, customText, onEffectiveJsonChange])

  const handleFormat = () => {
    try {
      const parsed = JSON.parse(customText)
      const formatted = JSON.stringify(parsed, null, 2)
      setCustomText(formatted)
      toast.success(t('JSON formatted successfully'))
    } catch {
      toast.error(t('Invalid JSON format'))
    }
  }

  const handleResetToPreview = () => {
    setCustomText(displayRequestJSON)
    toast.success(t('Reset to current form preview'))
  }

  const handleApplyToForm = () => {
    if (!props.onApplyToForm) return
    const res = parseVideoPayloadToForm(customText)
    if (!res.success) {
      toast.error(
        t('Failed to parse JSON: {{error}}', { error: res.error || '' })
      )
      return
    }
    props.onApplyToForm(customText)
    setEditMode(false)
    toast.success(t('Successfully parsed and applied JSON to form'))
  }

  const handleSendDirectly = () => {
    if (!props.onSendRawJson) return
    const trimmed = customText.trim()
    if (!trimmed) {
      toast.error(t('Please paste or enter a JSON request body'))
      return
    }
    try {
      JSON.parse(trimmed)
    } catch {
      toast.error(t('Invalid JSON format'))
      return
    }
    props.onSendRawJson(trimmed)
    setEditMode(false)
  }

  return (
    <div className='bg-card rounded-xl border p-3 shadow-sm'>
      <div className='mb-2 flex items-center justify-between gap-2'>
        <div className='text-xs font-semibold'>
          {t('Raw Request & Response JSON')}
        </div>
        {props.videoMetrics && (
          <Badge variant='secondary' className='h-4 px-1 text-[10px]'>
            {t('Results Ready')}
          </Badge>
        )}
      </div>
      <Tabs
        value={activeTab}
        onValueChange={(val) =>
          setActiveTab(val as 'request' | 'submit' | 'poll')
        }
        className='w-full'
      >
        <TabsList className='grid w-full grid-cols-3'>
          <TabsTrigger value='request' className='text-xs'>
            {t('1. Submit Request Body')}
          </TabsTrigger>
          <TabsTrigger value='submit' className='text-xs'>
            {t('2. Submit Response Body')}
          </TabsTrigger>
          <TabsTrigger value='poll' className='text-xs'>
            {t('3. Final Poll Response')}
          </TabsTrigger>
        </TabsList>

        <TabsContent value='request' className='mt-3 space-y-3'>
          <div className='flex flex-wrap items-center justify-between gap-2 border-b pb-2'>
            <div className='flex items-center gap-2'>
              <span className='text-muted-foreground text-xs'>
                {t('Live preview based on current form settings')}
              </span>
            </div>
            <div className='flex items-center gap-1.5'>
              <Button
                variant={editMode ? 'ghost' : 'secondary'}
                size='xs'
                className='h-7 gap-1 text-xs'
                onClick={() => setEditMode(false)}
              >
                <Eye className='size-3.5' />
                {t('Preview')}
              </Button>
              <Button
                variant={editMode ? 'secondary' : 'ghost'}
                size='xs'
                className='h-7 gap-1 text-xs'
                onClick={() => {
                  setEditMode(true)
                  if (!customText.trim()) setCustomText(displayRequestJSON)
                }}
              >
                <Edit3 className='size-3.5' />
                {t('Paste & Edit JSON')}
              </Button>
            </div>
          </div>

          {!editMode ? (
            <JsonCodeViewer
              code={displayRequestJSON}
              label={t('Submit Request Body')}
            />
          ) : (
            <div className='space-y-3'>
              <Textarea
                value={customText}
                onChange={(e) => setCustomText(e.target.value)}
                placeholder='{\n  "model": "doubao-seedance-1-0-pro",\n  "content": [\n    {"type": "text", "text": "..."}\n  ]\n}'
                rows={13}
                className='font-mono text-xs leading-relaxed break-all whitespace-pre'
              />
              <div className='flex flex-wrap items-center justify-between gap-2 pt-1'>
                <div className='flex items-center gap-2'>
                  <Button
                    variant='outline'
                    size='xs'
                    className='h-7 gap-1 text-xs'
                    onClick={handleFormat}
                  >
                    <Sparkles className='size-3.5' />
                    {t('Format JSON')}
                  </Button>
                  <Button
                    variant='ghost'
                    size='xs'
                    className='text-muted-foreground h-7 gap-1 text-xs'
                    onClick={handleResetToPreview}
                  >
                    <RotateCcw className='size-3.5' />
                    {t('Reset to Form')}
                  </Button>
                </div>
                <div className='flex items-center gap-2'>
                  {props.onApplyToForm && (
                    <Button
                      variant='outline'
                      size='xs'
                      className='h-7 gap-1 text-xs'
                      onClick={handleApplyToForm}
                    >
                      {t('Apply to Form')}
                    </Button>
                  )}
                  {props.onSendRawJson && (
                    <Button
                      variant='default'
                      size='xs'
                      className='h-7 gap-1 text-xs font-semibold'
                      disabled={props.busy}
                      onClick={handleSendDirectly}
                    >
                      <Play className='size-3.5 fill-current' />
                      {t('Send This JSON Directly')}
                    </Button>
                  )}
                </div>
              </div>
            </div>
          )}
        </TabsContent>

        <TabsContent value='submit' className='mt-3'>
          <JsonCodeViewer code={submitJSON} label={t('Submit Response Body')} />
        </TabsContent>

        <TabsContent value='poll' className='mt-3'>
          <JsonCodeViewer
            code={pollJSON}
            label={t('Final Poll Response Body')}
          />
        </TabsContent>
      </Tabs>
    </div>
  )
}
