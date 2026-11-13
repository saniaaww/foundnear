
package main

import (
	"log"

	"foundnear/config"
)

func main() {
	config.LoadEnv()
	config.InitLogger()

	app := config.NewApp()

	port := config.GetEnv("APP_PORT", "3000")

	log.Printf("FoundNear server running on port %s", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}

