import type { Customer } from "./customer"

export type TicketStatus = 'new' | 'open' | 'in_progress' | 'resolved' | 'closed'
export type TicketPriority = 'low' | 'medium' | 'high'

export interface Ticket {
  id: string
  subject: string
  description: string
  status: TicketStatus
  priority: TicketPriority
  customer: Customer
  customerId: string
  assignedToId?: string
  assignedTo?: {
    id: string
    firstName: string
    lastName: string
    email: string
  }
  createdAt: string
  updatedAt: string
}

export interface TicketListResponse {
  tickets: Ticket[]
  total: number
  page: number
  pageSize: number
  hasNextPage: boolean
}

export interface TicketStats {
  new: number
  open: number
  in_progress: number
  resolved: number
  closed: number
  total: number
}

export interface CreateTicketData {
  subject: string
  description: string
  customerId: string
  priority: TicketPriority
  status: TicketStatus
  assignedToId?: string
}

export type UpdateTicketData = Partial<CreateTicketData>