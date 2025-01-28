package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type UserDetails struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username      string             `bson:"login" json:"login"`
	Role          string             `bson:"role" json:"role"`
	FirstName     string             `bson:"name" json:"name"`
	LastName      string             `bson:"surname" json:"surname"`
	Patronymic    string             `bson:"patronymic" json:"patronymic"`
	RefreshToken  string             `bson:"refresh_token" json:"-"`
	RTTokenExpiry time.Time          `bson:"rt_token_expiry" json:"-"`
	ATTokenExpiry time.Time          `bson:"at_token_expiry" json:"-"`
	PasswordHash  string             `bson:"password_hash" json:"-"`
}

type Credentials struct {
	Username   string `json:"login"`
	Role       string `json:"role"`
	Password   string `json:"password"`
	FirstName  string `json:"name"`
	LastName   string `json:"surname"`
	Patronymic string `json:"patronymic"`
}

type UserResponce struct {
	UserDetails  UserDetails `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenData struct {
	Token     string    `bson:"token"`
	UserID    string    `bson:"user_id"`
	ExpiresAt time.Time `bson:"expires_at"`
}
