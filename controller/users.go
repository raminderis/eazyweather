package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/raminderis/lenslocked/context"
	"github.com/raminderis/lenslocked/errors"
	"github.com/raminderis/lenslocked/models"
)

type Users struct {
	Templates struct {
		New            Template
		SignIn         Template
		ForgotPassword Template
		CheckYourEmail Template
		ResetPassword  Template
		CityTemp       Template
		ShowCityTemp   Template
	}
	UserService          *models.UserService
	SessionService       *models.SessionService
	PasswordResetService *models.PasswordResetService
	EmailService         *models.EmailService
	CityTempS            *models.CityTempS
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
	user, err := u.UserService.Create(data.Email, data.Password)
	if err != nil {
		if errors.Is(err, models.ErrEmailTaken) {
			err = errors.Public(err, "That email address is already associated with an account.")
		}
		u.Templates.New.Execute(w, r, data, err)
		// fmt.Printf("User Create Failed : %v", err)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
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
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	fmt.Fprintf(w, "Current User: %s\n", user.Email)
}

func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {
	token, err := readCookie(r, CookieSession)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = u.SessionService.Delete(token)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	deleteCookie(w, CookieSession)
	http.Redirect(w, r, "/signin", http.StatusFound)
}

func (u Users) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.ForgotPassword.Execute(w, r, data)
}

func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	pwReset, err := u.PasswordResetService.Create(data.Email)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	vals := url.Values{
		"token": {pwReset.Token},
	}
	resetURL := "https://www.justmightthoughts.com/reset-pw?" + vals.Encode()
	fmt.Println(data.Email, resetURL)
	err = u.EmailService.ForgotPassword(data.Email, resetURL)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	u.Templates.CheckYourEmail.Execute(w, r, data)
}

func (u Users) CityTemp(w http.ResponseWriter, r *http.Request) {
	var data struct {
		City string
	}
	data.City = r.FormValue("city")
	u.Templates.CityTemp.Execute(w, r, data)
}

func (u Users) ProcessCityTemp(w http.ResponseWriter, r *http.Request) {
	var data struct {
		City         string
		ApiToken     string
		CityTemp     string
		CityHumidity string
		Time         string
	}
	data.City = r.FormValue("city")
	data.ApiToken = os.Getenv("OPENWEATHER_TOKEN")
	//strings.ToUpper(r.FormValue("city"))
	cityTemp, err := u.CityTempS.Communicate(data.City, data.ApiToken)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "processing city temp has an error "+fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	data.City = strings.ToUpper(r.FormValue("city")[:1]) + strings.ToLower(r.FormValue("city")[1:])
	data.CityHumidity = cityTemp.Humidity
	data.CityTemp = cityTemp.Temp
	data.Time = cityTemp.Time
	u.Templates.ShowCityTemp.Execute(w, r, data)
}

func (u Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string
	}
	data.Token = r.FormValue("token")
	u.Templates.ResetPassword.Execute(w, r, data)
}

func (u Users) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token    string
		Password string
	}
	data.Token = r.FormValue("token")
	data.Password = r.FormValue("password")
	// Steps:
	// consume token
	// if token consumes correctly, we are going to update user password
	// create a new session for the user, sign them in
	// redirect them to /user/me page
	user, err := u.PasswordResetService.Consume(data.Token)
	if err != nil {
		fmt.Println(err)
		// TODO: Distinguish between different types of errors.
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	//TODO: update user password
	err = u.UserService.UpdatePassword(user.ID, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
	// Sign the user in
	// any errors should redirect to signin page
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

type UserMiddleware struct {
	SessionService *models.SessionService
}

func (umw UserMiddleware) SetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := readCookie(r, CookieSession)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		user, err := umw.SessionService.User(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (umw UserMiddleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
