package services

import (
	"context"
	"crm-service/internal/core/domain"
	"crm-service/internal/core/ports"
)

type CustomerServiceImpl struct {
	repository ports.CustomerRepository
}

// NewCustomerService creates a new customer service
func NewCustomerService(repo ports.CustomerRepository) ports.CustomerService {
	return &CustomerServiceImpl{
		repository: repo,
	}
}

// GetCustomers retrieves customers with pagination
func (s *CustomerServiceImpl) GetCustomers(ctx context.Context, limit, offset int) ([]domain.Customer, error) {
	return s.repository.GetCustomers(ctx, limit, offset)
}

// GetCustomerByID retrieves a customer by ID
func (s *CustomerServiceImpl) GetCustomerByID(ctx context.Context, id string) (*domain.Customer, error) {
	return s.repository.GetCustomerByID(ctx, id)
}

// CreateCustomer creates a new customer
func (s *CustomerServiceImpl) CreateCustomer(ctx context.Context, customer *domain.Customer) error {
	// Set default status if not provided
	if customer.Status == "" {
		customer.Status = "active"
	}

	return s.repository.CreateCustomer(ctx, customer)
}

// UpdateCustomer updates a customer
func (s *CustomerServiceImpl) UpdateCustomer(ctx context.Context, customer *domain.Customer) error {
	// Get existing customer to preserve created_at
	existing, err := s.repository.GetCustomerByID(ctx, customer.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrNotFound
	}

	// Preserve original creation timestamp
	customer.CreatedAt = existing.CreatedAt

	return s.repository.UpdateCustomer(ctx, customer)
}

// DeleteCustomer removes a customer
func (s *CustomerServiceImpl) DeleteCustomer(ctx context.Context, id string) error {
	return s.repository.DeleteCustomer(ctx, id)
}

// SearchCustomers searches for customers
func (s *CustomerServiceImpl) SearchCustomers(ctx context.Context, query string) ([]domain.Customer, error) {
	return s.repository.SearchCustomers(ctx, query)
}
