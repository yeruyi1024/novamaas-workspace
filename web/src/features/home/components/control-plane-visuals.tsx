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
  ApiGatewayIcon,
  ArrowRight01Icon,
  CheckmarkCircle02Icon,
  Route02Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useTranslation } from 'react-i18next'

const providerNames = ['OpenAI', 'Claude', 'Gemini']

const settlementEntries = [
  { requestId: 'req_84C2', provider: 'OpenAI', usage: '12.4k', cost: '$0.84' },
  { requestId: 'req_12A7', provider: 'Claude', usage: '18.7k', cost: '$1.36' },
  { requestId: 'req_9D41', provider: 'Gemini', usage: '9.6k', cost: '$0.62' },
]

export function SupplyConvergenceVisual() {
  const { t } = useTranslation()

  return (
    <div
      role='img'
      aria-label={t('Unify heterogeneous supply')}
      className='maas-supply-visual grid grid-cols-1 items-center rounded-2xl p-3 sm:grid-cols-[minmax(0,1fr)_3.5rem_minmax(0,1.2fr)] sm:p-4'
    >
      <div aria-hidden='true' className='grid grid-cols-2 gap-2'>
        {providerNames.map((provider) => (
          <span
            key={provider}
            className='maas-supply-source flex min-w-0 items-center gap-2 rounded-lg px-2.5 py-2 text-[10px] font-medium sm:text-xs'
          >
            <span className='maas-provider-signal size-1.5 shrink-0 rounded-full' />
            <span className='truncate'>{provider}</span>
          </span>
        ))}
        <span className='maas-supply-source flex min-w-0 items-center gap-2 rounded-lg px-2.5 py-2 text-[10px] font-medium sm:text-xs'>
          <span className='maas-provider-signal size-1.5 shrink-0 rounded-full' />
          <span className='truncate'>{t('More supply')}</span>
        </span>
      </div>

      <div
        aria-hidden='true'
        className='maas-supply-connector relative h-8 sm:h-full'
      >
        <HugeiconsIcon
          icon={ArrowRight01Icon}
          className='maas-supply-arrow absolute size-4'
          strokeWidth={1.8}
        />
      </div>

      <div
        aria-hidden='true'
        className='maas-unified-supply-node min-w-0 rounded-xl p-3 sm:p-4'
      >
        <div className='flex items-center gap-3'>
          <span className='maas-unified-supply-icon flex size-9 shrink-0 items-center justify-center rounded-lg'>
            <HugeiconsIcon
              icon={ApiGatewayIcon}
              className='size-4'
              strokeWidth={1.8}
            />
          </span>
          <div className='min-w-0'>
            <p className='truncate text-xs font-semibold sm:text-sm'>
              {t('Unified token pool')}
            </p>
            <p className='maas-visual-code mt-0.5 font-mono text-[9px] sm:text-[10px]'>
              POST /v1
            </p>
          </div>
        </div>
        <div className='mt-3 flex flex-wrap gap-1.5'>
          <span className='maas-supply-output rounded-md px-2 py-1 text-[9px] sm:text-[10px]'>
            {t('Public API supply')}
          </span>
          <span className='maas-supply-output rounded-md px-2 py-1 text-[9px] sm:text-[10px]'>
            {t('Metered access')}
          </span>
        </div>
      </div>
    </div>
  )
}

export function PolicyRoutingVisual() {
  const { t } = useTranslation()

  return (
    <div
      role='img'
      aria-label={t('Route by policy')}
      className='maas-policy-visual rounded-2xl p-3 sm:p-4'
    >
      <div aria-hidden='true' className='flex flex-wrap gap-1.5'>
        {[t('Cost'), t('Latency'), t('Availability')].map((criterion) => (
          <span
            key={criterion}
            className='maas-policy-criterion rounded-full px-2.5 py-1 text-[9px] font-medium sm:text-[10px]'
          >
            {criterion}
          </span>
        ))}
      </div>

      <div
        aria-hidden='true'
        className='maas-routing-stage relative mt-3 grid h-28 grid-cols-[4.25rem_minmax(2.75rem,1fr)_6rem] items-center sm:grid-cols-[5rem_minmax(3.5rem,1fr)_7.5rem]'
      >
        <svg
          className='maas-routing-map absolute inset-0 size-full'
          viewBox='0 0 100 100'
          preserveAspectRatio='none'
        >
          <path d='M 13 50 H 54 C 66 50 68 17 84 17 H 100' />
          <path className='maas-selected-route' d='M 13 50 H 100' />
          <path d='M 13 50 H 54 C 66 50 68 83 84 83 H 100' />
        </svg>

        <span className='maas-route-source relative flex h-10 items-center justify-center rounded-lg px-2 text-[10px] font-semibold'>
          {t('Request')}
        </span>

        <span className='maas-route-engine relative mx-auto flex size-12 flex-col items-center justify-center rounded-xl'>
          <HugeiconsIcon
            icon={Route02Icon}
            className='size-4'
            strokeWidth={1.8}
          />
          <span className='mt-0.5 text-[8px] font-semibold'>{t('Policy')}</span>
        </span>

        <span className='relative flex flex-col gap-1.5'>
          {providerNames.map((provider, index) => (
            <span
              key={provider}
              className='maas-route-target flex h-7 items-center justify-between rounded-md px-2 text-[9px] font-medium sm:text-[10px]'
              data-selected={index === 1}
            >
              {provider}
              {index === 1 ? (
                <HugeiconsIcon
                  icon={CheckmarkCircle02Icon}
                  className='size-3.5'
                  strokeWidth={2}
                />
              ) : null}
            </span>
          ))}
        </span>
      </div>
    </div>
  )
}

export function SettlementLedgerVisual() {
  const { t } = useTranslation()

  return (
    <div
      role='img'
      aria-label={t('Settle with confidence')}
      className='maas-settlement-visual overflow-hidden rounded-2xl'
    >
      <div
        aria-hidden='true'
        className='maas-settlement-header flex flex-wrap items-center justify-between gap-2 px-3 py-2.5 sm:px-4'
      >
        <div className='flex flex-wrap gap-3 text-[9px] font-medium sm:text-[10px]'>
          <span>{t('Usage')}</span>
          <span>{t('Cost')}</span>
          <span>{t('Channel health')}</span>
        </div>
        <span className='maas-ledger-status flex items-center gap-1.5 rounded-full px-2 py-1 text-[9px] font-semibold'>
          <span className='size-1.5 rounded-full' />
          {t('Healthy')}
        </span>
      </div>

      <div aria-hidden='true'>
        {settlementEntries.map((entry) => (
          <div
            key={entry.requestId}
            className='maas-ledger-row grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-3 px-3 py-2.5 text-[10px] sm:grid-cols-[minmax(0,1fr)_minmax(0,0.75fr)_auto_auto] sm:px-4 sm:text-xs'
          >
            <span className='maas-ledger-request truncate font-mono'>
              {entry.requestId}
            </span>
            <span className='text-muted-foreground hidden truncate sm:block'>
              {entry.provider}
            </span>
            <span className='text-muted-foreground font-mono'>
              {entry.usage}
            </span>
            <span className='font-mono font-semibold'>{entry.cost}</span>
          </div>
        ))}
      </div>

      <div
        aria-hidden='true'
        className='maas-settlement-total flex items-center justify-between gap-3 px-3 py-3 sm:px-4'
      >
        <span className='flex items-center gap-2 text-[10px] font-semibold sm:text-xs'>
          <HugeiconsIcon
            icon={CheckmarkCircle02Icon}
            className='size-4'
            strokeWidth={2}
          />
          {t('Metered settlement')}
        </span>
        <span className='text-[10px] sm:text-xs'>
          <span className='text-muted-foreground'>{t('Total')}</span>{' '}
          <strong className='ml-1 font-mono'>$2.82</strong>
        </span>
      </div>
    </div>
  )
}
