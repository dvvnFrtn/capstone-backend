package handler

import (
	"log/slog"

	"github.com/dvvnFrtn/capstone-backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, logger *slog.Logger, uh UserHandler, fb middleware.Firebase) {
	// Auth
	r.POST(
		"/api/auth/signup",
		middleware.RequestContext(),
		uh.AdminSignup,
	)

	// Users
	r.POST(
		"/api/users",
		middleware.RequestContext(),
		fb.MustAuthenticated(logger),
		middleware.MustHaveRole(logger, "admin"),
		uh.AdminCreateUser,
	)
	r.GET(
		"/api/users",
		middleware.RequestContext(),
		fb.MustAuthenticated(logger),
		middleware.MustHaveRole(logger, "admin", "pengurus"),
		uh.GetUsersCommunity,
	)
	r.GET(
		"/api/users/:userID",
		middleware.RequestContext(),
		fb.MustAuthenticated(logger),
		middleware.MustHaveRole(logger, "admin", "pengurus", "warga"),
		uh.GetUser,
	)
	r.PATCH(
		"/api/users/:userID",
		middleware.RequestContext(),
		fb.MustAuthenticated(logger),
		middleware.MustHaveRole(logger, "admin"),
		uh.AdminUpdateUser,
	)
	r.DELETE(
		"/api/users/:userID",
		middleware.RequestContext(),
		fb.MustAuthenticated(logger),
		middleware.MustHaveRole(logger, "admin"),
		uh.AdminDeleteUser,
	)
}
