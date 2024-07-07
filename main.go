package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
	"github.com/raminderis/lenslocked/controller"
	"github.com/raminderis/lenslocked/migrations"
	"github.com/raminderis/lenslocked/models"
	"github.com/raminderis/lenslocked/templates"
	"github.com/raminderis/lenslocked/views"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key    string
		Secure bool
	}
	Server struct {
		Address string
	}
}

func loadEnvConfig() (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}
	//PSQL
	cfg.PSQL = models.DefaultPostgresConfig()

	//SMTP
	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")

	//CSRF
	cfg.CSRF.Key = "6ydtr6eyr76qwouyehdfgdhsywegwtqh"
	cfg.CSRF.Secure = false
	//Server
	cfg.Server.Address = ":3000"
	return cfg, nil
}

func main() {
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}
	//Setup the DB
	db, err := models.Open(cfg.PSQL)
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
	userService := &models.UserService{
		DB: db,
	}
	sessionService := &models.SessionService{
		DB:            db,
		BytesPerToken: 32,
	}
	emailService := models.NewEmailService(cfg.SMTP)
	passwordResetService := &models.PasswordResetService{
		DB: db,
	}

	//Setup Middleware
	umw := controller.UserMiddleware{
		SessionService: sessionService,
	}
	csrfMw := csrf.Protect([]byte(cfg.CSRF.Key), csrf.Secure(cfg.CSRF.Secure))

	//Setup controllers
	usersC := controller.Users{
		UserService:          userService,
		SessionService:       sessionService,
		EmailService:         emailService,
		PasswordResetService: passwordResetService,
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
	usersC.Templates.CheckYourEmail = views.Must(views.ParseFS(
		templates.FS,
		"check-your-email.gohtml", "tailwind.gohtml",
	))
	usersC.Templates.ResetPassword = views.Must(views.ParseFS(
		templates.FS,
		"reset-pw.gohtml", "tailwind.gohtml",
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
	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)
	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	//Start the Server
	port := cfg.Server.Address
	fmt.Printf("LISTENING now on: %s...\n", port)
	err = http.ListenAndServe(port, r)
	if err != nil {
		panic(err)
	}
}
