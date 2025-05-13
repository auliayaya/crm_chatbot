// src/types/auth.ts
export interface User {
  id: string
  email: string
  username: string
  firstName: string
  lastName: string
  role: 'admin' | 'agent' | 'customer' | 'user'
  avatar?: string
}

export interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
}

// src/types/customer.ts
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

// src/types/ticket.ts
export type TicketStatus =
  | 'new'
  | 'open'
  | 'in_progress'
  | 'resolved'
  | 'closed'
export type TicketPriority = 'low' | 'medium' | 'high' | 'critical'

export interface Ticket {
  id: string
  customerId: string
  agentId?: string
  subject: string
  description: string
  status: TicketStatus
  priority: TicketPriority
  tags: string[]
  createdAt: string
  updatedAt: string
  closedAt?: string
}

export interface TicketEvent {
  id: string
  ticketId: string
  userId: string
  eventType: 'created' | 'status_changed' | 'comment' | 'assigned' | 'closed'
  content: string
  timestamp: string
}

// src/types/agent.ts
export interface Agent {
  id: string
  email: string
  firstName: string
  lastName: string
  department: string
  status: 'active' | 'away' | 'offline'
  createdAt: string
  updatedAt: string
}

// src/types/chat.ts
export interface ChatSession {
  id: string
  customerId: string
  ticketId?: string
  status: 'active' | 'closed'
  createdAt: string
  updatedAt: string
  closedAt?: string
}

export interface ChatMessage {
  id: string
  sessionId: string
  userId: string
  userType: 'customer' | 'agent' | 'system'
  content: string
  type: 'text' | 'image' | 'file' | 'system'
  timestamp: string
}
