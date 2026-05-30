package main

import (
	"fmt"
	"log"
	"os"
	"recipe-web-server/internal/database"
	"recipe-web-server/internal/models"
	"recipe-web-server/internal/services"
)

func main() {
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatal("ERROR: ADMIN_PASSWORD environment variable is not set")
	}

	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.RunMigrations()
	if err != nil {
		log.Fatal(err)
	}

	userService := services.NewUserService(db)
	user, err := userService.CreateUser("admin", "Administrator", adminPassword, models.RoleAdmin)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created admin user: %s (ID: %d)\n", user.Username, user.ID)
}
