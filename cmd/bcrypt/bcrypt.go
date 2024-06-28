package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	switch os.Args[1] {
	case "hash":
		hash(os.Args[2])
	case "compare":
		compare(os.Args[2], os.Args[3])
	default:
		fmt.Println("i only support hash and compare")
	}
}

func hash(password string) {
	fmt.Println("hashing ", password)
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
		return
	}
	fmt.Printf("hashed to : %q", string(hashed))
}

func compare(hashedPassword string, password string) {
	fmt.Println("comparing ", password, hashedPassword)
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		fmt.Println("password doesnt match.")
		return
	}
	fmt.Println("password correct")
}
