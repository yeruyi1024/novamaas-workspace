import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import {
  getAssetRequestLogDetail,
  listAssetRequestLogs,
} from '@/features/assets/api'
import type { AssetRequestLog } from '@/features/assets/types'

import { AssetRequestLogDialog } from '../asset-request-log-dialog'

vi.mock('@/features/assets/api', () => ({
  listAssetRequestLogs: vi.fn(),
  getAssetRequestLogDetail: vi.fn(),
}))

let queryClient: QueryClient

beforeEach(() => {
  vi.clearAllMocks()
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
})

afterEach(() => queryClient.clear())

function renderDialog() {
  render(
    <QueryClientProvider client={queryClient}>
      <AssetRequestLogDialog
        open
        channelId={7}
        replicaId={91}
        onOpenChange={vi.fn()}
      />
    </QueryClientProvider>
  )
}

function requestLog(id: number, operation: string): AssetRequestLog {
  return {
    id,
    created_at: 1_790_000_000_000,
    channel_id: 7,
    replica_id: 91,
    asset_id: 123,
    request_id: `request-${id}`,
    source: 'sync',
    protocol: 'volc_action',
    operation,
    method: 'POST',
    path_template: '/',
    http_status: 502,
    duration_ms: 85,
    request_bytes: 120,
    response_bytes: 80,
    result: 'failure',
    error_kind: 'upstream',
  }
}

test('lists requests for one synchronization job and follows the next cursor', async () => {
  vi.mocked(listAssetRequestLogs).mockImplementation(async (params) => ({
    success: true,
    data: params.cursor
      ? {
          items: [requestLog(1, 'CreateAsset')],
          next_cursor: '',
          dropped_on_this_node: 0,
        }
      : {
          items: [requestLog(2, 'GetAsset')],
          next_cursor: '1790000000000:2',
          dropped_on_this_node: 3,
        },
  }))
  renderDialog()

  expect(await screen.findByText('GetAsset')).toBeVisible()
  expect(
    screen.getByText('Some request logs were dropped on this node: 3')
  ).toBeVisible()
  expect(listAssetRequestLogs).toHaveBeenCalledWith(
    expect.objectContaining({ channelId: 7, replicaId: 91, pageSize: 20 })
  )
  fireEvent.click(screen.getByRole('button', { name: 'Next' }))
  expect(await screen.findByText('CreateAsset')).toBeVisible()
  expect(listAssetRequestLogs).toHaveBeenCalledWith(
    expect.objectContaining({ cursor: '1790000000000:2' })
  )
  fireEvent.change(screen.getByRole('combobox', { name: 'Source' }), {
    target: { value: 'test' },
  })
  await waitFor(() =>
    expect(listAssetRequestLogs).toHaveBeenCalledWith(
      expect.objectContaining({ source: 'test', cursor: '' })
    )
  )
  fireEvent.change(screen.getByRole('combobox', { name: 'Time' }), {
    target: { value: '7' },
  })
  await waitFor(() => {
    const call = vi.mocked(listAssetRequestLogs).mock.lastCall?.[0]
    expect((call?.endMS ?? 0) - (call?.startMS ?? 0)).toBe(
      7 * 24 * 60 * 60 * 1000
    )
  })
})

test('distinguishes a failed log query from an empty log page', async () => {
  vi.mocked(listAssetRequestLogs).mockRejectedValue(
    new Error('log database unavailable')
  )
  renderDialog()

  expect(await screen.findByText('Unable to load request logs')).toBeVisible()
  expect(screen.getByText('log database unavailable')).toBeVisible()
  expect(screen.queryByText('No logs')).not.toBeInTheDocument()
})

test('loads sanitized request and response details only when expanded', async () => {
  vi.mocked(listAssetRequestLogs).mockResolvedValue({
    success: true,
    data: {
      items: [requestLog(2, 'ListAssetGroups')],
      next_cursor: '',
      dropped_on_this_node: 0,
    },
  })
  vi.mocked(getAssetRequestLogDetail).mockResolvedValue({
    success: true,
    data: {
      log: requestLog(2, 'ListAssetGroups'),
      detail: {
        log_id: 2,
        request_url: 'https://assets.example.com/?Action=ListAssetGroups',
        request_body: '{"PageNumber":1}',
        response_body: '{"error":"method not allowed"}',
        request_body_omitted: false,
        response_body_omitted: false,
      },
    },
  })
  renderDialog()
  expect(await screen.findByText('ListAssetGroups')).toBeVisible()
  expect(getAssetRequestLogDetail).not.toHaveBeenCalled()
  fireEvent.click(screen.getByRole('button', { name: 'Show details' }))
  expect(
    await screen.findByText(
      'https://assets.example.com/?Action=ListAssetGroups'
    )
  ).toBeVisible()
  expect(screen.getByText('{"PageNumber":1}')).toBeVisible()
  expect(screen.getByText('{"error":"method not allowed"}')).toBeVisible()
  expect(screen.getAllByText('502').length).toBeGreaterThan(1)
  expect(getAssetRequestLogDetail).toHaveBeenCalledWith(2)
  fireEvent.click(screen.getByRole('button', { name: 'Hide details' }))
  expect(screen.queryByText('{"PageNumber":1}')).not.toBeInTheDocument()
})

test('explains when a legacy request log has no captured detail', async () => {
  const log = requestLog(3, 'ListAssetGroups')
  vi.mocked(listAssetRequestLogs).mockResolvedValue({
    success: true,
    data: { items: [log], next_cursor: '', dropped_on_this_node: 0 },
  })
  vi.mocked(getAssetRequestLogDetail).mockResolvedValue({
    success: true,
    data: { log, detail: null },
  })
  renderDialog()
  fireEvent.click(await screen.findByRole('button', { name: 'Show details' }))
  expect(
    await screen.findByText('Details unavailable for this log')
  ).toBeVisible()
})
