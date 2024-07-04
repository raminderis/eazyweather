package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-mail/mail/v2"
	"github.com/joho/godotenv"
)

// const (
// 	host     = "sandbox.smtp.mailtrap.io"
// 	port     = 587
// 	username = "f7ead630decaaf"
// 	password = "463775e4042cbd"
// )

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	from := "test@lenslocked.com"
	to := "raminderis@live.com"
	subject := "This is a test email"
	plaintext := "This is the body of the email"
	htmltext := `<h1>Hello there buddy!</h1><p>This is the email</p><p>Hope you enjoy it</p>`
	msg := mail.NewMessage()
	msg.SetHeader("To", to)
	msg.SetHeader("From", from)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", plaintext)
	msg.AddAlternative("text/html", htmltext)
	msg.WriteTo(os.Stdout)

	dialer := mail.NewDialer(host, port, username, password)
	err = dialer.DialAndSend(msg)
	if err != nil {
		panic(err)
	}
	fmt.Println("Message Sent")
	// sender, err := dialer.Dial()
	// if err != nil {
	// 	panic(err)
	// }
	// err = sender.Send(from, []string{to}, msg)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Message Sent")

}
