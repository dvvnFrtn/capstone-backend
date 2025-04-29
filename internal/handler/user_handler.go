package handler

import (
	"log/slog"
	"net/http"

	"github.com/dvvnFrtn/capstone-backend/internal/handler/response"
	"github.com/dvvnFrtn/capstone-backend/internal/service"
	"github.com/dvvnFrtn/capstone-backend/pkg/errs"
	"github.com/gin-gonic/gin"
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
		response.SendRESTSuccess(ctx, http.StatusNoContent, "Silahkan cek email anda", nil)
		return
	}
}
