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
  ChartRelationshipIcon,
  Key01Icon,
  Route01Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export function ControlPlane() {
  const { t } = useTranslation()

  return (
    <section
      className='maas-control-section maas-deferred-section relative px-5 py-24 sm:px-6 md:py-32'
      aria-labelledby='control-plane-title'
    >
      <div aria-hidden className='maas-section-grid absolute inset-0' />
      <div className='relative mx-auto max-w-7xl'>
        <AnimateInView className='mx-auto max-w-3xl text-center'>
          <p className='maas-section-kicker'>{t('The operating layer')}</p>
          <h2
            id='control-plane-title'
            className='mt-4 text-3xl leading-tight font-semibold tracking-[-0.035em] text-balance md:text-5xl'
          >
            {t('Operate the economics, not just the API.')}
          </h2>
          <p className='text-muted-foreground mx-auto mt-5 max-w-2xl text-base leading-7 text-pretty'>
            {t(
              'Connect supply, package access, enforce policy and understand every unit of value from one control plane.'
            )}
          </p>
        </AnimateInView>

        <div className='mt-14 grid gap-4 md:grid-cols-12'>
          <AnimateInView className='md:col-span-7' animation='fade-right'>
            <Card className='maas-control-card h-full min-h-96 rounded-3xl'>
              <CardHeader className='px-6 md:px-7'>
                <span className='maas-control-icon mb-3 flex size-10 items-center justify-center rounded-xl'>
                  <HugeiconsIcon
                    icon={ApiGatewayIcon}
                    className='size-5'
                    strokeWidth={1.7}
                    aria-hidden='true'
                  />
                </span>
                <CardTitle className='text-xl'>
                  {t('Unify heterogeneous supply')}
                </CardTitle>
                <CardDescription className='max-w-xl leading-6'>
                  {t(
                    'Connect model providers and token inventories behind one compatible access layer.'
                  )}
                </CardDescription>
              </CardHeader>
              <CardContent className='mt-auto px-6 pb-4 md:px-7'>
                <div className='maas-provider-matrix grid grid-cols-2 gap-2 rounded-2xl p-3 sm:grid-cols-3'>
                  {[
                    'OpenAI',
                    'Claude',
                    'Gemini',
                    'DeepSeek',
                    'Qwen',
                    'More',
                  ].map((provider, index) => (
                    <div
                      key={provider}
                      className='maas-provider-cell flex items-center gap-2 rounded-xl px-3 py-3 text-xs font-medium'
                    >
                      <span
                        aria-hidden
                        className='maas-provider-signal size-1.5 rounded-full'
                        style={{ animationDelay: `${index * 180}ms` }}
                      />
                      {provider === 'More' ? t('More supply') : provider}
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </AnimateInView>

          <AnimateInView
            className='md:col-span-5'
            animation='fade-left'
            delay={80}
          >
            <Card className='maas-control-card h-full min-h-96 rounded-3xl'>
              <CardHeader className='px-6 md:px-7'>
                <span className='maas-control-icon mb-3 flex size-10 items-center justify-center rounded-xl'>
                  <HugeiconsIcon
                    icon={Key01Icon}
                    className='size-5'
                    strokeWidth={1.7}
                    aria-hidden='true'
                  />
                </span>
                <CardTitle className='text-xl'>
                  {t('Package access for every market')}
                </CardTitle>
                <CardDescription className='leading-6'>
                  {t(
                    'Turn capacity into controlled products for direct customers, partners and internal teams.'
                  )}
                </CardDescription>
              </CardHeader>
              <CardContent className='mt-auto px-6 pb-4 md:px-7'>
                <div className='flex flex-col gap-2.5'>
                  {[
                    t('Direct API access'),
                    t('Channel distribution'),
                    t('Enterprise workspaces'),
                  ].map((channel, index) => (
                    <div
                      key={channel}
                      className='maas-access-row flex items-center gap-3 rounded-xl px-3 py-3'
                    >
                      <span className='maas-access-index flex size-6 items-center justify-center rounded-md font-mono text-[10px]'>
                        {String(index + 1).padStart(2, '0')}
                      </span>
                      <span className='text-sm font-medium'>{channel}</span>
                      <span className='maas-access-line ml-auto h-px w-10 sm:w-16' />
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </AnimateInView>

          <AnimateInView
            className='md:col-span-5'
            animation='fade-right'
            delay={120}
          >
            <Card className='maas-control-card h-full rounded-3xl'>
              <CardHeader className='px-6 md:px-7'>
                <span className='maas-control-icon mb-3 flex size-10 items-center justify-center rounded-xl'>
                  <HugeiconsIcon
                    icon={Route01Icon}
                    className='size-5'
                    strokeWidth={1.7}
                    aria-hidden='true'
                  />
                </span>
                <CardTitle className='text-xl'>
                  {t('Route by policy')}
                </CardTitle>
                <CardDescription className='leading-6'>
                  {t(
                    'Balance availability, performance and cost without exposing upstream complexity.'
                  )}
                </CardDescription>
              </CardHeader>
              <CardContent className='px-6 pb-4 md:px-7'>
                <div className='maas-policy-orbit relative mt-4 flex h-28 items-center justify-center'>
                  <span className='maas-policy-core flex size-14 items-center justify-center rounded-full font-mono text-[10px] font-semibold'>
                    {t('Policy')}
                  </span>
                  {[t('Cost'), t('Latency'), t('Availability')].map(
                    (policy, index) => (
                      <span
                        key={policy}
                        className='maas-policy-node absolute rounded-full px-2.5 py-1 text-[10px]'
                        data-position={index}
                      >
                        {policy}
                      </span>
                    )
                  )}
                </div>
              </CardContent>
            </Card>
          </AnimateInView>

          <AnimateInView
            className='md:col-span-7'
            animation='fade-left'
            delay={160}
          >
            <Card className='maas-control-card h-full rounded-3xl'>
              <CardHeader className='px-6 md:px-7'>
                <span className='maas-control-icon mb-3 flex size-10 items-center justify-center rounded-xl'>
                  <HugeiconsIcon
                    icon={ChartRelationshipIcon}
                    className='size-5'
                    strokeWidth={1.7}
                    aria-hidden='true'
                  />
                </span>
                <CardTitle className='text-xl'>
                  {t('Settle with confidence')}
                </CardTitle>
                <CardDescription className='max-w-xl leading-6'>
                  {t(
                    'Keep usage, cost and channel performance traceable across the complete distribution chain.'
                  )}
                </CardDescription>
              </CardHeader>
              <CardContent className='mt-auto px-6 pb-4 md:px-7'>
                <div className='maas-ledger grid grid-cols-3 gap-px overflow-hidden rounded-2xl'>
                  {[t('Usage'), t('Cost'), t('Channel health')].map(
                    (metric, index) => (
                      <div key={metric} className='p-3 sm:p-4'>
                        <div className='maas-ledger-chart flex h-10 items-end gap-1'>
                          {[35, 55, 45, 78, 62].map((height) => (
                            <span
                              key={`${metric}-${height}`}
                              className='maas-ledger-bar flex-1 rounded-sm'
                              style={{
                                height: `${Math.max(18, height - index * 8)}%`,
                              }}
                            />
                          ))}
                        </div>
                        <p className='text-muted-foreground mt-3 text-[10px] sm:text-xs'>
                          {metric}
                        </p>
                      </div>
                    )
                  )}
                </div>
              </CardContent>
            </Card>
          </AnimateInView>
        </div>
      </div>
    </section>
  )
}
