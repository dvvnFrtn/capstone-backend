package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dvvnFrtn/capstone-backend/config"
	"github.com/dvvnFrtn/capstone-backend/infra/db"
	database "github.com/dvvnFrtn/capstone-backend/infra/db/sqlc"
	"github.com/dvvnFrtn/capstone-backend/internal/service"
	"github.com/dvvnFrtn/capstone-backend/internal/types"
	"github.com/dvvnFrtn/capstone-backend/pkg/testutil"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type TestSuiteUserService struct {
	suite.Suite
	cfg       *config.DB
	container *postgres.PostgresContainer
}

func (ts *TestSuiteUserService) SetupSuite() {
	err := godotenv.Load("../../.env")
	require.NoError(ts.T(), err)

	cfg := config.Database()
	container, err := testutil.SetupTestDatabase(context.Background(), &cfg)
	require.NoError(ts.T(), err)
	ts.cfg = &cfg
	ts.container = container
}

func (ts *TestSuiteUserService) TearDownTest() {
	err := ts.container.Restore(context.Background())
	require.NoError(ts.T(), err)
}

func dummyAdminRegistrationRequest(email, passw string) service.AdminRegistrationRequest {
	return service.AdminRegistrationRequest{
		Email:       email,
		Password:    passw,
		Fullname:    "user test",
		DOB:         time.Date(2004, 05, 28, 0, 0, 0, 0, time.Local),
		Gender:      "laki-laki",
		RtNumber:    1,
		RwNumber:    1,
		Subdistrict: "test",
		District:    "test",
		City:        "test",
		Province:    "test",
	}
}

func (ts *TestSuiteUserService) TestUserService_SignUp_Success() {
	var (
		ctx           = context.WithValue(context.Background(), types.RequestIDKey, uuid.New())
		ctrl          = gomock.NewController(ts.T())
		expectedAdmID = uuid.New()
		expectedEmail = "test@gmail.com"
	)
	defer ctrl.Finish()

	conn, err := db.NewPostgreConn(ctx, ts.cfg)
	require.NoError(ts.T(), err)
	defer conn.Close(ctx)

	mockAuth := service.NewMockAuthService(ctrl)
	mockAuth.EXPECT().
		Signup(service.SignupRequest{Email: expectedEmail, Password: "123"}).
		Return(&service.SignupResponse{
			ID:    expectedAdmID,
			Email: expectedEmail,
		}, nil)

	userService := service.NewUserService(conn, mockAuth)

	result, err := userService.AdminRegistration(
		ctx,
		dummyAdminRegistrationRequest(expectedEmail, "123"),
	)

	assert.NoError(ts.T(), err)
	assert.Equal(ts.T(), expectedAdmID, result.AdminID)
	assert.Equal(ts.T(), expectedEmail, result.Email)
}

func (ts *TestSuiteUserService) TestUserService_SignUp_Failed() {
	var (
		ctx  = context.WithValue(context.Background(), types.RequestIDKey, uuid.New())
		ctrl = gomock.NewController(ts.T())
	)
	defer ctrl.Finish()

	conn, err := db.NewPostgreConn(ctx, ts.cfg)
	require.NoError(ts.T(), err)
	defer conn.Close(ctx)

	mockAuth := service.NewMockAuthService(ctrl)
	mockAuth.EXPECT().
		Signup(gomock.Any()).
		Return(nil, errors.New("some-errors"))

	userService := service.NewUserService(conn, mockAuth)

	result, err := userService.AdminRegistration(
		ctx,
		dummyAdminRegistrationRequest("test@gmail.com", "123"),
	)

	assert.Error(ts.T(), err)
	assert.Nil(ts.T(), result)
}

func (ts *TestSuiteUserService) TestUserService_SignUp_ShouldNotCreateNewAdmin() {
	var (
		ctx           = context.WithValue(context.Background(), types.RequestIDKey, uuid.New())
		expectedAdmID = uuid.New()
		expectedComID = uuid.New()
		expectedEmail = "test@gmail.com"
		ctrl          = gomock.NewController(ts.T())
	)
	defer ctrl.Finish()

	conn, err := db.NewPostgreConn(context.Background(), ts.cfg)
	require.NoError(ts.T(), err)
	defer conn.Close(ctx)

	err = seedAdminCommunity(conn, expectedAdmID, expectedComID)
	require.NoError(ts.T(), err)

	mockAuth := service.NewMockAuthService(ctrl)
	mockAuth.EXPECT().
		Signup(service.SignupRequest{Email: expectedEmail, Password: "123"}).
		Return(&service.SignupResponse{ID: expectedAdmID, Email: expectedEmail}, nil)

	userService := service.NewUserService(conn, mockAuth)

	result, err := userService.AdminRegistration(
		ctx,
		dummyAdminRegistrationRequest(expectedEmail, "123"),
	)

	assert.NoError(ts.T(), err)
	assert.Nil(ts.T(), result)
}

func seedAdminCommunity(conn *pgx.Conn, admID, comID uuid.UUID) error {
	queries := database.New(conn)
	_, err := queries.InsertCommunity(context.Background(), database.InsertCommunityParams{
		ID:          comID,
		RtNumber:    3,
		RwNumber:    3,
		Subdistrict: "test",
		District:    "test",
		City:        "test",
		Province:    "test",
		IsConfirmed: false,
	})
	if err != nil {
		return err
	}
	_, err = queries.InsertUser(context.Background(), database.InsertUserParams{
		ID:          admID,
		CommunityID: comID,
		Fullname:    "test",
		Dob:         pgtype.Date{Time: time.Now(), Valid: true},
		Gender:      "test",
		Role:        "admin",
		IsConfirmed: false,
	})

	return err
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteUserService))
}
