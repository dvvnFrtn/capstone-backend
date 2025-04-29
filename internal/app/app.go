package app

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dvvnFrtn/capstone-backend/config"
	"github.com/dvvnFrtn/capstone-backend/infra/db"
	"github.com/dvvnFrtn/capstone-backend/internal/handler"
	"github.com/dvvnFrtn/capstone-backend/internal/service"
	"github.com/dvvnFrtn/capstone-backend/pkg/authx"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func Run() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env file: ", err)
	} else {
		log.Println("success loading .env file")
	}

	cfg := config.Database()
	conn, err := db.NewPostgreConn(context.Background(), &cfg)
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	var (
		router      = gin.Default()
		authClient  = authx.NewSupabase()
		authService = service.NewSupabaseAuthService(authClient)
		userService = service.NewUserService(conn, authService)
		userHandler = handler.NewUserHandler(slog.Default(), userService)
	)

	handler.Register(router, userHandler)
	server := &http.Server{
		Addr:    os.Getenv("APP_HOST"),
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Println("server shutdown error: ", err)
	}

	log.Println("server shutdown gracefully. bye!")
}
