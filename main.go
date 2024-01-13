package main

import (
	"fmt"
	"io"
	"net/http"

	"example/hello/db"
	"example/hello/email"
	"example/hello/helpers"

	"os"

	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

var maxDaysApiKey int
var maxDaysLibrary int

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

	dbFile := os.Getenv("DB_FILE")

	if dbFile == "" {
		dbFile = "libraries.db"
	}

	err = db.InitDb(dbFile)

	if err != nil {
		fmt.Println("Error: Could not open database")
		panic(err)
	}

	defer db.DB.Close()

	db.CreateTables()

	maxDayLibraries := os.Getenv("MAX_DAY_LIBRARIES")

	if maxDayLibraries == "" {
		maxDayLibraries = "180"
	}

	maxDayLibrariesInt, err := strconv.Atoi(maxDayLibraries)

	if err != nil {
		fmt.Println("Error: MAX_DAY_LIBRARIES must be an integer")
		maxDayLibrariesInt = 180
	}

	if maxDayLibrariesInt <= 0 {
		fmt.Println("Error: MAX_DAY_LIBRARIES must be greater than 0")
		maxDayLibrariesInt = 180
	}

	maxDayApiKey := os.Getenv("MAX_DAY_API_KEY")

	if maxDayApiKey == "" {
		maxDayApiKey = "7"
	}

	maxDayApiKeyInt, err := strconv.Atoi(maxDayApiKey)

	if err != nil {
		fmt.Println("Error: MAX_DAY_API_KEY must be an integer")
		maxDayApiKeyInt = 7
	}

	if maxDayApiKeyInt <= 0 {
		fmt.Println("Error: MAX_DAY_API_KEY must be greater than 0")
		maxDayApiKeyInt = 7
	}

	maxDaysApiKey = maxDayApiKeyInt
	maxDaysLibrary = maxDayLibrariesInt

	db.TidyUp(maxDaysLibrary, maxDaysApiKey)

	r := mux.NewRouter()

	r.HandleFunc("/", indexHandler)

	r.HandleFunc("/getToken", getTokenHandler).Methods("POST")

	http.ListenAndServe(fmt.Sprintf(":%s", port), r)

	fmt.Println(fmt.Sprintf("Server running on port %s", port))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	message := "This API provides information about nearby UK libraries. Check the GitHub Page https://github.com/jonesjacklewis/GoNearbyUkLibraries"

	io.WriteString(w, message)
}

func getTokenHandler(w http.ResponseWriter, r *http.Request) {
	// check is post request
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type RequestBody struct {
		Email string `json:"email"`
	}

	var body RequestBody

	err := helpers.DecodeJson(r.Body, &body)

	if err != nil {
		http.Error(w, "Error decoding json", http.StatusBadRequest)
		return
	}

	emailAddress := body.Email

	if emailAddress == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	db.TidyUp(maxDaysLibrary, maxDaysApiKey)

	apiKey := helpers.GenerateToken()

	db.InsertApiKey(emailAddress, apiKey)

	apiKey, err = db.GetApiKey(emailAddress)

	if err != nil {
		http.Error(w, "Error getting api key", http.StatusInternalServerError)
		return
	}

	err = email.SendEmail(emailAddress, "Authentication Key", fmt.Sprintf("Your authentication key is: %s", apiKey))

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error sending email", http.StatusInternalServerError)
		return
	}

	io.WriteString(w, apiKey)
}
