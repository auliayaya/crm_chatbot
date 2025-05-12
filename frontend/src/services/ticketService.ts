import type {
  CreateTicketData,
  Ticket,
  TicketListResponse,
  TicketStats,
  UpdateTicketData,
} from '../types/ticket'
import { api } from './api'

interface GetTicketsParams {
  search?: string
  status?: string
  priority?: string
  customerId?: string
  assignedToId?: string
  page?: number
  pageSize?: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

export const ticketService = {
  async getTickets({
    search = '',
    status = '',
    limit = 10,
    offset = 0,
    sortBy = 'createdAt',
    sortOrder = 'desc',
  } = {}): Promise<TicketListResponse> {
    try {
      const response = await api.get('/api/crm/tickets', {
        params: {
          search,
          status,
          limit,
          offset,
          sortBy,
          sortOrder,
        },
      })
      return response.data
    } catch (error) {
      console.error('Error fetching tickets:', error)
      throw error
    }
  },

  async getTicketById(id: string): Promise<Ticket> {
    const response = await api.get(`/api/crm/tickets/${id}`)
    return response.data
  },

  async createTicket(data: CreateTicketData): Promise<Ticket> {
    const response = await api.post('/api/crm/tickets', data)
    return response.data
  },

  async updateTicket(id: string, data: UpdateTicketData): Promise<Ticket> {
    const response = await api.put(`/api/crm/tickets/${id}`, data)
    return response.data
  },

  async deleteTicket(id: string): Promise<void> {
    await api.delete(`/api/crm/tickets/${id}`)
  },

  async getTicketStats(): Promise<TicketStats> {
    const response = await api.get('/api/crm/tickets/stats')
    return response.data
  },

  async getRecentTickets(): Promise<TicketListResponse> {
    const response = await api.get('/api/crm/tickets/recent')
    return response.data
  },
}
