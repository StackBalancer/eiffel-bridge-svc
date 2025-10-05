package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"eiffel-bridge-svc/internal/gitlab"
	"eiffel-bridge-svc/internal/publisher"
)

func main() {
	var pub *publisher.RabbitPublisher
	var err error

	// Retry connection to RabbitMQ
	maxRetries := 10
	for i := 1; i <= maxRetries; i++ {
		pub, err = publisher.NewRabbitPublisher()
		if err == nil {
			log.Printf("Connected to RabbitMQ on attempt %d", i)
			break
		}
		log.Printf("RabbitMQ connection failed (attempt %d/%d): %v", i, maxRetries, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after %d attempts: %v", maxRetries, err)
	}
	defer pub.Close()

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		gitlab.HandleWebhook(w, r, pub.Publish)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Adapter listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
