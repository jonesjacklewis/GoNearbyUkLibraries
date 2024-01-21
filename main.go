package main

import (
	"example/hello/db"
	"example/hello/email"
	"example/hello/helpers"
	"example/hello/thirdPartyIntegrations"
	"fmt"
	"io"
	"net/http"
	"time"

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

	r.HandleFunc("/requestToken", requestTokenHandler).Methods("POST")

	// endpoint /getLibraries/postcode/{postcode}/count/{count} post quest
	r.HandleFunc("/getLibraries/postcode/{postcode}/count/{count}", getLibrariesHandler).Methods("GET")

	fmt.Println(fmt.Sprintf("Server running on port %s", port))
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)

}

func populateDatabaseWithLibraries() {

	db.TidyUp(maxDaysLibrary, maxDaysApiKey)

	libraries, err := thirdPartyIntegrations.GetAllLibraries()

	if err != nil {
		println(err)
		panic(err)
	}

	for _, library := range libraries {
		db.InsertLibrary(library)
	}

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	message := "This API provides information about nearby UK libraries. Check the GitHub Page https://github.com/jonesjacklewis/GoNearbyUkLibraries"

	io.WriteString(w, message)
}

func requestTokenHandler(w http.ResponseWriter, r *http.Request) {
	// check is post request
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type RequestBody struct {
		Email string `json:"email"`
	}

	var body RequestBody

	err := helpers.DecodeJson(r.Body, &body, false)

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

func getLibrariesHandler(w http.ResponseWriter, r *http.Request) {
	// check is post request
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get params
	params := mux.Vars(r)

	// check params
	postcode := params["postcode"]
	count := params["count"]

	if postcode == "" {
		http.Error(w, "Postcode is required", http.StatusBadRequest)
		return
	}

	if count == "" {
		http.Error(w, "Count is required", http.StatusBadRequest)
		return
	}

	// check count is integer
	intCount, err := strconv.Atoi(count)

	if err != nil {
		http.Error(w, "Count must be an integer", http.StatusBadRequest)
		return
	}

	if intCount <= 0 {
		http.Error(w, "Count must be greater than 0", http.StatusBadRequest)
		return
	}

	// check api key
	apiKey := r.Header.Get("X-API-KEY")

	if apiKey == "" {
		http.Error(w, "API Key is required", http.StatusBadRequest)
		return
	}

	valid, err := db.ValidateApiKey(apiKey)

	if err != nil {
		http.Error(w, "Error validating api key", http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "Invalid API Key", http.StatusUnauthorized)
		return
	}

	OldestDateAddedLibraries := db.GetOldestDateAddedLibraries()
	now := time.Now()
	sixMonthsAgo := now.AddDate(0, -6, 0)

	if OldestDateAddedLibraries.Before(sixMonthsAgo) {
		fmt.Printf("Populating database with libraries\n")
		populateDatabaseWithLibraries()
	}

	// check postcode is valid
	valid = thirdPartyIntegrations.CheckPostcodeIsValid(postcode)

	if !valid {
		http.Error(w, "Invalid Postcode", http.StatusBadRequest)
		return
	}

	// get point for postcode
	point, err := thirdPartyIntegrations.GetPointForPostcode(postcode)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get libraries
	libraries, err := db.GetLibraries()

	if err != nil {
		http.Error(w, "Error getting libraries", http.StatusInternalServerError)
		return
	}

	distanceLibraries := helpers.GetDistanceLibraries(libraries, point)

	// n nearest libraries
	distanceLibraries = distanceLibraries[:intCount]

	// return json with count and postcode for testing
	value, err := helpers.EncodeJson(w, distanceLibraries)

	if err != nil {
		http.Error(w, "Error encoding json", http.StatusInternalServerError)
		return
	}

	w.Write(value)
}
