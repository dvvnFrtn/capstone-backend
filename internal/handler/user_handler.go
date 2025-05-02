package handler

import (
	"log/slog"
	"net/http"

	"github.com/dvvnFrtn/capstone-backend/internal/handler/response"
	"github.com/dvvnFrtn/capstone-backend/internal/service"
	"github.com/dvvnFrtn/capstone-backend/pkg/errs"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type UserHandler struct {
	userService service.UserService
	logger      *slog.Logger
}

func NewUserHandler(logger *slog.Logger, us service.UserService) UserHandler {
	return UserHandler{
		userService: us,
		logger:      logger,
	}
}

func (h *UserHandler) AdminSignup(ctx *gin.Context) {
	const op errs.Op = "handler.user.AdminSignup"

	var req service.AdminRegistrationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.SendRESTError(ctx, h.logger, errs.New(op, errs.BadRequest, errs.Msg("Request tidak valid"), err))
		return
	}

	res, err := h.userService.AdminRegistration(ctx, req)
	if err != nil {
		response.SendRESTError(ctx, h.logger, err)
		return
	}

	if res != nil {
		response.SendRESTSuccess(ctx, http.StatusOK, "Registrasi berhasil, silahkan cek email anda", res)
		return
	} else {
		response.SendRESTSuccess(ctx, http.StatusOK, "Registrasi berhasil, silahkan cek email anda", nil)
		return
	}
}

func (h *UserHandler) VerifyOTP(ctx *gin.Context) {
	const op errs.Op = "handler.user.VerifyOTP"

	var req service.VerifyOTPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.SendRESTError(ctx, h.logger, errs.New(op, errs.BadRequest, errs.Msg("Request tidak valid"), err))
		return
	}

	res, err := h.userService.VerifySignUp(ctx, req)
	if err != nil {
		response.SendRESTError(ctx, h.logger, err)
		return
	}

	response.SendRESTSuccess(ctx, http.StatusOK, "Verifikasi email telah berhasil", res)
}

func (h *UserHandler) UserLogin(ctx *gin.Context) {
	const op errs.Op = "handler.user.UserLogin"

	var req service.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.SendRESTError(ctx, h.logger, errs.New(op, errs.BadRequest, errs.Msg("Request tidak valid"), err))
		return
	}

	res, err := h.userService.UserLogin(ctx, req)
	if err != nil {
		response.SendRESTError(ctx, h.logger, err)
		return
	}

	response.SendRESTSuccess(ctx, http.StatusOK, "Login telah berhasil", res)
}

func (h *UserHandler) GetAuthUser(ctx *gin.Context) {
	const op errs.Op = "handler.user.GetAuthUser"

	raw, exists := ctx.Get("claims")
	if !exists {
		response.SendRESTError(ctx, h.logger, errs.New(op, errs.Unauthorize, "missing claims"))
		return
	}
	claims, ok := raw.(jwt.MapClaims)
	if !ok {
		response.SendRESTError(ctx, h.logger, errs.New(op, errs.Unauthorize, "invalid claims"))
		return
	}

	res, err := h.userService.GetAuthenticatedUser(ctx, claims)
	if err != nil {
		response.SendRESTError(ctx, h.logger, err)
		return
	}

	response.SendRESTSuccess(ctx, http.StatusOK, "Berhasil memuat data pengguna", res)
}
