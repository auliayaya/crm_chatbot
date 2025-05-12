import { api } from './api'
import type {
  Customer,
  CustomerListResponse,
  CreateCustomerData,
  UpdateCustomerData,
} from '../types/customer'

interface GetCustomersParams {
  search?: string
  page?: number
  pageSize?: number
  status?: 'active' | 'inactive'
}

export const customerService = {
  async getCustomers(
    params?: GetCustomersParams
  ): Promise<CustomerListResponse> {
    const response = await api.get('/api/crm/customers', { params })
    console.log('getCustomers response', response.data)
    return response.data
  },

  async getCustomerById(id: string): Promise<Customer> {
    const response = await api.get(`/api/crm/customers/${id}`)
    return response.data
  },

  async createCustomer(data: CreateCustomerData): Promise<Customer> {
    const response = await api.post('/api/crm/customers', data)
    return response.data
  },

  async updateCustomer(
    id: string,
    data: UpdateCustomerData
  ): Promise<Customer> {
    const response = await api.put(`/api/crm/customers/${id}`, data)
    return response.data
  },

  async deleteCustomer(id: string): Promise<void> {
    await api.delete(`/api/crm/customers/${id}`)
  },

  async getCustomerStats(): Promise<{
    total: number
    active: number
    inactive: number
  }> {
    const response = await api.get('/api/crm/customers/stats')
    return response.data
  },
}
