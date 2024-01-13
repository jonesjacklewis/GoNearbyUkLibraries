package helpers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
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

func DecodeJson(body io.ReadCloser, v interface{}) error {
	decoder := json.NewDecoder(body)

	decoder.DisallowUnknownFields() // Optional: prevent decoding if unknown fields are present

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
