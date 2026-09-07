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
import type { SVGProps } from 'react'

import { cn } from '@/lib/utils'

export function Logo({ className, ...props }: SVGProps<SVGSVGElement>) {
  return (
    <svg
      id='newapi-logo'
      viewBox='0 0 180 180'
      xmlns='http://www.w3.org/2000/svg'
      height='24'
      width='24'
      className={cn('size-6', className)}
      {...props}
    >
      <title>New API</title>
      <rect x='8' y='8' width='164' height='164' rx='46' fill='#17336F' />
      <path
        d='M21 125 125 21h1c25 0 45 20 45 45v60c0 25-20 45-45 45H66c-22 0-41-16-45-37v-9Z'
        fill='#3658A5'
        opacity='.68'
      />
      <rect
        x='13'
        y='13'
        width='154'
        height='154'
        rx='41'
        fill='none'
        stroke='#D9E3FF'
        strokeOpacity='.28'
        strokeWidth='2'
      />
      <path d='m40 50 9-9 39 39-9 9Z' fill='#AFC5E8' />
      <path d='m131 41 9 9-39 39-9-9Z' fill='#AFC5E8' />
      <path d='m40 130 39-39 9 9-39 39Z' fill='#AFC5E8' />
      <path d='m101 91 39 39-9 9-39-39Z' fill='#AFC5E8' />
      <circle cx='42' cy='43' r='10' fill='#F5F7FF' />
      <circle cx='138' cy='43' r='10' fill='#F5F7FF' />
      <circle cx='42' cy='137' r='10' fill='#F5F7FF' />
      <circle cx='138' cy='137' r='10' fill='#F5F7FF' />
      <circle cx='42' cy='43' r='4' fill='#6385E8' />
      <circle cx='138' cy='43' r='4' fill='#6385E8' />
      <circle cx='42' cy='137' r='4' fill='#6385E8' />
      <circle cx='138' cy='137' r='4' fill='#6385E8' />
      <path d='m90 51 34 20v38l-34 20-34-20V71Z' fill='#F8FAFF' />
      <path d='m90 68 19 11v22l-19 11-19-11V79Z' fill='#24438F' />
      <path d='m90 78 12 12-12 12-12-12Z' fill='#9FC2FF' />
    </svg>
  )
}
