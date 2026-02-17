package main

import (
	"log"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/app"
)

func main() {

	application, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	defer application.Close()

	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
