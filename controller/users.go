package controller

import (
	"fmt"
	"net/http"

	"github.com/raminderis/lenslocked/models"
)

type Users struct {
	Templates struct {
		New    Template
		SignIn Template
	}
	UserService *models.UserService
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.New.Execute(w, r, data)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := u.UserService.Create(email, password)
	if err != nil {
		fmt.Printf("User Create Failed : %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "User Created: %+v", user)
}

func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string
		Password string
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")
	user, err := u.UserService.Login(data.Email, data.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cookie := http.Cookie{
		Name:  "email",
		Value: user.Email,
		Path:  "/",
	}
	http.SetCookie(w, &cookie)
	fmt.Fprintf(w, "User Login Success: %+v", user)
}

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	emailCookie, err := r.Cookie("email")
	if err != nil {
		fmt.Fprint(w, "The email couldnt be read.")
		return
	}
	fmt.Fprintf(w, "Email cookie: %s\n", emailCookie.Value)
	fmt.Fprintf(w, "Headers: %+v\n", r.Header)
}
