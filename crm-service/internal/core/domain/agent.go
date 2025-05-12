package domain

import "time"

// Agent represents a support team member
type Agent struct {
	ID         string    `json:"id" db:"id"`
	Email      string    `json:"email" db:"email"`
	FirstName  string    `json:"first_name" db:"first_name"`
	LastName   string    `json:"last_name" db:"last_name"`
	Department string    `json:"department" db:"department"`
	Status     string    `json:"status" db:"status"` // active, away, offline
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// AgentWorkload represents agent's current workload statistics
type AgentWorkload struct {
	AgentID           string `json:"agent_id"`
	AgentName         string `json:"agent_name"`
	OpenTicketCount   int    `json:"open_ticket_count"`
	ResolvedLastWeek  int    `json:"resolved_last_week"`
	AvgResolutionTime int    `json:"avg_resolution_time_minutes"`
	Status            string `json:"status"`
}
