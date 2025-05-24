package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "backend/proto"
	"forum-service/internal/entity"
	"forum-service/internal/handler"
	"forum-service/internal/repository"
	"forum-service/internal/usecase"
	"forum-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/gin-contrib/cors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	migrationsPath = flag.String("migrations-path", "./migrations", "path to migrations files")
	grpcPort       = flag.String("grpc-port", ":50052", "gRPC server port for forum service")
)

// AuthClientAdapter adapts pb.AuthServiceClient to handler.AuthClient
type AuthClientAdapter struct {
	client pb.AuthServiceClient
}

func (a *AuthClientAdapter) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateSessionResponse, error) {
	return a.client.ValidateToken(ctx, req)
}

func (a *AuthClientAdapter) GetUser(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.UserResponse, error) {
	resp, err := a.client.GetUserProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.UserResponse{
		Id:       uint64(resp.User.Id),
		Username: resp.User.Username,
		Role:     resp.User.Role,
	}, nil
}

// CommentRepoAdapter adapts repository.CommentRepository to usecase.CommentRepository
type CommentRepoAdapter struct {
	repo repository.CommentRepository
}

func (a *CommentRepoAdapter) Create(ctx context.Context, comment *entity.Comment) error {
	return a.repo.CreateComment(ctx, comment)
}

func (a *CommentRepoAdapter) GetByID(ctx context.Context, id int64) (*entity.Comment, error) {
	return a.repo.GetCommentByID(ctx, id)
}

func (a *CommentRepoAdapter) Delete(ctx context.Context, id int64) error {
	return a.repo.DeleteComment(ctx, id, 0) // 0 as userID since it's not used in this method
}

func (a *CommentRepoAdapter) DeleteWithUserID(ctx context.Context, id int64, userID int64) error {
	return a.repo.DeleteComment(ctx, id, userID)
}

func (a *CommentRepoAdapter) GetCommentsByPostID(ctx context.Context, postID int64) ([]entity.Comment, error) {
	return a.repo.GetCommentsByPostID(ctx, postID)
}

// forumPostServiceServer implements PostServiceServer
type forumPostServiceServer struct {
	pb.UnimplementedPostServiceServer
	logger *logger.Logger
}

func (s *forumPostServiceServer) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error) {
	s.logger.Info("Received CreatePost request", zap.Any("request", req))
	return &pb.CreatePostResponse{Id: 123}, nil
}

func (s *forumPostServiceServer) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.GetPostResponse, error) {
	s.logger.Info("Received GetPost request", zap.Any("request", req))
	return &pb.GetPostResponse{
		Id:        req.Id,
		Title:     "Placeholder Post",
		Content:   "This is a placeholder post.",
		UserId:    1,
		TopicId:   1,
		CreatedAt: time.Now().String(),
		UpdatedAt: time.Now().String(),
	}, nil
}

func (s *forumPostServiceServer) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
	s.logger.Info("Received DeletePost request", zap.Any("request", req))
	return &pb.DeletePostResponse{Success: true}, nil
}

func (s *forumPostServiceServer) ListPosts(ctx context.Context, req *pb.ListPostsRequest) (*pb.ListPostsResponse, error) {
	s.logger.Info("Received ListPosts request", zap.Any("request", req))
	return &pb.ListPostsResponse{
		Posts: []*pb.Post{
			{Id: 1, Title: "Placeholder 1"},
			{Id: 2, Title: "Placeholder 2"},
		},
		Total: 2,
	}, nil
}

// forumCommentServiceServer implements CommentServiceServer
type forumCommentServiceServer struct {
	pb.UnimplementedCommentServiceServer
	logger *logger.Logger
}

func (s *forumCommentServiceServer) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CreateCommentResponse, error) {
	s.logger.Info("Received CreateComment request", zap.Any("request", req))
	return &pb.CreateCommentResponse{Id: 456}, nil
}

func (s *forumCommentServiceServer) GetComment(ctx context.Context, req *pb.GetCommentRequest) (*pb.GetCommentResponse, error) {
	s.logger.Info("Received GetComment request", zap.Any("request", req))
	return &pb.GetCommentResponse{
		Id:        req.Id,
		PostId:    1,
		UserId:    1,
		Content:   "Placeholder comment.",
		CreatedAt: time.Now().String(),
		UpdatedAt: time.Now().String(),
	}, nil
}

func (s *forumCommentServiceServer) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteCommentResponse, error) {
	s.logger.Info("Received DeleteComment request", zap.Any("request", req))
	return &pb.DeleteCommentResponse{Success: true}, nil
}

func (s *forumCommentServiceServer) ListComments(ctx context.Context, req *pb.ListCommentsRequest) (*pb.ListCommentsResponse, error) {
	s.logger.Info("Received ListComments request", zap.Any("request", req))
	return &pb.ListCommentsResponse{
		Comments: []*pb.Comment{
			{Id: 1, Content: "Comment 1"},
			{Id: 2, Content: "Comment 2"},
		},
		Total: 2,
	}, nil
}

func main() {
	flag.Parse()

	// Инициализация логгера
	logger, err := logger.NewLogger("development")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting forum service")

	// Подключение к базе данных
	conn, err := pgx.Connect(context.Background(), "host=localhost port=5432 user=postgres password=postgres dbname=forum_service sslmode=disable")
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer conn.Close(context.Background())
	
	// Create sqlx.DB connection
	db, err := sqlx.Connect("postgres", "postgres://postgres:postgres@localhost:5432/forum_service?sslmode=disable")
	if err != nil {
		logger.Fatal("Failed to connect to database with sqlx", zap.Error(err))
	}
	defer db.Close()
	
	logger.Info("Connected to database")

	if err := runMigrations("postgres://postgres:postgres@localhost:5432/forum_service?sslmode=disable", *migrationsPath, logger); err != nil {
		logger.Fatal("Migrations failed: %v", err)
	}

	// Инициализация репозиториев
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	// Инициализация auth клиента
	authConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Failed to connect to auth service", zap.Error(err))
	}
	defer authConn.Close()
	authClient := &AuthClientAdapter{client: pb.NewAuthServiceClient(authConn)}

	// Инициализация usecase
	postUC := usecase.NewPostUsecase(postRepo, authClient.client, logger)
	commentUC := usecase.NewCommentUseCase(&CommentRepoAdapter{repo: commentRepo}, postRepo)

	// Инициализация HTTP сервера
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
	}))
	postHandler := handler.NewPostHandler(postUC, logger, authClient)
	commentHandler := handler.NewCommentHandler(commentUC, authClient.client)
	postHandler.RegisterRoutes(router)
	commentHandler.RegisterRoutes(router)

	// Запуск HTTP сервера
	go func() {
		logger.Info("Starting HTTP server", zap.String("port", "8080"))
		if err := router.Run(":8080"); err != nil {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Инициализация и запуск gRPC сервера для Forum Service
	go func() {
		lis, err := net.Listen("tcp", *grpcPort)
		if err != nil {
			logger.Fatal("Failed to listen for gRPC", zap.Error(err))
		}

		s := grpc.NewServer()
		pb.RegisterPostServiceServer(s, &forumPostServiceServer{logger: logger})
		pb.RegisterCommentServiceServer(s, &forumCommentServiceServer{logger: logger})

		logger.Info("Starting gRPC server for Forum Service", zap.String("port", *grpcPort))
		if err := s.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC for Forum Service", zap.Error(err))
		}
	}()

	// Канал для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Пример фоновой задачи (например, чистка старых сообщений)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			before := time.Now().AddDate(0, 0, -7)
			// TODO: Вызвать usecase для удаления старых сообщений, если реализовано
			logger.Info("Cleanup job executed", zap.Time("before", before))
		}
	}()

	<-stop
	logger.Info("Shutting down server...")
	logger.Info("Server stopped")
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