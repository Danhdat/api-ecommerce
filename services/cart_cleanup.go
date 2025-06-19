package services

import (
	"log"
	"time"
)

// CartCleanupService handles periodic cleanup of expired carts
type CartCleanupService struct {
	cartService *CartService
	ticker      *time.Ticker
	done        chan bool
}

func NewCartCleanupService() *CartCleanupService {
	return &CartCleanupService{
		cartService: NewCartService(),
		done:        make(chan bool),
	}
}

// StartCleanupScheduler starts the periodic cleanup process
func (ccs *CartCleanupService) StartCleanupScheduler() {
	// Run cleanup every 6 hours
	ccs.ticker = time.NewTicker(6 * time.Hour)

	go func() {
		for {
			select {
			case <-ccs.done:
				return
			case <-ccs.ticker.C:
				if err := ccs.cartService.CleanExpiredCarts(); err != nil {
					log.Printf("Failed to clean expired carts: %v", err)
				} else {
					log.Println("Successfully cleaned expired carts")
				}
			}
		}
	}()

	log.Println("Cart cleanup scheduler started")
}

// StopCleanupScheduler stops the periodic cleanup process
func (ccs *CartCleanupService) StopCleanupScheduler() {
	if ccs.ticker != nil {
		ccs.ticker.Stop()
	}
	ccs.done <- true
	log.Println("Cart cleanup scheduler stopped")
}

// RunManualCleanup runs cleanup manually (useful for admin endpoints)
func (ccs *CartCleanupService) RunManualCleanup() error {
	log.Println("Running manual cart cleanup...")
	return ccs.cartService.CleanExpiredCarts()
}
