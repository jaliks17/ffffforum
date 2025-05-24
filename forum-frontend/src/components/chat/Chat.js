import React, { useState, useEffect, useRef } from 'react';

export default function Chat() {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const ws = useRef(null);

  useEffect(() => {
    ws.current = new window.WebSocket('ws://localhost:8080/ws?token=demo');
    ws.current.onmessage = (event) => {
      setMessages(prev => [...prev, event.data]);
    };
    return () => ws.current && ws.current.close();
  }, []);

  const sendMessage = () => {
    if (ws.current && input) {
      ws.current.send(input);
      setInput('');
    }
  };

  return (
    <div>
      <h3>Чат</h3>
      <div style={{border: '1px solid #ccc', height: 150, overflowY: 'auto', marginBottom: 10}}>
        {messages.map((msg, idx) => <div key={idx}>{msg}</div>)}
      </div>
      <input value={input} onChange={e => setInput(e.target.value)} />
      <button onClick={sendMessage}>Отправить</button>
    </div>
  );
}
