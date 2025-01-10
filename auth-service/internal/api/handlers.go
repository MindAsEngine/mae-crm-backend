package handlers;

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "strings"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"

    db "auth-service/internal/repository"
    models "auth-service/internal/domain"
)

// JWT secret key (in production, store this securely)
var jwtKey = []byte("supersecretkey")

// RegisterHandler creates a new user in MongoDB with a hashed password.
// Only an authorized admin can access this endpoint.
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return
    }

    // Check for Bearer token
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        http.Error(w, "Unauthorized: no token", http.StatusUnauthorized)
        return
    }
    // "Bearer <jwt>"
    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        http.Error(w, "Unauthorized: invalid token format", http.StatusUnauthorized)
        return
    }
    tokenStr := parts[1]

    // Parse and validate token
    claims := jwt.MapClaims{}
    token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })
    if err != nil || !token.Valid {
        http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
        return
    }

    // Check role claim
    role, ok := claims["role"].(string)
    if !ok || role != "admin" {
        http.Error(w, "Forbidden: admin only", http.StatusForbidden)
        return
    }

    var creds models.Credentials
    if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    // Hash password
    hashed, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Internal Error", http.StatusInternalServerError)
        return
    }

    // Insert new user into MongoDB
    _, err = db.UsersCollection().InsertOne(context.Background(), models.User{
        Username:     creds.Username,
        PasswordHash: string(hashed),
    })
    if err != nil {
        http.Error(w, "Conflict: user may already exist", http.StatusConflict)
        return
    }

    w.WriteHeader(http.StatusCreated)
    fmt.Fprint(w, "User registered successfully")
}

// LoginHandler verifies credentials and returns a JWT on success.
// The JWT includes a "role" claim that should be set appropriately, e.g., "admin" or "user".
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return
    }

    var creds models.Credentials
    if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    // Find user in MongoDB
    var dbUser models.User
    err := db.UsersCollection().FindOne(context.Background(), bson.M{"username": creds.Username}).Decode(&dbUser)
    if err == mongo.ErrNoDocuments {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    } else if err != nil {
        http.Error(w, "Internal Error", http.StatusInternalServerError)
        return
    }

    // Compare password
    if err := bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(creds.Password)); err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Generate JWT (hardcoded role for demo: "admin" or "user")
    userRole := "admin" // or "user", based on your own logic
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": creds.Username,
        "role":     userRole,
        "exp":      time.Now().Add(time.Hour).Unix(),
    })
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        http.Error(w, "Internal Error", http.StatusInternalServerError)
        return
    }

    fmt.Fprint(w, tokenString)
}