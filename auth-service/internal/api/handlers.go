package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	//"strings"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	models "auth-service/internal/domain"
	db "auth-service/internal/repository"
)

type Handler struct {
	logger *zap.Logger
	db     *db.MongoRepository
}

func NewHandler(db *db.MongoRepository, logger *zap.Logger) *Handler {
	return &Handler{
		logger: logger,
		db:     db,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {

	r.HandleFunc("/login", h.LoginHandler).Methods(http.MethodPost)
	r.HandleFunc("/register", h.RegisterHandler).Methods(http.MethodPost)
	r.HandleFunc("/validate", h.ValidateToken).Methods(http.MethodPost)

}

// RegisterHandler creates a new user in MongoDB with a hashed password.
// Only an authorized admin can access this endpoint.
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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

	_, err = h.db.UsersCollection().InsertOne(context.Background(), models.UserDetails{
		ID:            primitive.NewObjectID(),
		Username:      creds.Username,
		Role:          "user",
		FirstName:     creds.FirstName,
		LastName:      creds.LastName,
		Patronymic:    creds.Patronymic,
		PasswordHash:  string(hashed),
		RefreshToken:  "",
		RTTokenExpiry: time.Date(0,0,0,0,0,0,0, time.Now().Location()), //time.Now().Add(1 * 24 * time.Hour), // 7 days
		ATTokenExpiry: time.Date(0,0,0,0,0,0,0, time.Now().Location()), //time.Now().Add(7 * 24 * time.Hour), // 7 days
	})
	if err != nil {
		http.Error(w, "Conflict: user may already exist", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "UserDetails registered successfully")
}

// LoginHandler verifies credentials and returns a JWT on success.
// The JWT includes a "role" claim that should be set appropriately, e.g., "admin" or "user".
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
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
	var dbUser models.UserDetails
	err := h.db.UsersCollection().FindOne(context.Background(), bson.M{"login": creds.Username}).Decode(&dbUser)
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
	//userRole := "admin" // or "user", based on your own logic
	accessToken, err := generateAccessToken(&dbUser)
	if err != nil {
		h.jsonResponse(w, "Access token generation error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		h.jsonResponse(w, "Refresh token generation error", http.StatusInternalServerError)
		return
	}

	result, err := h.db.UsersCollection().UpdateOne(
		context.Background(),
		bson.M{"_id": dbUser.ID}, // Use _id instead of login
		bson.M{"$set": bson.M{
			"refresh_token": refreshToken,
			"rt_token_expiry": time.Now().Add(1 * 24 * time.Hour),
			"at_token_expiry": time.Now().Add(7 * 24 * time.Hour),
		},
	},

		
	)
	if err != nil {
		h.logger.Error("Failed to update user refresh token", zap.Error(err))
		h.jsonResponse(w, "Refresh token update error", http.StatusInternalServerError)
		return
	}
	if result.MatchedCount == 0 {
		h.jsonResponse(w, "User not found", http.StatusNotFound)
		return
	}

	// Then store in refresh_tokens collection
	refreshTokenData := &models.RefreshTokenData{
		Token:     refreshToken,
		UserID:    dbUser.ID.Hex(), // Convert ObjectID to string
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.db.StoreRefreshToken(refreshTokenData); err != nil {
		h.logger.Error("Failed to store refresh token", zap.Error(err))
		h.jsonResponse(w, "Refresh token store error", http.StatusInternalServerError)
		return
	}

	h.db.DeleteRefreshTokensByUser(dbUser.ID.Hex(), refreshToken)
	
	var response = models.UserResponce{
		UserDetails:  dbUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	h.jsonResponse(w, response, http.StatusOK)
	h.logger.Info("User logged in", zap.String("username", creds.Username))
}

func (h *Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "No token provided", http.StatusUnauthorized)
		return
	}

	// Validate token
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Return validation result
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":  true,
		"claims": claims,
	})
}

func (h *Handler) jsonResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response",
			zap.Error(err))
	}
}

func generateAccessToken(user *models.UserDetails) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // day
	})
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
