package main

import (
	"fmt"
	"log"
	"recipe-web-server/internal/database"
	"recipe-web-server/internal/services"
)

func main() {
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
	user, err := userService.CreateUser("admin", "Administrator", "pwd", true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created admin user: %s (ID: %d)\n", user.Username, user.ID)
}
