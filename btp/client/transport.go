package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// generic function to handle HTTP-related errors
func handleError(response *http.Response, err error, knownErrors map[int]string) error {

	if err != nil {
		return err
	}

	if response.StatusCode == 429 {
		return fmt.Errorf("rate limits exceeded")
	}

	for httpStatus, description := range knownErrors {
		if response.StatusCode == httpStatus {
			return fmt.Errorf("%s", description)
		}
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		detail := "(unknown)"
		buffer := new(strings.Builder)
		_, bufErr := io.Copy(buffer, response.Body)

		if bufErr == nil {
			detail = buffer.String()
		}

		return fmt.Errorf("unknown error, expected 2xx but was: %d. Detail: %s", response.StatusCode, detail)
	}

	return nil
}
