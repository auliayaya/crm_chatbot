package repository

import (
	"context"
	"crm-service/internal/core/domain"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type TicketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) GetTickets(ctx context.Context, limit, offset int) ([]domain.Ticket, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
        SELECT id, customer_id, agent_id, subject, description, 
               status, priority, created_at, updated_at, closed_at, tags 
        FROM tickets 
        ORDER BY updated_at DESC 
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(
			&t.ID, &t.CustomerID, &t.AgentID, &t.Subject, &t.Description,
			&t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt, &t.ClosedAt,
			pq.Array(&t.Tags),
		); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tickets, nil
}

func (r *TicketRepository) GetTicketByID(ctx context.Context, id string) (*domain.Ticket, error) {
	query := `
        SELECT id, customer_id, agent_id, subject, description, 
               status, priority, created_at, updated_at, closed_at, tags 
        FROM tickets 
        WHERE id = $1
    `

	var ticket domain.Ticket
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ticket.ID, &ticket.CustomerID, &ticket.AgentID, &ticket.Subject, &ticket.Description,
		&ticket.Status, &ticket.Priority, &ticket.CreatedAt, &ticket.UpdatedAt, &ticket.ClosedAt,
		pq.Array(&ticket.Tags),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, err
	}

	return &ticket, nil
}

func (r *TicketRepository) GetTicketsByCustomer(ctx context.Context, customerID string) ([]domain.Ticket, error) {
	query := `
        SELECT id, customer_id, agent_id, subject, description, 
               status, priority, created_at, updated_at, closed_at, tags 
        FROM tickets 
        WHERE customer_id = $1
        ORDER BY updated_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(
			&t.ID, &t.CustomerID, &t.AgentID, &t.Subject, &t.Description,
			&t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt, &t.ClosedAt,
			pq.Array(&t.Tags),
		); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tickets, nil
}

func (r *TicketRepository) GetTicketsByAgent(ctx context.Context, agentID string) ([]domain.Ticket, error) {
	query := `
        SELECT id, customer_id, agent_id, subject, description, 
               status, priority, created_at, updated_at, closed_at, tags 
        FROM tickets 
        WHERE agent_id = $1
        ORDER BY updated_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(
			&t.ID, &t.CustomerID, &t.AgentID, &t.Subject, &t.Description,
			&t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt, &t.ClosedAt,
			pq.Array(&t.Tags),
		); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tickets, nil
}

func (r *TicketRepository) CreateTicket(ctx context.Context, ticket *domain.Ticket) error {
	query := `
        INSERT INTO tickets (
            id, customer_id, agent_id, subject, description, 
            status, priority, created_at, updated_at, closed_at, tags
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
        )
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		ticket.ID, ticket.CustomerID, ticket.AgentID, ticket.Subject, ticket.Description,
		ticket.Status, ticket.Priority, ticket.CreatedAt, ticket.UpdatedAt, ticket.ClosedAt,
		pq.Array(ticket.Tags),
	)

	return err
}

func (r *TicketRepository) UpdateTicket(ctx context.Context, ticket *domain.Ticket) error {
	query := `
        UPDATE tickets 
        SET customer_id = $2, agent_id = $3, subject = $4, description = $5,
            status = $6, priority = $7, updated_at = $8, closed_at = $9, tags = $10
        WHERE id = $1
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		ticket.ID, ticket.CustomerID, ticket.AgentID, ticket.Subject, ticket.Description,
		ticket.Status, ticket.Priority, ticket.UpdatedAt, ticket.ClosedAt,
		pq.Array(ticket.Tags),
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("ticket not found")
	}

	return nil
}

func (r *TicketRepository) DeleteTicket(ctx context.Context, id string) error {
	// First delete any events
	_, err := r.db.ExecContext(ctx, "DELETE FROM ticket_events WHERE ticket_id = $1", id)
	if err != nil {
		return err
	}

	// Then delete the ticket
	result, err := r.db.ExecContext(ctx, "DELETE FROM tickets WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("ticket not found")
	}

	return nil
}

func (r *TicketRepository) AddTicketEvent(ctx context.Context, event *domain.TicketEvent) error {
	query := `
        INSERT INTO ticket_events (id, ticket_id, user_id, event_type, content, timestamp)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		event.ID, event.TicketID, event.UserID, event.EventType, event.Content, event.Timestamp,
	)

	return err
}

func (r *TicketRepository) GetTicketEvents(ctx context.Context, ticketID string) ([]domain.TicketEvent, error) {
	query := `
        SELECT id, ticket_id, user_id, event_type, content, timestamp
        FROM ticket_events
        WHERE ticket_id = $1
        ORDER BY timestamp ASC
    `

	rows, err := r.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.TicketEvent
	for rows.Next() {
		var e domain.TicketEvent
		if err := rows.Scan(
			&e.ID, &e.TicketID, &e.UserID, &e.EventType, &e.Content, &e.Timestamp,
		); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *TicketRepository) GetOpenTicketCountByAgent(ctx context.Context) (map[string]int, error) {
	query := `
        SELECT agent_id, COUNT(*) as ticket_count
        FROM tickets
        WHERE agent_id IS NOT NULL 
        AND status NOT IN ('closed', 'resolved')
        GROUP BY agent_id
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var agentID string
		var count int
		if err := rows.Scan(&agentID, &count); err != nil {
			return nil, err
		}
		counts[agentID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return counts, nil
}
