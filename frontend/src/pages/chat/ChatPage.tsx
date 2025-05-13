import { useState, useRef, useEffect, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import DashboardLayout from '../../layouts/DashboardLayout'
import Card from '../../components/common/Card'
import Button from '../../components/common/Button'
import { customerService } from '../../services/customerService'
import { useAuthStore } from '../../store/authStore'
import { formatDate } from '../../utils/formatters'
import { useChatWebSocket } from '../../hooks/useChatWebSocket' // Import the custom hook

// Keep Message interface here or move to a types file and import in both places
export interface Message {
  // Export if hook needs it from a separate file
  id?: string
  content: string
  user_id: string
  customer_id: string
  type: 'user' | 'customer' | 'system' | 'bot'
  metadata?: Record<string, any> // Changed to any for flexibility from backend
  created_at?: string
  conversation_id?: string
  status?: 'thinking' | 'complete'
}

export default function ChatPage() {
  const { customerId: routeCustomerId } = useParams<{ customerId: string }>()
  const { user } = useAuthStore()
  const [messages, setMessages] = useState<Message[]>([])
  const [newMessage, setNewMessage] = useState('')
  // const [isConnected, setIsConnected] = useState(false) // Managed by hook
  const [isLoading, setIsLoading] = useState(true) // This will be our primary page loader state
  const [hasReceivedInitialMessages, setHasReceivedInitialMessages] =
    useState(false)
  const [isThinking, setIsThinking] = useState(false)
  // const [error, setError] = useState<string | null>(null) // Combined with webSocketError
  const [appError, setAppError] = useState<string | null>(null) // For non-WebSocket errors
  const [conversationId, setConversationId] = useState<string>('')

  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)

  const {
    data: customer,
    isLoading: isCustomerLoading,
    error: customerError,
  } = useQuery({
    queryKey: ['customer', routeCustomerId],
    queryFn: () => customerService.getCustomerById(routeCustomerId as string),
    enabled: !!routeCustomerId,
  })

  useEffect(() => {
    if (!isCustomerLoading) {
      setIsLoading(false)
    }
    if (customerError) {
      setAppError(customerError.message)
      setIsLoading(false)
    }
  }, [isCustomerLoading, customerError])

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
  console.log('User ', user?.id, ' Customer: ', routeCustomerId)

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

  const handleReceivedMessage = useCallback(
    (data: string) => {
      let parsedData: any
      try {
        parsedData = JSON.parse(data)
        // console.log('WebSocket: Received raw data:', parsedData); // Keep for debugging
      } catch (err) {
        console.error('WebSocket: Failed to parse message data', err)
        // Potentially set an appError here if needed
        return
      }

      if (Array.isArray(parsedData)) {
        console.log(
          'WebSocket: Detected history array from backend.',
          parsedData
        )
        const historyMessages: Message[] = parsedData
          .map((histMsg: any): Message | null => {
            if (
              histMsg.type === 'ping' ||
              (!histMsg.content && histMsg.type !== 'system')
            ) {
              return null
            }
            let frontendType: Message['type']
            switch (histMsg.type) {
              case 'user_message':
              case 'user':
                frontendType = 'user'
                break
              case 'bot_message':
              case 'bot':
                frontendType = 'bot'
                break
              case 'customer_message':
              case 'customer':
                frontendType = 'customer'
                break
              case 'system_message':
              case 'system':
                frontendType = 'system'
                break
              default:
                if (!histMsg.type && histMsg.content) {
                  if (
                    histMsg.user_id &&
                    histMsg.user_id.toLowerCase().includes('bot')
                  ) {
                    frontendType = 'bot'
                  } else if (histMsg.user_id) {
                    frontendType = 'user'
                  } else {
                    return null
                  }
                } else {
                  return null
                }
            }
            const content =
              typeof histMsg.content === 'string' ? histMsg.content : ''
            if (!content && frontendType !== 'system') return null

            return {
              id:
                histMsg.id ||
                histMsg.ID ||
                `hist-${Date.now()}-${Math.random()}`,
              content: content,
              user_id: histMsg.user_id || histMsg.UserID,
              customer_id: histMsg.customer_id || histMsg.CustomerID,
              type: frontendType,
              created_at:
                histMsg.timestamp || histMsg.Timestamp || histMsg.CreatedAt,
              conversation_id:
                histMsg.conversation_id || histMsg.ConversationID,
              metadata: histMsg.metadata || histMsg.Metadata,
              status: 'complete',
            }
          })
          .filter((msg): msg is Message => msg !== null)

        console.log(
          'WebSocket: Processed and filtered history messages:',
          historyMessages
        )
        setMessages(historyMessages) // Assuming history replaces current messages
        if (
          historyMessages.length > 0 &&
          historyMessages[0].conversation_id &&
          !conversationId
        ) {
          setConversationId(historyMessages[0].conversation_id)
        }
        setHasReceivedInitialMessages(true) // Mark that history/initial messages are processed
        return
      }

      // Single message processing
      const receivedMessage: Message = {
        id: parsedData.id || `msg-${Date.now()}-${Math.random()}`,
        content: parsedData.content,
        user_id: parsedData.user_id,
        customer_id: parsedData.customer_id,
        type: parsedData.type as Message['type'],
        created_at: parsedData.timestamp,
        conversation_id: parsedData.conversation_id,
        metadata: parsedData.metadata,
        status: 'complete',
      }
      console.log('WebSocket: Processed single message:', receivedMessage)

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
        // Further system message specific logic can go here
      }
      // Add non-duplicate message
      setMessages((prevMessages) => {
        if (
          receivedMessage.id &&
          prevMessages.some((msg) => msg.id === receivedMessage.id)
        ) {
          return prevMessages
        }
        return [...prevMessages, receivedMessage]
      })

      // If this is the very first message and no history was received (e.g., a brand new chat)
      if (!hasReceivedInitialMessages) {
        setHasReceivedInitialMessages(true)
      }
      setIsThinking(false) // Ensure thinking is stopped
    },
    [conversationId, hasReceivedInitialMessages] // Add hasReceivedInitialMessages
  ) // Added conversationId dependency

  const {
    isConnected,
    sendMessage: sendWsMessage,
    webSocketError,
  } = useChatWebSocket({
    userId: user?.id,
    customerId: routeCustomerId,
    onMessageReceived: handleReceivedMessage,
    addSystemMessage,
    getWebSocketUrl,
  })

  // Effect to clear messages when customer/user changes
  useEffect(() => {
    setMessages([])
    setIsLoading(true) // Reset loading when dependencies for connection change
  }, [routeCustomerId, user?.id])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  useEffect(() => {
    if (isConnected && !isLoading && inputRef.current) {
      inputRef.current.focus()
    }
  }, [isConnected, isLoading])

  const sendMessage = () => {
    const currentUserId = user?.id
    const currentCustId = routeCustomerId

    if (!currentUserId || !currentCustId) {
      setAppError(
        'Cannot send message: User or customer information is missing.'
      )
      return
    }
    if (!newMessage.trim()) return
    if (!isConnected) {
      setAppError('Cannot send message: Connection lost.')
      return
    }

    const messageData: Omit<Message, 'id' | 'status'> = {
      // Omit fields that are added optimistically or by backend
      content: newMessage.trim(),
      user_id: currentUserId,
      customer_id: currentCustId,
      type: 'user',
      metadata: { conversation_id: conversationId },
      created_at: new Date().toISOString(),
      conversation_id: conversationId,
    }

    sendWsMessage(messageData) // Use sendMessage from hook

    const optimisticMessage: Message = {
      ...messageData,
      id: `temp-user-${Date.now()}`,
      status: 'complete',
    }
    setMessages((prev) => [...prev, optimisticMessage])
    setNewMessage('')

    setIsThinking(true)
    const thinkingMessage: Message = {
      id: `thinking-${Date.now()}`,
      content: '...',
      user_id: 'bot', // Assuming bot user_id is known or fixed
      customer_id: currentCustId,
      type: 'bot',
      status: 'thinking',
      created_at: new Date().toISOString(),
      conversation_id: conversationId,
    }
    setMessages((prev) => [...prev, thinkingMessage])
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  // Stop loading ONLY if customer data is ready AND initial messages have been processed
  useEffect(() => {
    if (!isCustomerLoading && hasReceivedInitialMessages) {
      setIsLoading(false)
    }

    // If there's a customer query error, stop loading and show that error
    if (customerError) {
      setAppError(customerError.message)
      setIsLoading(false)
    }
  }, [isCustomerLoading, hasReceivedInitialMessages, customerError])

  // If WebSocket has a persistent error, also stop loading to show the error
  useEffect(() => {
    if (webSocketError) {
      // webSocketError is already set by the hook, appError can be used for other errors
      // setAppError(webSocketError); // Or just let displayError handle it
      setIsLoading(false)
    }
  }, [webSocketError])

  const displayError = appError || webSocketError

  return (
    <DashboardLayout>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">
          Chat with {customer?.firstName}{' '}
          {customer?.lastName || routeCustomerId}
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
        {isLoading ? ( // Use the refined isLoading state
          <div className="flex justify-center items-center h-96">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500"></div>
          </div>
        ) : displayError ? (
          <div className="bg-red-50 border border-red-200 text-red-800 p-4 rounded-md">
            {displayError}
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
                    key={
                      message.id ||
                      `msg-${index}-${message.type}-${message.created_at}`
                    }
                    className={`flex w-full ${
                      // Added w-full for better control if needed
                      message.type === 'user' ? 'justify-end' : 'justify-start'
                    }`}
                  >
                    <div
                      className={`max-w-[70%] min-w-[80px] rounded-lg px-3 py-2 shadow-sm ${
                        // Adjusted padding, added min-width and shadow
                        message.type === 'user'
                          ? 'bg-primary-100 text-primary-800 rounded-br-none' // Example: Tweak corner for "tail"
                          : message.type === 'system'
                          ? 'bg-gray-100 text-gray-700 italic text-center w-full max-w-full' // System messages can be styled differently
                          : message.type === 'bot'
                          ? 'bg-green-100 text-green-800 rounded-bl-none' // Example: Tweak corner for "tail"
                          : 'bg-white border border-gray-200 text-gray-800 rounded-bl-none' // Customer messages
                      }`}
                    >
                      {/* Main message content */}
                      {message.type !== 'system' && ( // System messages might not have typical "content" display
                        <div className="text-sm whitespace-pre-wrap break-words">
                          {message.content}
                        </div>
                      )}
                      {/* System message content (if different styling) */}
                      {message.type === 'system' && (
                        <div className="text-xs italic whitespace-pre-wrap break-words">
                          {message.content}
                        </div>
                      )}

                      {/* Thinking indicator for bot */}
                      {message.status === 'thinking' &&
                        message.type === 'bot' && (
                          <div className="italic text-xs text-gray-500 flex items-center mt-1">
                            <span className="animate-pulse">
                              Bot is thinking
                            </span>
                            <span className="animate-bounce ml-1">...</span>
                          </div>
                        )}

                      {/* Metadata: Timestamp, Type, ID */}
                      <div
                        className={`mt-1.5 pt-1 ${
                          message.type !== 'system'
                            ? 'border-t border-gray-200 dark:border-gray-600'
                            : ''
                        } text-right`}
                      >
                        <span className="text-[10px] text-gray-400 dark:text-gray-500 mr-2">
                          {message.type.toUpperCase()}
                        </span>
                        {message.type !== 'system' && ( // Don't show ID for system messages unless desired
                          <span className="text-[10px] text-gray-400 dark:text-gray-500 mr-2">
                            ID:{' '}
                            {(message.type === 'customer'
                              ? message.customer_id
                              : message.user_id
                            ).substring(0, 8)}
                            ...
                          </span>
                        )}
                        <span className="text-[10px] text-gray-400 dark:text-gray-500">
                          {message.created_at
                            ? formatDate(message.created_at, 'p') // Using short time format
                            : 'sending...'}
                        </span>
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
                  disabled={!isConnected || isLoading || isCustomerLoading}
                ></textarea>
                <Button
                  onClick={sendMessage}
                  disabled={
                    !isConnected ||
                    !newMessage.trim() ||
                    isThinking ||
                    isLoading ||
                    isCustomerLoading
                  }
                  className="self-end"
                  type="button"
                  variant="secondary"
                >
                  {isThinking ? 'Waiting...' : 'Send'}
                </Button>
              </div>
              {!isConnected && !isLoading && !isCustomerLoading && (
                <p className="text-sm text-red-500 mt-2">
                  You are currently disconnected. Please wait for reconnection
                  or refresh.
                </p>
              )}
            </div>
          </>
        )}
      </Card>
    </DashboardLayout>
  )
}
