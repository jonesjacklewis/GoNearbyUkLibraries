package main

import (
	"fmt"
	"io"
	"net/http"

	"os"

	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file")
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	// check port is integer
	intPort, err := strconv.Atoi(port)

	if err != nil {
		fmt.Println("Error: Port must be an integer")
		port = "8080"
	}

	if intPort <= 1024 || intPort > 65535 {
		fmt.Println("Error: Port must be between 1024 and 65535")
		port = "8080"
	}

	r := mux.NewRouter()

	r.HandleFunc("/", indexHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", port), r)

	fmt.Println(fmt.Sprintf("Server running on port %s", port))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	message := "This API provides information about nearby UK libraries. Check the GitHub Page https://github.com/jonesjacklewis/GoNearbyUkLibraries"

	io.WriteString(w, message)
}
