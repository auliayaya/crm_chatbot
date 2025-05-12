package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// PostgresRepository provides base DB access functionality
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgresRepository
func NewPostgresRepository(host, user, password, dbname string) (*PostgresRepository, error) {
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		host, user, password, dbname)

	// Try to connect with retry logic
	var db *sql.DB
	var err error

	// Retry connection with exponential backoff
	maxRetries := 5
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			// Test the connection
			err = db.Ping()
			if err == nil {
				break // Successfully connected
			}
		}

		if i < maxRetries-1 {
			log.Printf("Database connection attempt %d failed: %v. Retrying in %v...",
				i+1, err, retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w",
			maxRetries, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Initialize schema
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

// GetDB returns the database connection
func (r *PostgresRepository) GetDB() *sql.DB {
	return r.db
}

// CheckHealth verifies database connectivity
func (r *PostgresRepository) CheckHealth(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// Close closes the database connection
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// initSchema creates required tables if they don't exist
func initSchema(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create customers table
	_, err := db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS customers (
            id VARCHAR(36) PRIMARY KEY,
            email VARCHAR(255) NOT NULL UNIQUE,
            first_name VARCHAR(100) NOT NULL,
            last_name VARCHAR(100) NOT NULL,
            phone_number VARCHAR(20),
            company_name VARCHAR(100),
            status VARCHAR(20) NOT NULL,
            created_at TIMESTAMP NOT NULL,
            updated_at TIMESTAMP NOT NULL,
            last_contact_at TIMESTAMP
        )
    `)
	if err != nil {
		return err
	}

	// Create agents table
	_, err = db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS agents (
            id VARCHAR(36) PRIMARY KEY,
            email VARCHAR(255) NOT NULL UNIQUE,
            first_name VARCHAR(100) NOT NULL,
            last_name VARCHAR(100) NOT NULL,
            department VARCHAR(100) NOT NULL,
            status VARCHAR(20) NOT NULL,
            created_at TIMESTAMP NOT NULL,
            updated_at TIMESTAMP NOT NULL
        )
    `)
	if err != nil {
		return err
	}

	// Create tickets table
	_, err = db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS tickets (
            id VARCHAR(36) PRIMARY KEY,
            customer_id VARCHAR(36) NOT NULL REFERENCES customers(id),
            agent_id VARCHAR(36) REFERENCES agents(id),
            subject VARCHAR(200) NOT NULL,
            description TEXT NOT NULL,
            status VARCHAR(20) NOT NULL,
            priority VARCHAR(20) NOT NULL,
            created_at TIMESTAMP NOT NULL,
            updated_at TIMESTAMP NOT NULL,
            closed_at TIMESTAMP,
            tags TEXT[]
        )
    `)
	if err != nil {
		return err
	}

	// Create ticket events table
	_, err = db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS ticket_events (
            id VARCHAR(36) PRIMARY KEY,
            ticket_id VARCHAR(36) NOT NULL REFERENCES tickets(id),
            user_id VARCHAR(36) NOT NULL,
            event_type VARCHAR(50) NOT NULL,
            content TEXT,
            timestamp TIMESTAMP NOT NULL
        )
    `)

	return err
}
