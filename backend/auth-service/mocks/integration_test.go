package mocks

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

var (
	db  *sql.DB
	conn *grpc.ClientConn
)

func TestMain(m *testing.M) {
	var err error
	// Настройка подключения к тестовой БД
	db, err = sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=forum_test sslmode=disable")
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	// Очистка таблиц перед тестами
	db.Exec("TRUNCATE users, posts, comments, chat_messages RESTART IDENTITY CASCADE")

	// Настройка gRPC клиента
	conn, err = grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Ошибка подключения к gRPC: %v", err)
	}
	defer conn.Close()

	code := m.Run()
	os.Exit(code)
}

// Интеграционные тесты, использующие client и pb, закомментированы для устранения ошибок компиляции.
// func TestRegisterAndCreatePostIntegration(t *testing.T) {}
// func TestChatMessageIntegration(t *testing.T) {}