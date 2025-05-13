import React, { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel, // If you implement pagination
  useReactTable,
} from '@tanstack/react-table'
import DashboardLayout from '../../layouts/DashboardLayout'
import Card from '../../components/common/Card'
import Button from '../../components/common/Button'
import { PlusIcon, PencilIcon, TrashIcon } from '@heroicons/react/24/outline'
import { agentService } from '../../services/agentService'
import type { Agent, AgentFormData, CreateAgentData, UpdateAgentData } from '../../types/agent'
import AgentFormModal from '../../components/agents/AgentFormModal' // Import the modal
import { formatDate } from '../../utils/formatters' // Assuming you have this

const columnHelper = createColumnHelper<Agent>()

export default function AgentsPage() {
  const queryClient = useQueryClient()
  const [showAddEditAgentModal, setShowAddEditAgentModal] = useState(false)
  const [editingAgent, setEditingAgent] = useState<Agent | null>(null)
  const [searchTerm, setSearchTerm] = useState('') // For future search implementation

  const { data, isLoading, error /*, refetch */ } = useQuery({
    queryKey: ['agents', { search: searchTerm }],
    queryFn: () => agentService.getAgents({ search: searchTerm }),
  })

  const agentsData = data?.agents || []

  // Create Agent Mutation
  const createAgentMutation = useMutation({
    mutationFn: (newData: CreateAgentData) => agentService.createAgent(newData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] })
      setShowAddEditAgentModal(false)
      // TODO: Add success toast notification
    },
    onError: (err: any) => {
      // TODO: Add error toast notification
      console.error('Failed to create agent', err)
      alert(
        `Error: ${err.response?.data?.message || err.message || 'Could not create agent.'}`
      )
    },
  })

  // Update Agent Mutation
  const updateAgentMutation = useMutation({
    mutationFn: ({ id, updateData }: { id: string; updateData: UpdateAgentData }) =>
      agentService.updateAgent(id, updateData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] })
      setShowAddEditAgentModal(false)
      // TODO: Add success toast notification
    },
    onError: (err: any) => {
      // TODO: Add error toast notification
      console.error('Failed to update agent', err)
      alert(
        `Error: ${err.response?.data?.message || err.message || 'Could not update agent.'}`
      )
    },
  })

  // Delete Agent Mutation
  const deleteMutation = useMutation({
    mutationFn: agentService.deleteAgent,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] })
      // Add toast notification
    },
    onError: (err: any) => {
      // Add toast notification for error
      console.error('Failed to delete agent', err)
      alert(
        `Error: ${err.response?.data?.message || err.message || 'Could not delete agent.'}`
      )
    },
  })

  const handleOpenAddAgentModal = () => {
    setEditingAgent(null)
    setShowAddEditAgentModal(true)
  }

  const handleOpenEditAgentModal = (agent: Agent) => {
    setEditingAgent(agent)
    setShowAddEditAgentModal(true)
  }

  const handleDeleteAgent = (agentId: string) => {
    if (
      window.confirm(
        'Are you sure you want to delete this agent? This action cannot be undone.'
      )
    ) {
      deleteMutation.mutate(agentId)
    }
  }

  const handleFormSubmit = (formData: AgentFormData) => {
    // Remove confirmPassword before sending to API
    const { confirmPassword, ...submissionData } = formData

    if (editingAgent) {
      // For updates, typically password is not sent or handled differently
      const { password, ...updatePayload } = submissionData as UpdateAgentData
      updateAgentMutation.mutate({ id: editingAgent.id, updateData: updatePayload })
    } else {
      // For create, ensure password is included if required by API
      createAgentMutation.mutate(submissionData as CreateAgentData)
    }
  }

  const columns = [
    columnHelper.accessor((row) => `${row.first_name} ${row.last_name}`, {
      id: 'fullName',
      header: 'Name',
      cell: (info) => info.getValue(),
    }),
    columnHelper.accessor('email', {
      header: 'Email',
      cell: (info) => info.getValue(),
    }),
    columnHelper.accessor('department', {
      header: 'Department',
      cell: (info) => info.getValue() || 'N/A',
    }),
    columnHelper.accessor('role', {
      // Assuming role is part of your agent data from API
      header: 'Role',
      cell: (info) => (
        <span
          className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
            info.getValue() === 'admin'
              ? 'bg-red-100 text-red-800'
              : info.getValue() === 'agent'
              ? 'bg-blue-100 text-blue-800'
              : 'bg-gray-100 text-gray-800'
          }`}
        >
          {info.getValue()
            ? info.getValue()!.charAt(0).toUpperCase() +
              info.getValue()!.slice(1)
            : 'N/A'}
        </span>
      ),
    }),
    columnHelper.accessor('status', {
      header: 'Status',
      cell: (info) => (
        <span
          className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
            info.getValue() === 'active'
              ? 'bg-green-100 text-green-800'
              : 'bg-yellow-100 text-yellow-800'
          }`}
        >
          {info.getValue().charAt(0).toUpperCase() + info.getValue().slice(1)}
        </span>
      ),
    }),
    columnHelper.accessor('lastLogin', {
      header: 'Last Login',
      cell: (info) => (info.getValue() ? formatDate(info.getValue()) : 'N/A'),
    }),
    columnHelper.accessor('id', {
      header: 'Actions',
      cell: (info) => (
        <div className="space-x-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleOpenEditAgentModal(info.row.original)}
            aria-label="Edit agent"
          >
            <PencilIcon className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleDeleteAgent(info.getValue())}
            aria-label="Delete agent"
            className="text-red-600 hover:text-red-700"
          >
            <TrashIcon className="h-4 w-4" />
          </Button>
        </div>
      ),
    }),
  ]

  const table = useReactTable({
    data: agentsData,
    columns,
    getCoreRowModel: getCoreRowModel(),
    // getPaginationRowModel: getPaginationRowModel(), // Enable if using pagination
    // manualPagination: true, // If API handles pagination
    // pageCount: data?.totalPages ?? -1, // If API provides total pages
    // state: {
    //   pagination,
    // },
    // onPaginationChange: setPagination,
  })

  const apiError = error as any // Type assertion for error object

  return (
    <DashboardLayout>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">Manage Agents</h1>
        <Button onClick={handleOpenAddAgentModal} variant="default">
          <PlusIcon className="h-5 w-5 mr-2" />
          Add Agent
        </Button>
      </div>

      {/* TODO: Add Search Bar similar to CustomerPage */}

      <Card>
        {isLoading && (
          <div className="p-6 text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500 mx-auto"></div>
            <p className="mt-2 text-sm text-gray-500">Loading agents...</p>
          </div>
        )}
        {apiError && (
          <div className="p-6 bg-red-50 border border-red-200 text-red-700 rounded-md">
            Failed to load agents: {apiError.message || 'Unknown error'}
          </div>
        )}
        {!isLoading && !apiError && (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                {table.getHeaderGroups().map((headerGroup) => (
                  <tr key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <th
                        key={header.id}
                        scope="col"
                        className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
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
              <tbody className="bg-white divide-y divide-gray-200">
                {table.getRowModel().rows.length === 0 ? (
                  <tr>
                    <td
                      colSpan={columns.length}
                      className="px-6 py-12 text-center text-sm text-gray-500"
                    >
                      No agents found.
                    </td>
                  </tr>
                ) : (
                  table.getRowModel().rows.map((row) => (
                    <tr key={row.id}>
                      {row.getVisibleCells().map((cell) => (
                        <td
                          key={cell.id}
                          className="px-6 py-4 whitespace-nowrap text-sm text-gray-700"
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
        )}
        {/* TODO: Add Pagination controls similar to CustomerPage if API supports it */}
      </Card>

      {showAddEditAgentModal && (
        <AgentFormModal
          isOpen={showAddEditAgentModal}
          onClose={() => setShowAddEditAgentModal(false)}
          agentToEdit={editingAgent}
          onSubmit={handleFormSubmit}
          isSubmitting={createAgentMutation.isLoading || updateAgentMutation.isLoading}
        />
      )}
    </DashboardLayout>
  )
}
