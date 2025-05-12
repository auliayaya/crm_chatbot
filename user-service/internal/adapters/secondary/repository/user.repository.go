package repository


import (
    "database/sql"
    "errors"
    "fmt"
    "user-service/internal/core/domain"
    
    _ "github.com/lib/pq"
)

type PostgresUserRepository struct {
    db *sql.DB
}

func NewPostgresUserRepository(host, user, password, dbname string) *PostgresUserRepository {
    // Connection string
    psqlInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
        host, user, password, dbname)
    
    // Open connection
    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        panic(err)
    }
    
    // Check connection
    err = db.Ping()
    if err != nil {
        panic(err)
    }
    
    // Create table if not exists
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id VARCHAR(50) PRIMARY KEY,
            email VARCHAR(100) UNIQUE NOT NULL,
            username VARCHAR(50) UNIQUE NOT NULL,
            password VARCHAR(100) NOT NULL,
            role VARCHAR(20) NOT NULL
        )
    `)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Successfully connected to database")
    
    return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(user *domain.User) error {
    // Check if user already exists
    var exists bool
    err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username=$1 OR email=$2)",
        user.Username, user.Email).Scan(&exists)
    if err != nil {
        return err
    }
    
    if exists {
        return errors.New("user with this username or email already exists")
    }
    
    // Insert user
    _, err = r.db.Exec("INSERT INTO users(id, email, username, password, role) VALUES($1, $2, $3, $4, $5)",
        user.ID, user.Email, user.Username, user.Password, user.Role)
    return err
}

func (r *PostgresUserRepository) GetUserByUsername(username string) (*domain.User, error) {
    var user domain.User
    err := r.db.QueryRow("SELECT id, email, username, password, role FROM users WHERE username=$1",
        username).Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("user not found")
        }
        return nil, err
    }
    return &user, nil
}

func (r *PostgresUserRepository) GetUserByID(id string) (*domain.User, error) {
    var user domain.User
    err := r.db.QueryRow("SELECT id, email, username, password, role FROM users WHERE id=$1",
        id).Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("user not found")
        }
        return nil, err
    }
    return &user, nil
}