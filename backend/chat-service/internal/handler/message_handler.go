// internal/handler/message_handler.go
package handler

import (
	"chat-service/internal/entity"
	"chat-service/internal/usecase"
	myWeb "chat-service/pkg/websocket"
	"context"

	pb "backend/proto"

	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type MessageHandler struct {
	Uc usecase.MessageUseCase
	AuthClient pb.AuthServiceClient
}

func NewMessageHandler(uc usecase.MessageUseCase, authClient pb.AuthServiceClient) *MessageHandler {
	return &MessageHandler{Uc: uc, AuthClient: authClient}
}

// Define a struct for the incoming message format from the frontend
type incomingMessage struct {
	Type string `json:"type"`
	Message string `json:"message"`
	Timestamp string `json:"timestamp"`
	Username string `json:"username"`
}

func (h *MessageHandler) HandleConnections(c *gin.Context) {
	token := c.Query("token")
	log.Printf("Received WebSocket connection request with token: %s", token)

	// Временная обработка тестового токена
	if token == "demo" {
		log.Printf("Using demo token for testing - bypassing Auth Service validation")
		ws, err := myWeb.Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}
		defer func() {
			log.Printf("WebSocket connection closed and cleaning up for demo user")
			myWeb.CloseConnection(ws)
			log.Printf("WebSocket connection closed and cleaned up complete for demo user")
		}()

		log.Printf("WebSocket connection upgraded successfully for demo user")
		if !myWeb.AddClient(ws, 1) {
			log.Printf("Failed to add client to pool or client already exists")
			return
		}

		// Создаем канал для сигнала завершения
		done := make(chan struct{})

		// Запускаем горутину для отправки пингов
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
						log.Printf("Error sending ping to demo user: %v", err)
						select {
						case <-done:
						default:
							close(done)
						}
						return
					}
				case <-done:
					log.Printf("Ping goroutine for demo user stopping")
					return
				}
			}
		}()

		// Горутина для чтения сообщений
		go func() {
			defer func() {
				log.Printf("Message reading goroutine for demo user stopping")
				select {
				case <-done:
				default:
					close(done)
				}
				myWeb.CloseConnection(ws)
			}()

			for {
				var incMsg incomingMessage
				err := ws.ReadJSON(&incMsg)
				if err != nil {
					if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
						log.Printf("Client closed connection normally")
					} else {
						log.Printf("Error reading message from demo user: %v", err)
					}
					return
				}

				// Создаем entity.Message для сохранения и рассылки
				msg := &entity.Message{
					UserID:   1,
					Username: "Demo User",
					Message:  incMsg.Message,
				}

				log.Printf("Received and parsed message from demo user: %+v", msg)
				if err := h.Uc.SaveMessage(msg); err != nil {
					log.Printf("Error saving message from demo user: %v", err)
					continue
				}
				myWeb.Broadcast <- *msg
			}
		}()

		<-done
		log.Printf("HandleConnections for demo user finished")
		return
	}

	// Валидация токена с помощью Auth Service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("Validating token with Auth Service...")
	validateReq := &pb.ValidateSessionRequest{Token: token}
	validateResp, err := h.AuthClient.ValidateSession(ctx, validateReq)
	if err != nil {
		log.Printf("Auth Service error: %v", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if validateResp == nil || !validateResp.GetValid() {
		log.Printf("Invalid token: %s", token)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userID := validateResp.GetUserId()
	userRole := validateResp.GetUserRole()
	log.Printf("Token validated successfully. User ID: %d, Role: %s", userID, userRole)

	// Получаем информацию о пользователе
	userProfileReq := &pb.GetUserProfileRequest{UserId: userID}
	userProfileResp, err := h.AuthClient.GetUserProfile(ctx, userProfileReq)
	if err != nil {
		log.Printf("Auth Service GetUserProfile error: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if userProfileResp.GetUser() == nil {
		log.Printf("User profile not found for user ID: %d", userID)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	username := userProfileResp.GetUser().GetUsername()
	log.Printf("Fetched username: %s for user ID: %d", username, userID)

	ws, err := myWeb.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer func() {
		log.Printf("WebSocket connection closed and cleaning up for user %d", userID)
		if ws != nil {
			myWeb.CloseConnection(ws)
		}
		log.Printf("WebSocket connection closed and cleaned up complete for user %d", userID)
	}()

	log.Printf("WebSocket connection upgraded successfully for user %d", userID)
	
	// Добавляем клиента в пул
	if !myWeb.AddClient(ws, int(userID)) {
		log.Printf("Failed to add client to pool for user %d", userID)
		ws.Close()
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Создаем канал для сигнала завершения
	done := make(chan struct{})

	// Запускаем горутину для отправки пингов
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
					log.Printf("Error sending ping to user %d: %v", userID, err)
					select {
					case <-done:
					default:
						close(done)
					}
					return
				}
			case <-done:
				log.Printf("Ping goroutine for user %d stopping", userID)
				return
			}
		}
	}()

	// Горутина для чтения сообщений
	go func() {
		defer func() {
			log.Printf("Message reading goroutine for user %d stopping", userID)
			select {
			case <-done:
			default:
				close(done)
			}
			myWeb.CloseConnection(ws)
		}()

		log.Printf("Starting message reading goroutine for user %d", userID)

		for {
			// Устанавливаем таймаут на чтение следующего сообщения
			ws.SetReadDeadline(time.Now().Add(120 * time.Second))
			// Устанавливаем таймаут на запись сообщения
			ws.SetWriteDeadline(time.Now().Add(30 * time.Second))

			log.Printf("Waiting to read message from user %d", userID)

			messageType, p, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("Client %d closed connection normally", userID)
				} else {
					log.Printf("Error reading message from user %d: %v", userID, err)
				}
				return
			}

			// Only process text messages
			if messageType == websocket.TextMessage {
				var incMsg incomingMessage
				err = json.Unmarshal(p, &incMsg)
				if err != nil {
					log.Printf("Error unmarshalling message from user %d: %v", userID, err)
					continue
				}

				log.Printf("Received valid message from user %d: %+v", userID, incMsg)

				// Обработка сообщения в зависимости от типа
				switch incMsg.Type {
				case "message":
					msg := &entity.Message{
						UserID:    int(userID),
						Username:  username,
						Message:   incMsg.Message,
						Timestamp: time.Now(),
					}

					if err := h.Uc.SaveMessage(msg); err != nil {
						log.Printf("Error saving message for user %d: %v", userID, err)
						continue
					}

					myWeb.Broadcast <- *msg
				default:
					log.Printf("Received unknown message type from user %d: %s", userID, incMsg.Type)
				}
			} else {
				log.Printf("Received non-text message from user %d. Type: %d", userID, messageType)
			}
		}
	}()

	<-done
	log.Printf("HandleConnections for user %d finished", userID)
}

func (h *MessageHandler) HandleMessages() {
	for {
		msg := <-myWeb.Broadcast
		log.Printf("Broadcasting message: %+v", msg)
		for client := range myWeb.Clients {
			if err := myWeb.SendMessage(client, msg); err != nil {
				log.Printf("Error broadcasting message: %v", err)
				myWeb.CloseConnection(client)
			}
		}
	}
}

// GetMessages godoc
// @Summary Get chat messages history
// @Description Get a list of recent chat messages
// @Tags messages
// @Accept json
// @Produce json
// @Success 200 {array} entity.Message "List of messages"
// @Failure 500 {object} entity.ErrorResponse "Internal Server Error"
// @Router /api/v1/messages [get]
func (h *MessageHandler) GetMessages(c *gin.Context) {
	messages, err := h.Uc.GetMessages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Временно преобразуем UserID и Username для соответствия фронтенду, ожидающему author_id и author_name
	// TODO: Обновить фронтенд для использования UserID и Username
	formattedMessages := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		formattedMessages[i] = map[string]interface{}{
			"id":         msg.ID,
			"author_id":  msg.UserID,
			"author_name": msg.Username,
			"message":    msg.Message,
			"timestamp":  msg.Timestamp,
		}
	}
	c.JSON(http.StatusOK, formattedMessages)
}