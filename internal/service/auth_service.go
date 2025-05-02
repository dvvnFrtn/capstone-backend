package service

import (
	"encoding/json"
	"strings"

	"github.com/dvvnFrtn/capstone-backend/pkg/errs"
	"github.com/google/uuid"
	"github.com/supabase-community/auth-go"
	"github.com/supabase-community/auth-go/types"
)

type AuthService interface {
	Signup(req SignupRequest) (*SignupResponse, error)
	VerifyOTP(req VerifyOTPRequest) (*VerifyOTPResponse, error)
}

type supabaseAuthService struct {
	client auth.Client
}

func NewSupabaseAuthService(client auth.Client) AuthService {
	return &supabaseAuthService{
		client: client,
	}
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func (s *supabaseAuthService) Signup(req SignupRequest) (*SignupResponse, error) {
	const op errs.Op = "service.auth.SignUp"

	result, err := s.client.Signup(types.SignupRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		parsedErr, parseErr := parseSupabaseAuthError(err)
		if parseErr != nil {
			return nil, errs.New(op, errs.Internal, err)
		} else {
			return nil, mapSupabaseAuthError(op, parsedErr, err)
		}
	}
	if len(result.Identities) == 0 {
		return nil, errs.New(op, errs.BadRequest, errs.Msg("Email sudah terdaftar, silahkan login"), err)
	}

	return &SignupResponse{
		ID:    result.ID,
		Email: result.Email,
	}, nil
}

type VerifyOTPRequest struct {
	Type  string `json:"type"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Token string `json:"token"`
}

type VerifyOTPResponse struct {
	ID           uuid.UUID `json:"_"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
}

func (s *supabaseAuthService) VerifyOTP(req VerifyOTPRequest) (*VerifyOTPResponse, error) {
	const op errs.Op = "service.auth.VerifyOTP"

	result, err := s.client.VerifyForUser(types.VerifyForUserRequest{
		Type:       types.VerificationType(req.Type),
		Token:      req.Token,
		Email:      req.Email,
		RedirectTo: "http://localhost:8000",
	})
	if err != nil {
		parsedErr, parseErr := parseSupabaseAuthError(err)
		if parseErr != nil {
			return nil, errs.New(op, errs.Internal, err)
		} else {
			return nil, mapSupabaseAuthError(op, parsedErr, err)
		}
	}
	return &VerifyOTPResponse{
		ID:           result.User.ID,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

type SupabaseAuthError struct {
	Code      int    `json:"code"`
	ErrorCode string `json:"error_code"`
	Message   string `json:"msg"`
}

func (sp *SupabaseAuthError) Error() string {
	return sp.Message
}

func parseSupabaseAuthError(err error) (*SupabaseAuthError, error) {
	if err == nil {
		return nil, nil
	}

	parts := strings.SplitN(err.Error(), ":", 2)
	if len(parts) < 2 {
		return nil, err
	}

	jsonPart := strings.TrimSpace(parts[1])
	var supabaseErr SupabaseAuthError
	if unmarshalErr := json.Unmarshal([]byte(jsonPart), &supabaseErr); unmarshalErr != nil {
		return nil, err
	}

	return &supabaseErr, nil
}

func mapSupabaseAuthError(op errs.Op, serr *SupabaseAuthError, original error) error {
	switch serr.ErrorCode {
	case "otp_expired":
		return errs.New(op, errs.OTPExpired, errs.Msg("OTP telah kadaluwarsa"), original)
	case "over_email_send_rate_limit":
		return errs.New(op, errs.RateLimit, errs.Msg("Terlalu banyak request, tunggu beberapa saat lagi"), original)
	case "over_request_rate_limit":
		return errs.New(op, errs.RateLimit, errs.Msg("Terlalu banyak request, tunggu beberapa saat lagi"), original)
	case "email_address_invalid":
		return errs.New(op, errs.BadRequest, errs.Msg("Alamat email tidak valid"), original)
	case "email_address_not_authorized":
		return errs.New(op, errs.BadRequest, errs.Msg("Alamat email tidak valid"), original)
	default:
		return errs.New(op, errs.Internal, original)
	}
}
