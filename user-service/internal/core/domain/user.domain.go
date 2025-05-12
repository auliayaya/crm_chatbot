package domain

type User struct {
    ID       string
    Email    string
    Username string
    Password string // Hashed
    Role     string
}