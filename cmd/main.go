package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hmdnu/okane/config"
	"github.com/hmdnu/okane/internal/handler"
	appmiddleware "github.com/hmdnu/okane/internal/middleware"
	"github.com/hmdnu/okane/internal/service"
	"github.com/hmdnu/okane/lib"
)

func BuildRouter(r *chi.Mux, db *sql.DB) {
	userService := service.UserServiceInit(db)
	transactionService := service.TransactionServiceInit(db)
	categoryService := service.CategoryServiceInit(db)
	userHandler := handler.UserHandlerInit(userService)
	transactionHandler := handler.TransactionHandlerInit(transactionService)
	categoryHandler := handler.CategoryHandlerInit(categoryService)

	r.Handle("/css/*", http.StripPrefix("/css/", http.FileServer(http.Dir("public/css"))))

	// Public auth routes
	r.Get("/login", userHandler.LoginView)
	r.Get("/register", userHandler.RegisterView)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)
	})

	// Protected routes — require a valid auth session
	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.AuthMiddleware(db))
		r.Post("/auth/logout", userHandler.Logout)
		r.Get("/", transactionHandler.DashboardView)
		r.Post("/transactions", transactionHandler.Create)
		r.Post("/transactions/{id}/delete", transactionHandler.Delete)
		r.Get("/categories", categoryHandler.CategoryManagementView)
		r.Post("/categories", categoryHandler.Create)
		r.Post("/categories/{id}/update", categoryHandler.Update)
		r.Post("/categories/{id}/delete", categoryHandler.Delete)
		r.Get("/settings", userHandler.UserSettingView)
		r.Post("/settings/salary", userHandler.SaveSalarySetting)
	})
}

func main() {
	db, err := config.StartDb()

	if err != nil {
		log.Fatal("Db failed to start", err)

	}

	r := chi.NewRouter()

	logFile, err := lib.InitLogger()
	if err != nil {
		log.Fatal("Logger failed to start", err)
	}
	defer logFile.Close()

	r.Use(lib.RequestLogger)

	BuildRouter(r, db)
	service.StartScheduledJobRunner(db)

	// csrf protection
	csrfProtection := http.NewCrossOriginProtection()

	fmt.Println("Server is running at http://localhost:" + config.PORT)
	http.ListenAndServe(":"+config.PORT, csrfProtection.Handler(r))
}
