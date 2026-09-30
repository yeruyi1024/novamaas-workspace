import {
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { render, screen } from '@testing-library/react'
import i18next from 'i18next'
import { afterEach, beforeAll, describe, expect, test, vi } from 'vitest'

import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import { usageLogSchema, type UsageLog } from '../../../data/schema'
import { UsageLogsProvider } from '../../usage-logs-provider'
import { useCommonLogsColumns } from '../common-logs-columns'

vi.mock('@/features/usage-logs/components/dialogs/details-dialog', () => ({
  DetailsDialog: () => null,
}))
vi.mock('@/features/usage-logs/components/model-badge', () => ({
  ModelBadge: () => null,
}))

const log = usageLogSchema.parse({
  id: 1,
  user_id: 7,
  created_at: 1,
  type: 2,
  content: '',
  revenue_quota: 1000,
  cost_quota: 800,
  profit_quota: 200,
})

function FinancialRow(props: { log: UsageLog }) {
  const columns = useCommonLogsColumns(true).filter((column) =>
    ['revenue_quota', 'cost_quota', 'profit_quota'].includes(
      'accessorKey' in column ? String(column.accessorKey) : ''
    )
  )
  const table = useReactTable({
    data: [props.log],
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  return (
    <table>
      <thead>
        <tr>
          {table.getHeaderGroups()[0]?.headers.map((header) => (
            <th key={header.id}>
              {flexRender(header.column.columnDef.header, header.getContext())}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {table.getRowModel().rows.map((row) => (
          <tr key={row.id}>
            {row.getVisibleCells().map((cell) => (
              <td key={cell.id}>
                {flexRender(cell.column.columnDef.cell, cell.getContext())}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  )
}

describe('common log financial rows', () => {
  beforeAll(async () => {
    await i18next.changeLanguage('en')
  })

  afterEach(() => {
    useAuthStore.getState().auth.setUser(null)
  })

  test('financial viewer sees revenue, cost and profit for a linked log', () => {
    useAuthStore.getState().auth.setUser({
      id: 1,
      username: 'finance-admin',
      role: ROLE.ADMIN,
      permissions: {
        admin_permissions: { financial_accounting: { view: true } },
      },
    })

    render(
      <UsageLogsProvider>
        <FinancialRow log={log} />
      </UsageLogsProvider>
    )

    expect(
      screen.getByRole('columnheader', { name: 'Turnover' })
    ).toBeInTheDocument()
    expect(
      screen.getByRole('columnheader', { name: 'Cost amount' })
    ).toBeInTheDocument()
    expect(
      screen.getByRole('columnheader', { name: 'Profit amount' })
    ).toBeInTheDocument()
    expect(screen.getByRole('cell', { name: '$0.002' })).toBeInTheDocument()
    expect(screen.getByRole('cell', { name: '$0.0016' })).toBeInTheDocument()
    expect(screen.getByRole('cell', { name: '$0.0004' })).toBeInTheDocument()
  })

  test('an administrator without financial permission has no financial columns', () => {
    useAuthStore.getState().auth.setUser({
      id: 2,
      username: 'operations-admin',
      role: ROLE.ADMIN,
      permissions: {
        admin_permissions: { financial_accounting: { view: false } },
      },
    })

    render(
      <UsageLogsProvider>
        <FinancialRow log={log} />
      </UsageLogsProvider>
    )

    expect(
      screen.queryByRole('columnheader', { name: 'Turnover' })
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('cell', { name: '$0.002' })
    ).not.toBeInTheDocument()
  })
})
