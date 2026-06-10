package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/maltedickson/tomaternas/internal/database"
	"github.com/maltedickson/tomaternas/internal/models"
	"github.com/maltedickson/tomaternas/internal/services"
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
	user, err := userService.CreateUser(context.Background(), "admin", "Administrator", adminPassword, models.RoleAdmin)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created admin user: %s (ID: %d)\n", user.Username, user.ID)
}
