package service

import (
	"context"
	"fmt"

	"github.com/dvvnFrtn/capstone-backend/infra/db"
	database "github.com/dvvnFrtn/capstone-backend/infra/db/sqlc"
	"github.com/dvvnFrtn/capstone-backend/pkg/errs"
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
