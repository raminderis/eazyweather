package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/raminderis/lenslocked/controller"
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

	usersC := controller.Users{}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"signup.gohtml", "tailwind.gohtml",
	))
	r.Get("/signup", usersC.New)

	r.Get("/signin", controller.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml"))))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	//var router Router
	fmt.Println("LISTENING now on GAE default port of 8080")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		panic(err)
	}
}
