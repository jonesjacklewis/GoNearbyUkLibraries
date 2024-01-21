package thirdPartyIntegrations

import (
	"encoding/json"
	"example/hello/helpers"
	"example/hello/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

var count int = 0

func CheckPostcodeIsValid(postcode string) bool {
	request, error := http.NewRequest("GET", "https://api.postcodes.io/postcodes/"+postcode, nil)

	if error != nil {
		return false
	}

	client := &http.Client{}

	response, error := client.Do(request)

	if error != nil {
		return false
	}

	if response.StatusCode != 200 {
		return false
	}

	return true
}

func GetPointForPostcode(postcode string) (models.Point, error) {
	request, error := http.NewRequest("GET", "https://api.postcodes.io/postcodes/"+postcode, nil)

	if error != nil {
		return models.Point{}, error
	}

	client := &http.Client{}

	response, error := client.Do(request)

	if error != nil {
		return models.Point{}, error
	}

	if response.StatusCode != 200 {
		return models.Point{}, error
	}

	var postcodeResponse models.PostcodeResponse

	error = helpers.DecodeJson(response.Body, &postcodeResponse, true)

	if error != nil {
		return models.Point{}, error
	}

	latitude := postcodeResponse.Result.Latitude
	longitude := postcodeResponse.Result.Longitude

	point := models.Point{
		Latitude:  latitude,
		Longitude: longitude,
	}

	return point, nil
}

func GetAllLibraries() ([]models.Library, error) {
	// Example constants and query (replace with actual values)
	const sparqlWikidataURL = "https://query.wikidata.org/sparql" // Replace with actual SPARQL Wikidata URL

	file, err := os.ReadFile("sparql/getLibraries.sparql")

	if err != nil {
		panic(err)
	}

	// Construct the user agent string
	version := runtime.Version()
	majorMinor := strings.Split(version, ".")[:2]
	userAgent := fmt.Sprintf("WDQS-example Go/%s.%s", majorMinor[0][2:], majorMinor[1])

	// Create headers
	headers := map[string][]string{
		"Accept":     {"application/sparql-results+json"},
		"User-Agent": {userAgent},
	}

	// Prepare the request
	client := &http.Client{}
	req, err := http.NewRequest("GET", sparqlWikidataURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	// Add headers to the request
	for key, value := range headers {
		req.Header[key] = value
	}

	// convert file to string
	sparqlQuery := string(file)

	// Add query parameters
	q := req.URL.Query()
	q.Add("query", sparqlQuery)
	req.URL.RawQuery = q.Encode()

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error performing request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	var libraries []models.Library

	// loop through results
	var response models.AutoGenerated
	err = json.Unmarshal(body, &response)

	if err != nil {

		if count < 5 {
			count++
			GetAllLibraries()
			time.Sleep(1 * time.Second)
		} else {
			panic(err)
		}
	}

	for _, binding := range response.Results.Bindings {

		pointString := binding.Coord.Value

		point, err := helpers.GetPointFromString(pointString)

		if err != nil {
			continue
		}

		library := models.Library{
			Name:  binding.ItemLabel.Value,
			Point: point,
		}

		libraries = append(libraries, library)
	}

	return libraries, nil

}
