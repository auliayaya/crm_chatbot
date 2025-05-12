package repository

import (
	"context"
	"crm-service/internal/core/domain"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type CustomerRepository struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) GetCustomers(ctx context.Context, limit, offset int) ([]domain.Customer, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
        SELECT id, email, first_name, last_name, phone_number, company_name, 
               status, created_at, updated_at, last_contact_at 
        FROM customers 
        ORDER BY created_at DESC 
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []domain.Customer
	for rows.Next() {
		var c domain.Customer
		if err := rows.Scan(
			&c.ID, &c.Email, &c.FirstName, &c.LastName,
			&c.PhoneNumber, &c.CompanyName, &c.Status,
			&c.CreatedAt, &c.UpdatedAt, &c.LastContactAt,
		); err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return customers, nil
}

func (r *CustomerRepository) GetCustomerByID(ctx context.Context, id string) (*domain.Customer, error) {
	query := `
        SELECT id, email, first_name, last_name, phone_number, company_name, 
               status, created_at, updated_at, last_contact_at 
        FROM customers 
        WHERE id = $1
    `

	var customer domain.Customer
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&customer.ID, &customer.Email, &customer.FirstName, &customer.LastName,
		&customer.PhoneNumber, &customer.CompanyName, &customer.Status,
		&customer.CreatedAt, &customer.UpdatedAt, &customer.LastContactAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, err
	}

	return &customer, nil
}

func (r *CustomerRepository) CreateCustomer(ctx context.Context, customer *domain.Customer) error {
	// Generate ID if not provided
	if customer.ID == "" {
		customer.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	customer.CreatedAt = now
	customer.UpdatedAt = now

	query := `
        INSERT INTO customers (
            id, email, first_name, last_name, phone_number, company_name, 
            status, created_at, updated_at, last_contact_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
        )
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		customer.ID, customer.Email, customer.FirstName, customer.LastName,
		customer.PhoneNumber, customer.CompanyName, customer.Status,
		customer.CreatedAt, customer.UpdatedAt, customer.LastContactAt,
	)

	return err
}

func (r *CustomerRepository) UpdateCustomer(ctx context.Context, customer *domain.Customer) error {
	customer.UpdatedAt = time.Now()

	query := `
        UPDATE customers 
        SET email = $2, first_name = $3, last_name = $4, 
            phone_number = $5, company_name = $6, status = $7, 
            updated_at = $8, last_contact_at = $9
        WHERE id = $1
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		customer.ID, customer.Email, customer.FirstName, customer.LastName,
		customer.PhoneNumber, customer.CompanyName, customer.Status,
		customer.UpdatedAt, customer.LastContactAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("customer not found")
	}

	return nil
}

func (r *CustomerRepository) DeleteCustomer(ctx context.Context, id string) error {
	// First check if the customer has tickets
	var ticketCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tickets WHERE customer_id = $1", id).Scan(&ticketCount)
	if err != nil {
		return err
	}

	if ticketCount > 0 {
		return errors.New("cannot delete customer with existing tickets")
	}

	query := "DELETE FROM customers WHERE id = $1"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("customer not found")
	}

	return nil
}

func (r *CustomerRepository) SearchCustomers(ctx context.Context, query string) ([]domain.Customer, error) {
	searchQuery := `
        SELECT id, email, first_name, last_name, phone_number, company_name, 
               status, created_at, updated_at, last_contact_at 
        FROM customers 
        WHERE email ILIKE $1 OR first_name ILIKE $1 OR last_name ILIKE $1 
              OR company_name ILIKE $1
        ORDER BY created_at DESC 
        LIMIT 100
    `

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []domain.Customer
	for rows.Next() {
		var c domain.Customer
		if err := rows.Scan(
			&c.ID, &c.Email, &c.FirstName, &c.LastName,
			&c.PhoneNumber, &c.CompanyName, &c.Status,
			&c.CreatedAt, &c.UpdatedAt, &c.LastContactAt,
		); err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return customers, nil
}
