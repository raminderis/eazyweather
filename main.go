package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/raminderis/lenslocked/controller"
	"github.com/raminderis/lenslocked/migrations"
	"github.com/raminderis/lenslocked/models"
	"github.com/raminderis/lenslocked/templates"
	"github.com/raminderis/lenslocked/views"
)

func main() {
	//Setup the DB
	cfg := models.DefaultPostgresConfig()
	fmt.Println(cfg.String())
	db, err := models.Open(cfg)
	// cfg := models.DefaultCloudSqlConfig()
	// db, err := models.ConnectWithConnector(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = models.MigrateFS(db, migrations.FS, "")
	if err != nil {
		panic(err)
	}

	//Setup Services
	userService := models.UserService{
		DB: db,
	}
	sessionService := models.SessionService{
		DB:            db,
		BytesPerToken: 32,
	}

	//Setup Middleware
	umw := controller.UserMiddleware{
		SessionService: &sessionService,
	}
	csrfKey := "6ydtr6eyr76qwouyehdfgdhsywegwtqh"
	csrfMw := csrf.Protect([]byte(csrfKey), csrf.Secure(false))

	//Setup controllers
	usersC := controller.Users{
		UserService:    &userService,
		SessionService: &sessionService,
		//EmailService:         &emailService,
		//PasswordResetService: &passwordResetService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"signup.gohtml", "tailwind.gohtml",
	))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS,
		"signin.gohtml", "tailwind.gohtml",
	))
	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(
		templates.FS,
		"forgot-pw.gohtml", "tailwind.gohtml",
	))
	//Setup Router and Routes
	r := chi.NewRouter()
	r.Use(csrfMw)
	r.Use(umw.SetUser)
	r.Get("/", controller.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))
	r.Get("/contact", controller.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))
	r.Get("/faq", controller.FAQ(
		views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))
	r.Get("/signup", usersC.New)
	r.Post("/users", usersC.Create)
	r.Get("/signin", usersC.SignIn)
	r.Post("/login", usersC.ProcessSignIn)
	r.Post("/signout", usersC.ProcessSignOut)
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)
	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	//Start the Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Println("LISTENING now on: " + port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		panic(err)
	}
}
