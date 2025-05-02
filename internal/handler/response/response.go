package response

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/dvvnFrtn/capstone-backend/pkg/errs"
	"github.com/gin-gonic/gin"
)

type RESTResponse struct {
	Message string      `json:"message"`
	Code    string      `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func SendRESTSuccess(ctx *gin.Context, status int, msg string, data interface{}) {
	ctx.JSON(status, RESTResponse{
		Message: msg,
		Data:    data,
	})
}

func SendRESTError(ctx *gin.Context, logger *slog.Logger, err error) {
	var (
		resp   RESTResponse
		apperr *errs.Error
		status int
	)

	if errors.As(err, &apperr) {
		resp.Message = string(apperr.Msg)
		resp.Code = apperr.Code.String()
		status = mapAppError(apperr)
		if resp.Message == "" && status != 500 {
			resp.Message = apperr.Error()
		}
	} else {
		resp.Message = "Terjadi kesalahan pada server"
		resp.Code = errs.Internal.String()
		status = http.StatusInternalServerError
	}

	logger.Error(resp.Message, "code", resp.Code, "stack", errs.OpStack(err), "err", err)
	ctx.JSON(status, resp)
}

func mapAppError(err *errs.Error) int {
	switch err.Code {
	case errs.RateLimit:
		return http.StatusTooManyRequests
	case errs.BadRequest:
		return http.StatusBadRequest
	case errs.Conflict:
		return http.StatusConflict
	case errs.Forbidden:
		return http.StatusForbidden
	case errs.NotFound:
		return http.StatusNotFound
	case errs.OTPExpired:
		return http.StatusUnauthorized
	case errs.Unauthorize:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
