package messaging

import (
	"crm-service/internal/core/domain"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// EventMessage represents a standard format for events
type EventMessage struct {
	EventType    string          `json:"event_type"`
	ResourceID   string          `json:"resource_id"`
	ResourceType string          `json:"resource_type"`
	Timestamp    time.Time       `json:"timestamp"`
	Data         json.RawMessage `json:"data"`
}

// RabbitMQAdapter adapts the RabbitMQ client to the MessagePublisher interface
type RabbitMQAdapter struct {
	client *RabbitMQClient
}

// NewRabbitMQAdapter creates a new RabbitMQAdapter
func NewRabbitMQAdapter(client *RabbitMQClient) *RabbitMQAdapter {
	return &RabbitMQAdapter{
		client: client,
	}
}

// PublishTicketEvent publishes an event when a ticket is created, updated, or its status changes
func (a *RabbitMQAdapter) PublishTicketEvent(ticket *domain.Ticket, eventType string) error {
	if ticket == nil {
		return fmt.Errorf("ticket cannot be nil")
	}

	// Serialize ticket data
	ticketData, err := json.Marshal(ticket)
	if err != nil {
		return fmt.Errorf("failed to marshal ticket: %w", err)
	}

	// Create event message
	event := EventMessage{
		EventType:    eventType,
		ResourceID:   ticket.ID,
		ResourceType: "ticket",
		Timestamp:    time.Now(),
		Data:         ticketData,
	}

	// Serialize event
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Define routing key: tickets.<event_type>
	routingKey := fmt.Sprintf("tickets.%s", eventType)

	// Publish to RabbitMQ
	err = a.client.PublishMessage(routingKey, eventJSON)
	if err != nil {
		return fmt.Errorf("failed to publish ticket event: %w", err)
	}

	log.Printf("Published ticket event: %s for ticket %s", eventType, ticket.ID)
	return nil
}

// PublishTicketComment publishes a comment added to a ticket
func (a *RabbitMQAdapter) PublishTicketComment(ticketID string, userID string, content string) error {
	// Create comment data
	commentData := struct {
		TicketID string    `json:"ticket_id"`
		UserID   string    `json:"user_id"`
		Content  string    `json:"content"`
		Time     time.Time `json:"time"`
	}{
		TicketID: ticketID,
		UserID:   userID,
		Content:  content,
		Time:     time.Now(),
	}

	// Serialize comment data
	commentJSON, err := json.Marshal(commentData)
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %w", err)
	}

	// Create event message
	event := EventMessage{
		EventType:    "comment_added",
		ResourceID:   ticketID,
		ResourceType: "ticket",
		Timestamp:    time.Now(),
		Data:         commentJSON,
	}

	// Serialize event
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Define routing key: tickets.comment_added
	routingKey := "tickets.comment_added"

	// Publish to RabbitMQ
	err = a.client.PublishMessage(routingKey, eventJSON)
	if err != nil {
		return fmt.Errorf("failed to publish ticket comment: %w", err)
	}

	log.Printf("Published comment event for ticket %s", ticketID)
	return nil
}

// PublishCustomerEvent publishes an event when a customer is created or updated
func (a *RabbitMQAdapter) PublishCustomerEvent(customer *domain.Customer, eventType string) error {
	if customer == nil {
		return fmt.Errorf("customer cannot be nil")
	}

	// Serialize customer data
	customerData, err := json.Marshal(customer)
	if err != nil {
		return fmt.Errorf("failed to marshal customer: %w", err)
	}

	// Create event message
	event := EventMessage{
		EventType:    eventType,
		ResourceID:   customer.ID,
		ResourceType: "customer",
		Timestamp:    time.Now(),
		Data:         customerData,
	}

	// Serialize event
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Define routing key: customers.<event_type>
	routingKey := fmt.Sprintf("customers.%s", eventType)

	// Publish to RabbitMQ
	err = a.client.PublishMessage(routingKey, eventJSON)
	if err != nil {
		return fmt.Errorf("failed to publish customer event: %w", err)
	}

	log.Printf("Published customer event: %s for customer %s", eventType, customer.ID)
	return nil
}

// PublishAgentEvent publishes an event when an agent is created, updated or status changes
func (a *RabbitMQAdapter) PublishAgentEvent(agent *domain.Agent, eventType string) error {
	if agent == nil {
		return fmt.Errorf("agent cannot be nil")
	}

	// Serialize agent data
	agentData, err := json.Marshal(agent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent: %w", err)
	}

	// Create event message
	event := EventMessage{
		EventType:    eventType,
		ResourceID:   agent.ID,
		ResourceType: "agent",
		Timestamp:    time.Now(),
		Data:         agentData,
	}

	// Serialize event
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Define routing key: agents.<event_type>
	routingKey := fmt.Sprintf("agents.%s", eventType)

	// Publish to RabbitMQ
	err = a.client.PublishMessage(routingKey, eventJSON)
	if err != nil {
		return fmt.Errorf("failed to publish agent event: %w", err)
	}

	log.Printf("Published agent event: %s for agent %s", eventType, agent.ID)
	return nil
}
