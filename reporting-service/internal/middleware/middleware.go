package middleware

import (
	"log"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
    authServiceURL string
}

func NewAuthMiddleware(authURL string) *AuthMiddleware {
    return &AuthMiddleware{authServiceURL: authURL}
}

func CustomMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Логирование начала запроса
		startTime := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path
		log.Printf("Started %s %s", method, path)

		// Проверка токена
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing Authorization token",
			})
			c.Abort()
			return
		}

		// Передать управление следующему middleware
		c.Next()

		// Логирование завершения запроса
		statusCode := c.Writer.Status()
		duration := time.Since(startTime)
		log.Printf("Completed %s %s with %d in %v", method, path, statusCode, duration)
	}
}

func (am *AuthMiddleware) Validate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get token from request
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Call auth service
        req, _ := http.NewRequest("POST", am.authServiceURL+"/validate", nil)
        req.Header.Set("Authorization", token)
        
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil || resp.StatusCode != http.StatusOK {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Token is valid, proceed
        next.ServeHTTP(w, r)
    })
}