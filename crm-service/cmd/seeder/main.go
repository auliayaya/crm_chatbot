package main

import (
	"context"
	"crm-service/internal/adapters/secondary/repository"
	"crm-service/internal/config"
	"crm-service/internal/core/domain"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

const (
	NUM_CUSTOMERS = 20
	NUM_AGENTS    = 5
	NUM_TICKETS   = 50
)

// Sample data for seeding
var (
	firstNames = []string{"James", "Mary", "John", "Patricia", "Robert", "Jennifer", "Michael", "Linda", "William", "Elizabeth", "David", "Susan", "Richard", "Jessica", "Joseph", "Sarah", "Thomas", "Karen", "Charles", "Nancy"}
	lastNames  = []string{"Smith", "Johnson", "Williams", "Jones", "Brown", "Davis", "Miller", "Wilson", "Moore", "Taylor", "Anderson", "Thomas", "Jackson", "White", "Harris", "Martin", "Thompson", "Garcia", "Martinez", "Robinson"}
	companies  = []string{"Acme Corp", "Wayne Industries", "Stark Enterprises", "Umbrella Corporation", "Globex", "Soylent Corp", "Initech", "Massive Dynamic", "Cyberdyne Systems", "Weyland-Yutani", "Rekall", "Stark Industries", "Oscorp", "LexCorp", "Waystar Royco"}
	domains    = []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "icloud.com", "company.com", "business.org", "example.net"}

	ticketSubjects = []string{
		"Cannot access my account",
		"Payment not processing",
		"How do I reset my password?",
		"Product delivered damaged",
		"Billing discrepancy",
		"Feature request",
		"Service outage reported",
		"Need to update my information",
		"Subscription cancellation",
		"Refund request",
		"Login issues",
		"Mobile app crashing",
		"Missing order",
		"Shipping delay",
		"Product not as described",
	}

	ticketDescriptions = []string{
		"I've been trying to log in for hours but it keeps saying my password is incorrect even though I know it's right.",
		"I attempted to make a payment but it's showing an error. My card works fine on other sites.",
		"I need assistance with resetting my password. The reset link in my email isn't working.",
		"The product arrived with visible damage to the packaging and the item inside is broken.",
		"I was charged twice for my last order. Please refund the duplicate charge.",
		"Could you add a dark mode to the application? It would be much easier on the eyes.",
		"Your service appears to be down. I can't access any features right now.",
		"I've moved to a new address and need to update my account information.",
		"I'd like to cancel my subscription. Can you tell me how to do that?",
		"I'd like to request a refund for my recent purchase as it didn't meet my expectations.",
	}

	departments = []string{"Customer Support", "Technical Support", "Billing", "Sales", "Account Management"}
	ticketTags  = []string{"account", "billing", "technical", "shipping", "refund", "login", "password", "mobile", "website", "payment", "feature", "bug", "question", "complaint", "urgent"}
	statuses    = []string{"new", "open", "in_progress", "resolved", "closed"}
	priorities  = []string{"low", "medium", "high", "critical"}
)

func main() {
	log.Println("Starting CRM database seeder...")

	// Load configuration
	cfg := config.LoadConfig()

	// Create database connection
	repo, err := repository.NewPostgresRepository(
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get DB connection
	db := repo.GetDB()

	// Initialize repositories
	customerRepo := repository.NewCustomerRepository(db)
	agentRepo := repository.NewAgentRepository(db)
	ticketRepo := repository.NewTicketRepository(db)

	// Set random seed
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()

	// Clear existing data (use with caution)
	if err := clearDatabase(db); err != nil {
		log.Fatalf("Failed to clear database: %v", err)
	}

	// Create customers
	customers, err := seedCustomers(ctx, customerRepo)
	if err != nil {
		log.Fatalf("Failed to seed customers: %v", err)
	}
	log.Printf("Created %d customers", len(customers))

	// Create agents
	agents, err := seedAgents(ctx, agentRepo)
	if err != nil {
		log.Fatalf("Failed to seed agents: %v", err)
	}
	log.Printf("Created %d agents", len(agents))

	// Create tickets
	tickets, err := seedTickets(ctx, ticketRepo, customers, agents)
	if err != nil {
		log.Fatalf("Failed to seed tickets: %v", err)
	}
	log.Printf("Created %d tickets", len(tickets))

	// Add ticket events and comments
	if err := seedTicketEvents(ctx, ticketRepo, tickets, agents); err != nil {
		log.Fatalf("Failed to seed ticket events: %v", err)
	}

	log.Println("Database seeding completed successfully!")
}

func clearDatabase(db *sql.DB) error {
	log.Println("Clearing existing data...")

	// Clear in order based on dependencies
	if _, err := db.Exec("DELETE FROM ticket_events"); err != nil {
		return fmt.Errorf("failed to clear ticket events: %w", err)
	}

	if _, err := db.Exec("DELETE FROM tickets"); err != nil {
		return fmt.Errorf("failed to clear tickets: %w", err)
	}

	if _, err := db.Exec("DELETE FROM customers"); err != nil {
		return fmt.Errorf("failed to clear customers: %w", err)
	}

	if _, err := db.Exec("DELETE FROM agents"); err != nil {
		return fmt.Errorf("failed to clear agents: %w", err)
	}

	return nil
}

func seedCustomers(ctx context.Context, repo *repository.CustomerRepository) ([]domain.Customer, error) {
	log.Println("Seeding customers...")

	var customers []domain.Customer
	var customer domain.Customer

	for i := 0; i < NUM_CUSTOMERS; i++ {
		firstName := firstNames[rand.Intn(len(firstNames))]
		lastName := lastNames[rand.Intn(len(lastNames))]
		domain := domains[rand.Intn(len(domains))]
		email := fmt.Sprintf("%s.%s@%s", firstName, lastName, domain)
		company := companies[rand.Intn(len(companies))]

		customer.ID = uuid.New().String()
		customer.Email = email
		customer.FirstName = firstName
		customer.LastName = lastName
		customer.PhoneNumber = generatePhoneNumber()
		customer.CompanyName = company
		customer.Status = "active"
		customer.CreatedAt = time.Now().Add(-time.Duration(rand.Intn(90)) * 24 * time.Hour)
		customer.UpdatedAt = time.Now()

		if err := repo.CreateCustomer(ctx, &customer); err != nil {
			return nil, fmt.Errorf("failed to create customer: %w", err)
		}

		customers = append(customers, customer)
	}

	return customers, nil
}

func seedAgents(ctx context.Context, repo *repository.AgentRepository) ([]domain.Agent, error) {
	log.Println("Seeding agents...")

	var agents []domain.Agent

	for i := 0; i < NUM_AGENTS; i++ {
		firstName := firstNames[rand.Intn(len(firstNames))]
		lastName := lastNames[rand.Intn(len(lastNames))]
		email := fmt.Sprintf("agent.%s.%s@support.com", firstName, lastName)
		department := departments[rand.Intn(len(departments))]

		agent := domain.Agent{
			ID:         uuid.New().String(),
			Email:      email,
			FirstName:  firstName,
			LastName:   lastName,
			Department: department,
			Status:     "active",
			CreatedAt:  time.Now().Add(-time.Duration(rand.Intn(90)) * 24 * time.Hour),
			UpdatedAt:  time.Now(),
		}

		if err := repo.CreateAgent(ctx, &agent); err != nil {
			return nil, fmt.Errorf("failed to create agent: %w", err)
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

func seedTickets(ctx context.Context, repo *repository.TicketRepository, customers []domain.Customer, agents []domain.Agent) ([]domain.Ticket, error) {
	log.Println("Seeding tickets...")

	var tickets []domain.Ticket

	for i := 0; i < NUM_TICKETS; i++ {
		customer := customers[rand.Intn(len(customers))]

		// Randomly assign an agent (or leave unassigned)
		var agentID *string
		if rand.Float32() < 0.8 { // 80% of tickets have an agent assigned
			agent := agents[rand.Intn(len(agents))]
			agentID = &agent.ID
		}

		// Generate random dates for ticket
		createdAt := time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour)
		updatedAt := createdAt.Add(time.Duration(rand.Intn(24)) * time.Hour)

		// Determine ticket status and closed date
		status := domain.TicketStatus(statuses[rand.Intn(len(statuses))])
		var closedAt *time.Time
		if status == domain.StatusClosed || status == domain.StatusResolved {
			closed := updatedAt.Add(time.Duration(rand.Intn(24)) * time.Hour)
			closedAt = &closed
		}

		// Generate random tags (1-3 tags)
		numTags := rand.Intn(3) + 1
		tags := make([]string, numTags)
		for j := 0; j < numTags; j++ {
			tags[j] = ticketTags[rand.Intn(len(ticketTags))]
		}

		subject := ticketSubjects[rand.Intn(len(ticketSubjects))]
		description := ticketDescriptions[rand.Intn(len(ticketDescriptions))]

		ticket := domain.Ticket{
			ID:          uuid.New().String(),
			CustomerID:  customer.ID,
			AgentID:     agentID,
			Subject:     subject,
			Description: description,
			Status:      status,
			Priority:    domain.TicketPriority(priorities[rand.Intn(len(priorities))]),
			Tags:        tags,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			ClosedAt:    closedAt,
		}

		if err := repo.CreateTicket(ctx, &ticket); err != nil {
			return nil, fmt.Errorf("failed to create ticket: %w", err)
		}

		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func seedTicketEvents(ctx context.Context, repo *repository.TicketRepository, tickets []domain.Ticket, agents []domain.Agent) error {
	log.Println("Seeding ticket events and comments...")

	// Sample comment texts
	comments := []string{
		"I've looked into this issue and will be working on it shortly.",
		"Could you provide more information about the problem you're experiencing?",
		"I've replicated the issue and found the root cause. Working on a fix now.",
		"This has been escalated to our development team.",
		"I've updated your account settings as requested.",
		"The issue should be resolved now. Please let me know if you experience any further problems.",
		"I'm checking with our billing department about this.",
		"We apologize for the inconvenience this has caused.",
		"Thanks for your patience while we work on this.",
		"I've added detailed notes to your account about this issue.",
	}

	// Resolution notes
	resolutions := []string{
		"Reset customer password and provided login instructions",
		"Issued refund for duplicate charge",
		"Replaced damaged product and expedited shipping",
		"Fixed account settings and confirmed with customer",
		"Updated billing information and applied discount for inconvenience",
		"Resolved technical issue with user's account",
		"Added feature request to product roadmap",
		"Sent detailed instructions on how to use the requested feature",
		"Confirmed shipping of replacement item",
		"Applied account credit as compensation",
	}

	for _, ticket := range tickets {
		// Create initial ticket creation event
		createEvent := domain.TicketEvent{
			ID:        uuid.New().String(),
			TicketID:  ticket.ID,
			UserID:    "system",
			EventType: "created",
			Content:   "Ticket created",
			Timestamp: ticket.CreatedAt,
		}

		if err := repo.AddTicketEvent(ctx, &createEvent); err != nil {
			return fmt.Errorf("failed to add ticket creation event: %w", err)
		}

		// Randomly add 0-5 comments to each ticket
		numComments := rand.Intn(6)
		for i := 0; i < numComments; i++ {
			// Fix here: Ensure time difference is at least 1
			timeDiff := int(ticket.UpdatedAt.Sub(ticket.CreatedAt))
			if timeDiff <= 0 {
				timeDiff = 1 // Default to 1 nanosecond if timestamps are identical
			}
			commentTime := ticket.CreatedAt.Add(time.Duration(rand.Intn(timeDiff)))

			// Comments can be from agents or customers
			var userID string
			var content string

			if rand.Float32() < 0.7 { // 70% from agents
				agent := agents[rand.Intn(len(agents))]
				userID = agent.ID
				content = comments[rand.Intn(len(comments))]
			} else { // 30% from customers
				userID = ticket.CustomerID
				content = "Customer: " + comments[rand.Intn(len(comments))]
			}

			commentEvent := domain.TicketEvent{
				ID:        uuid.New().String(),
				TicketID:  ticket.ID,
				UserID:    userID,
				EventType: "comment",
				Content:   content,
				Timestamp: commentTime,
			}

			if err := repo.AddTicketEvent(ctx, &commentEvent); err != nil {
				return fmt.Errorf("failed to add comment event: %w", err)
			}
		}

		// Add status change events if applicable
		if ticket.Status != domain.StatusNew {
			statusChange := domain.TicketEvent{
				ID:        uuid.New().String(),
				TicketID:  ticket.ID,
				UserID:    "system",
				EventType: "status_changed",
				Content:   fmt.Sprintf("Status changed from %s to %s", domain.StatusNew, ticket.Status),
				Timestamp: ticket.UpdatedAt.Add(-time.Hour * 2), // Some time before last update
			}

			if err := repo.AddTicketEvent(ctx, &statusChange); err != nil {
				return fmt.Errorf("failed to add status change event: %w", err)
			}
		}

		// Add assignment event if agent is assigned
		if ticket.AgentID != nil {
			// Fix here too
			timeDiff := int(ticket.UpdatedAt.Sub(ticket.CreatedAt))
			if timeDiff <= 0 {
				timeDiff = 1
			}
			assignTime := ticket.CreatedAt.Add(time.Duration(rand.Intn(timeDiff)))

			assignEvent := domain.TicketEvent{
				ID:        uuid.New().String(),
				TicketID:  ticket.ID,
				UserID:    "system",
				EventType: "assigned",
				Content:   fmt.Sprintf("Ticket assigned to agent %s", *ticket.AgentID),
				Timestamp: assignTime,
			}

			if err := repo.AddTicketEvent(ctx, &assignEvent); err != nil {
				return fmt.Errorf("failed to add assignment event: %w", err)
			}
		}

		// Add closed event if ticket is closed
		if ticket.ClosedAt != nil {
			resolution := resolutions[rand.Intn(len(resolutions))]

			closedEvent := domain.TicketEvent{
				ID:        uuid.New().String(),
				TicketID:  ticket.ID,
				UserID:    "system",
				EventType: "closed",
				Content:   fmt.Sprintf("Ticket closed with resolution: %s", resolution),
				Timestamp: *ticket.ClosedAt,
			}

			if err := repo.AddTicketEvent(ctx, &closedEvent); err != nil {
				return fmt.Errorf("failed to add closed event: %w", err)
			}
		}
	}

	return nil
}

func generatePhoneNumber() string {
	return fmt.Sprintf("+1%03d%03d%04d",
		rand.Intn(800)+100,
		rand.Intn(800)+100,
		rand.Intn(10000))
}
