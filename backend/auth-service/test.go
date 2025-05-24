package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "12345678"
	hashedPasswordFromDB := "$2a$10$Y7cmsAE.Jlap0r7YvNgp7uCQkJK3oogA4b599p4WgR34AQIkQX.LK" // <-- Вставьте хеш из pgAdmin сюда

	err := bcrypt.CompareHashAndPassword([]byte(hashedPasswordFromDB), []byte(password))
	if err != nil {
		fmt.Printf("Comparison failed: %v\n", err)
	} else {
		fmt.Println("Comparison successful!")
	}

	// Оставьте остальной код без изменений
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating new hash: %v\n", err)
		return
	}
	fmt.Printf("Newly generated hash for '%s': %s\n", password, string(newHashedPassword))

	err = bcrypt.CompareHashAndPassword(newHashedPassword, []byte(password))
	if err != nil {
		fmt.Printf("Comparison of new hash failed: %v\n", err)
	} else {
		fmt.Println("Comparison of new hash successful!")
	}
}