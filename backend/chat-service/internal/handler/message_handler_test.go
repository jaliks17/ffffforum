package handler

import (
	"github.com/Mandarinka0707/newRepoGOODarhit/chat/internal/entity"
	"github.com/Mandarinka0707/newRepoGOODarhit/chat/internal/usecase"
	myWeb "github.com/Mandarinka0707/newRepoGOODarhit/chat/pkg/websocket"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	Uc usecase.MessageUseCase
}

func NewMessageHandler(uc usecase.MessageUseCase) *MessageHandler {
	return &MessageHandler{Uc: uc}
}

func (h *MessageHandler) HandleConnections(c *gin.Context) {
	ws, err := myWeb.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	myWeb.Clients[ws] = true

	for {
		var msg entity.Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(myWeb.Clients, ws)
			break
		}
		h.Uc.SaveMessage(msg)
		myWeb.Broadcast <- msg
	}
}

func (h *MessageHandler) HandleMessages() {
	for {
		msg := <-myWeb.Broadcast
		for client := range myWeb.Clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(myWeb.Clients, client)
			}
		}
	}
}

// GetMessages получает список всех сообщений.
//
// @Summary Получить сообщения
// @Description Возвращает все сообщения из чата
// @Tags messages
// @Produce json
// @Success 200 {array} entity.Message
// @Failure 500 {object} entity.ErrorResponse
// @Router /messages [get]
func (h *MessageHandler) GetMessages(c *gin.Context) {
	messages, err := h.Uc.GetMessages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}