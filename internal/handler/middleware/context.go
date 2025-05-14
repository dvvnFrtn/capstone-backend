package middleware

import (
	"github.com/dvvnFrtn/capstone-backend/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestContext() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(string(types.RequestIDKey), uuid.New())

		ctx.Next()
	}
}
