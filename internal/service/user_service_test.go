package service_test

import (
	"context"
	"testing"

	"github.com/dvvnFrtn/capstone-backend/config"
	"github.com/dvvnFrtn/capstone-backend/infra/db"
	"github.com/dvvnFrtn/capstone-backend/internal/service"
	"github.com/dvvnFrtn/capstone-backend/internal/types"
	"github.com/dvvnFrtn/capstone-backend/pkg/authx"
	"github.com/dvvnFrtn/capstone-backend/pkg/testutil"
	"github.com/google/uuid"
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

func dummyAdminRegistrationRequest(email, phone, passw string) service.AdminRegistrationRequest {
	return service.AdminRegistrationRequest{
		Email:       email,
		Password:    passw,
		Phone:       phone,
		Fullname:    "user test",
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
		expectedEmail = "muhrizkifajar28@gmail.com"
		expectedPhone = "+6287819502098"
	)

	client, err := authx.InitFirebase(context.Background(), "../../firebase.json")
	require.NoError(ts.T(), err)

	conn, err := db.NewPostgreConn(ctx, ts.cfg)
	require.NoError(ts.T(), err)
	defer conn.Close(ctx)

	authService := service.NewFirebaseAuthService(client.Auth)
	userService := service.NewUserService(conn, authService)

	_, err = userService.AdminRegistration(
		ctx,
		dummyAdminRegistrationRequest(expectedEmail, expectedPhone, "password123"),
	)

	assert.NoError(ts.T(), err)
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteUserService))
}
