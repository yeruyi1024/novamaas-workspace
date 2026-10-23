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
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'

import { pollStudioRuns } from '../api'
import { studioPollInterval } from '../studio-utils'
import type { StudioRun } from '../types'

// One owner-scoped query retains its interval and 429 backoff as tasks change.
export function useStudioPolling(userId: number, runs?: StudioRun[]) {
  const client = useQueryClient()
  const [pendingIds, setPendingIds] = useState<string[]>([])
  useEffect(() => {
    const discovered =
      runs?.filter((run) => run.status === 'running').map((run) => run.id) ?? []
    if (!discovered.length) return
    setPendingIds((ids) =>
      discovered.some((id) => !ids.includes(id))
        ? [...new Set([...ids, ...discovered])]
        : ids
    )
  }, [runs])
  const pending = useQuery({
    queryKey: ['image-studio', 'pending', userId],
    queryFn: async () => ({
      ...(await pollStudioRuns(pendingIds)),
      ids: pendingIds,
    }),
    enabled: userId > 0 && pendingIds.length > 0,
    retry: false,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    refetchInterval: (query) =>
      studioPollInterval(query.state.error, Math.ceil(pendingIds.length / 50)),
  })
  useEffect(() => {
    if (!pending.data) return
    const running = new Set(
      pending.data.items
        .filter((run) => run.status === 'running')
        .map((run) => run.id)
    )
    const finished = new Set(pending.data.ids.filter((id) => !running.has(id)))
    for (const run of pending.data.items) {
      client.setQueryData(['image-studio', 'run', userId, run.id], run)
    }
    if (!finished.size) return
    setPendingIds((ids) => ids.filter((id) => !finished.has(id)))
    void client.invalidateQueries({
      queryKey: ['image-studio', 'runs', userId],
    })
  }, [pending.data, client, userId])

  const track = (run: StudioRun) => {
    if (run.status !== 'running') return
    setPendingIds((ids) => (ids.includes(run.id) ? ids : [...ids, run.id]))
  }
  return { track, runningCount: pendingIds.length }
}
