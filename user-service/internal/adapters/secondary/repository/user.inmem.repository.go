package repository

import (
	"errors"
	"sync"
	"user-service/internal/core/domain"
	"user-service/internal/core/ports"
)

// InMemoryUserRepo implements an in-memory user repository for testing
type InMemoryUserRepo struct {
	users     map[string]*domain.User // username -> user
	usersById map[string]*domain.User // id -> user
	emails    map[string]struct{}     // track emails for uniqueness
	mu        sync.RWMutex
}

// NewInMemoryUserRepo creates a new in-memory user repository
func NewInMemoryUserRepo() ports.UserRepository {
	return &InMemoryUserRepo{
		users:     make(map[string]*domain.User),
		usersById: make(map[string]*domain.User),
		emails:    make(map[string]struct{}),
	}
}

// CreateUser adds a user to the repository
func (r *InMemoryUserRepo) CreateUser(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if username exists
	if _, exists := r.users[user.Username]; exists {
		return errors.New("user with this username already exists")
	}

	// Check if email exists
	if _, exists := r.emails[user.Email]; exists {
		return errors.New("user with this email already exists")
	}

	// Store user
	r.users[user.Username] = user
	r.usersById[user.ID] = user
	r.emails[user.Email] = struct{}{}

	return nil
}

// GetUserByUsername retrieves a user by username
func (r *InMemoryUserRepo) GetUserByUsername(username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[username]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *InMemoryUserRepo) GetUserByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.usersById[id]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}
