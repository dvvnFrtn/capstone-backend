package service

import (
	"context"
	"errors"

	"github.com/dvvnFrtn/capstone-backend/infra/db"
	database "github.com/dvvnFrtn/capstone-backend/infra/db/sqlc"
	"github.com/dvvnFrtn/capstone-backend/internal/handler/middleware"
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

func (service *UserService) EnsureEmailOrPhoneUnique(ctx context.Context, email, phone string) error {
	const op errs.Op = "service.user.EnsureEmailOrPhoneUnique"
	queries := database.New(service.conn)

	if email != "" {
		emailExists, err := queries.IsEmailExists(ctx, pgtype.Text{String: email, Valid: true})
		if err != nil {
			return errs.New(op, errs.Internal, err)
		}
		if emailExists {
			return errs.New(op, errs.Conflict, "Email sudah terdaftar")
		}
	}

	if phone != "" {
		phoneExists, err := queries.IsPhoneExists(ctx, pgtype.Text{String: phone, Valid: true})
		if err != nil {
			return errs.New(op, errs.Internal, err)
		}
		if phoneExists {
			return errs.New(op, errs.Conflict, "Nomor sudah terdaftar")
		}
	}

	return nil
}

func (service *UserService) AdminRegistration(ctx context.Context, req AdminRegistrationRequest) (*AdminRegistrationResponse, error) {
	const op errs.Op = "service.user.AdminRegistration"

	if err := service.EnsureEmailOrPhoneUnique(ctx, req.Email, req.Phone); err != nil {
		return nil, errs.New(op, err)
	}

	admID, comID, err := service.createAdminCommunity(ctx, service.conn, uuid.New(), req)
	if err != nil {
		return nil, errs.New(op, err)
	}

	metadata := map[string]interface{}{
		"role":         "admin",
		"community_id": comID,
	}

	err = service.authService.CreateAccount(ctx, CreateAccountInput{UID: admID, Email: req.Email, Phone: req.Phone, Password: req.Password}, metadata)
	if err != nil {
		return nil, errs.New(op, err)
	}

	return &AdminRegistrationResponse{
		AdminID:     admID,
		CommunityID: comID,
		Email:       req.Email,
	}, nil
}

func (s *UserService) IsUserExists(ctx context.Context, req IsUserExistsInput) bool {
	queries := database.New(s.conn)

	if _, err := queries.FindUserByID(ctx, database.FindUserByIDParams{
		ID:    pgtype.UUID{Bytes: req.ID, Valid: req.ID != uuid.Nil},
		Email: pgtype.Text{String: req.Email, Valid: req.Email != ""},
		Phone: pgtype.Text{String: req.Phone, Valid: req.Phone != ""},
	}); err != nil {
		return false
	}

	return true
}

func (s *UserService) createAdminCommunity(ctx context.Context, conn *pgx.Conn, admID uuid.UUID, req AdminRegistrationRequest) (adm uuid.UUID, com uuid.UUID, err error) {
	const op errs.Op = "service.user.createAdminCommunity"

	var comID uuid.UUID
	err = db.RunTransaction(ctx, conn, func(queries *database.Queries) error {
		comID, err = queries.InsertCommunity(ctx, database.InsertCommunityParams{
			ID:          uuid.New(),
			RtNumber:    req.RtNumber,
			RwNumber:    req.RwNumber,
			Subdistrict: req.Subdistrict,
			District:    req.District,
			City:        req.City,
			Province:    req.City,
		})
		if err != nil {
			return errs.New(op, errs.Internal, err)
		}
		if _, err = queries.InsertUser(ctx, database.InsertUserParams{
			ID:          admID,
			CommunityID: comID,
			Fullname:    req.Fullname,
			Email:       pgtype.Text{String: req.Email, Valid: true},
			Phone:       pgtype.Text{String: req.Phone, Valid: true},
			Address:     pgtype.Text{String: req.Address, Valid: true},
			Role:        "admin",
		}); err != nil {
			return errs.New(op, errs.Internal, err)
		}

		return nil
	})
	if err != nil {
		return
	}

	return admID, comID, nil
}

func (service *UserService) AdminCreateUser(ctx context.Context, claims *middleware.UserClaims, req AdminCreateUserRequest) (*IDResponse, error) {
	const op errs.Op = "service.user.AdminCreateUser"

	if err := service.EnsureEmailOrPhoneUnique(ctx, "", req.Phone); err != nil {
		return nil, errs.New(op, err)
	}

	var createdID uuid.UUID
	if err := db.RunTransaction(ctx, service.conn, func(q *database.Queries) error {
		uID, err := q.InsertUser(ctx, database.InsertUserParams{
			ID:          uuid.New(),
			CommunityID: uuid.MustParse(claims.CommunityID),
			Fullname:    req.Fullname,
			Phone:       pgtype.Text{String: req.Phone, Valid: true},
			Address:     pgtype.Text{String: req.Address, Valid: req.Address != ""},
			Email:       pgtype.Text{String: req.Email, Valid: req.Email != ""},
			Role:        req.Role,
		})
		if err != nil {
			return errs.New(op, errs.Internal, err)
		}

		metadata := map[string]interface{}{
			"role":         req.Role,
			"community_id": uuid.MustParse(claims.CommunityID),
		}

		err = service.authService.CreateAccount(ctx, CreateAccountInput{UID: uID, Phone: req.Phone, Password: req.Password}, metadata)
		if err != nil {
			return errs.New(op, err)
		}

		createdID = uID

		return nil
	}); err != nil {
		return nil, err
	}

	return &IDResponse{ID: createdID}, nil
}

func (service *UserService) GetUser(ctx context.Context, claims *middleware.UserClaims, uID uuid.UUID) (*UserResponse, error) {
	const op errs.Op = "service.user.GetUser"

	queries := database.New(service.conn)

	var result database.FindUserByIDRow
	switch claims.Role {
	case "admin":
		row, err := queries.FindUserByID(ctx, database.FindUserByIDParams{
			ID:          pgtype.UUID{Bytes: uID, Valid: true},
			CommunityID: pgtype.UUID{Bytes: uuid.MustParse(claims.CommunityID), Valid: true},
		})
		if err != nil {
			return nil, errs.New(op, errs.Internal, err)
		} else {
			result = row
		}
	case "warga":
		if claims.UID != uID.String() {
			return nil, errs.New(op, errs.Forbidden, "Tidak dapat mengambil data pengguna lain")
		}
		row, err := queries.FindUserByID(ctx, database.FindUserByIDParams{
			ID: pgtype.UUID{Bytes: uID, Valid: true},
		})
		if err != nil {
			return nil, errs.New(op, errs.Internal, err)
		} else {
			result = row
		}
	}

	return toUserResponse(result), nil
}

func (service *UserService) AdminUpdateUser(ctx context.Context, claims *middleware.UserClaims, uID uuid.UUID, req AdminUpdateUserRequest) (*IDResponse, error) {
	const op errs.Op = "service.user.AdminUpdateUser"

	if err := service.EnsureEmailOrPhoneUnique(ctx, req.Email, req.Phone); err != nil {
		return nil, errs.New(op, err)
	}

	var updatedID uuid.UUID
	if err := db.RunTransaction(ctx, service.conn, func(q *database.Queries) error {
		if _, err := q.FindUserByID(ctx, database.FindUserByIDParams{
			ID:          pgtype.UUID{Bytes: uID, Valid: true},
			CommunityID: pgtype.UUID{Bytes: uuid.MustParse(claims.CommunityID), Valid: true},
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.New(op, errs.NotFound, "Pengguna tidak dapat ditemukan")
			}
			return errs.New(op, errs.Internal, err)
		}

		uID, err := q.UpdateUser(ctx, database.UpdateUserParams{
			ID:       uID,
			Fullname: pgtype.Text{String: req.Fullname, Valid: req.Fullname != ""},
			Email:    pgtype.Text{String: req.Email, Valid: req.Email != ""},
			Phone:    pgtype.Text{String: req.Phone, Valid: req.Phone != ""},
			Address:  pgtype.Text{String: req.Address, Valid: req.Address != ""},
			Role:     pgtype.Text{String: req.Role, Valid: req.Role != ""},
		})
		if err != nil {
			return errs.New(op, errs.Internal, err)
		}

		if err = service.authService.UpdateAccount(ctx, UpdateAccountInput{
			CreateAccountInput: CreateAccountInput{
				UID:      uID,
				Email:    req.Email,
				Phone:    req.Phone,
				Password: req.Password,
			},
		}); err != nil {
			return errs.New(op, err)
		}

		updatedID = uID

		return nil
	}); err != nil {
		return nil, err
	}

	return &IDResponse{ID: updatedID}, nil
}

func (service *UserService) AdminDeleteUser(ctx context.Context, claims *middleware.UserClaims, uID uuid.UUID) error {
	const op errs.Op = "service.user.AdminDeleteUser"

	if err := db.RunTransaction(ctx, service.conn, func(q *database.Queries) error {
		if _, err := q.FindUserByID(ctx, database.FindUserByIDParams{
			ID:          pgtype.UUID{Bytes: uID, Valid: true},
			CommunityID: pgtype.UUID{Bytes: uuid.MustParse(claims.CommunityID), Valid: true},
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.New(op, errs.NotFound, "Pengguna tidak dapat ditemukan")
			}
			return errs.New(op, errs.Internal, err)
		}

		if err := q.DeleteUser(ctx, uID); err != nil {
			return errs.New(op, errs.Internal, err)
		}

		if err := service.authService.DeleteAccount(ctx, uID); err != nil {
			return errs.New(op, err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (service *UserService) GetUserFromCommunity(ctx context.Context, claims *middleware.UserClaims) ([]*UserResponse, error) {
	const op errs.Op = "service.user.GetUserFromCommunity"

	queries := database.New(service.conn)

	rows, err := queries.FindUserByCommunityID(ctx, uuid.MustParse(claims.CommunityID))
	if err != nil {
		return nil, errs.New(op, errs.Internal, err)
	}

	var responses []*UserResponse
	for _, row := range rows {
		rowx := database.FindUserByIDRow(row)
		responses = append(responses, toUserResponse(rowx))
	}

	return responses, nil
}

type AdminUpdateUserRequest struct {
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Fullname string `json:"fullname"`
	Role     string `json:"role"`
}

type IDResponse struct {
	ID uuid.UUID `json:"id"`
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
	Phone     string            `json:"phone"`
	Address   string            `json:"address"`
	Role      string            `json:"role"`
	Community CommunityResponse `json:"community"`
}

func toUserResponse(row database.FindUserByIDRow) *UserResponse {
	return &UserResponse{
		ID:       row.ID,
		Fullname: row.Fullname,
		Email:    row.Email.String,
		Phone:    row.Phone.String,
		Address:  row.Address.String,
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
	}
}

type AdminRegistrationRequest struct {
	Email       string `json:"email" binding:"required,min=10"`
	Password    string `json:"password" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	Address     string `json:"address" binding:"required"`
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

type IsUserExistsInput struct {
	ID    uuid.UUID
	Email string
	Phone string
}

type AdminCreateUserRequest struct {
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Fullname string `json:"fullname" binding:"required"`
	Role     string `json:"role" binding:"required"`
}
