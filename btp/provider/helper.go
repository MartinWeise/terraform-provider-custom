package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

type ServiceKey struct {
	ClientId     string `json:"clientid"`
	ClientSecret string `json:"clientsecret"`
}

func ExtractAttributeValue(ctx context.Context, object string, key string) (*string, error) {
	var binding ServiceKey
	var err error

	reader := strings.NewReader(object)

	if err = json.NewDecoder(reader).Decode(&binding); err == nil || err == io.EOF {
		value := reflect.ValueOf(binding)
		res := value.FieldByName(key).String()
		return &res, nil
	}

	return nil, err
}

func ExtractHostname(ctx context.Context, object string) (*string, error) {
	var hostname string
	var err error

	url, err := url.Parse(object)

    if err != nil {
        return &hostname, errors.New("not an URL")
    }
	
    hostname = strings.TrimPrefix(url.Hostname(), "www.")

	return &hostname, nil
}

func ExtractIdpOrigin(ctx context.Context, object string) (*string, error) {
	var origin string
	var err error

	raw, err := base64.StdEncoding.DecodeString(object)

    if err != nil {
        return &origin, errors.New("not an URL")
    }

	re := regexp.MustCompile("entityID=\"([^/]+)/(.{7})")
	match := re.FindStringSubmatch(string(raw))
	
    origin = match[1] + match[2]

	return &origin, nil
}