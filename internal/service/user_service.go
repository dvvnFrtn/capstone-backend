package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/dvvnFrtn/capstone-backend/infra/db"
	database "github.com/dvvnFrtn/capstone-backend/infra/db/sqlc"
	"github.com/dvvnFrtn/capstone-backend/pkg/errs"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserService struct {
	authService AuthService
	conn        *pgx.Conn
}

func NewUserService(conn *pgx.Conn, as AuthService) UserService {
	return UserService{
		authService: as,
		conn:        conn,
	}
}

type AdminRegistrationRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Fullname    string `json:"fullname" binding:"required"`
	RtNumber    int32  `json:"rt_number" binding:"required"`
	RwNumber    int32  `json:"rw_number" binding:"required"`
	Subdistrict string `json:"subdistrict" binding:"required"`
	District    string `json:"district" binding:"required"`
	City        string `json:"city" binding:"required"`
	Province    string `json:"province" binding:"required"`
}

type AdminRegistrationResponse struct {
	AdminID     uuid.UUID `json:"admin_id"`
	CommunityID uuid.UUID `json:"community_id"`
	Email       string    `json:"email"`
}

func (s *UserService) AdminRegistration(ctx context.Context, req AdminRegistrationRequest) (*AdminRegistrationResponse, error) {
	const op errs.Op = "service.user.AdminRegistration"

	result, err := s.authService.Signup(SignupRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, errs.New(op, err)
	}

	exists := s.IsUserExists(ctx, result.ID)
	if !exists {
		admID, comID, err := s.createAdminCommunity(ctx, s.conn, result.ID, req)
		if err != nil {
			return nil, errs.New(op, err)
		}
		return &AdminRegistrationResponse{
			AdminID:     admID,
			CommunityID: comID,
			Email:       req.Email,
		}, nil
	} else {
		return nil, nil
	}
}

func (s *UserService) IsUserExists(ctx context.Context, uID uuid.UUID) bool {
	queries := database.New(s.conn)

	if _, err := queries.FindUserByID(ctx, uID); err != nil {
		return false
	}

	return true
}

func (s *UserService) createAdminCommunity(ctx context.Context, conn *pgx.Conn, admID uuid.UUID, req AdminRegistrationRequest) (adm uuid.UUID, com uuid.UUID, err error) {
	const op errs.Op = "service.user.createAdminCommunity"

	var comID uuid.UUID
	err = db.RunTransaction(ctx, conn, func(q *database.Queries) error {
		comID, err = q.InsertCommunity(ctx, database.InsertCommunityParams{
			ID:          uuid.New(),
			RtNumber:    req.RtNumber,
			RwNumber:    req.RwNumber,
			Subdistrict: req.Subdistrict,
			District:    req.District,
			City:        req.City,
			Province:    req.City,
			IsConfirmed: false,
		})
		if err != nil {
			return errs.New(op, errs.Internal, fmt.Errorf("failed insert community: %w", err))
		}
		if _, err = q.InsertUser(ctx, database.InsertUserParams{
			ID:          admID,
			CommunityID: comID,
			Fullname:    req.Fullname,
			Role:        "admin",
			IsConfirmed: false,
		}); err != nil {
			return errs.New(op, errs.Internal, fmt.Errorf("failed insert user: %w", err))
		}

		return nil
	})
	if err != nil {
		return
	}

	return admID, comID, nil
}

func (s *UserService) VerifySignUp(ctx context.Context, req VerifyOTPRequest) (*VerifyOTPResponse, error) {
	const op errs.Op = "service.user.VerifySignUp"

	result, err := s.authService.VerifyOTP(VerifyOTPRequest{
		Type:  "signup",
		Token: req.Token,
		Email: req.Email,
	})
	if err != nil {
		return nil, errs.New(op, err)
	}

	if err := db.RunTransaction(ctx, s.conn, func(q *database.Queries) error {
		if err := q.UpdateUserStatus(ctx, database.UpdateUserStatusParams{
			IsConfirmed: true,
			ID:          result.ID,
		}); err != nil {
			return errs.New(op, errs.Internal, fmt.Errorf("failed to update admin status: %w", err))
		}

		if err := q.UpdateCommunityStatus(ctx, database.UpdateCommunityStatusParams{
			IsConfirmed: true,
			ID:          result.ID,
		}); err != nil {
			return errs.New(op, errs.Internal, fmt.Errorf("failed to update community status: %w", err))
		}
		return nil
	}); err != nil {
		return nil, errs.New(op, err)
	}

	return result, nil
}

func (s *UserService) UserLogin(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	const op errs.Op = "service.user.UserLogin"

	result, err := s.authService.Login(LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, errs.New(op, err)
	}

	if !s.IsUserExists(ctx, result.ID) {
		return nil, errs.New(op, err, errs.Msg("Pengguna tidak ditemukan"))
	}

	return &LoginResponse{
		ID:           result.ID,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

type CommunityResponse struct {
	ID          uuid.UUID `json:"id"`
	RtNumber    int32     `json:"rt_number"`
	RwNumber    int32     `json:"rw_number"`
	Subdistrict string    `json:"subdistrict"`
	District    string    `json:"district"`
	City        string    `json:"city"`
	Province    string    `json:"province"`
}

type UserResponse struct {
	ID        uuid.UUID         `json:"id"`
	Fullname  string            `json:"fullname"`
	Email     string            `json:"email"`
	Role      string            `json:"role"`
	Community CommunityResponse `json:"community"`
}

func (s *UserService) GetAuthenticatedUser(ctx context.Context, claims jwt.MapClaims) (*UserResponse, error) {
	const op errs.Op = "service.user.GetAuthenticatedUser"

	uID := claims["sub"].(string)
	email := claims["email"].(string)

	queries := database.New(s.conn)

	row, err := queries.FindUserByID(ctx, uuid.MustParse(uID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.New(op, err, errs.NotFound, errs.Msg("Pengguna tidak dapat ditemukan"))
		}
		return nil, errs.New(op, err, errs.Internal, fmt.Errorf("failed to select user rows: %w", err))
	}

	return &UserResponse{
		ID:       row.ID,
		Fullname: row.Fullname,
		Email:    email,
		Role:     row.Role,
		Community: CommunityResponse{
			ID:          row.CommunityID,
			RtNumber:    row.RtNumber,
			RwNumber:    row.RwNumber,
			Subdistrict: row.Subdistrict,
			District:    row.District,
			City:        row.City,
			Province:    row.Province,
		},
	}, nil
}
