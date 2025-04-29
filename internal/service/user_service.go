package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dvvnFrtn/capstone-backend/infra/db"
	database "github.com/dvvnFrtn/capstone-backend/infra/db/sqlc"
	"github.com/dvvnFrtn/capstone-backend/pkg/errs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	Fullname    string    `json:"fullname"`
	DOB         time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	RtNumber    int32     `json:"rt_number"`
	RwNumber    int32     `json:"rw_number"`
	Subdistrict string    `json:"subdistrict"`
	District    string    `json:"district"`
	City        string    `json:"city"`
	Province    string    `json:"province"`
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
			Dob:         pgtype.Date{Time: req.DOB, Valid: !req.DOB.IsZero()},
			Gender:      req.Gender,
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
