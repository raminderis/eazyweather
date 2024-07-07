package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
