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
  Delete02Icon,
  FolderAddIcon,
  Image01Icon,
  Search01Icon,
  ShieldKeyIcon,
  Upload01Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { useDeferredValue, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from '@/components/ui/input-group'
import {
  Pagination,
  PaginationContent,
  PaginationItem,
} from '@/components/ui/pagination'
import { Skeleton } from '@/components/ui/skeleton'
import { useIsAdmin } from '@/hooks/use-admin'
import { useAuthStore } from '@/stores/auth-store'

import {
  createAssetGroup,
  deleteAssetGroup,
  deleteMediaAsset,
  getMediaAssetPreview,
  listAssetGroups,
  listAssetGroupsPage,
  listMediaAssets,
  uploadMediaAsset,
} from './api'
import { assertAssetSuccess, assetErrorMessage } from './asset-utils'
import { AssetApiAccessDialog } from './components/asset-api-access-dialog'
import { AssetCard } from './components/asset-card'
import {
  AssetGroupDialog,
  type AssetGroupFormValues,
} from './components/asset-group-dialog'
import { AssetGroupSelect } from './components/asset-group-select'
import { AssetPreviewDialog } from './components/asset-preview-dialog'
import { AssetUploadDialog } from './components/asset-upload-dialog'
import type { AssetGroup, MediaAsset } from './types'

const GROUPS_QUERY_KEY = ['asset-library', 'groups'] as const
const ASSETS_QUERY_KEY = ['asset-library', 'assets'] as const
const ASSET_PAGE_SIZE = 40

export function AssetLibrary() {
  const { t } = useTranslation()
  const isAdmin = useIsAdmin()
  const currentUserId = useAuthStore((state) => state.auth.user?.id ?? 0)
  const queryClient = useQueryClient()
  const [selectedGroup, setSelectedGroup] = useState('')
  const [selectedGroupInfo, setSelectedGroupInfo] = useState<
    AssetGroup | undefined
  >()
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [apiAccessDialogOpen, setApiAccessDialogOpen] = useState(false)
  const [groupDialogOpen, setGroupDialogOpen] = useState(false)
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false)
  const [reuploadTarget, setReuploadTarget] = useState<MediaAsset | null>(null)
  const [previewTarget, setPreviewTarget] = useState<MediaAsset | null>(null)
  const [uploadProgress, setUploadProgress] = useState(0)
  const [deleteTarget, setDeleteTarget] = useState<MediaAsset | null>(null)
  const [deleteGroupDialogOpen, setDeleteGroupDialogOpen] = useState(false)
  const deferredSearch = useDeferredValue(search.trim())

  const groupsQuery = useQuery({
    queryKey: [...GROUPS_QUERY_KEY, 'upload', currentUserId],
    queryFn: async () => assertAssetSuccess(await listAssetGroups(false)),
    enabled: uploadDialogOpen,
  })
  const ownGroupsQuery = useQuery({
    queryKey: [...GROUPS_QUERY_KEY, 'own-count', currentUserId],
    queryFn: async () =>
      assertAssetSuccess(
        await listAssetGroupsPage({
          includeAllOwners: false,
          page: 1,
          pageSize: 1,
          search: '',
        })
      ),
  })
  const assetsQuery = useQuery({
    queryKey: [
      ...ASSETS_QUERY_KEY,
      selectedGroup,
      deferredSearch,
      page,
      isAdmin,
    ],
    queryFn: async () =>
      assertAssetSuccess(
        await listMediaAssets({
          groupId: selectedGroup || undefined,
          search: deferredSearch || undefined,
          page,
          pageSize: ASSET_PAGE_SIZE,
          includeAllOwners: isAdmin,
        })
      ),
    placeholderData: keepPreviousData,
    refetchInterval: 30_000,
  })
  const uploadGroups = useMemo(() => groupsQuery.data ?? [], [groupsQuery.data])
  const hasOwnGroups = (ownGroupsQuery.data?.total ?? 0) > 0
  const assets = assetsQuery.data?.items ?? []
  const totalAssets = assetsQuery.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalAssets / ASSET_PAGE_SIZE))
  const selectedUploadGroup = uploadGroups.some(
    (group) => group.id === selectedGroup
  )
    ? selectedGroup
    : ''
  const invalidateGroups = () =>
    queryClient.invalidateQueries({ queryKey: GROUPS_QUERY_KEY })
  const invalidateAssets = () =>
    queryClient.invalidateQueries({ queryKey: ASSETS_QUERY_KEY })
  const groupMutation = useMutation({
    mutationFn: async (values: AssetGroupFormValues) =>
      assertAssetSuccess(await createAssetGroup(values)),
    onSuccess: async () => {
      await invalidateGroups()
      setGroupDialogOpen(false)
      toast.success(t('Asset group created'))
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })
  const uploadMutation = useMutation({
    mutationFn: async (formData: FormData) => {
      setUploadProgress(0)
      return assertAssetSuccess(
        await uploadMediaAsset(formData, setUploadProgress)
      )
    },
    onSuccess: async () => {
      const wasReupload = reuploadTarget !== null
      setPage(1)
      await invalidateAssets()
      setUploadDialogOpen(false)
      setReuploadTarget(null)
      setUploadProgress(0)
      toast.success(
        wasReupload
          ? t('Replacement uploaded. Use its new asset ID in future requests.')
          : t('Asset uploaded')
      )
    },
    onError: (error) => {
      setUploadProgress(0)
      toast.error(assetErrorMessage(error))
    },
  })
  const downloadMutation = useMutation({
    mutationFn: async (asset: MediaAsset) =>
      assertAssetSuccess(await getMediaAssetPreview(asset.id, 'download')),
    onSuccess: (signedURL) => {
      const link = document.createElement('a')
      link.href = signedURL.url
      link.rel = 'noopener'
      document.body.append(link)
      link.click()
      link.remove()
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })
  const deleteMutation = useMutation({
    mutationFn: async (id: string) =>
      assertAssetSuccess(await deleteMediaAsset(id)),
    onSuccess: async () => {
      await invalidateAssets()
      if (assets.length === 1 && page > 1) setPage(page - 1)
      setDeleteTarget(null)
      toast.success(t('Asset deletion queued'))
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })
  const deleteGroupMutation = useMutation({
    mutationFn: async (id: string) =>
      assertAssetSuccess(await deleteAssetGroup(id)),
    onSuccess: async () => {
      setSelectedGroup('')
      setSelectedGroupInfo(undefined)
      setDeleteGroupDialogOpen(false)
      await Promise.all([invalidateGroups(), invalidateAssets()])
      toast.success(t('Asset group deleted'))
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })

  const loading = assetsQuery.isLoading
  const error = assetsQuery.error
  const canDeleteSelectedGroup = Boolean(
    selectedGroupInfo &&
    !deferredSearch &&
    !assetsQuery.isFetching &&
    !assetsQuery.error &&
    totalAssets === 0
  )

  return (
    <>
      <SectionPageLayout>
        <SectionPageLayout.Title>{t('Asset Library')}</SectionPageLayout.Title>
        <SectionPageLayout.Actions>
          {canDeleteSelectedGroup && (
            <Button
              size='sm'
              variant='destructive'
              disabled={deleteGroupMutation.isPending}
              onClick={() => setDeleteGroupDialogOpen(true)}
            >
              <HugeiconsIcon icon={Delete02Icon} data-icon='inline-start' />
              {t('Delete group')}
            </Button>
          )}
          <Button
            size='sm'
            variant='outline'
            onClick={() => setApiAccessDialogOpen(true)}
          >
            <HugeiconsIcon icon={ShieldKeyIcon} data-icon='inline-start' />
            {t('AK/SK access')}
          </Button>
          <Button
            size='sm'
            variant='outline'
            onClick={() => setGroupDialogOpen(true)}
          >
            <HugeiconsIcon icon={FolderAddIcon} data-icon='inline-start' />
            {t('New group')}
          </Button>
          <Button
            size='sm'
            disabled={!hasOwnGroups}
            onClick={() => setUploadDialogOpen(true)}
          >
            <HugeiconsIcon icon={Upload01Icon} data-icon='inline-start' />
            {t('Upload asset')}
          </Button>
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          <div className='flex flex-col gap-4 pb-4'>
            <Card size='sm'>
              <CardContent className='grid gap-3 py-3 md:grid-cols-[minmax(260px,380px)_minmax(220px,1fr)_auto] md:items-center'>
                <AssetGroupSelect
                  value={selectedGroup}
                  selectedGroup={selectedGroupInfo}
                  isAdmin={isAdmin}
                  onValueChange={(value, group) => {
                    setSelectedGroup(value)
                    setSelectedGroupInfo(group)
                    setPage(1)
                  }}
                />
                <InputGroup>
                  <InputGroupAddon>
                    <HugeiconsIcon icon={Search01Icon} />
                  </InputGroupAddon>
                  <InputGroupInput
                    value={search}
                    aria-label={t('Search assets')}
                    placeholder={t('Search by name, asset ID, or uploader')}
                    onChange={(event) => {
                      setSearch(event.target.value)
                      setPage(1)
                    }}
                  />
                </InputGroup>
                <span className='text-muted-foreground text-xs tabular-nums md:text-right'>
                  {t('{{count}} assets', { count: totalAssets })}
                </span>
              </CardContent>
            </Card>

            <p className='text-muted-foreground text-xs'>
              {t(
                'Copy an asset reference and use it in image_url, video_url, or audio_url fields.'
              )}
            </p>

            {loading && (
              <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5'>
                {Array.from({ length: 8 }, (_, index) => (
                  <Skeleton key={index} className='h-72' />
                ))}
              </div>
            )}
            {error && (
              <div className='text-destructive py-8 text-center text-sm'>
                {assetErrorMessage(error)}
              </div>
            )}
            {!loading && !error && totalAssets === 0 && !deferredSearch && (
              <Empty className='min-h-72 border'>
                <EmptyHeader>
                  <EmptyMedia variant='icon'>
                    <HugeiconsIcon icon={Image01Icon} />
                  </EmptyMedia>
                  <EmptyTitle>{t('No assets yet')}</EmptyTitle>
                  <EmptyDescription>
                    {t(
                      'Create a group, then upload an image, video, or audio file.'
                    )}
                  </EmptyDescription>
                </EmptyHeader>
                <EmptyContent>
                  <Button
                    onClick={() =>
                      hasOwnGroups
                        ? setUploadDialogOpen(true)
                        : setGroupDialogOpen(true)
                    }
                  >
                    {hasOwnGroups ? t('Upload asset') : t('Create asset group')}
                  </Button>
                </EmptyContent>
              </Empty>
            )}
            {!loading &&
              !error &&
              totalAssets === 0 &&
              Boolean(deferredSearch) && (
                <Empty className='min-h-56 border'>
                  <EmptyHeader>
                    <EmptyTitle>{t('No matching assets')}</EmptyTitle>
                    <EmptyDescription>
                      {t('Try another group or search term.')}
                    </EmptyDescription>
                  </EmptyHeader>
                </Empty>
              )}
            {assets.length > 0 && (
              <div
                data-testid='asset-grid'
                data-density='compact'
                className='grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5'
              >
                {assets.map((asset) => (
                  <AssetCard
                    key={asset.id}
                    asset={asset}
                    groupName={`${asset.group_name || t('Unknown group')} · ${t('ID')}: ${asset.group_id}`}
                    onDelete={setDeleteTarget}
                    onPreview={setPreviewTarget}
                    onDownload={(target) => downloadMutation.mutate(target)}
                    downloading={
                      downloadMutation.isPending &&
                      downloadMutation.variables?.id === asset.id
                    }
                    onReupload={
                      asset.owner_user_id === currentUserId
                        ? (target) => {
                            setReuploadTarget(target)
                            setUploadDialogOpen(true)
                          }
                        : undefined
                    }
                  />
                ))}
              </div>
            )}

            {totalPages > 1 && (
              <Pagination aria-label={t('Asset pages')}>
                <PaginationContent>
                  <PaginationItem>
                    <Button
                      size='sm'
                      variant='outline'
                      disabled={page <= 1 || assetsQuery.isFetching}
                      onClick={() => setPage((current) => current - 1)}
                    >
                      {t('Previous')}
                    </Button>
                  </PaginationItem>
                  <PaginationItem>
                    <span
                      className='text-muted-foreground px-3 text-sm tabular-nums'
                      aria-live='polite'
                    >
                      {t('Page {{page}} of {{total}}', {
                        page,
                        total: totalPages,
                      })}
                    </span>
                  </PaginationItem>
                  <PaginationItem>
                    <Button
                      size='sm'
                      variant='outline'
                      disabled={page >= totalPages || assetsQuery.isFetching}
                      onClick={() => setPage((current) => current + 1)}
                    >
                      {t('Next')}
                    </Button>
                  </PaginationItem>
                </PaginationContent>
              </Pagination>
            )}
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <AssetApiAccessDialog
        open={apiAccessDialogOpen}
        onOpenChange={setApiAccessDialogOpen}
      />
      {previewTarget && (
        <AssetPreviewDialog
          asset={previewTarget}
          onOpenChange={(open) => !open && setPreviewTarget(null)}
          onDownload={(target) => downloadMutation.mutate(target)}
          downloading={
            downloadMutation.isPending &&
            downloadMutation.variables?.id === previewTarget.id
          }
        />
      )}
      <AssetGroupDialog
        open={groupDialogOpen}
        onOpenChange={setGroupDialogOpen}
        onSubmit={(values) => groupMutation.mutate(values)}
        pending={groupMutation.isPending}
      />
      <AssetUploadDialog
        open={uploadDialogOpen}
        onOpenChange={(open) => {
          setUploadDialogOpen(open)
          if (!open) setReuploadTarget(null)
        }}
        groups={uploadGroups}
        selectedGroup={reuploadTarget?.group_id || selectedUploadGroup}
        initialName={reuploadTarget?.name}
        replacing={reuploadTarget !== null}
        onSubmit={(values) => uploadMutation.mutate(values)}
        pending={uploadMutation.isPending}
        progress={uploadProgress}
      />
      <AlertDialog
        open={deleteGroupDialogOpen}
        onOpenChange={(open) =>
          !deleteGroupMutation.isPending && setDeleteGroupDialogOpen(open)
        }
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('Delete group')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t(
                'Are you sure you want to delete group "{{name}}"? This action cannot be undone.',
                { name: selectedGroupInfo?.name || '' }
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('Cancel')}</AlertDialogCancel>
            <AlertDialogAction
              variant='destructive'
              disabled={deleteGroupMutation.isPending}
              onClick={() =>
                selectedGroup && deleteGroupMutation.mutate(selectedGroup)
              }
            >
              {t('Delete')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <AlertDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('Delete asset?')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t(
                'The asset and its managed copies will be removed in the background. This action cannot be undone.'
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('Cancel')}</AlertDialogCancel>
            <AlertDialogAction
              variant='destructive'
              onClick={() =>
                deleteTarget && deleteMutation.mutate(deleteTarget.id)
              }
            >
              {t('Delete')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
