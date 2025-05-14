package handler

import (
	"log/slog"
	"net/http"

	"github.com/dvvnFrtn/capstone-backend/internal/handler/middleware"
	"github.com/dvvnFrtn/capstone-backend/internal/handler/response"
	"github.com/dvvnFrtn/capstone-backend/internal/service"
	"github.com/dvvnFrtn/capstone-backend/pkg/errs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	response.SendRESTSuccess(ctx, http.StatusOK, "Registrasi berhasil, silahkan cek email anda", res)
}

func (h *UserHandler) AdminCreateUser(ctx *gin.Context) {
	const op errs.Op = "handler.user.AdminCreateUser"

	var req service.AdminCreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.SendRESTError(ctx, h.logger, errs.New(op, errs.BadRequest, errs.Msg("Request tidak valid"), err))
		return
	}

	claims := middleware.GetUserClaims(ctx)

	res, err := h.userService.AdminCreateUser(ctx, claims, req)
	if err != nil {
		response.SendRESTError(ctx, h.logger, err)
		return
	}

	response.SendRESTSuccess(ctx, http.StatusCreated, "Akun berhasil disimpan", res)
}

func (h *UserHandler) GetUser(ctx *gin.Context) {
	const op errs.Op = "handler.user.GetUser"

	userID := ctx.Param("userID")
	claims := middleware.GetUserClaims(ctx)

	res, err := h.userService.GetUser(ctx, claims, uuid.MustParse(userID))
	if err != nil {
		response.SendRESTError(ctx, h.logger, err)
		return
	}

	response.SendRESTSuccess(ctx, http.StatusCreated, "Akun berhasil dimuat", res)
}

func (h *UserHandler) GetUsersCommunity(ctx *gin.Context) {
	const op errs.Op = "handler.user.GetUsersCommunity"

	claims := middleware.GetUserClaims(ctx)

	res, err := h.userService.GetUserFromCommunity(ctx, claims)
	if err != nil {
		response.SendRESTError(ctx, h.logger, err)
		return
	}

	response.SendRESTSuccess(ctx, http.StatusCreated, "Akun berhasil dimuat", res)
}

func (h *UserHandler) AdminUpdateUser(ctx *gin.Context) {
	const op errs.Op = "handler.user.AdminUpdateUser"

	var req service.AdminUpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.SendRESTError(ctx, h.logger, errs.New(op, errs.BadRequest, errs.Msg("Request tidak valid"), err))
		return
	}

	userID := ctx.Param("userID")
	claims := middleware.GetUserClaims(ctx)

	res, err := h.userService.AdminUpdateUser(ctx, claims, uuid.MustParse(userID), req)
	if err != nil {
		response.SendRESTError(ctx, h.logger, err)
		return
	}

	response.SendRESTSuccess(ctx, http.StatusCreated, "Akun berhasil diperbarui", res)
}

func (h *UserHandler) AdminDeleteUser(ctx *gin.Context) {
	const op errs.Op = "handler.user.AdminDeleteUser"

	userID := ctx.Param("userID")
	claims := middleware.GetUserClaims(ctx)

	if err := h.userService.AdminDeleteUser(ctx, claims, uuid.MustParse(userID)); err != nil {
		response.SendRESTError(ctx, h.logger, err)
		return
	}

	response.SendRESTSuccess(ctx, http.StatusOK, "Akun berhasil dihapus", nil)
}
