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
	r := chi.NewRouter()
	r.Get("/", controller.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))

	r.Get("/contact", controller.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))

	r.Get("/faq", controller.FAQ(
		views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))

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
	userService := models.UserService{
		DB: db,
	}
	sessionService := models.SessionService
		DB:            db,
		BytesPerToken: 32,
	}
	usersC := controller.Users{
		UserService:    &userService,
		SessionService: &sessionService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"signup.gohtml", "tailwind.gohtml",
	))
	r.Get("/signup", usersC.New)

	r.Post("/users", usersC.Create)

	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS,
		"signin.gohtml", "tailwind.gohtml",
	))
	r.Get("/signin", usersC.SignIn)
	r.Post("/login", usersC.ProcessSignIn)
	r.Post("/signout", usersC.ProcessSignOut)
	r.Get("/users/me", usersC.CurrentUser)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	//var router Router
	fmt.Println("LISTENING now on GAE default port of 8080")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	csrfKey := "6ydtr6eyr76qwouyehdfgdhsywegwtqh"
	csrfMw := csrf.Protect([]byte(csrfKey), csrf.Secure(false))
	err = http.ListenAndServe(":"+port, csrfMw(r))
	if err != nil {
		panic(err)
	}
}
