
interface TicketStatusBadgeProps {
  status: 'new' | 'open' | 'in_progress' | 'resolved' | 'closed'
}

export default function TicketStatusBadge({ status }: TicketStatusBadgeProps) {
  const getStatusStyle = () => {
    switch (status) {
      case 'new':
        return 'bg-blue-100 text-blue-800'
      case 'open':
        return 'bg-yellow-100 text-yellow-800'
      case 'in_progress':
        return 'bg-purple-100 text-purple-800'
      case 'resolved':
        return 'bg-green-100 text-green-800'
      case 'closed':
        return 'bg-gray-100 text-gray-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getStatusLabel = () => {
    switch (status) {
      case 'in_progress':
        return 'In Progress'
      default:
        return status.charAt(0).toUpperCase() + status.slice(1)
    }
  }

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusStyle()}`}
    >
      {getStatusLabel()}
    </span>
  )
}
