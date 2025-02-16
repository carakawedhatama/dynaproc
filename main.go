package main

import (
	"log"
	"time"
)

func main() {
	cfg := Config{}
	cfg.LoadConfig("config")

	InitDB(cfg)
	InitRabbitMQ(cfg)
	go ConsumeQueue(cfg)

	for {
		log.Println("Fetching purchase orders for sync...")
		orders, err := FetchPendingOrders()
		if err != nil {
			log.Printf("Error fetching orders: %v", err)
			continue
		}

		for _, order := range orders {
			err := PublishToQueue(order)
			if err != nil {
				log.Printf("Failed to publish order %s to queue: %v", order.ID, err)
				ReportErrorToGlitchTip(cfg, order.ID, err)
			}
		}
		time.Sleep(30 * time.Second)
	}
}
