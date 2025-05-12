package services

import (
	"context"
	"crm-service/internal/core/domain"
	"crm-service/internal/core/ports"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
)

// AgentServiceImpl implements the AgentService interface
type AgentServiceImpl struct {
	agentRepo  ports.AgentRepository
	ticketRepo ports.TicketRepository
}

// NewAgentService creates a new agent service
func NewAgentService(
	agentRepo ports.AgentRepository,
	ticketRepo ports.TicketRepository,
) ports.AgentService {
	return &AgentServiceImpl{
		agentRepo:  agentRepo,
		ticketRepo: ticketRepo,
	}
}

// GetAgents retrieves all agents
func (s *AgentServiceImpl) GetAgents(ctx context.Context) ([]domain.Agent, error) {
	return s.agentRepo.GetAgents(ctx)
}

// GetAgentByID retrieves an agent by ID
func (s *AgentServiceImpl) GetAgentByID(ctx context.Context, id string) (*domain.Agent, error) {
	return s.agentRepo.GetAgentByID(ctx, id)
}

// CreateAgent creates a new agent
func (s *AgentServiceImpl) CreateAgent(ctx context.Context, agent *domain.Agent) error {
	// Generate ID if not provided
	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}

	// Set default status if not provided
	if agent.Status == "" {
		agent.Status = "active"
	}

	// Set timestamps
	now := time.Now()
	agent.CreatedAt = now
	agent.UpdatedAt = now

	return s.agentRepo.CreateAgent(ctx, agent)
}

// UpdateAgent updates an existing agent
func (s *AgentServiceImpl) UpdateAgent(ctx context.Context, agent *domain.Agent) error {
	// Get existing agent to preserve created_at
	existing, err := s.agentRepo.GetAgentByID(ctx, agent.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("agent not found")
	}

	// Preserve original creation timestamp
	agent.CreatedAt = existing.CreatedAt
	agent.UpdatedAt = time.Now()

	return s.agentRepo.UpdateAgent(ctx, agent)
}

// GetAgentWorkloads retrieves workload statistics for all agents
func (s *AgentServiceImpl) GetAgentWorkloads(ctx context.Context) ([]domain.AgentWorkload, error) {
	// Get agent workload data from repository
	return s.agentRepo.GetAgentWorkloads(ctx)
}

// FindBestAgentForTicket finds the most suitable agent for a ticket
func (s *AgentServiceImpl) FindBestAgentForTicket(ctx context.Context, ticket *domain.Ticket) (*domain.Agent, error) {
	// Get all active agents
	agents, err := s.agentRepo.GetAgents(ctx)
	if err != nil {
		return nil, err
	}

	// Filter only active agents
	var activeAgents []domain.Agent
	for _, agent := range agents {
		if agent.Status == "active" {
			activeAgents = append(activeAgents, agent)
		}
	}

	if len(activeAgents) == 0 {
		return nil, errors.New("no active agents available")
	}

	// Get open ticket counts for each agent
	openTicketCounts, err := s.ticketRepo.GetOpenTicketCountByAgent(ctx)
	if err != nil {
		return nil, err
	}

	// Score agents based on workload and other factors
	type agentScore struct {
		agent domain.Agent
		score int
	}

	var scoredAgents []agentScore
	for _, agent := range activeAgents {
		// Base score
		score := 100

		// Subtract for open tickets
		score -= openTicketCounts[agent.ID] * 5

		// Check department match (higher score for matching department)
		// This assumes tickets might have a department field. If not, remove this logic.
		// if ticket.Department == agent.Department {
		//     score += 20
		// }

		// Consider high-priority tickets
		if ticket.Priority == domain.PriorityHigh || ticket.Priority == domain.PriorityCritical {
			// For high priority tickets, prefer agents with fewer open tickets
			score -= openTicketCounts[agent.ID] * 10
		}

		scoredAgents = append(scoredAgents, agentScore{
			agent: agent,
			score: score,
		})
	}

	// Sort agents by score (highest first)
	sort.Slice(scoredAgents, func(i, j int) bool {
		return scoredAgents[i].score > scoredAgents[j].score
	})

	// Return the highest scoring agent
	if len(scoredAgents) > 0 {
		return &scoredAgents[0].agent, nil
	}

	// Fallback to first active agent if scoring fails
	return &activeAgents[0], nil
}
