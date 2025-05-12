package domain

import (
	"time"
)

// TicketStatus represents the status of a ticket
type TicketStatus string

// TicketPriority represents the priority of a ticket
type TicketPriority string

// Ticket statuses
const (
	StatusNew        TicketStatus = "new"
	StatusOpen       TicketStatus = "open"
	StatusInProgress TicketStatus = "in_progress"
	StatusResolved   TicketStatus = "resolved"
	StatusClosed     TicketStatus = "closed"
)

// Ticket priorities
const (
	PriorityLow      TicketPriority = "low"
	PriorityMedium   TicketPriority = "medium"
	PriorityHigh     TicketPriority = "high"
	PriorityCritical TicketPriority = "critical"
)

// Ticket represents a customer support ticket in the CRM system
type Ticket struct {
	ID          string         `json:"id"`
	CustomerID  string         `json:"customer_id"`
	AgentID     *string        `json:"agent_id,omitempty"`
	Subject     string         `json:"subject"`
	Description string         `json:"description"`
	Status      TicketStatus   `json:"status"`
	Priority    TicketPriority `json:"priority"`
	Tags        []string       `json:"tags,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ClosedAt    *time.Time     `json:"closed_at,omitempty"`
}

// TicketEvent represents an event in the history of a ticket
type TicketEvent struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	UserID    string    `json:"user_id"`
	EventType string    `json:"event_type"` // created, status_changed, comment, assigned, closed
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
