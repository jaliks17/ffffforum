package websocket

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/jaliks17/ffffforum/backend/chat-service/internal/entity"

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
	// Добавляем обратную карту для быстрого поиска userID по соединению
	ConnUserIDs = make(map[*websocket.Conn]int) // Map to track userID by connection
	Broadcast  = make(chan entity.Message)
	mu         sync.Mutex
)

// CloseConnection закрывает WebSocket соединение и удаляет клиента из пула
func CloseConnection(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()

	// Получаем userID по соединению из обратной карты и удаляем запись
	if userID, exists := ConnUserIDs[conn]; exists {
		delete(UserConns, userID)
		delete(ConnUserIDs, conn)
	}

	// Удаляем из Clients map
	delete(Clients, conn)

	log.Printf("Client removed from pool. Total clients: %d", len(Clients))
	// Закрываем само соединение
	conn.Close()
}

// GetClientConnection returns the WebSocket connection for a given user ID
func GetClientConnection(userID int) *websocket.Conn {
	mu.Lock() // Добавляем блокировку при чтении
	defer mu.Unlock()
	return UserConns[userID] // Используем UserConns для прямого поиска
}

// AddClient добавляет нового клиента в пул
func AddClient(conn *websocket.Conn, userID int) bool {
	mu.Lock()
	defer mu.Unlock()

	// Check if user already has a connection
	if existingConn, exists := UserConns[userID]; exists {
		log.Printf("User %d already has an active connection. Closing old connection before adding new one.", userID)
		// Close the old connection gracefully
		existingConn.WriteControl(websocket.CloseMessage, 
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Connection replaced by new client"), 
			time.Now().Add(time.Second))
		existingConn.Close()
		// Remove old connection from maps
		delete(Clients, existingConn)
		delete(UserConns, userID)
		delete(ConnUserIDs, existingConn)
	}

	// Add new connection
	Clients[conn] = userID
	UserConns[userID] = conn
	ConnUserIDs[conn] = userID
	log.Printf("New client added to pool. Total clients: %d", len(Clients))
	return true
}

// SendMessage отправляет сообщение конкретному клиенту
func SendMessage(ws *websocket.Conn, msg entity.Message) error {
	// Добавляем блокировку при отправке, чтобы избежать состояния гонки при параллельном доступе к соединению
	// Это может потребовать изменений в Client структуре или Pool логике для более гранулярной блокировки,
	// но для начала попробуем простую блокировку здесь.
	// Внимание: Блокировка здесь может вызвать проблемы с производительностью/блокировками,
	// если SendMessage вызывается часто или из разных горутин без дополнительной синхронизации.
	// mu.Lock()
	// defer mu.Unlock()
	// Рекомендуется управлять Write на каждое соединение в отдельной горутине клиента.
	return ws.WriteJSON(msg)
}

// BroadcastMessage отправляет сообщение всем клиентам
func BroadcastMessage(message entity.Message) {
	mu.Lock()
	defer mu.Unlock()

	for client := range Clients {
		// Отправку сообщения клиенту лучше выполнять в отдельной горутине,
		// чтобы BroadcastMessage не блокировалась при проблемах с одним клиентом.
		// Также это позволяет избежать проблем с блокировкой мьютекса `mu` во время SendMessage.
		go func(c *websocket.Conn) {
			if err := SendMessage(c, message); err != nil {
				log.Printf("Error broadcasting message: %v", err)
				// Если при отправке возникает ошибка, закрываем соединение
				CloseConnection(c)
			}
		}(client) // Запускаем горутину для каждого клиента
	}
}