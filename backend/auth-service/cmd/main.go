package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	pb "github.com/jaliks17/ffffforum/backend/proto"

	"github.com/jaliks17/ffffforum/backend/auth-service/internal/config"
	"github.com/jaliks17/ffffforum/backend/auth-service/internal/controller"
	"github.com/jaliks17/ffffforum/backend/auth-service/internal/repository"
	"github.com/jaliks17/ffffforum/backend/auth-service/internal/usecase"
	"github.com/jaliks17/ffffforum/backend/auth-service/pkg/logger"

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
)

// @title Auth Microservice API
// @description This is an authentication and authorization microservice.
// @version 1.0
// @host localhost:8081
// @BasePath /
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

var (
	grpcPort        = flag.String("grpc-port", ":50051", "gRPC server port")
	httpPort        = flag.String("http-port", ":8081", "HTTP server port")
	dbURL           = flag.String("db-url", "postgres://postgres:postgres@localhost:5432/auth_service?sslmode=disable", "Database connection URL")
	migrationsPath  = flag.String("migrations-path", "./migrations", "path to migrations files")
	tokenSecret     = flag.String("token-secret", "secret", "JWT token secret")
	tokenExpiration = flag.Duration("token-expiration", 24*time.Hour, "JWT token expiration")
	logLevel        = flag.String("log-level", "info", "Logging level")
)

func main() {
	flag.Parse()

	logger, err := logger.NewLogger(*logLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	db, err := sqlx.Connect("postgres", *dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := runMigrations(*dbURL, *migrationsPath, logger); err != nil {
		logger.Fatal("Migrations failed: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	authConfig := &config.AuthConfig{
		Secret:     *tokenSecret,
		Expiration: *tokenExpiration,
	}

	authUseCase := usecase.NewAuthUseCase(
		userRepo,
		sessionRepo,
		authConfig,
		logger,
	)

	grpcController := controller.NewAuthGRPCController(authUseCase)
	httpController := controller.NewAuthHTTPController(authUseCase)

	go startGRPCServer(*grpcPort, grpcController, logger)
	startHTTPServer(*httpPort, httpController, logger)
}

func startGRPCServer(port string, controller *controller.AuthGRPCController, logger *logger.Logger) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Fatal("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, controller)

	logger.Info("Starting gRPC server on %s", port)
	if err := s.Serve(lis); err != nil {
		logger.Fatal("Failed to serve gRPC: %v", err)
	}
}

func startHTTPServer(port string, controller *controller.AuthHTTPController, logger *logger.Logger) {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowCredentials: true,
	}))

	api := router.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/signup", controller.SignUp)
			authGroup.POST("/signin", controller.SignIn)
			authGroup.GET("/users/:id", controller.GetUserProfile)
			authGroup.GET("/validate", controller.ValidateToken)
		}
	}

	// Add Swagger UI route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	logger.Info("Starting HTTP server on %s", port)
	if err := http.ListenAndServe(port, router); err != nil {
		logger.Fatal("Failed to start HTTP server: %v", err)
	}
}

func runMigrations(dbURL, migrationsPath string, logger *logger.Logger) error {
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

	logger.Info("Database migrations applied successfully")
	return nil
}