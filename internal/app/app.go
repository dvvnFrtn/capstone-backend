package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dvvnFrtn/capstone-backend/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func Run() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env file: ", err)
	} else {
		log.Println("success loading .env file")
	}

	router := gin.Default()
	handler.Register(router)
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
