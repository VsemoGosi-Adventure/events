package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	UserIDKey    = "user_id"
	RequestIDKey = "request_id"
)

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}
