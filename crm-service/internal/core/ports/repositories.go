package ports

import (
	"context"
	"crm-service/internal/core/domain"
)

// CustomerRepository defines operations for customer management
type CustomerRepository interface {
	GetCustomers(ctx context.Context, limit, offset int) ([]domain.Customer, error)
	GetCustomerByID(ctx context.Context, id string) (*domain.Customer, error)
	CreateCustomer(ctx context.Context, customer *domain.Customer) error
	UpdateCustomer(ctx context.Context, customer *domain.Customer) error
	DeleteCustomer(ctx context.Context, id string) error
	SearchCustomers(ctx context.Context, query string) ([]domain.Customer, error)
}

// TicketRepository defines operations for ticket management
type TicketRepository interface {
	GetTickets(ctx context.Context, limit, offset int) ([]domain.Ticket, error)
	GetTicketByID(ctx context.Context, id string) (*domain.Ticket, error)
	GetTicketsByCustomer(ctx context.Context, customerID string) ([]domain.Ticket, error)
	GetTicketsByAgent(ctx context.Context, agentID string) ([]domain.Ticket, error)
	CreateTicket(ctx context.Context, ticket *domain.Ticket) error
	UpdateTicket(ctx context.Context, ticket *domain.Ticket) error
	DeleteTicket(ctx context.Context, id string) error
	AddTicketEvent(ctx context.Context, event *domain.TicketEvent) error
	GetTicketEvents(ctx context.Context, ticketID string) ([]domain.TicketEvent, error)
	GetOpenTicketCountByAgent(ctx context.Context) (map[string]int, error)
}

// AgentRepository defines operations for agent management
type AgentRepository interface {
	GetAgents(ctx context.Context) ([]domain.Agent, error)
	GetAgentByID(ctx context.Context, id string) (*domain.Agent, error)
	CreateAgent(ctx context.Context, agent *domain.Agent) error
	UpdateAgent(ctx context.Context, agent *domain.Agent) error
	DeleteAgent(ctx context.Context, id string) error
	GetAgentWorkloads(ctx context.Context) ([]domain.AgentWorkload, error)
}
