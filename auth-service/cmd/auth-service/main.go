package main

import (
    "context"
    "log"
    "net/http"
    "os"

    db "auth-service/internal/repository"
    "auth-service/internal/api"
)

func main() {
    // Connect to Mongo
    mongoURI := os.Getenv("MONGO_URI")
    if mongoURI == "" {
        mongoURI = "mongodb://localhost:27017"
    }
    if err := db.ConnectMongo(context.Background(), mongoURI); err != nil {
        log.Fatal("MongoDB connection error:", err)
    }

    // HTTP routes
    http.HandleFunc("/register", handlers.RegisterHandler)
    http.HandleFunc("/login", handlers.LoginHandler)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }
    log.Printf("Auth service running on :%s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}