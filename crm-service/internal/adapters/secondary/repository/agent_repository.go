package repository

import (
	"context"
	"crm-service/internal/core/domain"
	"database/sql"
	"errors"
)

type AgentRepository struct {
	db *sql.DB
}

func NewAgentRepository(db *sql.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) GetAgents(ctx context.Context) ([]domain.Agent, error) {
	query := `
        SELECT id, email, first_name, last_name, department, status, created_at, updated_at
        FROM agents
        ORDER BY first_name, last_name
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []domain.Agent
	for rows.Next() {
		var a domain.Agent
		if err := rows.Scan(
			&a.ID, &a.Email, &a.FirstName, &a.LastName,
			&a.Department, &a.Status, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		agents = append(agents, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return agents, nil
}

func (r *AgentRepository) GetAgentByID(ctx context.Context, id string) (*domain.Agent, error) {
	query := `
        SELECT id, email, first_name, last_name, department, status, created_at, updated_at
        FROM agents
        WHERE id = $1
    `

	var agent domain.Agent
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&agent.ID, &agent.Email, &agent.FirstName, &agent.LastName,
		&agent.Department, &agent.Status, &agent.CreatedAt, &agent.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, err
	}

	return &agent, nil
}

func (r *AgentRepository) CreateAgent(ctx context.Context, agent *domain.Agent) error {
	query := `
        INSERT INTO agents (id, email, first_name, last_name, department, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		agent.ID, agent.Email, agent.FirstName, agent.LastName,
		agent.Department, agent.Status, agent.CreatedAt, agent.UpdatedAt,
	)

	return err
}

func (r *AgentRepository) UpdateAgent(ctx context.Context, agent *domain.Agent) error {
	query := `
        UPDATE agents
        SET email = $2, first_name = $3, last_name = $4, department = $5, status = $6, updated_at = $7
        WHERE id = $1
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		agent.ID, agent.Email, agent.FirstName, agent.LastName,
		agent.Department, agent.Status, agent.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("agent not found")
	}

	return nil
}

func (r *AgentRepository) DeleteAgent(ctx context.Context, id string) error {
	// Check if agent has assigned tickets
	var ticketCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tickets WHERE agent_id = $1", id).Scan(&ticketCount)
	if err != nil {
		return err
	}

	if ticketCount > 0 {
		return errors.New("cannot delete agent with assigned tickets")
	}

	// Delete the agent
	result, err := r.db.ExecContext(ctx, "DELETE FROM agents WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("agent not found")
	}

	return nil
}

func (r *AgentRepository) GetAgentWorkloads(ctx context.Context) ([]domain.AgentWorkload, error) {
	query := `
        WITH open_tickets AS (
            SELECT agent_id, COUNT(*) as open_count
            FROM tickets
            WHERE agent_id IS NOT NULL AND status NOT IN ('closed', 'resolved')
            GROUP BY agent_id
        ),
        resolved_last_week AS (
            SELECT agent_id, COUNT(*) as resolved_count
            FROM tickets
            WHERE agent_id IS NOT NULL 
              AND status = 'resolved' 
              AND closed_at >= NOW() - INTERVAL '7 days'
            GROUP BY agent_id
        ),
        resolution_times AS (
            SELECT agent_id, 
                   EXTRACT(EPOCH FROM AVG(closed_at - created_at))/60 as avg_minutes
            FROM tickets
            WHERE agent_id IS NOT NULL 
              AND status = 'resolved' 
              AND closed_at IS NOT NULL
            GROUP BY agent_id
        )
        SELECT a.id, a.first_name || ' ' || a.last_name as name, a.status,
               COALESCE(ot.open_count, 0) as open_tickets,
               COALESCE(rw.resolved_count, 0) as resolved_last_week,
               COALESCE(rt.avg_minutes, 0) as avg_resolution_minutes
        FROM agents a
        LEFT JOIN open_tickets ot ON a.id = ot.agent_id
        LEFT JOIN resolved_last_week rw ON a.id = rw.agent_id
        LEFT JOIN resolution_times rt ON a.id = rt.agent_id
        ORDER BY a.first_name, a.last_name
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workloads []domain.AgentWorkload
	for rows.Next() {
		var w domain.AgentWorkload
		if err := rows.Scan(
			&w.AgentID, &w.AgentName, &w.Status,
			&w.OpenTicketCount, &w.ResolvedLastWeek, &w.AvgResolutionTime,
		); err != nil {
			return nil, err
		}
		workloads = append(workloads, w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return workloads, nil
}
