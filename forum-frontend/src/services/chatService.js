export const API_URL = 'http://localhost:8082/api/v1';
export const WS_URL = 'ws://localhost:8082/ws';

export function connectToChat(token, onMessage) {
  const ws = new WebSocket(`${WS_URL}?token=${token}`);
  
  ws.onopen = () => {
    console.log('WebSocket connection established');
  };

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      if (onMessage) onMessage(data);
    } catch (error) {
      console.error('Error parsing WebSocket message:', error);
    }
  };

  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };

  ws.onclose = () => {
    console.log('WebSocket connection closed');
  };

  return ws;
}

export async function fetchChatHistory(token) {
  try {
    const response = await fetch(`${API_URL}/messages`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error('Failed to fetch chat history');
    }
    
    return await response.json();
  } catch (error) {
    console.error('Error fetching chat history:', error);
    throw error;
  }
}

export async function sendMessage(message, token) {
  try {
    const response = await fetch(`${API_URL}/messages`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(message),
    });
    
    if (!response.ok) {
      throw new Error('Failed to send message');
    }
    
    return await response.json();
  } catch (error) {
    console.error('Error sending message:', error);
    throw error;
  }
}

export default { fetchChatHistory, sendMessage };