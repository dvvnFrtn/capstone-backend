package handler

import (
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, uh UserHandler) {
	r.POST("/api/auth/signup", uh.AdminSignup)
	r.POST("/api/auth/signup/verify", uh.VerifyOTP)
}
