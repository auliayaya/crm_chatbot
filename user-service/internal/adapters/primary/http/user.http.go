package http

import (
	"encoding/json"
	"net/http"
	"user-service/internal/core/ports"
	"user-service/internal/core/services"
)

type Handler struct {
    authService *services.AuthService
    rabbitMQ    ports.MessagePublisher
}

func NewHandler(authService *services.AuthService, rabbitMQ ports.MessagePublisher) *Handler {
    return &Handler{
        authService: authService,
        rabbitMQ:    rabbitMQ,
    }
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        Email    string `json:"email"`
        Username string `json:"username"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if req.Email == "" || req.Username == "" || req.Password == "" {
        http.Error(w, "Email, username and password are required", http.StatusBadRequest)
        return
    }

    if err := h.authService.Register(req.Email, req.Username, req.Password); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Publish user created event
    h.rabbitMQ.PublishUserEvent("user_created", map[string]interface{}{
        "username": req.Username,
        "email":    req.Email,
    })

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    token, err := h.authService.Login(req.Username, req.Password)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Publish login event
    h.rabbitMQ.PublishUserEvent("user_login", map[string]interface{}{
        "username": req.Username,
    })

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "token": token,
    })
}

func (h *Handler) VerifyToken(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    tokenString := r.Header.Get("Authorization")
    if tokenString == "" {
        http.Error(w, "Authorization header required", http.StatusUnauthorized)
        return
    }

    // Remove "Bearer " prefix if present
    if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
        tokenString = tokenString[7:]
    }

    user, err := h.authService.VerifyToken(tokenString)
    if err != nil {
        http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "user_id":  user.ID,
        "username": user.Username,
        "email":    user.Email,
        "role":     user.Role,
    })
}