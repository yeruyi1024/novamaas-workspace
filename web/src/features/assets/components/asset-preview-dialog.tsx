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
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'

import { getMediaAssetPreview } from '../api'
import { assertAssetSuccess, formatAssetBytes } from '../asset-utils'
import type { MediaAsset } from '../types'

export function AssetPreviewDialog(props: {
  asset: MediaAsset
  onOpenChange: (open: boolean) => void
  onDownload: (asset: MediaAsset) => void
  downloading: boolean
}) {
  const { t } = useTranslation()
  const previewQuery = useQuery({
    queryKey: ['asset-library', 'preview', props.asset.id, 'original'],
    queryFn: async () =>
      assertAssetSuccess(
        await getMediaAssetPreview(props.asset.id, 'original')
      ),
    staleTime: 5 * 60 * 1000,
  })
  let previewContent = <Skeleton className='h-64 w-full' />
  if (!previewQuery.isLoading && !previewQuery.data?.url) {
    previewContent = (
      <div className='bg-muted text-muted-foreground flex min-h-64 items-center justify-center rounded-md'>
        {t('Preview unavailable')}
      </div>
    )
  } else if (previewQuery.data?.url && props.asset.type === 'image') {
    previewContent = (
      <img
        src={previewQuery.data.url}
        alt={`${props.asset.name} ${t('Preview')}`}
        className='mx-auto max-h-[70dvh] max-w-full object-contain'
      />
    )
  } else if (previewQuery.data?.url && props.asset.type === 'video') {
    previewContent = (
      <video
        src={previewQuery.data.url}
        controls
        preload='metadata'
        className='max-h-[70dvh] w-full rounded-md bg-black object-contain'
      />
    )
  } else if (previewQuery.data?.url) {
    previewContent = (
      <audio
        src={previewQuery.data.url}
        controls
        preload='metadata'
        className='my-10 w-full'
      />
    )
  }

  return (
    <Dialog open onOpenChange={props.onOpenChange}>
      <DialogContent className='max-h-[calc(100dvh-2rem)] overflow-y-auto sm:max-w-4xl'>
        <DialogHeader>
          <DialogTitle>{props.asset.name}</DialogTitle>
          <DialogDescription>
            {t('Preview asset')} · {formatAssetBytes(props.asset.size)}
          </DialogDescription>
        </DialogHeader>
        {previewContent}
        <DialogFooter>
          <Button
            type='button'
            disabled={props.downloading}
            onClick={() => props.onDownload(props.asset)}
          >
            {t('Download')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
