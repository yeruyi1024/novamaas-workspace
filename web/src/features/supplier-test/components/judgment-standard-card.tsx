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
import { ChevronDown, SlidersHorizontal } from 'lucide-react'
import { useState, type Dispatch, type SetStateAction } from 'react'
import { useTranslation } from 'react-i18next'

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { cn } from '@/lib/utils'

import {
  applyStandardEditorValue,
  getStandard,
  STANDARD_EDITOR_FIELDS,
  standardEditorValue,
  SUPPLIER_STANDARDS,
  type SupplierStandard,
} from '../baselines'
import { FieldSelect, NumberField } from './form-controls'

export function JudgmentStandardCard(props: {
  standard: SupplierStandard
  matchedStandardId: string | null
  onChange: Dispatch<SetStateAction<SupplierStandard>>
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <div className='bg-card rounded-xl border'>
        <CollapsibleTrigger
          render={
            <button
              type='button'
              className='hover:bg-muted/40 flex w-full items-center justify-between gap-3 rounded-xl px-4 py-3 text-left transition-colors'
            />
          }
        >
          <span className='flex min-w-0 items-center gap-3'>
            <span className='bg-muted text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-md'>
              <SlidersHorizontal className='size-4' aria-hidden='true' />
            </span>
            <span className='min-w-0'>
              <span className='block text-sm font-medium'>
                {t('Judgment standard')}
              </span>
              <span className='text-muted-foreground block text-xs'>
                {t(
                  'These numbers are the ruler after a run. They do not judge whether the model is smart. Change them for this vendor; tables update immediately.'
                )}
              </span>
            </span>
          </span>
          <span className='text-muted-foreground flex shrink-0 items-center gap-1 text-sm'>
            {open ? t('Collapse') : t('Expand')}
            <ChevronDown
              className={cn(
                'size-4 transition-transform',
                open && 'rotate-180'
              )}
              aria-hidden='true'
            />
          </span>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <div className='space-y-5 border-t px-4 py-4'>
            <div className='max-w-sm'>
              <FieldSelect
                label={t('Load preset')}
                value={props.matchedStandardId ?? 'custom'}
                disabled={false}
                items={[
                  ...SUPPLIER_STANDARDS.map((item) => ({
                    value: item.id,
                    label: t(item.labelKey),
                  })),
                  ...(props.matchedStandardId
                    ? []
                    : [
                        {
                          value: 'custom',
                          label: t('Custom standard'),
                        },
                      ]),
                ]}
                onChange={(id) => {
                  if (id === 'custom') return
                  props.onChange({ ...getStandard(id) })
                }}
              />
            </div>
            <div className='mt-4 space-y-5'>
              {(
                [
                  {
                    group: 'stress' as const,
                    title: t('From stress test'),
                    note: t(
                      'Error rate is the share of failed HTTP/API requests in that short run (timeouts, 5xx, refused). It is not wrong model answers.'
                    ),
                  },
                  {
                    group: 'cache' as const,
                    title: t('From cache test'),
                    note: t(
                      'These rulers score Cache test only. They do not use stress numbers.'
                    ),
                  },
                ]
              ).map((section) => (
                <div key={section.group}>
                  <p className='mb-1 text-sm font-medium'>{section.title}</p>
                  <p className='text-muted-foreground mb-3 text-xs'>
                    {section.note}
                  </p>
                  <div className='grid gap-4 md:grid-cols-2 lg:grid-cols-3'>
                    {STANDARD_EDITOR_FIELDS.filter(
                      (field) => field.group === section.group
                    ).map((field) => (
                      <NumberField
                        key={field.id}
                        id={field.id}
                        label={t(field.label)}
                        hint={t(field.hint)}
                        value={standardEditorValue(props.standard, field)}
                        disabled={false}
                        min={field.min}
                        max={field.max}
                        step={field.step}
                        onChange={(value) =>
                          props.onChange((current) =>
                            applyStandardEditorValue(current, field, value)
                          )
                        }
                      />
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </CollapsibleContent>
      </div>
    </Collapsible>
  )
}
