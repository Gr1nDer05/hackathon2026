package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/database"
	"github.com/Gr1nDer05/Hackathon2026/internal/api"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
	"github.com/Gr1nDer05/Hackathon2026/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.ConnectPostgres()
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer db.Close()

	log.Println("database connected")

	repo := repository.NewAppRepository()
	appService := service.NewAppService(repo)
	handler := api.NewHandler(appService, db)

	router := gin.Default()
	router.GET("/health", handler.Health)
	router.GET("/testdb", handler.TestDB)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("shutting down server")
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
