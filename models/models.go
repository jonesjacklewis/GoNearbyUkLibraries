package models

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type PostcodeResponseResult struct {
	Postcode  string  `json:"postcode"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type PostcodeResponse struct {
	Status int                    `json:"status"`
	Result PostcodeResponseResult `json:"result"`
}

type Library struct {
	Name  string
	Point Point
}

type DistanceLibrary struct {
	Library
	Distance float64
}
