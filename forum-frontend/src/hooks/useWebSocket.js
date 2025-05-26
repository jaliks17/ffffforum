import { useState, useEffect, useCallback, useRef } from 'react';

const useWebSocket = (webSocketUrl, apiUrl, token) => {
  const [messages, setMessages] = useState([]);
  const [status, setStatus] = useState('disconnected');
  const [error, setError] = useState(null);
  const ws = useRef(null);
  const reconnectAttempts = useRef(0);
  const messageQueue = useRef([]); // Очередь для сообщений, отправленных до установки соединения

  const loadHistory = useCallback(async () => {
    try {
      console.log('useWebSocket: Attempting to load history...');
      if (!token) {
        console.warn('useWebSocket: No token available for loading history.');
        setMessages([]);
        setStatus('disconnected');
        setError('Authentication token is required');
        return;
      }
      const historyUrl = `${apiUrl}/messages`;
      console.log('useWebSocket: Fetching history from URL:', historyUrl);
      const headers = { 'Authorization': `Bearer ${token}` };
      console.log('useWebSocket: Fetch history headers:', headers);
      const response = await fetch(historyUrl, {
        headers: headers
      });
      
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
      
      const history = await response.json();
      console.log('useWebSocket: History loaded successfully.', history);
      setMessages(history);
    } catch (err) {
      console.error('useWebSocket: History load error:', err);
      setError(err.message);
    }
  }, [apiUrl, token]);

  // Эффект для загрузки истории при монтировании или изменении токена/apiUrl
  useEffect(() => {
      loadHistory();
  }, [loadHistory]); // Зависимость только от loadHistory (которая зависит от token и apiUrl)


  // Эффект для управления жизненным циклом WebSocket
  useEffect(() => {
    console.log('useWebSocket Effect: Setting up WebSocket effect...');

    let currentWs = null; // Переменная для текущего экземпляра WebSocket в этом эффекте
    let reconnectTimeout = null; // Для очистки таймаута переподключения

    const connect = () => {
        // Если уже есть активное соединение, не пытаемся создать новое
        if (currentWs && currentWs.readyState === WebSocket.OPEN) {
            console.log('useWebSocket Effect: WebSocket instance already open, skipping creation.');
            setStatus('connected');
            return;
        }

        if (!token) {
            console.warn('useWebSocket Effect: No token for WebSocket connection.');
            setStatus('disconnected');
            setError('Authentication token is required');
            return;
        }

        console.log('useWebSocket Effect: Attempting to create WebSocket...');
        setStatus('connecting');
        setError(null); // Очищаем ошибки при попытке подключения

        const wsUrl = new URL(webSocketUrl);
        wsUrl.searchParams.set('token', token);

        try {
            currentWs = new WebSocket(wsUrl.href);
            ws.current = currentWs; // Обновляем ref для доступа извне эффекта

            currentWs.onopen = () => {
                console.log('useWebSocket Effect: WebSocket connected');
                reconnectAttempts.current = 0;
                setStatus('connected');
                setError(null);
                // Отправляем сообщения из очереди после успешного подключения
                while(messageQueue.current.length > 0) {
                    const messageToSend = messageQueue.current.shift();
                    if (ws.current?.readyState === WebSocket.OPEN) {
                         console.log('useWebSocket Effect: Sending queued message:', messageToSend);
                         ws.current.send(JSON.stringify(messageToSend));
                    }
                }
            };

            currentWs.onmessage = (event) => {
                try {
                    const newMessage = JSON.parse(event.data);
                    console.log('useWebSocket Effect: Message received:', newMessage);
                    setMessages(prev => [...prev, newMessage]);
                } catch (err) {
                    console.error('useWebSocket Effect: Message parse error:', err);
                }
            };

            currentWs.onclose = (event) => {
                console.log('useWebSocket Effect: WebSocket closed:', event.code, event.reason);
                setStatus('disconnected');
                ws.current = null; // Очищаем ref

                if (event.code === 1000) {
                    console.log('useWebSocket Effect: WebSocket closed normally.');
                    return; // Не переподключаемся при нормальном закрытии
                }

                if (event.code === 4001 || event.code === 4002) {
                    setError('Authentication required');
                    console.error('useWebSocket Effect: Authentication required, not attempting reconnect.');
                    return; // Не переподключаемся при ошибках аутентификации
                }

                // Логика автоматического переподключения с экспоненциальной выдержкой
                const timeout = Math.min(1000 * 2 ** reconnectAttempts.current, 30000);
                console.log(`useWebSocket Effect: Attempting reconnect in ${timeout}ms. Attempt: ${reconnectAttempts.current + 1}`);
                
                reconnectTimeout = setTimeout(() => {
                    reconnectAttempts.current++;
                    connect(); // Повторная попытка подключения
                }, timeout);
            };

            currentWs.onerror = (err) => {
                console.error('useWebSocket Effect: WebSocket error:', err);
                setError('WebSocket connection error');
                setStatus('error');
                // Обработчик onclose, скорее всего, будет вызван после onerror,
                // и там будет произведена попытка переподключения.
            };
        } catch (err) {
            console.error('useWebSocket Effect: Failed to create WebSocket:', err);
            setError('Failed to create WebSocket connection');
            setStatus('error');
            ws.current = null;
        }
    };

    // Закрываем существующее соединение перед попыткой создания нового
    // Это важно при перезапуске эффекта из-за изменения зависимостей
    if (ws.current && (ws.current.readyState === WebSocket.OPEN || ws.current.readyState === WebSocket.CONNECTING)) {
        console.log('useWebSocket Effect: Closing existing WebSocket before creating new...');
        try {
            ws.current.close(1000, 'Dependency change');
        } catch (e) {
            console.error('useWebSocket Effect: Error closing existing WS:', e);
        }
        ws.current = null; // Очищаем ref сразу после попытки закрытия
    }

    // Небольшая задержка перед попыткой подключения, чтобы дать время на очистку
    const connectAttemptTimeout = setTimeout(() => {
        connect(); // Запускаем попытку подключения
    }, 100); // Небольшая задержка, например 100ms

    // Функция очистки эффекта: закрываем WebSocket при размонтировании компонента
    return () => {
      console.log('useWebSocket Effect: Cleaning up effect...');
      clearTimeout(connectAttemptTimeout); // Очищаем таймаут попытки подключения
      clearTimeout(reconnectTimeout); // Очищаем таймаут переподключения

      // Закрываем WebSocket, если он существует и не находится в состоянии закрытия
      if (currentWs && currentWs.readyState !== WebSocket.CLOSING && currentWs.readyState !== WebSocket.CLOSED) {
        console.log('useWebSocket Effect: Closing WebSocket on cleanup.');
        try {
           currentWs.close(1000, 'Component unmount'); // Нормальное закрытие
        } catch (e) {
           console.error('useWebSocket Effect: Error closing WebSocket:', e);
        }
      }
      currentWs = null; // Убеждаемся, что локальная ссылка очищена
      ws.current = null; // Убеждаемся, что ref очищен
      reconnectAttempts.current = 0; // Сбрасываем счетчик попыток переподключения
    };

  }, [webSocketUrl, token]); // Зависимости эффекта WebSocket (изменение webSocketUrl или token перезапустит эффект)


  const sendMessage = useCallback((message) => {
    const messageWithTimestamp = {
       ...message,
       timestamp: new Date().toISOString()
    };

    if (ws.current?.readyState === WebSocket.OPEN) {
      console.log('useWebSocket: Sending message:', messageWithTimestamp);
      ws.current.send(JSON.stringify(messageWithTimestamp));
    } else {
      console.warn('useWebSocket: WebSocket not open. Queuing message.', messageWithTimestamp);
      // Ставим сообщение в очередь, если соединение не открыто
      messageQueue.current.push(messageWithTimestamp);
      // Опционально: попробовать переподключиться немедленно, если статус не "connecting"
      // В этой версии мы полагаемся на автоматическое переподключение в onclose
    }
  }, []); // Зависимости для sendMessage


  return {
    messages,
    sendMessage,
    status,
    error,
  };
};

export default useWebSocket;