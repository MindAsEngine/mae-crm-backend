package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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
