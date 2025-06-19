package main

import (
	"log"
	"os"
	"os/signal"
	"storelite/config"
	"storelite/models"
	"storelite/routes"
	"storelite/services"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	// Initialize database
	config.InitDatabase()

	// Auto migrate models
	if err := autoMigrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Start cart cleanup scheduler
	cleanupService := services.NewCartCleanupService()
	cleanupService.StartCleanupScheduler()
	orderCleanupService := services.NewOrderCleanupService()
	orderCleanupService.StartCleanupScheduler()

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down...")
		cleanupService.StopCleanupScheduler()
		orderCleanupService.StopCleanupScheduler()
		os.Exit(0)
	}()

	// Setup routes
	router := routes.SetupRoutes()

	// Get port from environment or use default
	env := godotenv.Load()
	if env != nil {
		log.Fatal("Lỗi khi đọc PORT tại file .env")
	}
	port := os.Getenv("PORT")

	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}

func autoMigrate() error {
	db := config.GetDB()

	// Auto migrate all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.RecoveryCode{},
		&models.LoginAttempt{},
		&models.Category{},
		&models.Product{},
		&models.ProductSize{},
		&models.Review{},
		&models.Cart{},
		&models.CartItem{},
		// Add more models here as you create them
	); err != nil {
		return err
	}

	// Add unique constraint for product_id + size
	if err := db.Exec("ALTER TABLE product_sizes ADD CONSTRAINT unique_product_size UNIQUE (product_id, size)").Error; err != nil {
		// Ignore error if constraint already exists
		log.Println("Unique constraint may already exist:", err)
	}

	// Add unique constraint for cart_id + product_id + product_size_id
	if err := db.Exec("ALTER TABLE cart_items ADD CONSTRAINT unique_cart_item UNIQUE (cart_id, product_id, product_size_id)").Error; err != nil {
		// Ignore error if constraint already exists
		log.Println("Cart item unique constraint may already exist:", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}
