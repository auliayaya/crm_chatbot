package ports

import (
	"context"
	"crm-service/internal/core/domain"
)

// CustomerService defines business operations for customer management
type CustomerService interface {
	GetCustomers(ctx context.Context, limit, offset int) ([]domain.Customer, error)
	GetCustomerByID(ctx context.Context, id string) (*domain.Customer, error)
	CreateCustomer(ctx context.Context, customer *domain.Customer) error
	UpdateCustomer(ctx context.Context, customer *domain.Customer) error
	DeleteCustomer(ctx context.Context, id string) error
	SearchCustomers(ctx context.Context, query string) ([]domain.Customer, error)
	// Add this new method
	GetCustomersCount(ctx context.Context) (int64, error)
}

// TicketService defines business operations for ticket management
type TicketService interface {
	GetTickets(ctx context.Context, limit, offset int) ([]domain.Ticket, error)
	GetTicketByID(ctx context.Context, id string) (*domain.Ticket, error)
	CreateTicket(ctx context.Context, ticket *domain.Ticket) error
	UpdateTicket(ctx context.Context, ticket *domain.Ticket) error
	AssignTicketToAgent(ctx context.Context, ticketID, agentID string) error
	AddTicketComment(ctx context.Context, ticketID, userID, content string) error
	CloseTicket(ctx context.Context, ticketID string, resolution string) error
	GetTicketHistory(ctx context.Context, ticketID string) ([]domain.TicketEvent, error)
	GetTicketsByCustomer(ctx context.Context, customerID string) ([]domain.Ticket, error)
	GetTicketsCount(ctx context.Context) (int64, error)
}

// AgentService defines business operations for agent management
type AgentService interface {
	GetAgents(ctx context.Context) ([]domain.Agent, error)
	GetAgentByID(ctx context.Context, id string) (*domain.Agent, error)
	CreateAgent(ctx context.Context, agent *domain.Agent) error
	UpdateAgent(ctx context.Context, agent *domain.Agent) error
	GetAgentWorkloads(ctx context.Context) ([]domain.AgentWorkload, error)
	FindBestAgentForTicket(ctx context.Context, ticket *domain.Ticket) (*domain.Agent, error)
}

// MessagePublisher defines operations for publishing events to a message broker
type MessagePublisher interface {
	// PublishTicketEvent publishes an event when a ticket is created, updated, or its status changes
	PublishTicketEvent(ticket *domain.Ticket, eventType string) error

	// PublishTicketComment publishes a comment added to a ticket
	PublishTicketComment(ticketID string, userID string, content string) error

	// PublishCustomerEvent publishes an event when a customer is created or updated
	PublishCustomerEvent(customer *domain.Customer, eventType string) error

	// PublishAgentEvent publishes an event when an agent is created, updated or status changes
	PublishAgentEvent(agent *domain.Agent, eventType string) error
}