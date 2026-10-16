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
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Field, FieldLabel } from '@/components/ui/field'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useAuthStore } from '@/stores/auth-store'

import { getUpstreamGroups } from './api'
import { GroupEditor } from './components/group-editor'
import { ReportPanel } from './components/report-panel'
import { StatementsPanel } from './components/statements-panel'
import { canViewUpstreamBilling } from './permissions'

export function UpstreamBilling() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.auth.user)
  if (!canViewUpstreamBilling(user)) {
    return <p role='alert'>{t('Access denied')}</p>
  }
  return <UpstreamWorkspace />
}

function UpstreamWorkspace() {
  const { t } = useTranslation()
  const [groupId, setGroupId] = useState<number | null>(null)
  const [editing, setEditing] = useState<'new' | 'selected' | null>(null)
  const groups = useQuery({
    queryKey: ['upstream-billing', 'groups'],
    queryFn: getUpstreamGroups,
  })
  const selected =
    groups.data?.find((group) => group.id === groupId) || groups.data?.[0]
  const options = (groups.data || []).map((group) => ({
    value: group.id,
    label: group.name,
  }))
  return (
    <div className='flex flex-col gap-5 overflow-auto p-1'>
      <p className='text-muted-foreground text-sm'>
        {t(
          'Combine channels by upstream and reconcile their recorded costs and usage.'
        )}
      </p>
      <Card>
        <CardContent className='flex flex-col gap-4 py-4 sm:flex-row sm:items-end'>
          <Field className='sm:max-w-md'>
            <FieldLabel htmlFor='upstream-group'>
              {t('Upstream accounting group')}
            </FieldLabel>
            <Select
              items={options}
              value={selected?.id || null}
              onValueChange={setGroupId}
            >
              <SelectTrigger
                id='upstream-group'
                className='w-full'
                disabled={groups.isPending || options.length === 0}
              >
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {options.map((item) => (
                    <SelectItem key={item.value} value={item.value}>
                      {item.label}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
          <div className='flex flex-wrap gap-2'>
            <Button onClick={() => setEditing('new')}>
              {t('Create upstream group')}
            </Button>
            <Button
              variant='outline'
              disabled={!selected}
              onClick={() => setEditing('selected')}
            >
              {t('Edit upstream group')}
            </Button>
          </div>
        </CardContent>
      </Card>
      {groups.isPending && <p role='status'>{t('Loading...')}</p>}
      {groups.isError && (
        <p role='alert'>
          {t('Failed to load data')}{' '}
          <Button variant='link' onClick={() => void groups.refetch()}>
            {t('Retry')}
          </Button>
        </p>
      )}
      {groups.data?.length === 0 && (
        <p role='status'>
          {t(
            'Create an upstream group and select its channels to begin reconciliation.'
          )}
        </p>
      )}
      {selected && !groups.isError && (
        <Tabs key={selected.id} defaultValue='report'>
          <TabsList>
            <TabsTrigger value='report'>{t('Consumption lookup')}</TabsTrigger>
            <TabsTrigger value='statements'>
              {t('Archived statements')}
            </TabsTrigger>
          </TabsList>
          <TabsContent value='report'>
            <ReportPanel groupId={selected.id} />
          </TabsContent>
          <TabsContent value='statements'>
            <StatementsPanel groupId={selected.id} />
          </TabsContent>
        </Tabs>
      )}
      {editing && (
        <GroupEditor
          key={
            editing === 'new' ? 'new' : `${selected?.id}:${selected?.version}`
          }
          group={editing === 'selected' ? selected || null : null}
          onClose={() => setEditing(null)}
          onSaved={setGroupId}
        />
      )}
    </div>
  )
}
