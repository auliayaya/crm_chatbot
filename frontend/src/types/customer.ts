export interface Customer {
  id: string
  email: string
  firstName: string
  lastName: string
  phoneNumber: string
  companyName: string
  status: 'active' | 'inactive'
  createdAt: string
  updatedAt: string
}

export interface CustomerListResponse {
  customers: Customer[]
  total: number
  page: number
  pageSize: number
}

export interface CreateCustomerData {
  firstName: string
  lastName: string
  email: string
  phoneNumber: string
  companyName?: string
  status: 'active' | 'inactive'
}

export type UpdateCustomerData = Partial<CreateCustomerData>