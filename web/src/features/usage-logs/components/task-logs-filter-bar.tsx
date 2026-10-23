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
import { useIsFetching } from '@tanstack/react-query'
import { useNavigate, getRouteApi } from '@tanstack/react-router'
import type { Table } from '@tanstack/react-table'
import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'

import { buildSearchParams } from '../lib/filter'
import { getDefaultTimeRange } from '../lib/utils'
import type { LogCategory, TaskLogFilters } from '../types'
import { CompactDateTimeRangePicker } from './compact-date-time-range-picker'
import {
  LogsFilterField,
  LogsFilterInput,
  LogsFilterToolbar,
} from './logs-filter-toolbar'
import { useLogsViewScope } from './usage-logs-provider'

const route = getRouteApi('/_authenticated/usage-logs/$section')

interface TaskLogsFilterBarProps<TData> {
  table: Table<TData>
  logCategory: Extract<LogCategory, 'drawing' | 'task'>
}

export function TaskLogsFilterBar<TData>(props: TaskLogsFilterBarProps<TData>) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const searchParams = route.useSearch()
  const { isAdminView: isAdmin } = useLogsViewScope()
  const fetchingLogs = useIsFetching({ queryKey: ['logs'] })
  const isTask = props.logCategory === 'task'
  const isStudio =
    props.logCategory === 'drawing' &&
    (searchParams.drawingSource ?? 'studio') === 'studio'

  const [filters, setFilters] = useState<TaskLogFilters>(() => {
    const { start, end } = getDefaultTimeRange()
    return { startTime: start, endTime: end }
  })

  useEffect(() => {
    const { start, end } = getDefaultTimeRange()
    setFilters({
      startTime: searchParams.startTime
        ? new Date(searchParams.startTime)
        : start,
      endTime: searchParams.endTime ? new Date(searchParams.endTime) : end,
      channel: searchParams.channel ? String(searchParams.channel) : '',
      taskId: searchParams.filter || '',
      model: isTask ? searchParams.model || '' : '',
      username: isTask && isAdmin ? searchParams.username || '' : '',
    })
  }, [
    isTask,
    searchParams.startTime,
    searchParams.endTime,
    searchParams.channel,
    searchParams.filter,
    searchParams.model,
    searchParams.username,
    isAdmin,
  ])

  const handleChange = useCallback(
    (field: keyof TaskLogFilters, value: Date | string | undefined) => {
      setFilters((prev) => ({ ...prev, [field]: value }))
    },
    []
  )

  const applyFilters = useCallback(
    (next: TaskLogFilters) => {
      // Both views edit one task ID; the legacy drawing filter calls it mjId.
      const search = buildSearchParams(
        { ...next, mjId: next.taskId },
        props.logCategory
      )
      void navigate({
        to: '/usage-logs/$section',
        params: { section: props.logCategory },
        search: {
          ...search,
          drawingSource: searchParams.drawingSource,
          page: 1,
        },
      })
    },
    [navigate, props.logCategory, searchParams.drawingSource]
  )

  const handleReset = useCallback(() => {
    const { start, end } = getDefaultTimeRange()
    const next = { startTime: start, endTime: end }
    setFilters(next)
    applyFilters(next)
  }, [applyFilters])

  const handleKeyDown = useCallback(
    (event: React.KeyboardEvent) => {
      if (event.key === 'Enter') applyFilters(filters)
    },
    [applyFilters, filters]
  )

  let placeholder = t('Filter by task ID')
  if (isStudio) {
    placeholder = t('Filter by task or request ID')
  } else if (!isTask) {
    placeholder = t('Filter by MjProxy task ID')
  }
  const filterCount = [
    filters.taskId,
    filters.channel,
    isTask ? filters.model : undefined,
    isTask && isAdmin ? filters.username : undefined,
  ].filter(Boolean).length
  const dateRangeFilter = (
    <LogsFilterField wide>
      <CompactDateTimeRangePicker
        start={filters.startTime}
        end={filters.endTime}
        onChange={({ start, end }) => {
          handleChange('startTime', start)
          handleChange('endTime', end)
        }}
      />
    </LogsFilterField>
  )
  const taskIdFilter = (
    <LogsFilterField>
      <LogsFilterInput
        aria-label={t('Task ID')}
        placeholder={placeholder}
        value={filters.taskId || ''}
        onChange={(event) => handleChange('taskId', event.target.value)}
        onKeyDown={handleKeyDown}
      />
    </LogsFilterField>
  )
  const channelFilter = isAdmin ? (
    <LogsFilterField>
      <LogsFilterInput
        placeholder={t('Channel ID')}
        value={filters.channel || ''}
        onChange={(event) => handleChange('channel', event.target.value)}
        onKeyDown={handleKeyDown}
      />
    </LogsFilterField>
  ) : null
  const modelFilter = isTask ? (
    <LogsFilterField>
      <LogsFilterInput
        aria-label={t('Model name')}
        placeholder={t('Model name')}
        value={filters.model || ''}
        onChange={(event) => handleChange('model', event.target.value)}
        onKeyDown={handleKeyDown}
      />
    </LogsFilterField>
  ) : null
  const usernameFilter =
    isTask && isAdmin ? (
      <LogsFilterField>
        <LogsFilterInput
          aria-label={t('Username')}
          placeholder={t('Filter by username')}
          value={filters.username || ''}
          onChange={(event) => handleChange('username', event.target.value)}
          onKeyDown={handleKeyDown}
        />
      </LogsFilterField>
    ) : null

  return (
    <LogsFilterToolbar
      table={props.table}
      primaryFilters={
        <>
          {dateRangeFilter}
          {taskIdFilter}
          {modelFilter}
          {usernameFilter}
          {channelFilter}
        </>
      }
      mobilePinnedFilters={dateRangeFilter}
      mobileFilters={
        <>
          {taskIdFilter}
          {modelFilter}
          {usernameFilter}
          {channelFilter}
        </>
      }
      mobileFilterCount={filterCount}
      hasActiveFilters={filterCount > 0}
      onSearch={() => applyFilters(filters)}
      searchLoading={fetchingLogs > 0}
      onReset={handleReset}
    />
  )
}
