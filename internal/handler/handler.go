package handler

import (
	"github.com/dvvnFrtn/capstone-backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, uh UserHandler) {
	r.POST("/api/auth/signup", uh.AdminSignup)
	r.POST("/api/auth/signup/verify", uh.VerifyOTP)
	r.POST("/api/auth/signin", uh.UserLogin)
	r.GET("/api/auth/me", middleware.MustAuthenticated(uh.logger), uh.GetAuthUser)
}
