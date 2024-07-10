package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken = errors.New("models: email address is already in use")
)

type User struct {
	ID           uint
	Email        string
	PasswordHash string
}

type UserService struct {
	DB *sql.DB
}

type NewUser struct {
	Email    string
	Password string
}

func (us *UserService) Create(email, password string) (*User, error) {
	email = strings.ToLower(email)
	hashedBytePassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user : %w", err)
	}
	hashedPassword := string(hashedBytePassword)
	user := User{
		Email:        email,
		PasswordHash: hashedPassword,
	}
	row := us.DB.QueryRow(`
		INSERT INTO users (email, password_hash)
		VALUES ($1,$2) RETURNING id`, email, hashedPassword)

	// row := us.DB.QueryRow(`
	// 	INSERT INTO users (email, password_hash)
	// 	OUTPUT INSERTED.id
	// 	VALUES (@email, @hashedPassword)`, sql.Named("email", email), sql.Named("hashedPassword", hashedPassword))

	err = row.Scan(&user.ID)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.UniqueViolation {
				return nil, ErrEmailTaken
			}
		}
		return nil, fmt.Errorf("create user : %w", err)
	}
	return &user, nil
}

func (us *UserService) Login(email, password string) (*User, error) {
	email = strings.ToLower(email)
	user := User{
		Email: email,
	}
	row := us.DB.QueryRow(`
		SELECT id, password_hash FROM users WHERE
		email = $1`, email)
	// row := us.DB.QueryRow(`
	// 	SELECT id, password_hash FROM users WHERE
	// 	email = @email`, sql.Named("email", email))
	err := row.Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("LOGIN : %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("login : %w", err)
	}
	fmt.Println("User login successful")
	return &user, nil
}

func (us *UserService) UpdatePassword(userID uint, password string) error {
	hashedBytePassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("update password : %w", err)
	}
	hashedPassword := string(hashedBytePassword)
	_, err = us.DB.Exec(`
		UPDATE users
		SET password_hash = $2
		WHERE id = $1;`, userID, hashedPassword)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

const (
	WeatherUrl     = "www.weather.com"
	WeatherUrlAuth = ""
)

type CityTemp struct {
	Temp     string
	Humidity string
	Time     string
}

type CityTempService struct {
	DB            *sql.DB
	BytesPerToken int
	Duration      time.Duration
}

func (us *CityTempService) Communicate(city string) (*CityTemp, error) {
	cityTemp := CityTemp{}
	cityTemp.Time = time.Now().Format(time.RFC3339)

	//send query to openweahter
	requestURL := "https://api.openweathermap.org/data/2.5/weather?q=" + city + "&appid=9682880224b6a80bc7359f1bc0d27224"
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		fmt.Printf("Communicate: could not create request: %s\n", err)
		return nil, fmt.Errorf("Communicate : %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Communicate: error making http request: %s\n", err)
		return nil, fmt.Errorf("Communicate : %w", err)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Communicate: could not read response body: %s\n", err)
		return nil, fmt.Errorf("Communicate : %w", err)
	}

	fmt.Printf("Communicate: response body: %s\n", resBody)
	var jsonResponse map[string]interface{}
	err = json.Unmarshal(resBody, &jsonResponse)
	if err != nil {
		fmt.Printf("Communicate: could not unmarshal response body: %s\n", err)
		return nil, fmt.Errorf("Communicate : %w", err)
	}
	for key, value := range jsonResponse {
		if key == "cod" {
			if fmt.Sprint(reflect.TypeOf(value)) == "string" {
				fmt.Println("no 200  OK")
				originalErr := errors.New("no city by this name")
				return nil, fmt.Errorf("\nCommunicate : , %v", originalErr)
			}
		}
	}
	for key, value := range jsonResponse {
		//fmt.Println("Open Weather Response : ", key, value)
		if key == "main" {
			// Type assertion to extract the map[string]interface{} value
			mainData, ok := value.(map[string]interface{})
			if !ok {
				fmt.Println("Error: Unable to assert type for 'main'")
				return nil, fmt.Errorf("Communicate : %w", err)
			}

			// Now you can access specific fields within the 'main' data
			temperature, tempExists := mainData["temp"].(float64)
			if tempExists {
				fmt.Printf("Temperature: %.2f°C\n", temperature-273.15)
				cityTemp.Temp = fmt.Sprintf("%.2f", temperature-273.15)

			} else {
				fmt.Println("Temperature data not found")
				return nil, fmt.Errorf("Communicate : %w", err)
			}

			humidity, humidityExists := mainData["humidity"].(float64)
			if humidityExists {
				fmt.Printf("Humidity: %.2f\n", humidity)
				cityTemp.Humidity = fmt.Sprintf("%.2f", humidity)

			} else {
				fmt.Println("Humidity data not found")
				return nil, fmt.Errorf("Communicate : %w", err)
			}
			// Handle other fields similarly (e.g., humidity, pressure, etc.)
		}
	}
	fmt.Println(cityTemp.Temp)
	return &cityTemp, nil
}
