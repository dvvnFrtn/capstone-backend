package service

import (
	"context"

	"firebase.google.com/go/v4/auth"
	"github.com/dvvnFrtn/capstone-backend/pkg/errs"
	"github.com/google/uuid"
)

type AuthService interface {
	CreateAccount(ctx context.Context, req CreateAccountInput, claims map[string]interface{}) error
	DeleteAccount(ctx context.Context, uID uuid.UUID) error
	UpdateAccount(ctx context.Context, req UpdateAccountInput) error
}

type firebaseAuthService struct {
	client *auth.Client
}

func NewFirebaseAuthService(client *auth.Client) AuthService {
	return &firebaseAuthService{
		client: client,
	}
}

type CreateAccountInput struct {
	UID      uuid.UUID
	Email    string
	Phone    string
	Password string
}

type UpdateAccountInput struct {
	CreateAccountInput
}

func (s *firebaseAuthService) CreateAccount(ctx context.Context, req CreateAccountInput, claims map[string]interface{}) error {
	const op errs.Op = "service.auth.CreateAccount"

	params := (&auth.UserToCreate{}).PhoneNumber(req.Phone).Password(req.Password).UID(req.UID.String()).EmailVerified(false)
	if req.Email != "" {
		params = params.Email(req.Email)
	}

	user, err := s.client.CreateUser(ctx, params)
	if err != nil {
		return errs.New(op, err, errs.Internal)
	}

	err = s.client.SetCustomUserClaims(ctx, user.UID, claims)
	if err != nil {
		return errs.New(op, err, errs.Internal)
	}

	return nil
}

func (s *firebaseAuthService) DeleteAccount(ctx context.Context, uID uuid.UUID) error {
	const op errs.Op = "service.auth.DeleteAccount"

	if err := s.client.DeleteUser(ctx, uID.String()); err != nil {
		return errs.New(op, err, errs.Internal)
	} else {
		return nil
	}
}

func (s *firebaseAuthService) UpdateAccount(ctx context.Context, req UpdateAccountInput) error {
	const op errs.Op = "service.auth.UpdateAccount"

	params := &auth.UserToUpdate{}

	if req.Email != "" {
		params = params.Email(req.Email)
	}

	if req.Password != "" {
		params = params.Password(req.Password)
	}

	if req.Phone != "" {
		params = params.PhoneNumber(req.Phone)
	}

	if _, err := s.client.UpdateUser(ctx, req.UID.String(), params); err != nil {
		return errs.New(op, err, errs.Internal)
	} else {
		return nil
	}
}
