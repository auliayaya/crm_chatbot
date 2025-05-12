// internal/adapters/secondary/repository/postgres_repository_test.go
package repository_test

import (
	"chat-service/internal/adapters/secondary/repository"
	"chat-service/internal/core/domain"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func init() {
	// Hardcode the correct Colima Docker socket path
	dockerSock := "/Users/auliaillahi/.colima/default/docker.sock"
	os.Setenv("DOCKER_HOST", "unix://"+dockerSock)
	fmt.Println("Set DOCKER_HOST to:", os.Getenv("DOCKER_HOST"))
}


type RepositoryTestSuite struct {
	suite.Suite
	db         *sql.DB
	repository *repository.PostgresRepository
	pool       *dockertest.Pool
	resource   *dockertest.Resource
}


func (suite *RepositoryTestSuite) SetupSuite() {
	// Setup Docker pool
	pool, err := dockertest.NewPool("unix:///Users/auliaillahi/.colima/default/docker.sock")
	if err != nil {
		suite.T().Fatalf("Could not connect to docker: %s", err)
	}
	suite.pool = pool

	// Start postgres container
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			// No need to set POSTGRES_DB here, we'll create it manually
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		suite.T().Fatalf("Could not start resource: %s", err)
	}
	suite.resource = resource

	// Get the mapped port
	port := resource.GetPort("5432/tcp")

	// Retry until the database is ready
	if err := pool.Retry(func() error {
		// First connect to the default 'postgres' database
		defaultConnStr := fmt.Sprintf("postgres://postgres:postgres@localhost:%s/postgres?sslmode=disable", port)
		db, err := sql.Open("postgres", defaultConnStr)
		if err != nil {
			return err
		}
		defer db.Close()

		// Test connection
		if err = db.Ping(); err != nil {
			return err
		}

		// Create the testdb database
		_, err = db.Exec("CREATE DATABASE testdb")
		if err != nil {
			// If already exists, that's fine
			if strings.Contains(err.Error(), "already exists") {
				return nil
			}
			return err
		}

		return nil
	}); err != nil {
		suite.T().Fatalf("Could not initialize database: %s", err)
	}

	// Now that testdb exists, connect to it
	testdbConnStr := fmt.Sprintf("postgres://postgres:postgres@localhost:%s/testdb?sslmode=disable", port)
	suite.db, err = sql.Open("postgres", testdbConnStr)
	if err != nil {
		suite.T().Fatalf("Could not connect to testdb: %s", err)
	}

	// Initialize repository with the same connection string
	suite.repository, err = repository.NewPostgresRepository(
		"localhost",
		"postgres",
		"postgres",
		"testdb",
		testdbConnStr,
	)
	if err != nil {
		suite.T().Fatalf("Could not create repository: %s", err)
	}
}

func (suite *RepositoryTestSuite) TearDownSuite() {
	// Clean up
	if err := suite.pool.Purge(suite.resource); err != nil {
		suite.T().Fatalf("Could not purge resource: %s", err)
	}
}

func (suite *RepositoryTestSuite) TestSaveAndGetMessage() {
	// Create a conversation first
	customerID := uuid.New().String()
	conversation := &domain.Conversation{
		ID:         uuid.New().String(),
		CustomerID: customerID,
		StartedAt:  time.Now(),
		Status:     "active",
	}
	ctx := context.Background()

	err := suite.repository.CreateConversation(ctx, conversation)
	assert.NoError(suite.T(), err)

	// Create a message
	message := &domain.Message{
		ID:         uuid.New().String(),
		Content:    "Test message",
		UserID:     "user123",
		CustomerID: customerID,
		Type:       domain.UserMessage,
		Timestamp:  time.Now(),
	}

	// Save the message
	err = suite.repository.SaveMessage(ctx,message)
	assert.NoError(suite.T(), err)

	// Get messages by customer
	messages, err := suite.repository.GetMessagesByCustomer(ctx,customerID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), messages, 1)
	assert.Equal(suite.T(), message.Content, messages[0].Content)

	// Get messages by conversation
	messages, err = suite.repository.GetMessagesByConversation(ctx, conversation.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), messages, 1)
	assert.Equal(suite.T(), message.Content, messages[0].Content)
}

func (suite *RepositoryTestSuite) TestConversationOperations() {
	customerID := uuid.New().String()

	// Create a conversation
	conversation := &domain.Conversation{
		ID:         uuid.New().String(),
		CustomerID: customerID,
		StartedAt:  time.Now(),
		Status:     "active",
	}
	ctx := context.Background()

	err := suite.repository.CreateConversation(ctx, conversation)
	assert.NoError(suite.T(), err)

	// Get conversation by ID
	retrieved, err := suite.repository.GetConversation(ctx, conversation.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), conversation.ID, retrieved.ID)
	assert.Equal(suite.T(), conversation.CustomerID, retrieved.CustomerID)
	assert.Equal(suite.T(), conversation.Status, retrieved.Status)

	// Get active conversation by customer
	active, err := suite.repository.GetActiveConversationByCustomer(ctx, customerID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), conversation.ID, active.ID)

	// Update conversation
	conversation.Status = "closed"
	conversation.EndedAt = time.Now()
	err = suite.repository.UpdateConversation(ctx, conversation)
	assert.NoError(suite.T(), err)

	// Verify updated status
	updated, err := suite.repository.GetConversation(ctx, conversation.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "closed", updated.Status)
	assert.False(suite.T(), updated.EndedAt.IsZero())

	// Should not find active conversation after closing it
	_, err = suite.repository.GetActiveConversationByCustomer(ctx, customerID)
	assert.Error(suite.T(), err)
}

func TestRepositorySuite(t *testing.T) {
	// t.Skip("Skipping due to Docker socket issues with Colima - TO BE FIXED")

	// Original code stays here
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(RepositoryTestSuite))
}
