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

const mobileSessionExpiration = 30 * 24 * time.Hour // 30 days

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
	appAuthRepo := repository.NewAppAuthRepository(db)
	shiftRepo := repository.NewShiftRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	promotionRepo := repository.NewPromotionRepository(db)
	stockTransferRepo := repository.NewStockTransferRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	pointsRepo := repository.NewPointsRepository(db)

	authService := service.NewAuthService(staffUserRepo, cfg.JWT.Secret, cfg.JWT.Expiration)
	memberService := service.NewMemberService(memberRepo)
	transactionService := service.NewTransactionService(transactionRepo, memberRepo)
	appAuthService := service.NewAppAuthService(appAuthRepo, mobileSessionExpiration)
	shiftService := service.NewShiftService(shiftRepo)
	orderService := service.NewOrderService(orderRepo)
	promotionService := service.NewPromotionService(promotionRepo)
	stockTransferService := service.NewStockTransferService(stockTransferRepo)
	inventoryService := service.NewInventoryService(inventoryRepo)
	pointsService := service.NewPointsService(pointsRepo, orderRepo)

	authHandler := handler.NewAuthHandler(authService)
	memberHandler := handler.NewMemberHandler(memberService)
	transactionHandler := handler.NewTransactionHandler(transactionService)
	appAuthHandler := handler.NewAppAuthHandler(appAuthService)
	shiftHandler := handler.NewShiftHandler(shiftService, appAuthService)
	orderHandler := handler.NewOrderHandler(orderService, appAuthService, shiftService, pointsService)
	promotionHandler := handler.NewPromotionHandler(promotionService, appAuthService)
	stockTransferHandler := handler.NewStockTransferHandler(stockTransferService, appAuthService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService, appAuthService)
	pointsHandler := handler.NewPointsHandler(pointsService, appAuthService)

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
				transactions.GET("/branch", transactionHandler.ListByBranch)
			}
		}
	}

	mobileV1 := router.Group("/api/v2")
	{
		mobileAuth := mobileV1.Group("/auth")
		{
			mobileAuth.POST("/login", appAuthHandler.LoginStore)
			mobileAuth.POST("/register", appAuthHandler.RegisterBusiness)
			mobileAuth.POST("/verify-pin", appAuthHandler.VerifyPin)
			mobileAuth.GET("/session", appAuthHandler.ValidateSession)
			mobileAuth.POST("/logout", appAuthHandler.Logout)
			mobileAuth.POST("/generate-hash", appAuthHandler.GenerateHash)
		}

		branches := mobileV1.Group("/branches")
		{
			branches.GET("", shiftHandler.ListBranches)
			branches.POST("/select", shiftHandler.SelectBranch)
		}

		shifts := mobileV1.Group("/shifts")
		{
			shifts.POST("/open", shiftHandler.OpenShift)
			shifts.GET("/current", shiftHandler.GetCurrentShift)
			shifts.GET("/summary", shiftHandler.GetShiftSummary)
			shifts.POST("/close", shiftHandler.CloseShift)
		}

		products := mobileV1.Group("/products")
		{
			products.GET("", orderHandler.ListProducts)
		}

		customers := mobileV1.Group("/customers")
		{
			customers.GET("/search", orderHandler.SearchCustomers)
		}

		orders := mobileV1.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.GetOrdersByShift)
			orders.GET("/:id", orderHandler.GetOrderByID)
			orders.POST("/:id/cancel", orderHandler.CancelOrder)
		}

		promotions := mobileV1.Group("/promotions")
		{
			promotions.GET("", promotionHandler.GetActivePromotions)
			promotions.POST("/calculate", promotionHandler.CalculateDiscount)
			promotions.POST("/detect", promotionHandler.DetectPromotions)
		}

		stockTransfers := mobileV1.Group("/stock-transfers")
		{
			stockTransfers.POST("", stockTransferHandler.CreateTransfer)
			stockTransfers.POST("/withdraw", stockTransferHandler.WithdrawGoods)
			stockTransfers.GET("", stockTransferHandler.GetTransfers)
			stockTransfers.GET("/pending", stockTransferHandler.GetPendingTransfers)
			stockTransfers.GET("/:id", stockTransferHandler.GetTransfer)
			stockTransfers.POST("/:id/receive", stockTransferHandler.ReceiveTransfer)
			stockTransfers.POST("/:id/cancel", stockTransferHandler.CancelTransfer)
		}

		inventory := mobileV1.Group("/inventory")
		{
			inventory.POST("/adjust", inventoryHandler.AdjustStock)
			inventory.GET("/movements", inventoryHandler.GetMovements)
			inventory.GET("/low-stock", inventoryHandler.GetLowStockItems)
		}

		points := mobileV1.Group("/points")
		{
			points.GET("/customer/:customer_id", pointsHandler.GetCustomerPoints)
			points.GET("/customer/:customer_id/history", pointsHandler.GetPointHistory)
			points.GET("/redeemable-products", pointsHandler.GetRedeemableProducts)
			points.POST("/redeem", pointsHandler.RedeemPoints)
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
