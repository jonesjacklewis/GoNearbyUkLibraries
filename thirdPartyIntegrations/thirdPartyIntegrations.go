package thirdPartyIntegrations

import (
	"example/hello/helpers"
	"example/hello/models"
	"net/http"
)

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
