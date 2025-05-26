import React, { useState, useEffect, useRef, useCallback } from 'react';
import { WS_URL, API_URL } from '../../services/chatService';
import useWebSocket from '../../hooks/useWebSocket'; // Import the hook

export default function Chat() {
    const [input, setInput] = useState('');
    const messagesEndRef = useRef(null);

    // Use the useWebSocket hook
    const token = localStorage.getItem('token'); // Get token here to pass to hook
    const { messages, sendMessage, status, error, retry } = useWebSocket(
        WS_URL,
        API_URL,
        token
    );

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    };

    useEffect(() => {
        scrollToBottom();
    }, [messages]);

    const handleSendMessage = useCallback(() => {
        if (!input.trim()) {
            console.log('Cannot send empty message');
            return;
        }

        // Use the sendMessage from the hook
        sendMessage({
            type: 'message',
            message: input.trim(),
            username: localStorage.getItem('username') || 'User' // Assuming username is stored here
        });

        setInput('');
    }, [input, sendMessage]);

    const isConnected = status === 'connected';
    const isConnecting = status === 'connecting';

    return (
        <div style={{ maxWidth: '600px', margin: '0 auto', padding: '20px' }}>
            <h3>Чат {isConnected ? '(Подключено)' : isConnecting ? '(Подключение...)' : '(Отключено)'}</h3>
            {error && <p style={{ color: 'red' }}>Ошибка: {error}</p>}
            <div style={{
                border: '1px solid #ccc',
                height: '300px',
                overflowY: 'auto',
                marginBottom: '10px',
                padding: '10px',
                backgroundColor: '#f9f9f9'
            }}>
                {messages.map((msg, index) => (
                    <div key={msg.id || msg.clientId || index} style={{ marginBottom: '8px' }}>
                        <strong style={{ color: '#2196F3' }}>{msg.username || 'Unknown'}:</strong>{' '}
                        <span>{msg.message}</span>
                        <div style={{ fontSize: '0.8em', color: '#666' }}>
                            {msg.timestamp ? new Date(msg.timestamp).toLocaleTimeString() : 'Нет времени'}
                        </div>
                    </div>
                ))}
                <div ref={messagesEndRef} />
            </div>
            <div style={{ display: 'flex', gap: '10px' }}>
                <input
                    value={input}
                    onChange={e => setInput(e.target.value)}
                    onKeyPress={e => e.key === 'Enter' && handleSendMessage()}
                    disabled={!isConnected}
                    placeholder={isConnected ? "Введите сообщение..." : isConnecting ? "Подключение..." : "Отключено"}
                    style={{
                        flex: 1,
                        padding: '8px',
                        borderRadius: '4px',
                        border: '1px solid #ccc'
                    }}
                />
                <button
                    onClick={handleSendMessage}
                    disabled={!isConnected}
                    style={{
                        padding: '8px 16px',
                        backgroundColor: isConnected ? '#2196F3' : '#ccc',
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: isConnected ? 'pointer' : 'not-allowed'
                    }}
                >
                    Отправить
                </button>
            </div>
        </div>
    );
}