package messaging

import (
    "crm-service/internal/core/domain"
    "log"
)

// NoOpMessagePublisher is a fallback implementation that doesn't publish messages
// This is used when RabbitMQ connection fails
type NoOpMessagePublisher struct{}

// NewNoOpMessagePublisher creates a new no-op message publisher
func NewNoOpMessagePublisher() *NoOpMessagePublisher {
    log.Println("Using no-op message publisher (messages will be logged but not sent)")
    return &NoOpMessagePublisher{}
}

// PublishTicketEvent logs but doesn't publish events
func (p *NoOpMessagePublisher) PublishTicketEvent(ticket *domain.Ticket, eventType string) error {
    log.Printf("[NOOP] Would publish ticket event: %s for ticket %s", eventType, ticket.ID)
    return nil
}

// PublishTicketComment logs but doesn't publish comments
func (p *NoOpMessagePublisher) PublishTicketComment(ticketID string, userID string, content string) error {
    log.Printf("[NOOP] Would publish comment for ticket %s by user %s", ticketID, userID)
    return nil
}

// PublishCustomerEvent logs but doesn't publish events
func (p *NoOpMessagePublisher) PublishCustomerEvent(customer *domain.Customer, eventType string) error {
    log.Printf("[NOOP] Would publish customer event: %s for customer %s", eventType, customer.ID)
    return nil
}

// PublishAgentEvent logs but doesn't publish events
func (p *NoOpMessagePublisher) PublishAgentEvent(agent *domain.Agent, eventType string) error {
    log.Printf("[NOOP] Would publish agent event: %s for agent %s", eventType, agent.ID)
    return nil
}