// src/pages/dashboard/DashboardPage.tsx
import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import DashboardLayout from '../../layouts/DashboardLayout'
import Card from '../../components/common/Card'
import { formatDate } from '../../utils/formatters'
import TicketStatusBadge from '../../components/tickets/TicketStatusBadge'
import { ticketService } from '../../services/ticketService'
import { customerService } from '../../services/customerService'


export default function DashboardPage() {
  const { data: ticketsData, isLoading: ticketsLoading } = useQuery({
    queryKey: ['tickets', 'recent'],
    queryFn: () => ticketService.getRecentTickets(),
  })

  const { data: customersData, isLoading: customersLoading } = useQuery({
    queryKey: ['customers', 'count'],
    queryFn: () => customerService.getCustomerStats(),
  })

  const { data: ticketStats, isLoading: statsLoading } = useQuery({
    queryKey: ['tickets', 'stats'],
    queryFn: () => ticketService.getTicketStats(),
  })

  // Prepare chart data
  const chartData = useMemo(() => {
    if (!ticketStats) return []
    return [
      { name: 'New', value: ticketStats.new || 0 },
      { name: 'Open', value: ticketStats.open || 0 },
      { name: 'In Progress', value: ticketStats.in_progress || 0 },
      { name: 'Resolved', value: ticketStats.resolved || 0 },
      { name: 'Closed', value: ticketStats.closed || 0 },
    ]
  }, [ticketStats])

  return (
    <DashboardLayout>
      <h1 className="text-2xl font-semibold text-gray-900">Dashboard</h1>

      {/* Stats Cards */}
      <div className="mt-6 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        <Card className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0 bg-primary-500 rounded-md p-3">
                {/* Icon */}
                <svg
                  className="h-6 w-6 text-white"
                  viewBox="0 0 24 24"
                  fill="none"
                >
                  <path
                    stroke="currentColor"
                    strokeWidth="2"
                    d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                  />
                </svg>
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Total Customers
                  </dt>
                  <dd className="flex items-baseline">
                    <div className="text-2xl font-semibold text-gray-900">
                      {customersLoading
                        ? 'Loading...'
                        : customersData?.total || 0}
                    </div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </Card>

        <Card className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0 bg-primary-500 rounded-md p-3">
                {/* Icon */}
                <svg
                  className="h-6 w-6 text-white"
                  viewBox="0 0 24 24"
                  fill="none"
                >
                  <path
                    stroke="currentColor"
                    strokeWidth="2"
                    d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 012-2h10a2 2 0 012 2v14a2 2 0 01-2 2H7a2 2 0 01-2-2V5z"
                  />
                </svg>
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Active Tickets
                  </dt>
                  <dd className="flex items-baseline">
                    <div className="text-2xl font-semibold text-gray-900">
                      {statsLoading
                        ? 'Loading...'
                        : (ticketStats?.open || 0) +
                          (ticketStats?.in_progress || 0)}
                    </div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </Card>

        {/* More stat cards */}
      </div>

      <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Ticket Status Chart */}
        <Card title="Ticket Status Distribution">
          <div className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                data={chartData}
                margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
              >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Bar dataKey="value" fill="#3b82f6" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </Card>

        {/* Recent Tickets */}
        <Card title="Recent Tickets">
          <div className="flow-root">
            <ul className="-my-5 divide-y divide-gray-200">
              {ticketsLoading ? (
                <p>Loading...</p>
              ) : (
                ticketsData?.tickets?.map((ticket) => (
                  <li key={ticket.id} className="py-4">
                    <div className="flex items-center space-x-4">
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-gray-900 truncate">
                          {ticket.subject}
                        </p>
                        <p className="text-sm text-gray-500 truncate">
                          {formatDate(ticket.createdAt)}
                        </p>
                      </div>
                      <div>
                        <TicketStatusBadge status={ticket.status} />
                      </div>
                    </div>
                  </li>
                ))
              )}
            </ul>
          </div>
        </Card>
      </div>
    </DashboardLayout>
  )
}
