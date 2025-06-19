package services

import (
	"log"
	"time"
)

// OrderCleanupService handles periodic cleanup of expired orders
type OrderCleanupService struct {
	orderService *OrderService
	ticker       *time.Ticker
	done         chan bool
}

func NewOrderCleanupService() *OrderCleanupService {
	return &OrderCleanupService{
		orderService: NewOrderService(),
		done:         make(chan bool),
	}
}

// StartCleanupScheduler starts the periodic cleanup process
func (ocs *OrderCleanupService) StartCleanupScheduler() {
	// Run cleanup every 15 minutes
	ocs.ticker = time.NewTicker(15 * time.Minute)

	go func() {
		for {
			select {
			case <-ocs.done:
				return
			case <-ocs.ticker.C:
				if err := ocs.orderService.CleanExpiredOrders(); err != nil {
					log.Printf("Failed to clean expired orders: %v", err)
				} else {
					log.Println("Successfully cleaned expired orders")
				}
			}
		}
	}()

	log.Println("Order cleanup scheduler started")
}

// StopCleanupScheduler stops the periodic cleanup process
func (ocs *OrderCleanupService) StopCleanupScheduler() {
	if ocs.ticker != nil {
		ocs.ticker.Stop()
	}
	ocs.done <- true
	log.Println("Order cleanup scheduler stopped")
}

// RunManualCleanup runs cleanup manually (useful for admin endpoints)
func (ocs *OrderCleanupService) RunManualCleanup() error {
	log.Println("Running manual order cleanup...")
	return ocs.orderService.CleanExpiredOrders()
}
