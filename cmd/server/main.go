package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
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

	uploadsFileServer := http.FileServer(http.Dir("./data/uploads/"))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads", uploadsFileServer))

	os.MkdirAll("data/uploads/recipes", 0755)

	renderer, err := templates.NewRenderer()
	if err != nil {
		log.Fatal(err)
	}

	authService := services.NewAuthService(db)
	userService := services.NewUserService(db)
	recipeService := services.NewRecipeService(db)

	handler := handlers.NewHandler(authService, userService, recipeService, renderer)

	mux.HandleFunc("/", handler.ViewHome)
	mux.HandleFunc("GET /recipes", handler.ViewRecipes)
	mux.HandleFunc("GET /recipes/new", middleware.RequireAuth(handler.ViewCreateRecipe))
	mux.HandleFunc("POST /recipes", middleware.RequireAuth(handler.CreateRecipe))

	mux.HandleFunc("GET /recipes/{id}", handler.ViewRecipe)
	mux.HandleFunc("GET /recipes/{id}/edit", middleware.RequireAuth(handler.ViewEditRecipe))
	mux.HandleFunc("POST /recipes/{id}/delete", middleware.RequireAuth(handler.DeleteRecipe))
	mux.HandleFunc("POST /recipes/{id}", middleware.RequireAuth(handler.UpdateRecipe))

	mux.HandleFunc("GET /login", handler.ViewLogin)
	mux.HandleFunc("POST /login", handler.Login)
	mux.HandleFunc("POST /logout", handler.Logout)

	mux.HandleFunc("GET /settings", middleware.RequireAuth(handler.ViewSettings))

	adminMux := http.NewServeMux()

	adminMux.HandleFunc("GET /admin/dashboard", handler.ViewAdminDashboard)
	adminMux.HandleFunc("GET /admin/users", handler.ViewUsers)
	adminMux.HandleFunc("GET /admin/users/new", handler.ViewCreateUser)
	adminMux.HandleFunc("POST /admin/users", handler.CreateUser)
	adminMux.HandleFunc("GET /admin/users/{id}/edit", handler.ViewUpdateUser)
	adminMux.HandleFunc("POST /admin/users/{id}/username", handler.UpdateUsername)
	adminMux.HandleFunc("POST /admin/users/{id}/display-name", handler.UpdateDisplayName)
	adminMux.HandleFunc("POST /admin/users/{id}/password", handler.UpdatePassword)
	adminMux.HandleFunc("POST /admin/users/{id}/role", handler.UpdateRole)

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
