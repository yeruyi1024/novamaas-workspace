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

import { SectionPageLayout } from '@/components/layout'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { UpstreamBilling } from '@/features/upstream-billing'
import { canViewUpstreamBilling } from '@/features/upstream-billing/permissions'
import { useDebounce } from '@/hooks/use-debounce'
import {
  ADMIN_PERMISSION_ACTIONS,
  ADMIN_PERMISSION_RESOURCES,
  hasPermission,
} from '@/lib/admin-permissions'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import { billingToday, getBillingAccount, searchBillingAccounts } from './api'
import { BillingProfileCard } from './components/billing-profile-card'
import { ConsumptionPanel } from './components/consumption-panel'
import { CorrectionPanel } from './components/correction-panel'
import { StatementsPanel } from './components/statements-panel'
import type { BillingAccountOption } from './types'

export function Billing(
  props: {
    area?: 'customer' | 'upstream'
    onAreaChange?: (area: 'customer' | 'upstream') => void
  } = {}
) {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.auth.user)
  const allowed = canViewUpstreamBilling(user)
  const [localArea, setLocalArea] = useState<'customer' | 'upstream'>(
    'customer'
  )
  const activeArea = allowed ? props.area || localArea : 'customer'
  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        {t('Billing statements')}
      </SectionPageLayout.Title>
      <SectionPageLayout.Content>
        {allowed ? (
          <Tabs
            value={activeArea}
            className='min-h-0 flex-1'
            onValueChange={(value) => {
              if (value !== 'customer' && value !== 'upstream') return
              setLocalArea(value)
              props.onAreaChange?.(value)
            }}
          >
            <TabsList>
              <TabsTrigger value='customer'>
                {t('Customer statements')}
              </TabsTrigger>
              <TabsTrigger value='upstream'>
                {t('Upstream statements')}
              </TabsTrigger>
            </TabsList>
            <TabsContent value='customer' className='min-h-0 overflow-auto'>
              {activeArea === 'customer' && <CustomerBilling />}
            </TabsContent>
            <TabsContent value='upstream' className='min-h-0 overflow-auto'>
              {activeArea === 'upstream' && <UpstreamBilling />}
            </TabsContent>
          </Tabs>
        ) : (
          <CustomerBilling />
        )}
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}

function CustomerBilling() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.auth.user)
  const admin = (user?.role ?? 0) >= ROLE.ADMIN
  const financialAccounting =
    admin &&
    hasPermission(
      user,
      ADMIN_PERMISSION_RESOURCES.FINANCIAL_ACCOUNTING,
      ADMIN_PERMISSION_ACTIONS.VIEW
    )
  const canSwitchAccount = admin && financialAccounting
  const [selectedAccount, setSelectedAccount] =
    useState<BillingAccountOption | null>(null)
  const [search, setSearch] = useState('')
  const keyword = useDebounce(search, 300)
  const [date, setDate] = useState(billingToday)
  const [tab, setTab] = useState('daily')
  const userId =
    (canSwitchAccount ? selectedAccount?.id : null) ?? user?.id ?? 0
  const accounts = useQuery({
    queryKey: ['billing', 'users', keyword],
    queryFn: () => searchBillingAccounts(keyword),
    enabled: canSwitchAccount,
  })
  const ownAccount = useQuery({
    queryKey: ['billing', 'account', user?.id ?? 0],
    queryFn: () => getBillingAccount(user?.id ?? 0),
    enabled: canSwitchAccount && Boolean(user?.id),
  })
  const activeAccount = useQuery({
    queryKey: ['billing', 'account', userId],
    queryFn: () => getBillingAccount(userId),
    enabled: canSwitchAccount && userId > 0,
  })
  const candidates: BillingAccountOption[] = [
    {
      id: user?.id ?? 0,
      username: user?.username ?? t('My account'),
      company_title: ownAccount.data?.company_title ?? '',
      accounting_start_at: ownAccount.data?.accounting_start_at ?? 0,
    },
    ...(accounts.data?.items || []).filter((item) => item.id !== user?.id),
  ]
  if (selectedAccount && !candidates.some((item) => item.id === userId)) {
    candidates.push(selectedAccount)
  }
  const userOptions = candidates.map((item) => {
    const profile = item.id === userId ? activeAccount.data : undefined
    const start = profile?.accounting_start_at ?? item.accounting_start_at
    const title = profile?.company_title ?? item.company_title
    let label = `${item.username} (#${item.id})`
    if (start > 0) {
      label += ` ${t('(Formal accounting - {{title}})', { title })}`
    }
    return { value: item.id, label }
  })
  return (
    <div className='flex flex-col gap-5 overflow-auto p-1'>
      {canSwitchAccount && (
        <FieldGroup className='grid gap-4 sm:grid-cols-2'>
          <Field>
            <FieldLabel htmlFor='billing-search'>
              {t('Search accounts')}
            </FieldLabel>
            <Input
              id='billing-search'
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </Field>
          <Field>
            <FieldLabel htmlFor='billing-account'>{t('Account')}</FieldLabel>
            <Select
              items={userOptions}
              value={userId}
              onValueChange={(value) =>
                value !== null &&
                setSelectedAccount(
                  candidates.find((item) => item.id === value) ?? null
                )
              }
            >
              <SelectTrigger id='billing-account' className='w-full'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {userOptions.map((item) => (
                    <SelectItem key={item.value} value={item.value}>
                      {item.label}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
        </FieldGroup>
      )}
      <p className='text-sm break-words'>
        {t('Selected billing customer')}:{' '}
        {userOptions.find((item) => item.value === userId)?.label}
      </p>
      <Tabs value={tab} onValueChange={(value) => setTab(String(value))}>
        <TabsList>
          <TabsTrigger value='daily'>{t('Consumption lookup')}</TabsTrigger>
          <TabsTrigger value='statements'>
            {t('Billing statements')}
          </TabsTrigger>
          <TabsTrigger value='identity'>{t('Billing identity')}</TabsTrigger>
          {financialAccounting && (
            <TabsTrigger value='corrections'>
              {t('Billing adjustments')}
            </TabsTrigger>
          )}
        </TabsList>
        <TabsContent value='daily'>
          <ConsumptionPanel
            key={userId}
            userId={userId}
            date={date}
            onDateChange={setDate}
          />
        </TabsContent>
        <TabsContent value='statements'>
          <StatementsPanel
            key={userId}
            userId={userId}
            currentUserId={user?.id ?? 0}
            admin={admin}
            onConfigureIdentity={() => setTab('identity')}
            onSelectDay={(day) => {
              setDate(day)
              setTab('daily')
            }}
          />
        </TabsContent>
        <TabsContent value='identity'>
          <BillingProfileCard userId={userId} admin={admin} />
        </TabsContent>
        {financialAccounting && (
          <TabsContent value='corrections'>
            <CorrectionPanel
              key={userId}
              userId={userId}
              actorId={user?.id ?? 0}
              canManage={user?.role === ROLE.SUPER_ADMIN}
            />
          </TabsContent>
        )}
      </Tabs>
    </div>
  )
}
