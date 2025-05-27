package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jaliks17/ffffforum/backend/chat-service/internal/entity"

	_ "github.com/lib/pq"
)

type MessageRepository interface {
	SaveMessage(msg *entity.Message) error
	GetMessages() ([]entity.Message, error)
	DeleteOldMessages(before time.Time) error
}

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (repo *messageRepository) SaveMessage(msg *entity.Message) error {
	log.Printf("Executing SaveMessage query...")

	// Start a new transaction
	tx, err := repo.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback() // Rollback in case of error

	query := "INSERT INTO chat_messages (user_id, username, content, timestamp) VALUES ($1, $2, $3, $4) RETURNING id"

	var id int
	// Use the transaction's QueryRow
	err = tx.QueryRow(query, msg.UserID, msg.Username, msg.Message, msg.Timestamp).Scan(&id)
	if err != nil {
		log.Printf("Error saving message in transaction: %v", err)
		return fmt.Errorf("error saving message: %w", err)
	}

	msg.ID = id
	log.Printf("Message saved successfully in transaction. About to commit. ID: %d", id)

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("CRITICAL ERROR: Failed to commit transaction for message ID %d: %v", id, err)
		return fmt.Errorf("error committing transaction: %w", err)
	}

	log.Printf("Transaction successfully committed for message ID: %d", id)

	return nil
}

func (repo *messageRepository) GetMessages() ([]entity.Message, error) {
	log.Println("Executing GetMessages query...")

	// Remove temporary code for closing and re-opening connection
	
	// Add a ping to check connection health
	if err := repo.db.Ping(); err != nil {
		log.Printf("Database ping failed in GetMessages: %v", err)
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Log the query we're about to execute
	query := "SELECT id, user_id, username, content as message, timestamp FROM chat_messages ORDER BY timestamp ASC"
	log.Printf("Executing query: %s", query)

	rows, err := repo.db.Query(query)
	if err != nil {
		log.Printf("Query error in GetMessages: %v", err)
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var messages []entity.Message
	for rows.Next() {
		var msg entity.Message
		log.Println("Scanning message row...")
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.Username, &msg.Message, &msg.Timestamp)
		if err != nil {
			log.Printf("Scan error in GetMessages: %v", err)
			return nil, fmt.Errorf("scan error: %w", err)
		}
		messages = append(messages, msg)
	}
	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		log.Printf("Rows iteration error in GetMessages: %v", err)
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	log.Printf("Successfully fetched %d messages.", len(messages))
	// Log the fetched messages before returning
	log.Printf("Fetched messages content: %+v", messages)
	return messages, nil
}

func (repo *messageRepository) DeleteOldMessages(before time.Time) error {
	log.Printf("Deleting messages before: %v", before)

	query := "DELETE FROM chat_messages WHERE timestamp < $1"

	_, err := repo.db.Exec(query, before)
	if err != nil {
		log.Printf("Error deleting old messages: %v", err)
		return fmt.Errorf("error deleting old messages: %w", err)
	}

	log.Println("Old messages deleted successfully")

	return nil
}