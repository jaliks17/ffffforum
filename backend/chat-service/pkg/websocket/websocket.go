package websocket

import (
	"log"
	"net/http"
	"sync"
	"time"

	"chat-service/internal/entity"

	"github.com/gorilla/websocket"
)

var (
	Upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Настройка таймаутов
		HandshakeTimeout: 10 * time.Second,
		// Проверка origin
		CheckOrigin: func(r *http.Request) bool {
			return true // В продакшене нужно настроить правильную проверку origin
		},
	}

	Clients    = make(map[*websocket.Conn]int)
	UserConns  = make(map[int]*websocket.Conn) // Map to track user connections
	Broadcast  = make(chan entity.Message)
	mu         sync.Mutex
)

// CloseConnection закрывает WebSocket соединение и удаляет клиента из пула
func CloseConnection(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()

	// Remove from Clients map
	delete(Clients, conn)

	// Remove from UserConns map
	for userID, conn := range UserConns {
		if conn == conn {
			delete(UserConns, userID)
			break
		}
	}

	log.Printf("Client removed from pool. Total clients: %d", len(Clients))
	conn.Close()
}

// GetClientConnection returns the WebSocket connection for a given user ID
func GetClientConnection(userID int) *websocket.Conn {
	for conn, id := range Clients {
		if id == userID {
			return conn
		}
	}
	return nil
}

// AddClient добавляет нового клиента в пул
func AddClient(conn *websocket.Conn, userID int) bool {
	mu.Lock()
	defer mu.Unlock()

	// Check if user already has a connection
	if existingConn, exists := UserConns[userID]; exists {
		log.Printf("User %d already has a connection, closing old one", userID)
		CloseConnection(existingConn)
	}

	// Add new connection
	Clients[conn] = userID
	UserConns[userID] = conn
	log.Printf("New client added to pool. Total clients: %d", len(Clients))
	return true
}

// SendMessage отправляет сообщение конкретному клиенту
func SendMessage(ws *websocket.Conn, msg entity.Message) error {
	return ws.WriteJSON(msg)
}

// BroadcastMessage отправляет сообщение всем клиентам
func BroadcastMessage(message entity.Message) {
	mu.Lock()
	defer mu.Unlock()

	for client := range Clients {
		if err := SendMessage(client, message); err != nil {
			log.Printf("Error broadcasting message: %v", err)
			CloseConnection(client)
		}
	}
}