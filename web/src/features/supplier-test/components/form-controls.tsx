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
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
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
import { Textarea } from '@/components/ui/textarea'

import { CORPORA, estimateTokens, resolveCorpusPrompt } from '../constants'

export function FieldSelect(props: {
  label: string
  value: string
  disabled: boolean
  items: Array<{ value: string; label: string }>
  onChange: (value: string) => void
}) {
  return (
    <div className='space-y-2'>
      <Label>{props.label}</Label>
      <Select
        items={props.items}
        value={props.value}
        disabled={props.disabled}
        onValueChange={(value) => {
          if (value) props.onChange(value)
        }}
      >
        <SelectTrigger className='w-full'>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {props.items.map((item) => (
            <SelectItem key={item.value} value={item.value}>
              {item.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}

export function StreamSwitch(props: {
  id: string
  checked: boolean
  disabled: boolean
  onChange: (checked: boolean) => void
}) {
  const { t } = useTranslation()
  return (
    <div className='space-y-1.5'>
      <Label htmlFor={props.id}>{t('Stream')}</Label>
      <div className='flex h-8 items-center'>
        <Switch
          id={props.id}
          checked={props.checked}
          disabled={props.disabled}
          onCheckedChange={(checked) => props.onChange(Boolean(checked))}
        />
      </div>
    </div>
  )
}

export function NumberField(props: {
  id: string
  label: string
  hint?: string
  value: number
  min: number
  max: number
  step?: number
  disabled: boolean
  presets?: Array<{ label: string; value: number }>
  onChange: (value: number) => void
}) {
  return (
    <div className='space-y-1.5'>
      <Label htmlFor={props.id}>{props.label}</Label>
      <Input
        id={props.id}
        type='number'
        min={props.min}
        max={props.max}
        step={props.step}
        value={props.value}
        disabled={props.disabled}
        onChange={(event) => {
          const raw = event.target.value
          const next =
            props.step && props.step < 1
              ? Number.parseFloat(raw)
              : Number.parseInt(raw, 10)
          props.onChange(Number.isNaN(next) ? 0 : next)
        }}
      />
      {props.hint ? (
        <p className='text-muted-foreground text-xs leading-snug'>{props.hint}</p>
      ) : null}
      {props.presets && props.presets.length > 0 ? (
        <div className='flex flex-wrap gap-1 pt-0.5'>
          {props.presets.map((preset) => {
            const active = props.value === preset.value
            return (
              <Button
                key={preset.label}
                type='button'
                variant={active ? 'secondary' : 'outline'}
                size='xs'
                disabled={props.disabled}
                className='h-5.5 px-1.5 text-[11px] font-normal'
                onClick={() => props.onChange(preset.value)}
              >
                {preset.label}
              </Button>
            )
          })}
        </div>
      ) : null}
    </div>
  )
}

export function OptionalNumberField(props: {
  id: string
  label: string
  value: string
  min: number
  max: number
  step?: number
  disabled: boolean
  placeholder?: string
  onChange: (value: string) => void
}) {
  return (
    <div className='space-y-2'>
      <Label htmlFor={props.id}>{props.label}</Label>
      <Input
        id={props.id}
        type='number'
        min={props.min}
        max={props.max}
        step={props.step}
        value={props.value}
        disabled={props.disabled}
        placeholder={props.placeholder}
        onChange={(event) => props.onChange(event.target.value)}
      />
    </div>
  )
}

export function CorpusPicker(props: {
  id: string
  form: { corpus: string; prompt: string }
  disabled: boolean
  onChange: (next: { corpus: string; prompt: string }) => void
}) {
  const { t } = useTranslation()
  return (
    <div className='space-y-4'>
      <FieldSelect
        label={t('Corpus')}
        value={props.form.corpus}
        disabled={props.disabled}
        items={CORPORA.map((item) => ({
          value: item.id,
          label:
            item.tokens > 0
              ? t('{{label}} ({{tokens}} tokens)', {
                  label: t(item.labelKey),
                  tokens: item.tokens,
                })
              : t(item.labelKey),
        }))}
        onChange={(corpus) =>
          props.onChange({ corpus, prompt: props.form.prompt })
        }
      />
      {props.form.corpus === 'custom' ? (
        <div className='space-y-2'>
          <Label htmlFor={props.id}>{t('Prompt')}</Label>
          <Textarea
            id={props.id}
            rows={3}
            placeholder={t('Write a short prompt, or pick a built-in corpus')}
            value={props.form.prompt}
            disabled={props.disabled}
            onChange={(event) =>
              props.onChange({
                corpus: props.form.corpus,
                prompt: event.target.value,
              })
            }
          />
          <p className='text-muted-foreground text-sm'>
            {t('{{tokens}} tokens', {
              tokens: estimateTokens(props.form.prompt),
            })}
          </p>
        </div>
      ) : (
        <p className='text-muted-foreground text-sm'>
          {t(
            'Using built-in corpus ({{tokens}} tokens, {{chars}} characters). Replace files in supplier-test/corpora to change the text.',
            {
              tokens: estimateTokens(resolveCorpusPrompt(props.form)),
              chars: resolveCorpusPrompt(props.form).length,
            }
          )}
        </p>
      )}
    </div>
  )
}
