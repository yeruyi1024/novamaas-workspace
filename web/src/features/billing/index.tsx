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
import { searchUsers } from '@/features/users/api'
import { useDebounce } from '@/hooks/use-debounce'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import { billingToday } from './api'
import { BillingProfileCard } from './components/billing-profile-card'
import { ConsumptionPanel } from './components/consumption-panel'
import { StatementsPanel } from './components/statements-panel'

export function Billing() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.auth.user)
  const admin = (user?.role ?? 0) >= ROLE.ADMIN
  const [selectedAccount, setSelectedAccount] = useState<{
    value: number
    label: string
  } | null>(null)
  const [search, setSearch] = useState('')
  const keyword = useDebounce(search, 300)
  const [date, setDate] = useState(billingToday)
  const [tab, setTab] = useState('daily')
  const userId = (admin ? selectedAccount?.value : null) ?? user?.id ?? 0
  const accounts = useQuery({
    queryKey: ['billing', 'users', keyword],
    queryFn: () => searchUsers({ keyword, page_size: 30 }),
    enabled: admin,
  })
  const userOptions = [
    {
      value: user?.id ?? 0,
      label: `${user?.username ?? t('My account')} (#${user?.id ?? 0})`,
    },
    ...(accounts.data?.data?.items || [])
      .filter((item) => item.id !== user?.id)
      .map((item) => ({
        value: item.id,
        label: `${item.username} (#${item.id})`,
      })),
  ]
  if (userId > 0 && !userOptions.some((item) => item.value === userId)) {
    userOptions.push(
      selectedAccount ?? { value: userId, label: `${t('Account')} #${userId}` }
    )
  }
  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        {t('Billing statements')}
      </SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='flex flex-col gap-5 overflow-auto p-1'>
          {admin && (
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
                <FieldLabel htmlFor='billing-account'>
                  {t('Account')}
                </FieldLabel>
                <Select
                  items={userOptions}
                  value={userId}
                  onValueChange={(value) =>
                    value !== null &&
                    setSelectedAccount(
                      userOptions.find((item) => item.value === value) ?? null
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
              <TabsTrigger value='identity'>
                {t('Billing identity')}
              </TabsTrigger>
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
          </Tabs>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
