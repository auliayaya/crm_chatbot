// src/pages/customers/CustomersPage.tsx
import { useState } from 'react'
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'

import DashboardLayout from '../../layouts/DashboardLayout'

import Button from '../../components/common/Button'
import type { Customer } from '../../types/auth'
import { customerService } from '../../services/customerService'
import { formatDate } from '../../utils/formatters'
import CustomerForm from '../../components/customers/CustomerForm'
import Card from '../../components/common/Card'

const columnHelper = createColumnHelper<Customer>()

export default function CustomersPage() {
  const [showAddCustomer, setShowAddCustomer] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const navigate = useNavigate()
  const [editingCustomerId, setEditingCustomerId] = useState<string | null>(
    null
  )
  const [editingCustomerData, setEditingCustomerData] =
    useState<Customer | null>(null)
  const [pagination, setPagination] = useState({
    pageIndex: 0,
    pageSize: 10,
  })

  // Fetch customers
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['customers', { search: searchTerm }],
    queryFn: () => customerService.getCustomers({ search: searchTerm }),
  })

  const columns = [
    columnHelper.accessor('firstName', {
      header: 'First Name',
      cell: (info) => info.getValue(),
    }),
    columnHelper.accessor('lastName', {
      header: 'Last Name',
      cell: (info) => info.getValue(),
    }),
    columnHelper.accessor('email', {
      header: 'Email',
      cell: (info) => info.getValue(),
    }),
    columnHelper.accessor('companyName', {
      header: 'Company',
      cell: (info) => info.getValue(),
    }),
    columnHelper.accessor('phoneNumber', {
      header: 'Phone',
      cell: (info) => info.getValue(),
    }),
    columnHelper.accessor('status', {
      header: 'Status',
      cell: (info) => (
        <span
          className={`px-2 py-1 text-xs rounded-full ${
            info.getValue() === 'active'
              ? 'bg-green-100 text-green-800'
              : 'bg-red-100 text-red-800'
          }`}
        >
          {info.getValue()}
        </span>
      ),
    }),
    columnHelper.accessor('createdAt', {
      header: 'Created',
      cell: (info) => formatDate(info.getValue()),
    }),
    columnHelper.accessor('id', {
      header: 'Actions',
      cell: (info) => (
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => handleViewDetails(info.getValue())}
          >
            View
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => handleEditCustomer(info.getValue())}
          >
            Edit
          </Button>
        </div>
      ),
    }),
  ]

  const table = useReactTable({
    data: Array.isArray(data) ? data : data?.customers || [],
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    state: {
      pagination,
    },
    onPaginationChange: setPagination,
  })

  const handleViewDetails = (customerId: string) => {
    navigate(`/customers/${customerId}`)
  }

  const handleEditCustomer = (customerId: string) => {
    // Find the customer data
    const customerToEdit = data?.customers.find(
      (customer) => customer.id === customerId
    )

    if (customerToEdit) {
      setEditingCustomerId(customerId)
      setEditingCustomerData(customerToEdit)
      setShowAddCustomer(true) // Reuse the same modal
    }
  }

  const handleCustomerAdded = () => {
    setShowAddCustomer(false)
    setEditingCustomerId(null)
    setEditingCustomerData(null)
    refetch()
  }

  // Debugging logs
  console.log('API Response:', data)
  console.log('Customers array:', data?.customers)
  console.log('Table rows:', table.getRowModel().rows)

  return (
    <DashboardLayout>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">Customers</h1>
        <Button onClick={() => setShowAddCustomer(true)}>Add Customer</Button>
      </div>

      {/* Search Bar */}
      <div className="mb-4">
        <label htmlFor="search" className="sr-only">
          Search
        </label>
        <div className="relative">
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
            <svg
              className="h-5 w-5 text-gray-400"
              viewBox="0 0 20 20"
              fill="currentColor"
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
            placeholder="Search customers..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
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
              ) : !data?.customers || data.customers.length === 0 ? (
                <tr>
                  <td colSpan={8} className="text-center py-4">
                    No customers found
                  </td>
                </tr>
              ) : (
                table.getRowModel().rows.map((row) => (
                  <tr key={row.id}>
                    {row.getVisibleCells().map((cell) => (
                      <td
                        key={cell.id}
                        className="whitespace-nowrap px-3 py-4 text-sm text-gray-500"
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

        {/* Pagination */}
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
                  {table.getState().pagination.pageIndex *
                    table.getState().pagination.pageSize +
                    1}
                </span>{' '}
                to{' '}
                <span className="font-medium">
                  {Math.min(
                    (table.getState().pagination.pageIndex + 1) *
                      table.getState().pagination.pageSize,
                    data?.customers?.length || 0
                  )}
                </span>{' '}
                of{' '}
                <span className="font-medium">
                  {data?.customers?.length || 0}
                </span>{' '}
                results
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
              </nav>
            </div>
          </div>
        </div>
      </Card>

      {/* Customer Form Modal */}
      {showAddCustomer && (
        <div className="fixed inset-0 z-10 overflow-y-auto">
          <div className="flex min-h-screen items-end justify-center px-4 pt-4 pb-20 text-center sm:block sm:p-0">
            <div
              className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"
              onClick={() => {
                setShowAddCustomer(false)
                setEditingCustomerId(null)
                setEditingCustomerData(null)
              }}
            />
            <div className="inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full sm:p-6">
              <div>
                <div className="mt-3 text-center sm:mt-0 sm:text-left">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">
                    {editingCustomerId ? 'Edit Customer' : 'Add New Customer'}
                  </h3>
                  <CustomerForm
                    customerId={editingCustomerId || undefined}
                    initialValues={editingCustomerData || undefined}
                    onSuccess={handleCustomerAdded}
                    onCancel={() => {
                      setShowAddCustomer(false)
                      setEditingCustomerId(null)
                      setEditingCustomerData(null)
                    }}
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
