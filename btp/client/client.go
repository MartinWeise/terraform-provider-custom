package client

import (
	"bytes"
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/oauth2/clientcredentials"
)

type Error struct {
	Code        string `json:"errorCode,omitempty"`
	Description string `json:"errorDescription,omitempty"`
	TrackingID  string `json:"trackingId,omitempty"`
}

type RetryConfig struct {
	Enabled      bool
	RetryMax     int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
}

type CredStoreBindingParameters struct {
	Url                  *url.URL
	Username             string
	Password             string
	EncryptionPrivateKey *rsa.PrivateKey
	EncryptionPublicKey  *rsa.PublicKey
}

type OAuthTokenFlowBindingParameters struct {
	TokenUrl     *url.URL
	ClientId     string
	ClientSecret string
	ServiceUrl   *url.URL
}

type OAuthTokenFlowStubBindingParameters struct {
	TokenStubUrl *url.URL
	ClientId     string
	ClientSecret string
	ServiceUrl   *url.URL
}

func NewRetryableHttpClient(cfg *RetryConfig) *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	if cfg == nil {
		cfg = &RetryConfig{
			Enabled:      true,
			RetryMax:     6,
			RetryWaitMin: 1 * time.Second,
			RetryWaitMax: 120 * time.Second,
		}
	}
	retryClient.RetryMax = cfg.RetryMax
	retryClient.RetryWaitMin = cfg.RetryWaitMin
	retryClient.RetryWaitMax = cfg.RetryWaitMax
	retryClient.Logger = nil

	if !cfg.Enabled {
		retryClient.RetryMax = 0
		retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			return false, nil
		}
		retryClient.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
			return 0
		}
		return retryClient
	}

	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		// Retry on transient network errors and specific HTTP status codes (429, 500, 502, 503)
		if err != nil {
			return true, nil
		}
		if resp == nil {
			return false, nil
		}

		switch resp.StatusCode {
		case http.StatusTooManyRequests, // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout:      // 504
			// retry only these
			return true, nil

		case http.StatusBadRequest: // 400
			// Peek into the body to check for specific error codes/messages
			const maxBodyPeek = 4096
			var buf bytes.Buffer
			tee := io.TeeReader(io.LimitReader(resp.Body, maxBodyPeek), &buf)
			peekBytes, _ := io.ReadAll(tee)

			resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(buf.Bytes()), resp.Body))

			if strings.Contains(string(peekBytes), "[Error: 30004/400]") {
				return true, nil // for locking scenario API call must be retried
			}
			return false, nil
		default:
			// do not retry on 4xx cli errors, or other 5xx errors
			return false, nil
		}
	}
	return retryClient
}

func NewV2ClientWithHttpClient(client *http.Client, credStoreParams *CredStoreBindingParameters,
	destinationParams *OAuthTokenFlowBindingParameters, cslnCoreParams *OAuthTokenFlowBindingParameters,
	cslnParticipantsParams *OAuthTokenFlowStubBindingParameters, terraformVersion string) *v2Client {
	retryClient := NewRetryableHttpClient(nil)
	retryClient.HTTPClient = client
	return &v2Client{
		httpClient:             retryClient.StandardClient(),
		CredStoreParams:        credStoreParams,
		DestinationParams:      destinationParams,
		CSLNCoreParams:         cslnCoreParams,
		CSLNParticipantsParams: cslnParticipantsParams,
		UserAgent:              fmt.Sprintf("Terraform/%s terraform-provider-custom/dev", terraformVersion),
	}
}

const (
	HeaderAccept             string = "Accept"
	HeaderUserAgent          string = "User-Agent"
	HeaderContentType        string = "Content-Type"
	HeaderCredStoreNamespace string = "sapcp-credstore-namespace"
	HeaderJoseValue          string = "application/jose"
	HeaderJsonValue          string = "application/json"
)

type v2Client struct {
	httpClient *http.Client
	hanaClient *sql.DB

	CredStoreParams        *CredStoreBindingParameters
	DestinationParams      *OAuthTokenFlowBindingParameters
	CSLNCoreParams         *OAuthTokenFlowBindingParameters
	CSLNParticipantsParams *OAuthTokenFlowStubBindingParameters

	UserAgent string
}

func (v2 *v2Client) DoCredStoreRequest(ctx context.Context, method string, extraHeaders map[string]string, path string,
	queryParams map[string]string, body any) (*http.Response, error) {
	fullyQualifiedUrl := v2.CredStoreParams.Url.JoinPath(path)
	query := fullyQualifiedUrl.Query()
	payload := new(bytes.Buffer) // will crash when attempting to pass <nil> to http

	for key, value := range queryParams {
		query.Set(key, value)
	}

	fullyQualifiedUrl.RawQuery = query.Encode()

	var err error

	if body != nil {
		payload, err = EncryptPayload(ctx, body, *v2.CredStoreParams.EncryptionPublicKey)

		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullyQualifiedUrl.String(), payload)

	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(v2.CredStoreParams.Username, v2.CredStoreParams.Password)
	req.Header.Set(HeaderUserAgent, v2.UserAgent)

	for name, value := range extraHeaders {
		req.Header.Set(name, value)
	}

	tflog.Debug(ctx, "Do request with config:", map[string]any{
		"method":  method,
		"headers": req.Header,
		"url":     fullyQualifiedUrl.String(),
	})

	return v2.httpClient.Do(req)
}

func (v2 *v2Client) DoDestinationTrustRequest(ctx context.Context, method string, path string, query *string,
	body any) (*http.Response, error) {
	return v2.genericOAuthTokenFlowRequest(ctx, method, v2.DestinationParams.ServiceUrl, path, query, body,
		v2.DestinationParams.ClientId, v2.DestinationParams.ClientSecret, v2.DestinationParams.TokenUrl)
}

func (v2 *v2Client) DoServiceCoreRequest(ctx context.Context, method string, path string, query *string,
	body any) (*http.Response, error) {

	if v2.CSLNParticipantsParams == nil {
		return nil, fmt.Errorf("missing CSLN Core params")
	}

	return v2.genericOAuthTokenFlowRequest(ctx, method, v2.CSLNCoreParams.ServiceUrl, path, query, body,
		v2.CSLNCoreParams.ClientId, v2.CSLNCoreParams.ClientSecret, v2.CSLNCoreParams.TokenUrl)
}

func (v2 *v2Client) DoServiceTenantRequest(ctx context.Context, method string, tenantSubdomain string,
	path string, query *string, body any) (*http.Response, error) {

	if v2.CSLNParticipantsParams == nil {
		return nil, fmt.Errorf("missing CSLN Participants params")
	}

	tokenUrl, err := url.Parse(v2.CSLNParticipantsParams.TokenStubUrl.String())

	if err != nil {
		return nil, err
	}

	tokenUrl.Host = tenantSubdomain + "." + tokenUrl.Hostname()

	return v2.genericOAuthTokenFlowRequest(ctx, method, v2.CSLNParticipantsParams.ServiceUrl, path, query, body,
		v2.CSLNParticipantsParams.ClientId, v2.CSLNParticipantsParams.ClientSecret, tokenUrl)
}

func (v2 *v2Client) genericOAuthTokenFlowRequest(ctx context.Context, method string, url *url.URL, path string,
	query *string, body any, clientId string, clientSecret string, tokenUrl *url.URL) (*http.Response, error) {
	fullyQualifiedUrl := url.JoinPath(path)

	if query != nil {
		fullyQualifiedUrl.RawQuery = *query
	}

	// request access token
	// #ref: https://cs.opensource.google/go/x/oauth2/+/refs/tags/v0.35.0:clientcredentials/clientcredentials_test.go
	//
	conf := &clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       []string{},
		TokenURL:     tokenUrl.String(),
	}

	configDetail := map[string]any{
		"client_id":     clientId,
		"client_secret": "(hidden)",
		"token_url":     tokenUrl.String(),
		"method":        method,
		"url":           fullyQualifiedUrl.String(),
	}

	var payload io.Reader

	if body != nil {
		raw, err := encodeJson(body)

		if err != nil {
			return nil, err
		}

		configDetail["body"] = *raw

		payload = strings.NewReader(*raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullyQualifiedUrl.String(), payload)

	if err != nil {
		return nil, err
	}

	req.Header.Set(HeaderAccept, HeaderJsonValue)
	req.Header.Set(HeaderUserAgent, v2.UserAgent)

	if body != nil {
		req.Header.Set(HeaderContentType, "application/json")
	}

	configDetail["headers"] = req.Header

	tflog.Debug(ctx, "Do request with config:", configDetail)

	return conf.Client(ctx).Do(req)
}
