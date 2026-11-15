package main

import (
	"context"
	"log"

	"foundnear/config"
	"foundnear/database"
)

func main() {
	config.LoadEnv()
	config.InitLogger()

	// Initialize PostgreSQL connection
	ctx := context.Background()

	db, err := database.NewPostgresPool(
		ctx,
		config.GetDatabaseURL(),
	)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Initialize Fiber application
	app := config.NewApp(db)

	port := config.GetEnv("APP_PORT", "3000")

	log.Printf("FoundNear server running on port %s", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
