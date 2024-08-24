package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")

	fmt.Println(port)
}
