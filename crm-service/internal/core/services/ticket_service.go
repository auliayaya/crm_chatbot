package services

import (
	"context"
	"crm-service/internal/core/domain"
	"crm-service/internal/core/ports"
	"errors"
	"time"

	"github.com/google/uuid"
)

// TicketServiceImpl implements the TicketService interface
type TicketServiceImpl struct {
	ticketRepo   ports.TicketRepository
	customerRepo ports.CustomerRepository
	agentRepo    ports.AgentRepository
	publisher    ports.MessagePublisher
}

// NewTicketService creates a new ticket service
func NewTicketService(
	ticketRepo ports.TicketRepository,
	customerRepo ports.CustomerRepository,
	agentRepo ports.AgentRepository,
	publisher ports.MessagePublisher,
) ports.TicketService {
	return &TicketServiceImpl{
		ticketRepo:   ticketRepo,
		customerRepo: customerRepo,
		agentRepo:    agentRepo,
		publisher:    publisher,
	}
}

// GetTickets retrieves tickets with pagination
func (s *TicketServiceImpl) GetTickets(ctx context.Context, limit, offset int) ([]domain.Ticket, error) {
	return s.ticketRepo.GetTickets(ctx, limit, offset)
}

// GetTicketByID retrieves a ticket by ID
func (s *TicketServiceImpl) GetTicketByID(ctx context.Context, id string) (*domain.Ticket, error) {
	return s.ticketRepo.GetTicketByID(ctx, id)
}

// CreateTicket creates a new ticket
func (s *TicketServiceImpl) CreateTicket(ctx context.Context, ticket *domain.Ticket) error {
	// Verify the customer exists
	customer, err := s.customerRepo.GetCustomerByID(ctx, ticket.CustomerID)
	if err != nil {
		return err
	}
	if customer == nil {
		return errors.New("customer not found")
	}

	// Generate ID if not provided
	if ticket.ID == "" {
		ticket.ID = uuid.New().String()
	}

	// Set default values
	if ticket.Status == "" {
		ticket.Status = domain.StatusNew
	}
	if ticket.Priority == "" {
		ticket.Priority = domain.PriorityMedium
	}

	// Set timestamps
	now := time.Now()
	ticket.CreatedAt = now
	ticket.UpdatedAt = now

	// Create the ticket
	if err := s.ticketRepo.CreateTicket(ctx, ticket); err != nil {
		return err
	}

	// Create a ticket created event
	event := &domain.TicketEvent{
		ID:        uuid.New().String(),
		TicketID:  ticket.ID,
		UserID:    "system", // Replace with actual user if available
		EventType: "created",
		Content:   "Ticket created",
		Timestamp: now,
	}
	if err := s.ticketRepo.AddTicketEvent(ctx, event); err != nil {
		return err
	}

	// Publish event for ticket creation
	if s.publisher != nil {
		s.publisher.PublishTicketEvent(ticket, "created")
	}

	return nil
}

// UpdateTicket updates an existing ticket
func (s *TicketServiceImpl) UpdateTicket(ctx context.Context, ticket *domain.Ticket) error {
	// Get existing ticket to preserve created_at
	existing, err := s.ticketRepo.GetTicketByID(ctx, ticket.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("ticket not found")
	}

	// Preserve original creation timestamp and customer ID
	ticket.CreatedAt = existing.CreatedAt
	ticket.CustomerID = existing.CustomerID
	ticket.UpdatedAt = time.Now()

	// Create event for status change if applicable
	if existing.Status != ticket.Status {
		event := &domain.TicketEvent{
			ID:        uuid.New().String(),
			TicketID:  ticket.ID,
			UserID:    "system", // Replace with actual user if available
			EventType: "status_changed",
			Content:   "Status changed from " + string(existing.Status) + " to " + string(ticket.Status),
			Timestamp: time.Now(),
		}
		if err := s.ticketRepo.AddTicketEvent(ctx, event); err != nil {
			return err
		}
	}

	// Update the ticket
	if err := s.ticketRepo.UpdateTicket(ctx, ticket); err != nil {
		return err
	}

	// Publish event for ticket update
	if s.publisher != nil {
		s.publisher.PublishTicketEvent(ticket, "updated")
	}

	return nil
}

// AssignTicketToAgent assigns a ticket to an agent
func (s *TicketServiceImpl) AssignTicketToAgent(ctx context.Context, ticketID, agentID string) error {
	// Get the ticket
	ticket, err := s.ticketRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return err
	}
	if ticket == nil {
		return errors.New("ticket not found")
	}

	// Verify the agent exists
	if agentID != "" {
		agent, err := s.agentRepo.GetAgentByID(ctx, agentID)
		if err != nil {
			return err
		}
		if agent == nil {
			return errors.New("agent not found")
		}
	}

	// Get previous agent ID for event logging
	previousAgentID := "none"
	if ticket.AgentID != nil {
		previousAgentID = *ticket.AgentID
	}

	// Update the ticket's agent
	ticket.AgentID = &agentID
	ticket.UpdatedAt = time.Now()

	// If ticket was new, change status to open when assigned
	if ticket.Status == domain.StatusNew {
		ticket.Status = domain.StatusOpen
	}

	// Update the ticket
	if err := s.ticketRepo.UpdateTicket(ctx, ticket); err != nil {
		return err
	}

	// Create assignment event
	event := &domain.TicketEvent{
		ID:        uuid.New().String(),
		TicketID:  ticket.ID,
		UserID:    "system", // Replace with actual user if available
		EventType: "assigned",
		Content:   "Ticket assigned from agent " + previousAgentID + " to " + agentID,
		Timestamp: time.Now(),
	}
	if err := s.ticketRepo.AddTicketEvent(ctx, event); err != nil {
		return err
	}

	// Publish event for ticket assignment
	if s.publisher != nil {
		s.publisher.PublishTicketEvent(ticket, "assigned")
	}

	return nil
}

// AddTicketComment adds a comment to a ticket
func (s *TicketServiceImpl) AddTicketComment(ctx context.Context, ticketID, userID, content string) error {
	// Verify the ticket exists
	ticket, err := s.ticketRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return err
	}
	if ticket == nil {
		return errors.New("ticket not found")
	}

	// Create comment event
	event := &domain.TicketEvent{
		ID:        uuid.New().String(),
		TicketID:  ticketID,
		UserID:    userID,
		EventType: "comment",
		Content:   content,
		Timestamp: time.Now(),
	}
	if err := s.ticketRepo.AddTicketEvent(ctx, event); err != nil {
		return err
	}

	// Update ticket's last modified time
	ticket.UpdatedAt = time.Now()
	if err := s.ticketRepo.UpdateTicket(ctx, ticket); err != nil {
		return err
	}

	// Publish event for comment added
	if s.publisher != nil {
		s.publisher.PublishTicketComment(ticketID, userID, content)
	}

	return nil
}

// CloseTicket closes a ticket with resolution
func (s *TicketServiceImpl) CloseTicket(ctx context.Context, ticketID string, resolution string) error {
	// Get the ticket
	ticket, err := s.ticketRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return err
	}
	if ticket == nil {
		return errors.New("ticket not found")
	}

	// Update ticket status and closed time
	now := time.Now()
	ticket.Status = domain.StatusClosed
	ticket.UpdatedAt = now
	ticket.ClosedAt = &now

	// Update the ticket
	if err := s.ticketRepo.UpdateTicket(ctx, ticket); err != nil {
		return err
	}

	// Create resolution event
	event := &domain.TicketEvent{
		ID:        uuid.New().String(),
		TicketID:  ticketID,
		UserID:    "system", // Replace with actual user if available
		EventType: "closed",
		Content:   "Ticket closed with resolution: " + resolution,
		Timestamp: now,
	}
	if err := s.ticketRepo.AddTicketEvent(ctx, event); err != nil {
		return err
	}

	// Publish event for ticket closure
	if s.publisher != nil {
		s.publisher.PublishTicketEvent(ticket, "closed")
	}

	return nil
}

// GetTicketHistory retrieves the history of events for a ticket
func (s *TicketServiceImpl) GetTicketHistory(ctx context.Context, ticketID string) ([]domain.TicketEvent, error) {
	// Verify the ticket exists
	ticket, err := s.ticketRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, errors.New("ticket not found")
	}

	// Get ticket events
	return s.ticketRepo.GetTicketEvents(ctx, ticketID)
}

// GetTicketsByCustomer retrieves all tickets for a customer
func (s *TicketServiceImpl) GetTicketsByCustomer(ctx context.Context, customerID string) ([]domain.Ticket, error) {
	// Verify the customer exists
	customer, err := s.customerRepo.GetCustomerByID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, domain.ErrNotFound
	}

	// Get tickets for customer
	return s.ticketRepo.GetTicketsByCustomer(ctx, customerID)
}
