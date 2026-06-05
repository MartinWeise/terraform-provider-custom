package client

import (
	"io"

	"github.com/go-jose/go-jose/v4/json"
)

func encodeJson(content any) (*string, error) {
	bytes, err := json.Marshal(content)

	if err != nil {
		return nil, err
	}

	payload := string(bytes)

	return &payload, nil
}

func Parse[T interface{}](content io.Reader) (T, error) {
	var obj T
	var err error

	if err = json.NewDecoder(content).Decode(&obj); err == nil || err == io.EOF {
		return obj, nil
	}

	return obj, err
}
