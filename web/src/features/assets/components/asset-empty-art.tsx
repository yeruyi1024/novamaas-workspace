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
import { AudioLines, Film, Image, Search } from 'lucide-react'

import { cn } from '@/lib/utils'

const TILES = [
  {
    Icon: Image,
    className:
      '-rotate-12 -translate-x-9 translate-y-1 bg-card text-muted-foreground',
  },
  {
    Icon: AudioLines,
    className:
      'rotate-12 translate-x-9 translate-y-1 bg-card text-muted-foreground',
  },
  {
    Icon: Film,
    className: 'z-10 -translate-y-1 bg-card text-foreground',
  },
] as const

/** Decorative stack of media tiles for the library's empty states. */
export function AssetEmptyArt(props: { variant: 'library' | 'search' }) {
  return (
    <div
      aria-hidden='true'
      className='relative mb-2 grid h-24 w-40 place-items-center'
    >
      {TILES.map((tile) => {
        const TileIcon = tile.Icon
        return (
          <span
            key={tile.className}
            className={cn(
              'ring-border/70 absolute grid size-15 place-items-center rounded-2xl shadow-sm ring-1',
              tile.className
            )}
          >
            <TileIcon className='size-6' strokeWidth={1.6} />
          </span>
        )
      })}
      {props.variant === 'search' && (
        <span className='bg-background ring-border absolute right-6 -bottom-1 z-20 grid size-8 place-items-center rounded-full shadow-sm ring-1'>
          <Search className='text-muted-foreground size-4' strokeWidth={1.75} />
        </span>
      )}
    </div>
  )
}
