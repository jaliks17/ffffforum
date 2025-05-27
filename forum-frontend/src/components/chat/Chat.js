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
        <div className="chat-container">
            <h3>Чат {isConnected ? '(Подключено)' : isConnecting ? '(Подключение...)' : '(Отключено)'}</h3>
            {error && <p style={{ color: 'red' }}>Ошибка: {error}</p>}
            <div className="messages-window">
                {messages.map((msg, index) => (
                    <div key={msg.id || msg.clientId || index} className={msg.author_id === parseInt(localStorage.getItem('userId'), 10) ? 'message own-message' : 'message'}>
                        <div className="message-header">
                          <span className="message-username">{msg.author_name || 'Unknown'}</span>
                          <span className="message-time">{msg.timestamp ? new Date(msg.timestamp).toLocaleTimeString() : 'Нет времени'}</span>
                        </div>
                        <div className="message-content">{msg.message}</div>
                    </div>
                ))}
                <div ref={messagesEndRef} />
            </div>
            <div className="message-input-area">
                <input
                    value={input}
                    onChange={e => setInput(e.target.value)}
                    onKeyPress={e => e.key === 'Enter' && handleSendMessage()}
                    disabled={!isConnected}
                    placeholder={isConnected ? "Введите сообщение..." : isConnecting ? "Подключение..." : "Отключено"}
                    className="message-input"
                />
                <button
                    onClick={handleSendMessage}
                    disabled={!isConnected || !input.trim()}
                    className="message-send-button"
                >
                    Отправить
                </button>
            </div>
        </div>
    );
}