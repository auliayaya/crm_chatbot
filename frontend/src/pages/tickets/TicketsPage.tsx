import { useEffect, useState, useCallback } from 'react'
import debounce from 'lodash/debounce'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  useReactTable,
} from '@tanstack/react-table'
import DashboardLayout from '../../layouts/DashboardLayout'
import Card from '../../components/common/Card'
import Button from '../../components/common/Button'
import { ticketService } from '../../services/ticketService'
import { formatDate } from '../../utils/formatters'
import TicketStatusBadge from '../../components/tickets/TicketStatusBadge'
import TicketForm from '../../components/tickets/TicketForm'
import type { Ticket } from '../../types/ticket'

const columnHelper = createColumnHelper<Ticket>()

export default function TicketsPage() {
  const navigate = useNavigate()
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [showAddTicket, setShowAddTicket] = useState(false)
  const [pagination, setPagination] = useState({
    pageIndex: 0,
    pageSize: 10,
  })
  const [sortBy, setSortBy] = useState('createdAt')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')

  // Convert pageIndex (0-based) to page (1-based) and calculate offset
  const page = pagination.pageIndex + 1
  const offset = pagination.pageIndex * pagination.pageSize

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: [
      'tickets',
      searchTerm,
      statusFilter,
      pagination.pageIndex,
      pagination.pageSize,
      sortBy,
      sortOrder,
    ],
    queryFn: () =>
      ticketService.getTickets({
        search: searchTerm,
        status: statusFilter || undefined,
        limit: pagination.pageSize,
        offset: offset,
        sortBy,
        sortOrder,
      }),
  })

  // Calculate if there's a next page
  const hasNextPage = data
    ? (pagination.pageIndex + 1) * pagination.pageSize < data.total
    : false

  const deleteMutation = useMutation({
    mutationFn: (ticketId: string) => ticketService.deleteTicket(ticketId),
    onSuccess: () => refetch(),
  })

  const handleDelete = (ticketId: string) => {
    if (window.confirm('Are you sure you want to delete this ticket?')) {
      deleteMutation.mutate(ticketId)
    }
  }

  const columns = [
    columnHelper.accessor('id', {
      header: 'ID',
      cell: (info) => (
        <span className="text-gray-500 text-sm">{`#${info
          .getValue()
          .slice(-6)}`}</span>
      ),
    }),
    columnHelper.accessor('subject', {
      header: 'Subject',
      cell: (info) => (
        <div className="font-medium text-gray-900 truncate max-w-xs">
          {info.getValue()}
        </div>
      ),
    }),
    columnHelper.accessor('customer', {
      header: 'Customer',
      cell: (info) => (
        <div className="truncate max-w-xs">
          {info.getValue()?.firstName} {info.getValue()?.lastName}
        </div>
      ),
    }),
    columnHelper.accessor('status', {
      header: 'Status',
      cell: (info) => <TicketStatusBadge status={info.getValue()} />,
    }),
    columnHelper.accessor('priority', {
      header: 'Priority',
      cell: (info) => {
        const priority = info.getValue()
        const colors = {
          low: 'bg-blue-100 text-blue-800',
          medium: 'bg-yellow-100 text-yellow-800',
          high: 'bg-red-100 text-red-800',
        }
        return (
          <span
            className={`px-2 py-1 text-xs font-medium rounded-full ${colors[priority]}`}
          >
            {priority.charAt(0).toUpperCase() + priority.slice(1)}
          </span>
        )
      },
    }),
    columnHelper.accessor('createdAt', {
      header: 'Created',
      cell: (info) => formatDate(info.getValue()),
    }),
    columnHelper.accessor('updatedAt', {
      header: 'Updated',
      cell: (info) => formatDate(info.getValue()),
    }),
    columnHelper.display({
      id: 'actions',
      header: 'Actions',
      cell: (info) => (
        <div className="flex space-x-2">
          <Button
            variant="default"
            size="sm"
            onClick={() => handleViewDetails(info.row.original.id)}
          >
            View
          </Button>
          <Button
            variant="default"
            size="sm"
            onClick={() => handleDelete(info.row.original.id)}
          >
            Delete
          </Button>
        </div>
      ),
    }),
  ]

  const table = useReactTable({
    data: data && data.tickets ? data.tickets : [],
    columns,
    pageCount: data?.total ? Math.ceil(data.total / pagination.pageSize) : 0,
    state: {
      pagination,
    },
    onPaginationChange: setPagination,
    manualPagination: true, // Important: Tell the table we're handling pagination manually
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
  })

  useEffect(() => {
    if (data?.total) {
      table.setPageCount(Math.ceil(data.total / pagination.pageSize))
    }
  }, [data?.total, pagination.pageSize, table])

  const handleViewDetails = (ticketId: string) => {
    navigate(`/tickets/${ticketId}`)
  }

  const handleTicketAdded = () => {
    setShowAddTicket(false)
    refetch()
  }

  // Replace your search input handler
  const handleSearchChange = useCallback(
    debounce((value: string) => {
      setSearchTerm(value)
    }, 500),
    []
  )

  // Add before return statement
  console.log('API Response:', data)
  console.log('Pagination state:', pagination)
  console.log('Table data:', table.getRowModel().rows)

  return (
    <DashboardLayout>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">Tickets</h1>
        <Button onClick={() => setShowAddTicket(true)}>Create Ticket</Button>
      </div>

      <div className="mb-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg
              className="h-5 w-5 text-gray-400"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fillRule="evenodd"
                d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
                clipRule="evenodd"
              />
            </svg>
          </div>
          <input
            type="text"
            id="search"
            name="search"
            className="block w-full rounded-md border border-gray-300 bg-white py-2 pl-10 pr-3 text-sm placeholder-gray-500 focus:border-primary-500 focus:text-gray-900 focus:placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-primary-500 sm:text-sm"
            placeholder="Search tickets..."
            defaultValue={searchTerm}
            onChange={(e) => handleSearchChange(e.target.value)}
          />
        </div>

        <select
          id="statusFilter"
          name="statusFilter"
          className="block w-full rounded-md border border-gray-300 bg-white py-2 px-3 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
        >
          <option value="">All Statuses</option>
          <option value="new">New</option>
          <option value="open">Open</option>
          <option value="in_progress">In Progress</option>
          <option value="resolved">Resolved</option>
          <option value="closed">Closed</option>
        </select>
      </div>

      <Card>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-300">
            <thead>
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <th
                      key={header.id}
                      scope="col"
                      className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody className="divide-y divide-gray-200">
              {isLoading ? (
                <tr>
                  <td colSpan={8} className="text-center py-4">
                    Loading...
                  </td>
                </tr>
              ) : isError ? (
                <tr>
                  <td colSpan={8} className="text-center py-4">
                    <div className="bg-red-50 border border-red-200 text-red-800 p-4 rounded-md mb-4">
                      An error occurred:{' '}
                      {error instanceof Error ? error.message : 'Unknown error'}
                    </div>
                  </td>
                </tr>
              ) : table.getRowModel().rows.length === 0 ? (
                <tr>
                  <td colSpan={8} className="text-center py-4">
                    No tickets found
                  </td>
                </tr>
              ) : (
                table.getRowModel().rows.map((row) => (
                  <tr
                    key={row.id}
                    className="hover:bg-gray-50 cursor-pointer"
                    onClick={() => handleViewDetails(row.original.id)}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <td
                        key={cell.id}
                        className="whitespace-nowrap px-3 py-4 text-sm text-gray-500"
                        onClick={(e) => {
                          // Prevent navigation when clicking action buttons
                          if (cell.column.id === 'actions') {
                            e.stopPropagation()
                          }
                        }}
                      >
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </td>
                    ))}
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Updated Pagination */}
        <div className="flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 sm:px-6">
          <div className="flex flex-1 justify-between sm:hidden">
            <Button
              variant="outline"
              size="sm"
              onClick={() => table.previousPage()}
              disabled={!table.getCanPreviousPage()}
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => table.nextPage()}
              disabled={!table.getCanNextPage()}
            >
              Next
            </Button>
          </div>
          <div className="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
            <div>
              <p className="text-sm text-gray-700">
                Showing{' '}
                <span className="font-medium">
                  {pagination.pageIndex * pagination.pageSize + 1}
                </span>{' '}
                to{' '}
                <span className="font-medium">
                  {Math.min(
                    (pagination.pageIndex + 1) * pagination.pageSize,
                    data?.total || 0
                  )}
                </span>{' '}
                of <span className="font-medium">{data?.total}</span> results
              </p>
            </div>
            <div>
              <nav
                className="isolate inline-flex -space-x-px rounded-md shadow-sm"
                aria-label="Pagination"
              >
                <Button
                  variant="outline"
                  size="sm"
                  className="rounded-l-md"
                  onClick={() => table.previousPage()}
                  disabled={!table.getCanPreviousPage()}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="rounded-r-md"
                  onClick={() => table.nextPage()}
                  disabled={!table.getCanNextPage()}
                >
                  Next
                </Button>
              </nav>
            </div>
          </div>
        </div>
      </Card>

      {/* Add Ticket Modal */}
      {showAddTicket && (
        <div className="fixed inset-0 z-10 overflow-y-auto">
          <div className="flex min-h-screen items-end justify-center px-4 pt-4 pb-20 text-center sm:block sm:p-0">
            <div
              className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"
              onClick={() => setShowAddTicket(false)}
            />
            <div className="inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full sm:p-6">
              <div>
                <div className="mt-3 text-center sm:mt-0 sm:text-left">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">
                    Create New Ticket
                  </h3>
                  <TicketForm
                    onSuccess={handleTicketAdded}
                    onCancel={() => setShowAddTicket(false)}
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </DashboardLayout>
  )
}
