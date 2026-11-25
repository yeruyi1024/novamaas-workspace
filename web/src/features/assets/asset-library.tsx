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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { FolderPlus, Upload } from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from '@/components/ui/empty'
import {
  Pagination,
  PaginationContent,
  PaginationItem,
} from '@/components/ui/pagination'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { WebhookDialog } from '@/features/webhooks/components/webhook-dialog'
import { useIsAdmin } from '@/hooks/use-admin'
import { useDebounce } from '@/hooks/use-debounce'
import { useAuthStore } from '@/stores/auth-store'

import {
  createAssetGroup,
  deleteAssetGroup,
  deleteMediaAsset,
  deleteMediaAssets,
  deleteMediaAssetsByFilter,
  getMediaAssetPreview,
  listAssetGroups,
  listAssetGroupsPage,
  listMediaAssets,
  previewMediaAssets,
  summarizeMediaAssets,
  updateAssetGroup,
  uploadMediaAsset,
} from './api'
import { assertAssetSuccess, assetErrorMessage, assetKey } from './asset-utils'
import { AssetApiAccessDialog } from './components/asset-api-access-dialog'
import { AssetCard } from './components/asset-card'
import { AssetConfirmDialog } from './components/asset-confirm-dialog'
import { AssetCopyDialog } from './components/asset-copy-dialog'
import { AssetDropOverlay } from './components/asset-drop-overlay'
import { AssetEmptyArt } from './components/asset-empty-art'
import { AssetFailedNotice } from './components/asset-failed-notice'
import { AssetFilterBar } from './components/asset-filter-bar'
import { AssetGroupDeleteDialog } from './components/asset-group-delete-dialog'
import {
  AssetGroupDialog,
  type AssetGroupFormValues,
} from './components/asset-group-dialog'
import { AssetGroupHeader } from './components/asset-group-header'
import { AssetLibraryActions } from './components/asset-library-actions'
import { AssetOwnerHeader } from './components/asset-owner-header'
import { AssetPreviewDialog } from './components/asset-preview-dialog'
import { AssetSelectionBar } from './components/asset-selection-bar'
import { AssetSelectionStrip } from './components/asset-selection-strip'
import { AssetUploadDialog } from './components/asset-upload-dialog'
import {
  AssetUserPurgeDialog,
  type InitialPurgeOwner,
} from './components/asset-user-purge-dialog'
import { useAssetSelection } from './hooks/use-asset-selection'
import { useMarqueeSelection } from './hooks/use-marquee-selection'
import type {
  AssetGroup,
  AssetLibraryFilters,
  AssetOwner,
  AssetRef,
  MediaAsset,
} from './types'

const GROUPS_QUERY_KEY = ['asset-library', 'groups'] as const
const ASSETS_QUERY_KEY = ['asset-library', 'assets'] as const
const PREVIEWS_QUERY_KEY = ['asset-library', 'previews'] as const
const OWNERS_QUERY_KEY = ['asset-library', 'owners'] as const
const DEFAULT_ASSET_PAGE_SIZE = 24
const PAGE_SIZE_OPTIONS = [24, 40, 80] as const
const BATCH_DELETE_CHUNK_SIZE = 100
const HIGH_IMPACT_DELETE_THRESHOLD = 10
const SEARCH_DEBOUNCE_MS = 300
const PREVIEW_REFRESH_MARGIN_MS = 60_000
const MAX_PREVIEW_STALE_MS = 24 * 60 * 60 * 1000

type GroupDialogMode = 'create' | 'rename' | null

export function AssetLibrary(props: {
  filters?: AssetLibraryFilters
  onFiltersChange?: (patch: Partial<AssetLibraryFilters>) => void
}) {
  const { t } = useTranslation()
  const isAdmin = useIsAdmin()
  const queryClient = useQueryClient()
  const currentUserId = useAuthStore((state) => state.auth.user?.id ?? 0)
  const [localFilters, setLocalFilters] = useState<AssetLibraryFilters>({})
  const [pageSize, setPageSize] = useState<number>(DEFAULT_ASSET_PAGE_SIZE)
  const filters = props.filters ?? localFilters
  const page = filters.page ?? 1
  const onFiltersChange = props.onFiltersChange
  const changeFilters = useCallback(
    (patch: Partial<AssetLibraryFilters>) => {
      const next = { page: undefined, ...patch }
      if (onFiltersChange) onFiltersChange(next)
      else setLocalFilters((previous) => ({ ...previous, ...next }))
    },
    [onFiltersChange]
  )

  const [searchText, setSearchText] = useState(filters.q ?? '')
  const lastExternalQuery = useRef(filters.q)
  useEffect(() => {
    if (filters.q === lastExternalQuery.current) return
    lastExternalQuery.current = filters.q
    setSearchText(filters.q ?? '')
  }, [filters.q])
  const debouncedSearch = useDebounce(searchText, SEARCH_DEBOUNCE_MS)
  useEffect(() => {
    if (debouncedSearch !== searchText) return
    const q = debouncedSearch.trim() || undefined
    if (q !== filters.q) {
      lastExternalQuery.current = q
      changeFilters({ q })
    }
  }, [changeFilters, debouncedSearch, filters.q, searchText])
  const changeSearch = (value: string) => {
    setSearchText(value)
    if (!value.trim() && filters.q) {
      lastExternalQuery.current = undefined
      changeFilters({ q: undefined })
    }
  }

  const [groupInfo, setGroupInfo] = useState<AssetGroup | undefined>()
  const selectedGroup =
    groupInfo &&
    groupInfo.id === filters.group &&
    (!filters.groupOwner || groupInfo.owner_user_id === filters.groupOwner)
      ? groupInfo
      : undefined
  const [ownerInfo, setOwnerInfo] = useState<AssetOwner | undefined>()
  const changeGroup = (value: string, group?: AssetGroup) => {
    setGroupInfo(group)
    changeFilters({
      group: value || undefined,
      groupOwner: isAdmin && value ? group?.owner_user_id : undefined,
    })
  }
  const changeOwner = (owner?: AssetOwner) => {
    setOwnerInfo(owner)
    setGroupInfo(undefined)
    changeFilters({
      owner: owner?.user_id,
      group: undefined,
      groupOwner: undefined,
    })
  }
  const clearAllFilters = () => {
    lastExternalQuery.current = undefined
    setSearchText('')
    setGroupInfo(undefined)
    setOwnerInfo(undefined)
    changeFilters({
      owner: undefined,
      group: undefined,
      groupOwner: undefined,
      type: undefined,
      status: undefined,
      q: undefined,
    })
  }

  const [apiAccessDialogOpen, setApiAccessDialogOpen] = useState(false)
  const [webhookDialogOpen, setWebhookDialogOpen] = useState(false)
  const [copyDialogOpen, setCopyDialogOpen] = useState(false)
  const [purgeDialogOpen, setPurgeDialogOpen] = useState(false)
  const [purgeInitialOwner, setPurgeInitialOwner] = useState<
    InitialPurgeOwner | undefined
  >()
  const [groupDialogMode, setGroupDialogMode] = useState<GroupDialogMode>(null)
  const [deleteGroupDialogOpen, setDeleteGroupDialogOpen] = useState(false)
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false)
  const [droppedFiles, setDroppedFiles] = useState<File[] | undefined>()
  const [uploadQueue, setUploadQueue] = useState<
    { index: number; total: number } | undefined
  >()
  const [reuploadTarget, setReuploadTarget] = useState<MediaAsset | null>(null)
  const [uploadProgress, setUploadProgress] = useState(0)
  const [previewKey, setPreviewKey] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<MediaAsset | null>(null)
  const [batchDeleteOpen, setBatchDeleteOpen] = useState(false)
  const [clearFailedOpen, setClearFailedOpen] = useState(false)
  const [batchMode, setBatchMode] = useState(false)

  const scope = useMemo(
    () => ({
      groupId: filters.group,
      ownerUserId: isAdmin
        ? (filters.owner ?? (filters.group ? filters.groupOwner : undefined))
        : undefined,
      search: filters.q,
      includeAllOwners: isAdmin,
    }),
    [filters.group, filters.groupOwner, filters.owner, filters.q, isAdmin]
  )
  // Scope key intentionally omits `page` so netdisk selections persist across
  // pages within the same filter scope and reset whenever filters change.
  const scopeKey = useMemo(
    () =>
      JSON.stringify({
        isAdmin,
        owner: filters.owner,
        group: filters.group,
        groupOwner: filters.groupOwner,
        type: filters.type,
        status: filters.status,
        q: filters.q,
      }),
    [
      filters.group,
      filters.groupOwner,
      filters.owner,
      filters.q,
      filters.status,
      filters.type,
      isAdmin,
    ]
  )
  const assetsQuery = useQuery({
    queryKey: [
      ...ASSETS_QUERY_KEY,
      'list',
      scopeKey,
      filters.sort,
      page,
      pageSize,
    ],
    queryFn: async () =>
      assertAssetSuccess(
        await listMediaAssets({
          ...scope,
          type: filters.type,
          status: filters.status,
          sort: filters.sort,
          page,
          pageSize,
        })
      ),
    staleTime: 30_000,
    refetchInterval: (query) => {
      if (query.state.status === 'error') return 30_000
      const reviewing = query.state.data?.items.some(
        (asset) => asset.status === 'processing'
      )
      return reviewing ? 3_000 : false
    },
  })
  const hasProcessingAssets = Boolean(
    assetsQuery.data?.items.some((asset) => asset.status === 'processing')
  )
  const summaryQuery = useQuery({
    queryKey: [...ASSETS_QUERY_KEY, 'summary', scope],
    queryFn: async () => assertAssetSuccess(await summarizeMediaAssets(scope)),
    staleTime: 30_000,
    refetchInterval: (query) => {
      if (query.state.status === 'error') return 30_000
      return hasProcessingAssets ? 6_000 : false
    },
  })
  const assets = useMemo(
    () => assetsQuery.data?.items ?? [],
    [assetsQuery.data?.items]
  )

  // Only invoke the batch preview endpoint for assets that did not receive an
  // inline thumbnail URL from the list response (legacy callers / fallbacks).
  const missingPreviewRefs = useMemo(
    () =>
      assets
        .filter((asset) => !asset.thumbnail_url)
        .map((asset) => ({
          id: asset.id,
          owner_user_id: asset.owner_user_id,
        })),
    [assets]
  )
  const previewsQuery = useQuery({
    queryKey: [
      ...PREVIEWS_QUERY_KEY,
      missingPreviewRefs.map(assetKey).join(','),
    ],
    queryFn: async () =>
      assertAssetSuccess(await previewMediaAssets(missingPreviewRefs)),
    enabled: missingPreviewRefs.length > 0,
    staleTime: (query) => {
      const expiries = Object.values(query.state.data ?? {}).map(
        (preview) => preview.expires_at * 1000
      )
      if (expiries.length === 0) return 0
      return Math.min(
        MAX_PREVIEW_STALE_MS,
        Math.max(
          0,
          Math.min(...expiries) - Date.now() - PREVIEW_REFRESH_MARGIN_MS
        )
      )
    },
  })

  // Preconnect to the object storage origin as soon as the first thumbnail URL
  // is known so subsequent thumbnails avoid repeated TLS handshakes.
  const firstThumbnailOrigin = useMemo(() => {
    for (const asset of assets) {
      const candidate =
        asset.thumbnail_url || previewsQuery.data?.[assetKey(asset)]?.url
      if (!candidate) continue
      try {
        const parsed = new URL(candidate)
        if (parsed.protocol === 'https:' || parsed.protocol === 'http:') {
          return parsed.origin
        }
      } catch {
        continue
      }
    }
    return undefined
  }, [assets, previewsQuery.data])
  useEffect(() => {
    if (!firstThumbnailOrigin || typeof document === 'undefined') return
    const existing = document.head.querySelector(
      `link[rel="preconnect"][href="${firstThumbnailOrigin}"]`
    )
    if (existing) return
    const link = document.createElement('link')
    link.rel = 'preconnect'
    link.href = firstThumbnailOrigin
    link.crossOrigin = 'anonymous'
    document.head.append(link)
  }, [firstThumbnailOrigin])

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
  const uploadGroupsQuery = useQuery({
    queryKey: [...GROUPS_QUERY_KEY, 'upload', currentUserId],
    queryFn: async () => assertAssetSuccess(await listAssetGroups(false)),
    enabled: uploadDialogOpen,
  })
  const groupDetailQuery = useQuery({
    queryKey: [
      ...GROUPS_QUERY_KEY,
      'detail',
      isAdmin,
      scope.ownerUserId,
      filters.group,
    ],
    queryFn: async () =>
      assertAssetSuccess(
        await listAssetGroupsPage({
          includeAllOwners: isAdmin,
          ownerUserId: scope.ownerUserId,
          groupId: filters.group,
          page: 1,
          pageSize: 1,
          search: '',
        })
      ),
    enabled: Boolean(filters.group),
  })
  const ownerGroupsCountQuery = useQuery({
    queryKey: [...GROUPS_QUERY_KEY, 'owner-count', filters.owner],
    queryFn: async () =>
      assertAssetSuccess(
        await listAssetGroupsPage({
          includeAllOwners: true,
          ownerUserId: filters.owner,
          page: 1,
          pageSize: 1,
          search: '',
        })
      ),
    enabled:
      isAdmin &&
      Boolean(filters.owner) &&
      !filters.group &&
      ownerInfo?.user_id !== filters.owner,
  })

  const groupDetail =
    groupDetailQuery.data?.items[0]?.id === filters.group
      ? groupDetailQuery.data?.items[0]
      : undefined
  const totalAssets = assetsQuery.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalAssets / pageSize))
  const hasOwnGroups = (ownGroupsQuery.data?.total ?? 0) > 0
  const uploadGroups = useMemo(
    () => uploadGroupsQuery.data ?? [],
    [uploadGroupsQuery.data]
  )
  const activeGroup = groupDetail ?? selectedGroup
  const groupLabel =
    activeGroup?.name ||
    assets.find((asset) => asset.group_id === filters.group)?.group_name ||
    filters.group
  const ownerLabel =
    (ownerInfo?.user_id === filters.owner ? ownerInfo?.username : undefined) ||
    (activeGroup?.owner_user_id === filters.owner
      ? activeGroup?.owner_name
      : undefined) ||
    assets.find((asset) => asset.owner_user_id === filters.owner)?.owner_name
  const ownerGroupCount =
    ownerInfo?.user_id === filters.owner
      ? ownerInfo?.group_count
      : ownerGroupsCountQuery.data?.total
  const resolvedGroupOwner = activeGroup?.owner_user_id ?? scope.ownerUserId
  const ownSelectedGroup =
    filters.group && (!isAdmin || resolvedGroupOwner === currentUserId)
      ? filters.group
      : ''
  const hasFilters = Boolean(
    filters.owner ||
    filters.group ||
    filters.type ||
    filters.status ||
    filters.q
  )

  useEffect(() => {
    setBatchMode(false)
  }, [scopeKey])

  const selection = useAssetSelection(scopeKey, assets, totalAssets)
  const selectionMode = batchMode || selection.count > 0
  const { selectPage, clear: clearSelection } = selection
  const hasAssets = assets.length > 0

  const anyDialogOpen =
    apiAccessDialogOpen ||
    webhookDialogOpen ||
    copyDialogOpen ||
    purgeDialogOpen ||
    groupDialogMode !== null ||
    deleteGroupDialogOpen ||
    uploadDialogOpen ||
    previewKey !== null ||
    deleteTarget !== null ||
    batchDeleteOpen ||
    clearFailedOpen

  const gridRef = useRef<HTMLDivElement>(null)
  const marqueeBox = useMarqueeSelection(
    gridRef,
    hasAssets && !anyDialogOpen,
    selection.selectKeys
  )

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.defaultPrevented || event.repeat) return
      const target = event.target
      if (
        (target instanceof HTMLElement &&
          target.closest(
            'input, textarea, select, [contenteditable="true"]'
          )) ||
        document.querySelector('[role="dialog"], [role="alertdialog"]')
      ) {
        return
      }
      if (event.key === 'Escape' && selectionMode) {
        clearSelection()
        setBatchMode(false)
      } else if (
        (event.key === 'Delete' || event.key === 'Backspace') &&
        selection.count > 0
      ) {
        event.preventDefault()
        setBatchDeleteOpen(true)
      } else if (
        event.key.toLowerCase() === 'a' &&
        (event.metaKey || event.ctrlKey) &&
        hasAssets
      ) {
        event.preventDefault()
        selectPage()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [clearSelection, hasAssets, selectPage, selection.count, selectionMode])

  const prefetchNextPage = useCallback(() => {
    if (page >= totalPages) return
    const nextPage = page + 1
    void queryClient.prefetchQuery({
      queryKey: [
        ...ASSETS_QUERY_KEY,
        'list',
        scopeKey,
        filters.sort,
        nextPage,
        pageSize,
      ],
      queryFn: async () =>
        assertAssetSuccess(
          await listMediaAssets({
            ...scope,
            type: filters.type,
            status: filters.status,
            sort: filters.sort,
            page: nextPage,
            pageSize,
          })
        ),
      staleTime: 30_000,
    })
  }, [
    filters.sort,
    filters.status,
    filters.type,
    page,
    pageSize,
    queryClient,
    scope,
    scopeKey,
    totalPages,
  ])

  const previewIndex = previewKey
    ? assets.findIndex((asset) => assetKey(asset) === previewKey)
    : -1
  const previewTarget = previewIndex >= 0 ? assets[previewIndex] : null
  const groupAssetCount =
    activeGroup?.asset_count ??
    (filters.q ? undefined : summaryQuery.data?.total)

  const invalidateAll = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: ASSETS_QUERY_KEY }),
      queryClient.invalidateQueries({ queryKey: GROUPS_QUERY_KEY }),
      queryClient.invalidateQueries({ queryKey: OWNERS_QUERY_KEY }),
    ])

  const groupMutation = useMutation({
    mutationFn: async (values: AssetGroupFormValues) => {
      if (groupDialogMode === 'rename' && filters.group) {
        return assertAssetSuccess(
          await updateAssetGroup(
            filters.group,
            values,
            isAdmin ? resolvedGroupOwner : undefined
          )
        )
      }
      return assertAssetSuccess(await createAssetGroup(values))
    },
    onSuccess: async (group) => {
      const renamed = groupDialogMode === 'rename'
      if (renamed) setGroupInfo(group)
      setGroupDialogMode(null)
      await queryClient.invalidateQueries({ queryKey: GROUPS_QUERY_KEY })
      toast.success(
        renamed ? t('Asset group renamed') : t('Asset group created')
      )
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })
  const uploadMutation = useMutation({
    mutationFn: async (forms: FormData[]) => {
      const failures: string[] = []
      let firstError: unknown
      for (const [index, formData] of forms.entries()) {
        setUploadQueue({ index, total: forms.length })
        setUploadProgress(0)
        try {
          assertAssetSuccess(
            await uploadMediaAsset(formData, setUploadProgress)
          )
        } catch (error) {
          firstError ??= error
          const file = formData.get('file')
          failures.push(file instanceof File ? file.name : String(index + 1))
        }
      }
      if (failures.length === forms.length) throw firstError
      return { uploaded: forms.length - failures.length, failures }
    },
    onSuccess: async (result) => {
      const wasReupload = reuploadTarget !== null
      changeFilters({ page: undefined })
      setUploadDialogOpen(false)
      setReuploadTarget(null)
      setDroppedFiles(undefined)
      setUploadProgress(0)
      setUploadQueue(undefined)
      await invalidateAll()
      if (wasReupload) {
        toast.success(
          t('Replacement uploaded. Use its new asset ID in future requests.')
        )
      } else if (result.uploaded === 1 && result.failures.length === 0) {
        toast.success(t('Asset uploaded'))
      } else {
        toast.success(
          t('{{count}} assets uploaded', { count: result.uploaded })
        )
      }
      if (result.failures.length > 0) {
        toast.error(
          t('Failed to upload: {{names}}', {
            names: result.failures.join(', '),
          })
        )
      }
    },
    onError: (error) => {
      setUploadProgress(0)
      setUploadQueue(undefined)
      toast.error(assetErrorMessage(error))
    },
  })
  const downloadMutation = useMutation({
    mutationFn: async (asset: MediaAsset) =>
      assertAssetSuccess(
        isAdmin
          ? await getMediaAssetPreview(
              asset.id,
              'download',
              asset.owner_user_id
            )
          : await getMediaAssetPreview(asset.id, 'download')
      ),
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
  const afterDeletion = async (removed: number, resetToFirstPage = false) => {
    if (resetToFirstPage && page > 1) {
      changeFilters({ page: undefined })
    } else if (removed >= assets.length && page > 1) {
      changeFilters({ page: page - 1 })
    }
    await invalidateAll()
  }
  const deleteMutation = useMutation({
    mutationFn: async (asset: MediaAsset) =>
      assertAssetSuccess(
        isAdmin
          ? await deleteMediaAsset(asset.id, asset.owner_user_id)
          : await deleteMediaAsset(asset.id)
      ),
    onSuccess: async () => {
      setDeleteTarget(null)
      await afterDeletion(1)
      toast.success(t('Asset deletion queued'))
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })
  const batchDeleteMutation = useMutation({
    mutationFn: async (
      payload:
        | { mode: 'filter'; exclude: AssetRef[] }
        | { mode: 'explicit'; targets: MediaAsset[] }
        // Failed assets (rejected by review or failed to synchronize) in the
        // current scope, whatever the type and status filters.
        | { mode: 'failed' }
    ) => {
      if (payload.mode !== 'explicit') {
        const failedOnly = payload.mode === 'failed'
        const res = assertAssetSuccess(
          await deleteMediaAssetsByFilter({
            scope: isAdmin ? 'all' : undefined,
            owner_user_id: scope.ownerUserId,
            group_id: scope.groupId,
            type: failedOnly ? undefined : filters.type,
            status: failedOnly ? 'unavailable' : filters.status,
            search: scope.search,
            exclude:
              payload.mode === 'filter' && payload.exclude.length > 0
                ? payload.exclude
                : undefined,
          })
        )
        return { deleted_count: res.deleted_count, resetToFirstPage: true }
      }
      const refs = payload.targets.map((asset) => ({
        id: asset.id,
        owner_user_id: asset.owner_user_id,
      }))
      let deletedCount = 0
      for (let i = 0; i < refs.length; i += BATCH_DELETE_CHUNK_SIZE) {
        const chunk = refs.slice(i, i + BATCH_DELETE_CHUNK_SIZE)
        try {
          const res = assertAssetSuccess(await deleteMediaAssets(chunk))
          deletedCount += res.deleted_count
          selection.removeKeys(chunk.map(assetKey))
        } catch (error) {
          if (deletedCount > 0) {
            void invalidateAll()
          }
          throw error
        }
      }
      return { deleted_count: deletedCount, resetToFirstPage: false }
    },
    onSuccess: async (result) => {
      setBatchDeleteOpen(false)
      setClearFailedOpen(false)
      setBatchMode(false)
      selection.clear()
      await afterDeletion(result.deleted_count, result.resetToFirstPage)
      toast.success(
        t('{{count}} assets queued for deletion', {
          count: result.deleted_count,
        })
      )
    },
    onError: (error) => toast.error(assetErrorMessage(error)),
  })
  // The summary counts failed assets regardless of type or status filters.
  const failedCount = summaryQuery.data?.by_status.unavailable ?? 0
  const deleteGroupMutation = useMutation({
    mutationFn: async (withAssets: boolean) =>
      assertAssetSuccess(
        await deleteAssetGroup(
          filters.group ?? '',
          isAdmin ? resolvedGroupOwner : undefined,
          {
            withAssets,
          }
        )
      ),
    onSuccess: async () => {
      setDeleteGroupDialogOpen(false)
      setGroupInfo(undefined)
      changeFilters({ group: undefined, groupOwner: undefined })
      await invalidateAll()
      toast.success(t('Asset group deleted'))
    },
    onError: async (error) => {
      toast.error(assetErrorMessage(error))
      await invalidateAll()
    },
  })

  const openUpload = () => {
    setDroppedFiles(undefined)
    setUploadDialogOpen(true)
  }
  const openPurgeDialog = (preselected?: InitialPurgeOwner) => {
    setPurgeInitialOwner(preselected)
    setPurgeDialogOpen(true)
  }
  const openReupload = useCallback((target: MediaAsset) => {
    setPreviewKey(null)
    setReuploadTarget(target)
    setUploadDialogOpen(true)
  }, [])
  const openPreview = useCallback(
    (asset: MediaAsset) => setPreviewKey(assetKey(asset)),
    []
  )
  const download = downloadMutation.mutate
  const copySelectedReferences = async () => {
    const references = selection.assets
      .filter((asset) => asset.status === 'ready')
      .map((asset) => `asset://${asset.id}`)
    try {
      await navigator.clipboard.writeText(references.join('\n'))
      toast.success(
        t('{{count}} asset references copied', { count: references.length })
      )
    } catch {
      toast.error(t('Copy failed'))
    }
  }

  let content = (
    <div className='grid grid-cols-1 gap-3 sm:grid-cols-[repeat(auto-fill,minmax(256px,1fr))] sm:gap-4'>
      {Array.from({ length: 12 }, (_, index) => (
        <div
          key={index}
          className='bg-card/50 ring-border/60 rounded-2xl p-1.5 ring-1'
        >
          <div className='skeleton-shimmer aspect-[4/3] rounded-xl' />
          <div className='flex flex-col gap-2 px-1.5 pt-3 pb-1.5'>
            <div className='skeleton-shimmer h-3.5 w-3/4 rounded' />
            <div className='skeleton-shimmer h-6 w-full rounded-md' />
            <div className='skeleton-shimmer h-3.5 w-2/3 rounded' />
            <div className='skeleton-shimmer h-3 w-1/2 rounded' />
          </div>
        </div>
      ))}
    </div>
  )
  if (assetsQuery.error) {
    content = (
      <div role='alert' className='text-destructive py-12 text-center text-sm'>
        {assetErrorMessage(assetsQuery.error)}
      </div>
    )
  } else if (!assetsQuery.isLoading && totalAssets === 0 && hasFilters) {
    content = (
      <Empty className='min-h-80 rounded-xl border border-dashed'>
        <EmptyHeader>
          <AssetEmptyArt variant='search' />
          <EmptyTitle>{t('No matching assets')}</EmptyTitle>
          <EmptyDescription>
            {t('Try another group or search term.')}
          </EmptyDescription>
        </EmptyHeader>
        <EmptyContent>
          <Button variant='outline' onClick={clearAllFilters}>
            {t('Clear filters')}
          </Button>
        </EmptyContent>
      </Empty>
    )
  } else if (!assetsQuery.isLoading && totalAssets === 0) {
    content = (
      <Empty className='min-h-96 rounded-xl border border-dashed'>
        <EmptyHeader>
          <AssetEmptyArt variant='library' />
          <EmptyTitle>{t('No assets yet')}</EmptyTitle>
          <EmptyDescription>
            {hasOwnGroups
              ? t('Drop files anywhere on this page, or upload them here.')
              : t(
                  'Create a group, then upload an image, video, or audio file.'
                )}
          </EmptyDescription>
        </EmptyHeader>
        <EmptyContent>
          <Button
            onClick={() =>
              hasOwnGroups ? openUpload() : setGroupDialogMode('create')
            }
          >
            {hasOwnGroups ? (
              <Upload
                className='size-4'
                strokeWidth={1.75}
                aria-hidden='true'
              />
            ) : (
              <FolderPlus
                className='size-4'
                strokeWidth={1.75}
                aria-hidden='true'
              />
            )}
            {hasOwnGroups ? t('Upload asset') : t('Create asset group')}
          </Button>
        </EmptyContent>
      </Empty>
    )
  } else if (assets.length > 0) {
    content = (
      <div
        data-testid='asset-grid'
        data-density='gallery'
        className='grid grid-cols-1 gap-3 select-none sm:grid-cols-[repeat(auto-fill,minmax(256px,1fr))] sm:gap-4'
      >
        {assets.map((asset, index) => {
          const key = assetKey(asset)
          const preview = previewsQuery.data?.[key]
          const thumbnailUrl = asset.thumbnail_url || preview?.url
          const posterUrl = asset.poster_url || preview?.poster_url
          let priority: 'high' | 'eager' | 'lazy' = 'lazy'
          if (index < 4) priority = 'high'
          else if (index < 8) priority = 'eager'
          return (
            <AssetCard
              key={key}
              asset={asset}
              isAdmin={isAdmin}
              thumbnailUrl={thumbnailUrl}
              posterUrl={posterUrl}
              priority={priority}
              previewLoading={!thumbnailUrl && previewsQuery.isLoading}
              selected={selection.isSelected(asset)}
              selectionMode={selectionMode}
              canReupload={asset.owner_user_id === currentUserId}
              downloading={
                downloadMutation.isPending &&
                downloadMutation.variables?.id === asset.id &&
                downloadMutation.variables?.owner_user_id ===
                  asset.owner_user_id
              }
              onToggleSelect={selection.toggle}
              onPreview={openPreview}
              onDownload={download}
              onDelete={setDeleteTarget}
              onReupload={openReupload}
            />
          )
        })}
      </div>
    )
  }

  const pageSizeItems = PAGE_SIZE_OPTIONS.map((size) => ({
    value: String(size),
    label: t('{{count}} / page', { count: size }),
  }))

  const activeOwnerForPurge = filters.owner
    ? {
        id: filters.owner,
        username: ownerLabel || `#${filters.owner}`,
      }
    : undefined

  return (
    <>
      <SectionPageLayout>
        <SectionPageLayout.Title>{t('Asset Library')}</SectionPageLayout.Title>
        <SectionPageLayout.Actions>
          <AssetLibraryActions
            isAdmin={isAdmin}
            canUpload={hasOwnGroups}
            onUpload={openUpload}
            onCreateGroup={() => setGroupDialogMode('create')}
            onOpenAccessKeys={() => setApiAccessDialogOpen(true)}
            onOpenWebhooks={() => setWebhookDialogOpen(true)}
            onCopyGroups={() => setCopyDialogOpen(true)}
            onPurgeUser={() => openPurgeDialog(activeOwnerForPurge)}
          />
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          <div
            className={
              selection.count > 0
                ? 'flex flex-col gap-4 pb-16'
                : 'flex flex-col gap-4 pb-4'
            }
          >
            <AssetFilterBar
              filters={filters}
              summary={summaryQuery.data}
              total={totalAssets}
              isAdmin={isAdmin}
              search={searchText}
              selectedGroup={activeGroup}
              selectedGroupLabel={groupLabel}
              ownerLabel={ownerLabel}
              onSearchChange={changeSearch}
              onOwnerChange={changeOwner}
              onGroupChange={changeGroup}
              onClearAll={clearAllFilters}
              onFiltersChange={changeFilters}
            />
            {failedCount > 0 && (
              <AssetFailedNotice
                count={failedCount}
                viewing={filters.status === 'unavailable' && !filters.type}
                pending={batchDeleteMutation.isPending}
                onView={() =>
                  changeFilters({ status: 'unavailable', type: undefined })
                }
                onClear={() => setClearFailedOpen(true)}
              />
            )}
            {isAdmin && filters.owner && !filters.group && (
              <AssetOwnerHeader
                ownerId={filters.owner}
                ownerName={ownerLabel}
                groupCount={ownerGroupCount}
                summary={summaryQuery.data}
                onCopyGroups={() => setCopyDialogOpen(true)}
                onPurgeOwner={() => openPurgeDialog(activeOwnerForPurge)}
              />
            )}
            {filters.group && (
              <AssetGroupHeader
                groupId={filters.group}
                group={activeGroup}
                fallbackName={groupLabel}
                isAdmin={isAdmin}
                canUpload={Boolean(ownSelectedGroup)}
                onUpload={openUpload}
                onRename={() => setGroupDialogMode('rename')}
                onDelete={() => setDeleteGroupDialogOpen(true)}
              />
            )}
            {assets.length > 0 && (
              <AssetSelectionStrip
                pageCount={assets.length}
                totalCount={totalAssets}
                selectedCount={selection.count}
                allOnPageSelected={selection.allOnPageSelected}
                someOnPageSelected={selection.someOnPageSelected}
                allMatching={selection.allMatching}
                canSelectAllMatching={selection.canSelectAllMatching}
                batchMode={batchMode}
                onTogglePage={selection.togglePage}
                onSelectAllMatching={selection.selectAllMatching}
                onClearSelection={() => {
                  selection.clear()
                  setBatchMode(false)
                }}
                onToggleBatchMode={() => {
                  setBatchMode((prev) => {
                    if (prev) selection.clear()
                    return !prev
                  })
                }}
              />
            )}
            <div ref={gridRef} className='-mx-2 px-2 py-1'>
              {content}
            </div>
            {(totalPages > 1 || totalAssets > DEFAULT_ASSET_PAGE_SIZE) && (
              <div className='flex flex-wrap items-center justify-center gap-3'>
                {totalPages > 1 && (
                  <Pagination aria-label={t('Asset pages')} className='w-auto'>
                    <PaginationContent>
                      <PaginationItem>
                        <Button
                          size='sm'
                          variant='outline'
                          disabled={page <= 1 || assetsQuery.isFetching}
                          onClick={() => changeFilters({ page: page - 1 })}
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
                          disabled={
                            page >= totalPages || assetsQuery.isFetching
                          }
                          onMouseEnter={prefetchNextPage}
                          onFocus={prefetchNextPage}
                          onClick={() => changeFilters({ page: page + 1 })}
                        >
                          {t('Next')}
                        </Button>
                      </PaginationItem>
                    </PaginationContent>
                  </Pagination>
                )}
                <Select
                  items={pageSizeItems}
                  value={String(pageSize)}
                  onValueChange={(value) => {
                    const next = Number(value)
                    if (!Number.isFinite(next) || next === pageSize) return
                    setPageSize(next)
                    changeFilters({ page: undefined })
                  }}
                >
                  <SelectTrigger
                    aria-label={t('Page size')}
                    className='h-8 w-28 text-xs'
                  >
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {pageSizeItems.map((item) => (
                      <SelectItem key={item.value} value={item.value}>
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
            <AssetSelectionBar
              count={selection.count}
              totalMatching={totalAssets}
              allOnPageSelected={selection.allOnPageSelected}
              allMatching={selection.allMatching}
              canSelectAllMatching={selection.canSelectAllMatching}
              copyableCount={
                selection.assets.filter((asset) => asset.status === 'ready')
                  .length
              }
              deleting={batchDeleteMutation.isPending}
              onTogglePage={selection.togglePage}
              onSelectAllMatching={selection.selectAllMatching}
              onCopyReferences={copySelectedReferences}
              onDelete={() => setBatchDeleteOpen(true)}
              onClear={() => {
                selection.clear()
                setBatchMode(false)
              }}
            />
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      {marqueeBox && (
        <div
          aria-hidden='true'
          data-testid='asset-marquee-box'
          className='border-primary bg-primary/10 pointer-events-none fixed z-40 rounded-md border'
          style={{
            left: marqueeBox.left,
            top: marqueeBox.top,
            width: marqueeBox.width,
            height: marqueeBox.height,
          }}
        />
      )}

      <AssetDropOverlay
        enabled={hasOwnGroups && !uploadDialogOpen}
        groupName={ownSelectedGroup ? groupLabel : undefined}
        onFiles={(files) => {
          setReuploadTarget(null)
          setDroppedFiles(files)
          setUploadDialogOpen(true)
        }}
      />
      <AssetApiAccessDialog
        open={apiAccessDialogOpen}
        onOpenChange={setApiAccessDialogOpen}
      />
      <WebhookDialog
        open={webhookDialogOpen}
        onOpenChange={setWebhookDialogOpen}
      />
      {previewTarget && (
        <AssetPreviewDialog
          asset={previewTarget}
          isAdmin={isAdmin}
          position={{ index: previewIndex, total: assets.length }}
          onPrevious={
            previewIndex > 0
              ? () => setPreviewKey(assetKey(assets[previewIndex - 1]))
              : undefined
          }
          onNext={
            previewIndex < assets.length - 1
              ? () => setPreviewKey(assetKey(assets[previewIndex + 1]))
              : undefined
          }
          onOpenChange={(open) => !open && setPreviewKey(null)}
          onDownload={download}
          onDelete={(asset) => {
            setPreviewKey(null)
            setDeleteTarget(asset)
          }}
          onReupload={
            previewTarget.owner_user_id === currentUserId
              ? openReupload
              : undefined
          }
          downloading={
            downloadMutation.isPending &&
            downloadMutation.variables?.id === previewTarget.id &&
            downloadMutation.variables?.owner_user_id ===
              previewTarget.owner_user_id
          }
        />
      )}
      <AssetGroupDialog
        open={groupDialogMode !== null}
        onOpenChange={(open) => !open && setGroupDialogMode(null)}
        onSubmit={(values) => groupMutation.mutate(values)}
        pending={groupMutation.isPending}
        initialValues={
          groupDialogMode === 'rename'
            ? {
                name: groupLabel ?? '',
                description: activeGroup?.description ?? '',
              }
            : undefined
        }
      />
      <AssetGroupDeleteDialog
        open={deleteGroupDialogOpen}
        groupName={groupLabel ?? ''}
        assetCount={groupAssetCount}
        pending={deleteGroupMutation.isPending}
        onOpenChange={setDeleteGroupDialogOpen}
        onConfirm={(withAssets) => deleteGroupMutation.mutate(withAssets)}
      />
      {isAdmin && (
        <>
          <AssetCopyDialog
            open={copyDialogOpen}
            onOpenChange={setCopyDialogOpen}
          />
          <AssetUserPurgeDialog
            open={purgeDialogOpen}
            initialOwner={purgeInitialOwner}
            onOpenChange={(open) => {
              setPurgeDialogOpen(open)
              if (!open) setPurgeInitialOwner(undefined)
            }}
            onPurged={(mode) => {
              if (mode === 'all') {
                setGroupInfo(undefined)
                setOwnerInfo(undefined)
                changeFilters({
                  group: undefined,
                  groupOwner: undefined,
                  owner: undefined,
                })
              }
            }}
          />
        </>
      )}
      <AssetUploadDialog
        open={uploadDialogOpen}
        onOpenChange={(open) => {
          setUploadDialogOpen(open)
          if (open) return
          setReuploadTarget(null)
          setDroppedFiles(undefined)
        }}
        groups={uploadGroups}
        selectedGroup={reuploadTarget?.group_id || ownSelectedGroup}
        initialName={reuploadTarget?.name}
        initialFiles={droppedFiles}
        replacing={reuploadTarget !== null}
        onSubmit={(values) => uploadMutation.mutate(values)}
        pending={uploadMutation.isPending}
        progress={uploadProgress}
        queue={uploadQueue}
      />
      <AssetConfirmDialog
        open={deleteTarget !== null}
        title={t('Delete asset?')}
        description={t(
          'The asset and its managed copies will be removed in the background. This action cannot be undone.'
        )}
        pending={deleteMutation.isPending}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        onConfirm={() => {
          if (deleteTarget && !deleteMutation.isPending) {
            deleteMutation.mutate(deleteTarget)
          }
        }}
      />
      <AssetConfirmDialog
        open={batchDeleteOpen}
        title={
          selection.allMatching
            ? t('Delete all {{count}} matching assets?', {
                count: selection.count,
              })
            : t('Delete {{count}} assets?', {
                count: selection.count,
              })
        }
        description={
          selection.allMatching
            ? t(
                'Every asset matching the current filters (excluding deselected items) and its managed copies will be removed in the background. This action cannot be undone.'
              )
            : t(
                'The selected assets and their managed copies will be removed in the background. This action cannot be undone.'
              )
        }
        acknowledgeLabel={
          selection.allMatching ||
          selection.count > HIGH_IMPACT_DELETE_THRESHOLD
            ? t('I confirm permanently deleting these {{count}} assets.', {
                count: selection.count,
              })
            : undefined
        }
        pending={batchDeleteMutation.isPending}
        onOpenChange={setBatchDeleteOpen}
        onConfirm={() => {
          if (batchDeleteMutation.isPending) return
          if (selection.allMatching) {
            batchDeleteMutation.mutate({
              mode: 'filter',
              exclude: selection.excludedRefs,
            })
          } else {
            batchDeleteMutation.mutate({
              mode: 'explicit',
              targets: selection.assets,
            })
          }
        }}
      />
      <AssetConfirmDialog
        open={clearFailedOpen}
        title={t('Clear {{count}} failed assets?', { count: failedCount })}
        description={t(
          'Assets rejected by review or failed to synchronize cannot be used. They and their managed copies will be removed in the background. This action cannot be undone.'
        )}
        acknowledgeLabel={
          isAdmin && !scope.ownerUserId
            ? t('I confirm clearing failed assets of all users in this view.')
            : undefined
        }
        checkboxId='asset-clear-failed-ack'
        pending={batchDeleteMutation.isPending}
        onOpenChange={setClearFailedOpen}
        onConfirm={() => {
          if (!batchDeleteMutation.isPending) {
            batchDeleteMutation.mutate({ mode: 'failed' })
          }
        }}
      />
    </>
  )
}
