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

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import type { AssetLibraryConnection } from './types'

export function AssetSyncLibraryOverview(props: {
  libraries: AssetLibraryConnection[]
  onSelect: (id: number) => void
}) {
  const { t } = useTranslation()
  const groups = new Map<string, AssetLibraryConnection[]>()
  const retired: AssetLibraryConnection[] = []
  for (const library of props.libraries) {
    if (
      library.channels.length === 0 &&
      (library.legacy_channel_id != null || library.summary.total > 0)
    ) {
      retired.push(library)
      continue
    }
    const provider = library.provider_name || library.name
    const group = groups.get(provider) ?? []
    group.push(library)
    groups.set(provider, group)
  }
  return (
    <div className='grid gap-4'>
      {[...groups].map(([provider, libraries]) => (
        <Card key={provider} role='region' aria-label={provider}>
          <CardHeader>
            <CardTitle>{provider}</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('Resource account / library name')}</TableHead>
                  <TableHead>{t('Linked channels')}</TableHead>
                  <TableHead>{t('Synchronized')}</TableHead>
                  <TableHead>{t('Synchronizing')}</TableHead>
                  <TableHead>{t('Synchronization failed')}</TableHead>
                  <TableHead className='text-right'>{t('Actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {libraries.map((library) => (
                  <TableRow key={library.id}>
                    <TableCell>
                      <span className='font-medium'>{library.name}</span>
                      <span className='text-muted-foreground mt-1 block text-xs'>
                        ID {library.id}
                      </span>
                    </TableCell>
                    <TableCell>{library.channels.length}</TableCell>
                    <TableCell>{library.summary.active}</TableCell>
                    <TableCell>
                      {library.summary.pending + library.summary.processing}
                    </TableCell>
                    <TableCell>
                      {library.summary.failed > 0 ? (
                        <Badge variant='destructive'>
                          {library.summary.failed}
                        </Badge>
                      ) : (
                        '—'
                      )}
                    </TableCell>
                    <TableCell className='text-right'>
                      <Button
                        variant='outline'
                        size='sm'
                        onClick={() => props.onSelect(library.id)}
                      >
                        {t('View details')}
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      ))}
      {groups.size === 0 && (
        <p className='text-muted-foreground py-8 text-center text-sm'>
          {t('No linked asset libraries')}
        </p>
      )}
      {retired.length > 0 && (
        <details className='rounded-lg border p-4'>
          <summary className='cursor-pointer text-sm font-medium'>
            {t('Unlinked / historical library')} · {retired.length}
          </summary>
          <div className='mt-3 grid gap-2'>
            {retired.map((library) => (
              <div
                key={library.id}
                className='flex flex-wrap items-center justify-between gap-2 rounded-md border p-3 text-sm'
              >
                <span>
                  {library.provider_name} · {library.name} · ID {library.id}
                </span>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={() => props.onSelect(library.id)}
                >
                  {t('View details')}
                </Button>
              </div>
            ))}
          </div>
        </details>
      )}
    </div>
  )
}
