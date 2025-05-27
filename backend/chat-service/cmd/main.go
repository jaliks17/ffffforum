package main

import (
	"database/sql"
	"log"
	"time"

	pb "backend/proto"
	_ "chat-service/docs"
	"chat-service/internal/handler"
	"chat-service/internal/repository"
	"chat-service/internal/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title Chat Microservice API
// @version 1.0
// @description This is a chat microservice with WebSocket support
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8082
// @BasePath /
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// Строка подключения к базе данных PostgreSQL
	connStr := "postgres://postgres:postgres@localhost:5432/auth_service?sslmode=disable&client_encoding=UTF8"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Установление gRPC соединения с Auth Service
	authConn, err := grpc.Dial(
		"localhost:50051", // Убедитесь, что это правильный адрес Auth Service
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to Auth Service: %v", err)
	}
	defer authConn.Close()

	// Создание клиента Auth Service
	authClient := pb.NewAuthServiceClient(authConn)

	repo := repository.NewMessageRepository(db)
	uc := usecase.NewMessageUseCase(repo)
	h := handler.NewMessageHandler(uc, authClient) // Передача authClient в обработчик

	go h.HandleMessages()

	// Горутина для удаления старых сообщений каждые 24 часа
	go func() {
		log.Println("Starting old message cleanup routine")
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			// Определяем время, старше которого сообщения должны быть удалены (например, 7 дней)
			cutoff := time.Now().Add(-7 * 24 * time.Hour)
			log.Printf("Running old message cleanup for messages before: %v", cutoff)
			if err := uc.DeleteOldMessages(cutoff); err != nil {
				log.Printf("Error during old message cleanup: %v", err)
			}
		}
	}()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Upgrade", "Connection"},
		ExposeHeaders:    []string{"Content-Length", "Upgrade", "Connection"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}))

	// Добавляем middleware для логирования WebSocket запросов
	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/ws" {
			log.Printf("WebSocket request headers: %v", c.Request.Header)
		}
		c.Next()
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// WebSocket endpoint
	// @Summary WebSocket connection
	// @Description Establishes a WebSocket connection for real-time chat
	// @Tags chat
	// @Accept json
	// @Produce json
	// @Success 101 {string} string "Switching Protocols"
	// @Router /ws [get]
	r.GET("/ws", h.HandleConnections)

	// Create an API group
	api := r.Group("/api/v1")
	{
		// Get messages endpoint
		// @Summary Get messages
		// @Description Get all chat messages
		// @Tags chat
		// @Accept json
		// @Produce json
		// @Success 200 {array} models.Message
		// @Router /api/v1/messages [get]
		api.GET("/messages", h.GetMessages)
	}

	log.Println("Listening on :8082...")
	log.Fatal(r.Run(":8082"))
}