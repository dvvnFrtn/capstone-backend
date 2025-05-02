package middleware

import (
	"log/slog"
	"os"
	"strings"

	"github.com/dvvnFrtn/capstone-backend/internal/handler/response"
	"github.com/dvvnFrtn/capstone-backend/pkg/errs"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func MustAuthenticated(logger *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		const op errs.Op = "middleware.MustAuthenticated"
		authorization := ctx.GetHeader("Authorization")
		if authorization == "" {
			response.SendRESTError(ctx, logger, errs.New(op, errs.Unauthorize, "authorization header is missing"))
			ctx.Abort()
			return
		}

		if !strings.HasPrefix(authorization, "Bearer ") {
			response.SendRESTError(ctx, logger, errs.New(op, errs.Unauthorize, "authorization header format must be 'Bearer <token>'"))
			ctx.Abort()
			return
		}

		accessToken := authorization[7:]

		claims, err := VerifyToken(accessToken)
		if err != nil {
			response.SendRESTError(ctx, logger, err)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}

func VerifyToken(t string) (jwt.MapClaims, error) {
	const op errs.Op = "middleware.VerifyToken"

	secret := []byte(os.Getenv("SUPABASE_SECRET_KEY"))

	token, err := jwt.Parse(t, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errs.New(op, errs.Internal, "unexpected token siging method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, errs.New(op, errs.Unauthorize, err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errs.New(op, errs.Unauthorize, "invalid token given")
	}
}
