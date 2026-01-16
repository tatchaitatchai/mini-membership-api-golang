package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katom-membership/api/config"
	"github.com/katom-membership/api/internal/handler"
	"github.com/katom-membership/api/internal/middleware"
	"github.com/katom-membership/api/internal/repository"
	"github.com/katom-membership/api/internal/service"
	"github.com/katom-membership/api/pkg/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewPostgresDB(&database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	staffUserRepo := repository.NewStaffUserRepository(db)
	memberRepo := repository.NewMemberRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	authService := service.NewAuthService(staffUserRepo, cfg.JWT.Secret, cfg.JWT.Expiration)
	memberService := service.NewMemberService(memberRepo)
	transactionService := service.NewTransactionService(transactionRepo, memberRepo)

	authHandler := handler.NewAuthHandler(authService)
	memberHandler := handler.NewMemberHandler(memberService)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	gin.SetMode(cfg.Server.Mode)
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.CreateStaffUser)
		}

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			members := protected.Group("/members")
			{
				members.POST("", memberHandler.Create)
				members.GET("", memberHandler.List)
				members.GET("/:id", memberHandler.GetByID)
				members.PUT("/:id", memberHandler.Update)
			}

			transactions := protected.Group("/transactions")
			{
				transactions.POST("", transactionHandler.Create)
				transactions.GET("/member/:member_id", transactionHandler.ListByMember)
				transactions.GET("/	", transactionHandler.ListByBranch)
			}
		}
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
