export type Message = {
  id?: string
  content: string
  user_id: string
  customer_id: string
  conversation_id?: string
  sender_type: 'user' | 'customer' | 'system'
  metadata?: Record<string, string>
  created_at?: string
}

export type InitResponse = {
  conversation_id: string
  history: Message[]
}
