package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	PORT           = ""
	SESSION_SECRET = ""
)

func init() {
	err := godotenv.Load()

	if err != nil {
		log.Println("Loading .env failed", err.Error())
	}

	PORT = os.Getenv("PORT")
	SESSION_SECRET = os.Getenv("SESSION_SECRET")
}
