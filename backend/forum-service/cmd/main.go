package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "backend/proto"
	"forum-service/config"
	_ "forum-service/docs"
	"forum-service/internal/handler"
	"forum-service/internal/repository"
	"forum-service/internal/usecase"
	"forum-service/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log, err := logger.NewLogger("debug")
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	// Загрузка конфигурации
	cfg := config.NewConfig()

	// Подключение к PostgreSQL
	db, err := sqlx.Connect("postgres", cfg.GetDBConnString())
	if err != nil {
		log.Error("Failed to connect to database", err)
		os.Exit(1)
	}
	defer db.Close()

	// Применение миграций
	if err := runMigrations(cfg.GetDBConnString(), "./migrations", log); err != nil {
		log.Error("Failed to run migrations", err)
		os.Exit(1)
	}

	// Инициализация Gin
	router := gin.Default()

	// Настройка CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Подключение к Auth Service
	authConn, err := grpc.Dial(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("Failed to connect to auth service", err)
		os.Exit(1)
	}
	defer authConn.Close()

	// Инициализация репозиториев и usecases
	authClient := pb.NewAuthServiceClient(authConn)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	postUsecase := usecase.NewPostUsecase(postRepo, authClient, log)
	commentUC := usecase.NewCommentUseCase(commentRepo, postRepo, authClient)

	// Регистрация обработчиков
	postHandler := handler.NewPostHandler(postUsecase, log)
	commentHandler := handler.NewCommentHandler(commentUC)

	// Группировка роутов
	api := router.Group("/api/v1")
	{
		// Роуты для постов
		posts := api.Group("/posts")
		{
			posts.POST("", postHandler.CreatePost)
			posts.GET("", postHandler.GetPosts)
			posts.DELETE("/:id", postHandler.DeletePost)
			posts.PUT("/:id", postHandler.UpdatePost)
		}

		// Роуты для комментариев, привязанных к посту
		comments := api.Group("/posts/:id/comments")
		{
			comments.POST("", commentHandler.CreateComment)
			comments.GET("", commentHandler.GetCommentsByPostID)
		}

		// Роут для лайка комментария
		api.POST("/comments/:id/like", commentHandler.LikeComment)
	}

	// Запуск сервера
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server error", err)
			os.Exit(1)
		}
	}()

	log.Info("Server started on :8080")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", err)
	}

	log.Info("Server stopped")
}

func runMigrations(dbURL, migrationsPath string, log *logger.Logger) error {
	m, err := migrate.New(
		"file://"+migrationsPath,
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("Database migrations applied successfully")
	return nil
}