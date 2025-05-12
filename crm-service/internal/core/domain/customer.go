package domain

import (
	"time"
)

// Customer represents a customer in the CRM system
type Customer struct {
	ID            string     `json:"id" db:"id"`
	Email         string     `json:"email" db:"email"`
	FirstName     string     `json:"first_name" db:"first_name"`
	LastName      string     `json:"last_name" db:"last_name"`
	PhoneNumber   string     `json:"phone_number" db:"phone_number"`
	CompanyName   string     `json:"company_name" db:"company_name"`
	Status        string     `json:"status" db:"status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	LastContactAt *time.Time `json:"last_contact_at,omitempty" db:"last_contact_at"`
}

// CustomerSummary provides a condensed view of a customer
type CustomerSummary struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	CompanyName string `json:"company_name"`
	Status      string `json:"status"`
	OpenTickets int    `json:"open_tickets"`
}
