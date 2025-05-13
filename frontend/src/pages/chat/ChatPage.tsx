import { useState, useRef, useEffect, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import DashboardLayout from '../../layouts/DashboardLayout'
import Card from '../../components/common/Card'
import Button from '../../components/common/Button'
import { customerService } from '../../services/customerService'
import { useAuthStore } from '../../store/authStore'
import { formatDate } from '../../utils/formatters'

interface Message {
  id?: string
  content: string
  user_id: string
  customer_id: string
  type: 'user' | 'customer' | 'system' | 'bot'
  metadata?: Record<string, string>
  created_at?: string
  conversation_id?: string
  status?: 'thinking' | 'complete'
}

export default function ChatPage() {
  const { customerId: routeCustomerId } = useParams<{ customerId: string }>()
  const { user } = useAuthStore()
  const [messages, setMessages] = useState<Message[]>([])
  const [newMessage, setNewMessage] = useState('')
  const [isConnected, setIsConnected] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [isThinking, setIsThinking] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [conversationId, setConversationId] = useState<string>('')
  const ws = useRef<WebSocket | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const pingIntervalRef = useRef<number | null>(null)
  const reconnectAttempts = useRef(0)
  const maxReconnectAttempts = 5

  const { data: customer } = useQuery({
    queryKey: ['customer', routeCustomerId],
    queryFn: () => customerService.getCustomerById(routeCustomerId as string),
    enabled: !!routeCustomerId,
  })

  const addSystemMessage = useCallback(
    (content: string) => {
      const systemMessage: Message = {
        id: `system-${Date.now()}`,
        content,
        user_id: 'system',
        customer_id: routeCustomerId || '',
        type: 'system',
        created_at: new Date().toISOString(),
      }
      setMessages((prev) => [...prev, systemMessage])
    },
    [routeCustomerId]
  )

  const getWebSocketUrl = useCallback(
    (currentUserId: string, currentCustId: string) => {
      const apiBaseUrl =
        import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
      let wsUrl = apiBaseUrl.replace(/^http(s)?:\/\//, (match) => {
        return match.includes('https') ? 'wss://' : 'ws://'
      })
      if (wsUrl.endsWith('/')) {
        wsUrl = wsUrl.slice(0, -1)
      }
      return `${wsUrl}/chat/ws?user_id=${currentUserId}&customer_id=${currentCustId}`
    },
    []
  )

  useEffect(() => {
    const currentUserId = user?.id
    const currentCustId = routeCustomerId

    if (!currentCustId || !currentUserId) {
      setIsLoading(false)
      return
    }

    if (ws.current && ws.current.readyState !== WebSocket.CLOSED) {
      ws.current.onclose = null
      ws.current.close()
    }

    const wsUrl = getWebSocketUrl(currentUserId, currentCustId)
    setIsLoading(true)
    setError(null)

    const socket = new WebSocket(wsUrl)
    ws.current = socket

    socket.onopen = () => {
      setIsConnected(true)
      setIsLoading(false)
      setError(null)
      if (reconnectAttempts.current > 0) {
        addSystemMessage('Connection restored')
      }
      reconnectAttempts.current = 0
      if (pingIntervalRef.current) {
        clearInterval(pingIntervalRef.current)
      }
      pingIntervalRef.current = window.setInterval(() => {
        if (ws.current?.readyState === WebSocket.OPEN) {
          ws.current.send(JSON.stringify({ type: 'ping' }))
        }
      }, 30000)
    }

    socket.onmessage = (event) => {
      let parsedData: any
      try {
        if (typeof event.data === 'string') {
          parsedData = JSON.parse(event.data)
        } else {
          return
        }

        const receivedMessage: Message = {
          id: parsedData.id,
          content: parsedData.content,
          user_id: parsedData.user_id,
          customer_id: parsedData.customer_id,
          type: parsedData.type as Message['type'],
          created_at: parsedData.timestamp,
          conversation_id: parsedData.conversation_id,
          metadata: parsedData.metadata,
          status: 'complete',
        }

        setMessages((prevMessages) =>
          prevMessages.filter(
            (msg) => !(msg.type === 'bot' && msg.status === 'thinking')
          )
        )
        setIsThinking(false)

        if (receivedMessage.type === 'system') {
          if (receivedMessage.metadata?.conversation_id && !conversationId) {
            setConversationId(receivedMessage.metadata.conversation_id)
          }
          if (
            receivedMessage.metadata?.type === 'history' &&
            Array.isArray(receivedMessage.metadata.messages)
          ) {
            const historyMessages = receivedMessage.metadata.messages.map(
              (histMsg: any) => ({
                ...histMsg,
                created_at: histMsg.timestamp || histMsg.created_at,
              })
            )
            setMessages(historyMessages)
          } else if (receivedMessage.content) {
            setMessages((prevMessages) => {
              if (
                receivedMessage.id &&
                prevMessages.some((msg) => msg.id === receivedMessage.id)
              ) {
                return prevMessages
              }
              return [...prevMessages, receivedMessage]
            })
          }
        } else if (
          receivedMessage.type === 'customer' ||
          receivedMessage.type === 'user' ||
          receivedMessage.type === 'bot'
        ) {
          setMessages((prevMessages) => {
            if (
              receivedMessage.id &&
              prevMessages.some((msg) => msg.id === receivedMessage.id)
            ) {
              return prevMessages
            }
            return [...prevMessages, receivedMessage]
          })
        }
      } catch (err) {
        setIsThinking(false)
        setMessages((prevMessages) =>
          prevMessages.filter(
            (msg) => !(msg.type === 'bot' && msg.status === 'thinking')
          )
        )
      }
    }

    socket.onerror = (event) => {
      setError(
        'A WebSocket connection error occurred. Please check the console and server logs for more details.'
      )
      setIsConnected(false)
      setIsLoading(false)
    }

    socket.onclose = (event) => {
      setIsConnected(false)
      if (pingIntervalRef.current) {
        clearInterval(pingIntervalRef.current)
        pingIntervalRef.current = null
      }
      if (!event.wasClean && reconnectAttempts.current < maxReconnectAttempts) {
        setError(
          `Connection closed unexpectedly (Code: ${event.code}). Attempting to reconnect...`
        )
        addSystemMessage('Connection lost. Attempting to reconnect...')
        attemptReconnect()
      } else if (!event.wasClean) {
        setError(
          `Connection closed unexpectedly (Code: ${event.code}). Max reconnect attempts reached.`
        )
        addSystemMessage(
          'Connection lost. Max reconnect attempts reached. Please refresh the page.'
        )
      }
    }

    return () => {
      if (pingIntervalRef.current) {
        clearInterval(pingIntervalRef.current)
      }
      if (ws.current) {
        ws.current.onopen = null
        ws.current.onmessage = null
        ws.current.onerror = null
        ws.current.onclose = null
        if (
          ws.current.readyState === WebSocket.OPEN ||
          ws.current.readyState === WebSocket.CONNECTING
        ) {
          ws.current.close(1000, 'Component unmounting')
        }
      }
      ws.current = null
    }
  }, [routeCustomerId, user?.id, conversationId])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  useEffect(() => {
    if (isConnected && !isLoading && inputRef.current) {
      inputRef.current.focus()
    }
  }, [isConnected, isLoading])

  const fetchMessageHistory = async (convId: string) => {
    try {
      console.log('Would fetch history for conversation:', convId)
    } catch (err) {
      console.error('Failed to fetch message history:', err)
    }
  }

  const sendMessage = () => {
    const currentUserId = user?.id
    const currentCustId = routeCustomerId

    if (!currentUserId || !currentCustId) {
      setError(
        'Cannot send message: User or customer information is missing. Please refresh.'
      )
      return
    }

    if (!newMessage.trim()) {
      return
    }

    if (
      !isConnected ||
      !ws.current ||
      ws.current.readyState !== WebSocket.OPEN
    ) {
      setError(
        'Cannot send message: Connection lost. Please wait for reconnection or refresh.'
      )
      if (!isConnected && reconnectAttempts.current < maxReconnectAttempts) {
        attemptReconnect()
      }
      return
    }

    const messageData: Message = {
      content: newMessage.trim(),
      user_id: currentUserId,
      customer_id: currentCustId,
      type: 'user',
      metadata: {
        conversation_id: conversationId,
      },
      created_at: new Date().toISOString(),
    }

    try {
      ws.current.send(JSON.stringify(messageData))
      setNewMessage('')

      const optimisticMessage: Message = {
        ...messageData,
        id: messageData.id || `temp-user-${Date.now()}`,
        status: 'complete',
      }
      setMessages((prev) => [...prev, optimisticMessage])

      setIsThinking(true)
      const thinkingMessage: Message = {
        id: `thinking-${Date.now()}`,
        content: '...',
        user_id: 'bot',
        customer_id: currentCustId,
        type: 'bot',
        status: 'thinking',
        created_at: new Date().toISOString(),
        conversation_id: conversationId,
      }
      setMessages((prev) => [...prev, thinkingMessage])
    } catch (error) {
      setError('Failed to send message. Please try again.')
      setIsThinking(false)
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  const attemptReconnect = useCallback(() => {
    if (
      ws.current &&
      (ws.current.readyState === WebSocket.OPEN ||
        ws.current.readyState === WebSocket.CONNECTING)
    ) {
      return
    }

    const currentUserId = user?.id
    const currentCustId = routeCustomerId

    if (reconnectAttempts.current >= maxReconnectAttempts) {
      setError(
        'Failed to connect after multiple attempts. Please refresh the page.'
      )
      addSystemMessage('Max reconnection attempts reached. Please refresh.')
      return
    }

    if (!currentUserId || !currentCustId) {
      return
    }

    const timeout = Math.min(1000 * 2 ** reconnectAttempts.current, 30000)
    reconnectAttempts.current += 1

    setTimeout(() => {
      if (ws.current && ws.current.readyState !== WebSocket.CLOSED) {
        ws.current.close(1001, 'Attempting reconnect')
      }
    }, timeout)
  }, [routeCustomerId, user?.id, conversationId, maxReconnectAttempts])

  return (
    <DashboardLayout>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">
          Chat with {customer?.firstName} {customer?.lastName}
        </h1>
        <div className="flex items-center space-x-2">
          <span
            className={`inline-block w-3 h-3 rounded-full ${
              isConnected ? 'bg-green-500' : 'bg-red-500'
            }`}
          ></span>
          <span className="text-sm text-gray-600">
            {isConnected ? 'Connected' : 'Disconnected'}
          </span>
        </div>
      </div>

      <Card>
        {isLoading ? (
          <div className="flex justify-center items-center h-96">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500"></div>
          </div>
        ) : error ? (
          <div className="bg-red-50 border border-red-200 text-red-800 p-4 rounded-md">
            {error}
          </div>
        ) : (
          <>
            <div className="h-96 overflow-y-auto p-4 space-y-4 bg-gray-50 rounded-t-lg">
              {messages.length === 0 ? (
                <div className="flex justify-center items-center h-full text-gray-500">
                  No messages yet. Start the conversation!
                </div>
              ) : (
                messages.map((message, index) => (
                  <div
                    key={message.id || `msg-${index}-${message.type}`}
                    className={`flex ${
                      message.type === 'user' ? 'justify-end' : 'justify-start'
                    }`}
                  >
                    <div
                      className={`max-w-[70%] rounded-lg p-3 ${
                        message.type === 'user'
                          ? 'bg-primary-100 text-primary-800'
                          : message.type === 'system'
                          ? 'bg-gray-100 text-gray-800 italic'
                          : message.type === 'bot'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-white border border-gray-200'
                      }`}
                    >
                      <div className="text-sm">{message.content}</div>
                      {message.status === 'thinking' &&
                        message.type === 'bot' && (
                          <div className="italic text-xs text-gray-500 flex items-center mt-1">
                            <span className="animate-pulse">
                              Bot is thinking
                            </span>
                            <span className="animate-bounce ml-1">...</span>
                          </div>
                        )}
                      <div className="text-xs text-gray-500 mt-1">
                        {message.created_at
                          ? formatDate(message.created_at)
                          : 'Just now'}
                      </div>
                    </div>
                  </div>
                ))
              )}
              <div ref={messagesEndRef} />
            </div>

            <div className="p-4 border-t border-gray-200">
              <div className="flex space-x-2">
                <textarea
                  ref={inputRef}
                  className="flex-1 min-h-[80px] rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                  placeholder="Type your message..."
                  value={newMessage}
                  onChange={(e) => setNewMessage(e.target.value)}
                  onKeyDown={handleKeyPress}
                  disabled={!isConnected || isLoading}
                ></textarea>
                <Button
                  onClick={sendMessage}
                  disabled={
                    !isConnected ||
                    !newMessage.trim() ||
                    isThinking ||
                    isLoading
                  }
                  className="self-end bg-blue"
                  type="button"
                  variant="secondary"
                >
                  {isThinking ? 'Waiting...' : 'Send'}
                </Button>
              </div>
              {!isConnected &&
                !isLoading && (
                  <p className="text-sm text-red-500 mt-2">
                    You are currently disconnected. Please refresh the page or
                    wait for reconnection.
                  </p>
                )}
            </div>
          </>
        )}
      </Card>
    </DashboardLayout>
  )
}
