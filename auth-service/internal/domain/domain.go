package domain

type User struct {
    Username     string `bson:"username"`
    PasswordHash string `bson:"password_hash"`
}

type Credentials struct {
    Username string `json:"username"`
    Password string `json:"password"`
}