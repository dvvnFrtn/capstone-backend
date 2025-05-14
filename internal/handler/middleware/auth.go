package middleware

import (
	"log/slog"
	"slices"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/dvvnFrtn/capstone-backend/internal/handler/response"
	"github.com/dvvnFrtn/capstone-backend/pkg/errs"
	"github.com/gin-gonic/gin"
)

type UserClaims struct {
	UID         string
	Role        string
	CommunityID string
}

type Firebase struct {
	Client *auth.Client
}

func NewFirebaseAuthMiddleware(client *auth.Client) *Firebase {
	return &Firebase{Client: client}
}

func (f *Firebase) MustAuthenticated(logger *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		const op errs.Op = "middleware.firebase.MustAuthenticated"

		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.SendRESTError(ctx, logger, errs.New(op, errs.Unauthorize, "missing or invalid authorization header"))
			ctx.Abort()
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := f.Client.VerifyIDToken(ctx, idToken)
		if err != nil {
			response.SendRESTError(ctx, logger, errs.New(op, errs.Unauthorize, "invalid token"))
			ctx.Abort()
			return
		}

		userClaims := &UserClaims{
			UID:         token.UID,
			Role:        toString(token.Claims["role"]),
			CommunityID: toString(token.Claims["community_id"]),
		}

		ctx.Set("claims", userClaims)
		ctx.Next()
	}
}

func MustHaveRole(logger *slog.Logger, allowedRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		const op errs.Op = "middleware.auth.MustHaveRole"
		user := GetUserClaims(ctx)

		if slices.Contains(allowedRoles, user.Role) {
			ctx.Next()
			return
		}

		response.SendRESTError(ctx, logger, errs.New(op, errs.Forbidden, "insufficient role"))
		ctx.Abort()
	}
}

func toString(val interface{}) string {
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func GetUserClaims(ctx *gin.Context) *UserClaims {
	user, exists := ctx.Get("claims")
	if !exists {
		return nil
	} else {
		return user.(*UserClaims)
	}
}
