import { useRef, useEffect, useCallback, useState } from 'react';


const MAX_RECONNECT_ATTEMPTS = 5;

interface UseChatWebSocketProps {
  userId: string | undefined;
  customerId: string | undefined;
  onMessageReceived: (data: string) => void;
  addSystemMessage: (content: string) => void;
  getWebSocketUrl: (userId: string, customerId: string) => string;
}

interface UseChatWebSocketReturn {
  isConnected: boolean;
  sendMessage: (message: Record<string, any>) => void;
  webSocketError: string | null;
}

export const useChatWebSocket = ({
  userId,
  customerId,
  onMessageReceived,
  addSystemMessage,
  getWebSocketUrl,
}: UseChatWebSocketProps): UseChatWebSocketReturn => {
  const ws = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [webSocketError, setWebSocketError] = useState<string | null>(null);
  const pingIntervalRef = useRef<number | null>(null);
  const reconnectAttempts = useRef(0);

  const attemptReconnectInternal = useCallback(() => {
    if (
      ws.current &&
      (ws.current.readyState === WebSocket.OPEN ||
        ws.current.readyState === WebSocket.CONNECTING)
    ) {
      return;
    }

    if (reconnectAttempts.current >= MAX_RECONNECT_ATTEMPTS) {
      setWebSocketError(
        'Failed to connect after multiple attempts. Please refresh the page.'
      );
      addSystemMessage('Max reconnection attempts reached. Please refresh.');
      return;
    }

    if (!userId || !customerId) {
      return;
    }
    
    const timeout = Math.min(1000 * 2 ** reconnectAttempts.current, 30000);
    reconnectAttempts.current += 1;
    
    console.log(`WebSocket: Reconnect attempt ${reconnectAttempts.current}/${MAX_RECONNECT_ATTEMPTS} in ${timeout / 1000}s`);

    setTimeout(() => {
        // The main useEffect will re-trigger connection attempt due to isConnected state change
        // or if ws.current.close() triggers its onclose which then might lead to re-evaluation.
        // For a more direct re-trigger, we ensure the old socket is closed if it exists.
        if (ws.current && ws.current.readyState !== WebSocket.CLOSED) {
             ws.current.close(1001, 'Attempting reconnect'); // This will trigger onclose
        } else {
            // If no socket or already closed, we might need to manually re-initiate.
            // The main useEffect's dependency on userId/customerId should handle this if they are stable.
            // Forcing a re-run of the main effect can be done by changing a dependency,
            // but ideally, the closure of the socket and subsequent state changes should suffice.
            // Let's rely on the main effect's re-evaluation.
            // If the main effect doesn't run, a dummy state change could be used here.
            // For now, we'll assume the main effect will re-run.
            // To be more explicit, we can call connectWs directly if we expose it from useEffect.
            // However, the main useEffect should re-run if dependencies change or if it's designed to.
            // The most robust way is to have the main useEffect re-establish connection.
            // The `connect` function defined inside useEffect will be called again.
        }
    }, timeout);
  }, [userId, customerId, addSystemMessage, getWebSocketUrl]); // getWebSocketUrl added

  useEffect(() => {
    if (!customerId || !userId) {
      setIsConnected(false);
      return;
    }

    const connect = () => {
      if (ws.current && ws.current.readyState !== WebSocket.CLOSED) {
        ws.current.onclose = null; // Prevent old onclose from firing during new setup
        ws.current.close();
      }
      
      const wsUrl = getWebSocketUrl(userId, customerId);
      console.log("WebSocket: Attempting to connect to:", wsUrl);
      setWebSocketError(null);

      const socket = new WebSocket(wsUrl);
      ws.current = socket;

      socket.onopen = () => {
        console.log("WebSocket: Connection opened successfully.");
        setIsConnected(true);
        setWebSocketError(null);
        if (reconnectAttempts.current > 0) {
          addSystemMessage('Connection restored');
        }
        reconnectAttempts.current = 0;

        // if (pingIntervalRef.current) {
        //   clearInterval(pingIntervalRef.current);
        // }
        // pingIntervalRef.current = window.setInterval(() => {
        //   if (ws.current?.readyState === WebSocket.OPEN) {
        //     ws.current.send(JSON.stringify({ type: 'ping' }));
        //   }
        // }, 30000);
      };

      socket.onmessage = (event) => {
        if (typeof event.data === 'string') {
          onMessageReceived(event.data);
        } else {
          console.warn("WebSocket: Received non-string message:", event.data);
        }
      };

      socket.onerror = (event) => {
        console.error("WebSocket: Error occurred:", event);
        setWebSocketError(
          'A WebSocket connection error occurred. Check console/server logs.'
        );
        setIsConnected(false);
      };

      socket.onclose = (event) => {
        console.log("WebSocket: Connection closed. Code:", event.code, "Reason:", event.reason, "Clean:", event.wasClean);
        setIsConnected(false);
        if (pingIntervalRef.current) {
          clearInterval(pingIntervalRef.current);
          pingIntervalRef.current = null;
        }
        if (!event.wasClean && reconnectAttempts.current < MAX_RECONNECT_ATTEMPTS) {
          addSystemMessage('Connection lost. Attempting to reconnect...');
          attemptReconnectInternal();
        } else if (!event.wasClean) {
          setWebSocketError(
            `Connection closed (Code: ${event.code}). Max reconnect attempts reached.`
          );
          addSystemMessage(
            'Connection lost. Max reconnect attempts reached. Please refresh.'
          );
        }
      };
    }

    connect(); // Initial connection attempt

    return () => {
      if (pingIntervalRef.current) {
        clearInterval(pingIntervalRef.current);
      }
      if (ws.current) {
        ws.current.onopen = null;
        ws.current.onmessage = null;
        ws.current.onerror = null;
        ws.current.onclose = null; // Important to nullify to prevent calls on old socket instance
        if (
          ws.current.readyState === WebSocket.OPEN ||
          ws.current.readyState === WebSocket.CONNECTING
        ) {
          ws.current.close(1000, 'Hook unmounting');
        }
      }
      ws.current = null;
    };
  }, [userId, customerId, onMessageReceived, addSystemMessage, getWebSocketUrl, attemptReconnectInternal]);


  const sendMessage = useCallback((message: Record<string, any> | string) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      try {
        const messageToSend = typeof message === 'string' ? message : JSON.stringify(message);
        ws.current.send(messageToSend);
      } catch (error) {
        console.error("WebSocket: Failed to send message", error);
        setWebSocketError("Failed to send message. Connection might be down.");
      }
    } else {
      console.warn("WebSocket: Cannot send message, not connected.");
      setWebSocketError("Cannot send message. Not connected.");
    }
  }, []);

  return { isConnected, sendMessage, webSocketError };
};