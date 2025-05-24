import { useState, useEffect, useCallback, useRef } from 'react';

const useWebSocket = (webSocketUrl, apiUrl, token) => {
  const [messages, setMessages] = useState([]);
  const [status, setStatus] = useState('disconnected');
  const [error, setError] = useState(null);
  const ws = useRef(null);
  const reconnectAttempts = useRef(0);

  const loadHistory = useCallback(async () => {
    try {
      const response = await fetch(apiUrl, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
      
      const history = await response.json();
      setMessages(history);
    } catch (err) {
      setError(err.message);
      console.error('History load error:', err);
    }
  }, [apiUrl, token]);

  const connectWebSocket = useCallback(() => {
    if (!token || status === 'connected') return;

    const wsUrl = new URL(webSocketUrl);
    wsUrl.searchParams.set('token', token);

    ws.current = new WebSocket(wsUrl.href);
    setStatus('connecting');

    ws.current.onopen = () => {
      console.log('WebSocket connected');
      reconnectAttempts.current = 0;
      setStatus('connected');
      setError(null);
    };

    ws.current.onmessage = (event) => {
      try {
        const newMessage = JSON.parse(event.data);
        setMessages(prev => [...prev, newMessage]);
      } catch (err) {
        console.error('Message parse error:', err);
      }
    };

    ws.current.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason);
      setStatus('disconnected');
      
      if (event.code === 4001 || event.code === 4002) {
        setError('Authentication required');
        return;
      }

      // Exponential backoff reconnect
      const timeout = Math.min(1000 * 2 ** reconnectAttempts.current, 30000);
      setTimeout(() => {
        reconnectAttempts.current++;
        connectWebSocket();
      }, timeout);
    };

    ws.current.onerror = (err) => {
      console.error('WebSocket error:', err);
      setStatus('error');
      setError('Connection error');
    };
  }, [webSocketUrl, token, status]);

  useEffect(() => {
    if (!token) {
      setError('Authentication token is required');
      return;
    }

    loadHistory();
    connectWebSocket();

    return () => {
      if (ws.current?.readyState === WebSocket.OPEN) {
        ws.current.close();
      }
    };
  }, [token, loadHistory, connectWebSocket]);

  const sendMessage = useCallback((message) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({
        ...message,
        timestamp: new Date().toISOString()
      }));
    }
  }, []);

  return {
    messages,
    sendMessage,
    status,
    error,
    retry: loadHistory
  };
};

export default useWebSocket;