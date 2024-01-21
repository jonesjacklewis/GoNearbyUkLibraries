package helpers

import (
	"crypto/rand"
	"encoding/json"
	"example/hello/models"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func GenerateToken() string {
	length := 32

	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-" // 64 characters

	bytes := make([]byte, length)

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))

		if err != nil {
			panic(err)
		}

		bytes[i] = letters[n.Int64()]
	}

	return fmt.Sprintf("%s", bytes)
}

func DecodeJson(body io.ReadCloser, v interface{}, allowUnknownFields bool) error {
	decoder := json.NewDecoder(body)

	if !allowUnknownFields {
		decoder.DisallowUnknownFields() // Optional: prevent decoding if unknown fields are present
	}

	err := decoder.Decode(v)

	if err != nil {
		return err
	}

	return nil
}

func EncodeJson(h http.ResponseWriter, v interface{}) ([]byte, error) {
	h.Header().Set("Content-Type", "application/json")

	b, err := json.Marshal(v)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func GetPointFromString(pointString string) (models.Point, error) {
	// e.g. Point(-3.56812 55.8521)
	var point models.Point

	pointString = strings.TrimPrefix(pointString, "Point(")
	pointString = strings.TrimSuffix(pointString, ")")

	parts := strings.Split(pointString, " ") // e.g. [-3.56812 55.8521]

	if len(parts) != 2 {
		return models.Point{}, fmt.Errorf("Invalid point string")
	}

	longitudeString := parts[0]
	latitudeString := parts[1]

	longitude, err := strconv.ParseFloat(longitudeString, 64)

	if err != nil {
		return models.Point{}, err
	}

	latitude, err := strconv.ParseFloat(latitudeString, 64)

	if err != nil {
		return models.Point{}, err
	}

	point = models.Point{
		Latitude:  latitude,
		Longitude: longitude,
	}

	return point, nil
}

func GetDistanceBetweenTwoPoints(point1 models.Point, point2 models.Point) float64 {
	earthRadius := 6371.0 // km

	// Convert to radians
	lat1 := point1.Latitude * math.Pi / 180.0
	lon1 := point1.Longitude * math.Pi / 180.0

	lat2 := point2.Latitude * math.Pi / 180.0
	lon2 := point2.Longitude * math.Pi / 180.0

	// Haversine formula
	dlon := lon2 - lon1
	dlat := lat2 - lat1

	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dlon/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := earthRadius * c

	return distance
}

func SortDistanceLibraries(distanceLibraries []models.DistanceLibrary) []models.DistanceLibrary {
	// Sort by distance
	// https://stackoverflow.com/a/28937386/1063392

	sort.Slice(distanceLibraries, func(i, j int) bool {
		return distanceLibraries[i].Distance < distanceLibraries[j].Distance
	})

	return distanceLibraries
}

func GetDistanceLibraries(libraries []models.Library, point models.Point) []models.DistanceLibrary {
	var distanceLibraries []models.DistanceLibrary

	for _, library := range libraries {
		distance := GetDistanceBetweenTwoPoints(library.Point, point)

		distanceLibrary := models.DistanceLibrary{
			Library:  library,
			Distance: distance,
		}

		distanceLibraries = append(distanceLibraries, distanceLibrary)

	}

	// sort by distance

	distanceLibraries = SortDistanceLibraries(distanceLibraries)

	return distanceLibraries
}
