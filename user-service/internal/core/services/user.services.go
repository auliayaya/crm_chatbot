package services

import (
    "time"
    "user-service/internal/core/domain"
    "user-service/internal/core/ports"
    
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

type AuthService struct {
    repo      ports.UserRepository
    jwtSecret []byte
}

func NewAuthService(repo ports.UserRepository, jwtSecret []byte) *AuthService {
    return &AuthService{repo, jwtSecret}
}

func (s *AuthService) Register(email, username, password string) error {
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    
    // Create user
    user := &domain.User{
        ID:       generateUUID(), // Implement UUID generation
        Email:    email,
        Username: username,
        Password: string(hashedPassword),
        Role:     "user",
    }
    
    return s.repo.CreateUser(user)
}

func (s *AuthService) Login(username, password string) (string, error) {
    // Get user
    user, err := s.repo.GetUserByUsername(username)
    if err != nil {
        return "", err
    }
    
    // Compare passwords
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
    if err != nil {
        return "", err
    }
    
    // Generate JWT
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub": user.ID,
        "username": user.Username,
        "role": user.Role,
        "exp": time.Now().Add(time.Hour * 24).Unix(),
    })
    
    return token.SignedString(s.jwtSecret)
}

func (s *AuthService) VerifyToken(tokenString string) (*domain.User, error) {
    // Parse token
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return s.jwtSecret, nil
    })
    if err != nil || !token.Valid {
        return nil, err
    }
    
    // Get claims
    claims := token.Claims.(jwt.MapClaims)
    userID := claims["sub"].(string)
    
    // Get user
    return s.repo.GetUserByID(userID)
}

func generateUUID() string {
    // Implement UUID generation
    return "user-" + time.Now().Format("20060102150405")
}