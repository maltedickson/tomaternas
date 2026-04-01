package main

import (
	"fmt"
	"log"
	"net/http"
	"recipe-web-server/internal/database"
	"recipe-web-server/internal/handlers"
	"recipe-web-server/internal/middleware"
	"recipe-web-server/internal/services"
	"recipe-web-server/internal/templates"
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

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./web/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	renderer, err := templates.NewRenderer()
	if err != nil {
		log.Fatal(err)
	}

	authService := services.NewAuthService(db)
	userService := services.NewUserService(db)

	authHandler := handlers.NewAuthHandler(authService, renderer)
	homeHandler := handlers.NewHomeHandler(renderer)
	adminHandler := handlers.NewAdminHandler(userService, renderer)
	loggedInHandler := handlers.NewLoggedInHandler(userService, renderer)

	mux.HandleFunc("/", homeHandler.HomePage)

	mux.HandleFunc("GET /login", authHandler.LoginPage)
	mux.HandleFunc("POST /login", authHandler.Login)

	mux.HandleFunc("POST /logout", authHandler.Logout)

	loggedInMux := http.NewServeMux()

	loggedInMux.HandleFunc("GET /settings", loggedInHandler.SettingsPage)
	loggedInMux.HandleFunc("GET /manage/new", loggedInHandler.NewRecipePage)

	adminMux := http.NewServeMux()

	adminMux.HandleFunc("GET /admin/dashboard", adminHandler.DashboardPage)

	adminMux.HandleFunc("GET /admin/users", adminHandler.UsersPage)

	adminMux.HandleFunc("GET /admin/users/create", adminHandler.CreateUserPage)
	adminMux.HandleFunc("POST /admin/users/create", adminHandler.CreateUser)

	adminMux.HandleFunc("GET /admin/users/manage/{id}", adminHandler.ManageUserPage)
	adminMux.HandleFunc("POST /admin/users/manage/{id}/username", adminHandler.UpdateUsername)
	adminMux.HandleFunc("POST /admin/users/manage/{id}/display-name", adminHandler.UpdateDisplayName)
	adminMux.HandleFunc("POST /admin/users/manage/{id}/password", adminHandler.UpdatePassword)
	adminMux.HandleFunc("POST /admin/users/manage/{id}/role", adminHandler.UpdateRole)

	mux.Handle("GET /settings", middleware.RequireAuth(loggedInMux))
	mux.Handle("/manage/", middleware.RequireAuth(loggedInMux))
	mux.Handle("/admin/", middleware.RequireAdmin(adminMux))

	middlewareStack := middleware.CreateStack(
		middleware.AuthMiddleware(authService),
	)
	fmt.Println("Server starting on :8080...")
	err = http.ListenAndServe(":8080", middlewareStack(mux))
	if err != nil {
		log.Fatal(err)
	}
}

/*

===============================================================================
Routes:
===============================================================================

PUBLIC (Anonymous access)
  GET  /                     -> Home: List all recipes (paged/searchable)
  GET  /r/{id}               -> View: Show single recipe details
  GET  /search               -> Search: List recipes matching query params
  GET  /static/*             -> Assets: Serve CSS, icons, and client JS
  GET  /uploads/*            -> Media: Serve uploaded recipe photos

AUTHENTICATION
  GET  /login                -> Show login form
  POST /login                -> Process credentials & create session
  POST /logout               -> Terminate session (POST for security)

MANAGEMENT (Requires 'Editor' or 'Admin' role)
  GET  /manage/new           -> Form to create a new recipe
  POST /manage/new           -> Save new recipe -> Redirect to /r/{id}
  GET  /manage/edit/{id}     -> Form to edit existing recipe
  POST /manage/edit/{id}     -> Update recipe data -> Redirect to /r/{id}
  GET  /manage/delete/{id}   -> "Are you sure?" confirmation page
  POST /manage/delete/{id}   -> Permanent delete action -> Redirect to /

ADMINISTRATION (Requires 'Admin' role)
  GET  /admin/users          -> List of people with access
  POST /admin/users          -> Create a new user account (supervised)

===============================================================================
Design Decisions
===============================================================================

1.
Keep recipes under /r/{id} so that I won't have to worry about a recipe id
colliding with another route.

2.
Use id instead of slug for recipes because:
  1. It's easier to implement. For example no need to create a slug, and
     easy to retrieve the right recipe from the database.
  2. If recipe title changes, the recipe's url can stay the same without the
     url containing an outdated slug.

3.
Have a /manage/ route for managing recipes and /admin/ for managing users so
that it is easy to implement access control.

*/
