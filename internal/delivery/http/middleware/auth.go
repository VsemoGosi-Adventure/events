package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const UserIDHeader = "X-User-Id"

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader(UserIDHeader)
		if userIDStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized: missing user identification header",
			})
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized: invalid user id format",
			})
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}

func AuthOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		if userIDStr := c.GetHeader(UserIDHeader); userIDStr != "" {
			if userID, err := uuid.Parse(userIDStr); err == nil {
				c.Set(UserIDKey, userID)
			}
		}
		c.Next()
	}
}
