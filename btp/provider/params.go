package provider

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"terraform-provider-custom/btp/client"

	"github.com/go-jose/go-jose/v4/json"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type CredStoreCredentials struct {
	Password   string     `json:"password"`
	ExpiresAt  string     `json:"expires_at"`
	Encryption Encryption `json:"encryption"`
	Parameters Parameters `json:"parameters"`
	Uri        string     `json:"uri"`
	Url        string     `json:"url"`
	Username   string     `json:"username"`
}

type DestinationCredentials struct {
	ClientId     string `json:"clientid"`
	ClientSecret string `json:"clientsecret"`
	Uri          string `json:"uri"`
	TokenUrl     string `json:"url"`
}

type UaaCredentials struct {
	ClientId     string `json:"clientid"`
	ClientSecret string `json:"clientsecret"`
	Url          string `json:"url"`
	IdentityZone string `json:"identityzone"`
}

type Encryption struct {
	ClientPrivateKey string `json:"client_private_key"`
	ServerPublicKey  string `json:"server_public_key"`
}

type CredStoreCredentialBinding struct {
	Credentials CredStoreCredentials `json:"credentials"`
}

type DestinationCredentialBinding struct {
	Credentials DestinationCredentials `json:"credentials"`
}

type CSLNCoreCredentialBinding struct {
	Credentials UaaCredentials `json:"credentials"`
}

type Parameters struct {
	Authorization Authorization `json:"authorization"`
}

type Authorization struct {
	DefaultPermissions []string `json:"default_permissions"`
}

func parseCredStoreParams(ctx context.Context, raw string) (*client.CredStoreBindingParameters, error) {
	var credentialBinding CredStoreCredentialBinding
	err := json.Unmarshal([]byte(raw), &credentialBinding)

	if err != nil {
		return nil, err
	}

	apiUrl, err := url.Parse(credentialBinding.Credentials.Url)

	if err != nil {
		return nil, err
	}

	privKey, err := parsePrivateKey(credentialBinding.Credentials.Encryption.ClientPrivateKey)

	if err != nil {
		return nil, err
	}

	pubKey, err := parsePublicKey(credentialBinding.Credentials.Encryption.ServerPublicKey)

	if err != nil {
		return nil, err
	}

	params := client.CredStoreBindingParameters{
		Url:                  apiUrl,
		Username:             credentialBinding.Credentials.Username,
		Password:             credentialBinding.Credentials.Password,
		EncryptionPrivateKey: privKey,
		EncryptionPublicKey:  pubKey,
	}

	tflog.Debug(ctx, "configured the credential store from the binding parameters")

	return &params, nil
}

func parseDestinationParams(ctx context.Context, raw string) (*client.OAuthTokenFlowBindingParameters, error) {
	re := regexp.MustCompile(`\r?\n`)
	sanitized := re.ReplaceAllString(raw, "")

	var credentialBinding DestinationCredentialBinding
	err := json.Unmarshal([]byte(sanitized), &credentialBinding)

	if err != nil {
		return nil, err
	}

	serviceUrl, err := url.Parse(credentialBinding.Credentials.Uri)

	if err != nil {
		return nil, err
	}

	accessTokenUrl, err := url.Parse(credentialBinding.Credentials.TokenUrl + "/oauth/token")

	if err != nil {
		return nil, err
	}

	params := client.OAuthTokenFlowBindingParameters{
		TokenUrl:     accessTokenUrl,
		ServiceUrl:   serviceUrl,
		ClientId:     credentialBinding.Credentials.ClientId,
		ClientSecret: credentialBinding.Credentials.ClientSecret,
	}

	tflog.Debug(ctx, "successfully parsed the OAuth token flow parameters")

	return &params, nil
}

func parseUaaCredentials(raw string) (*CSLNCoreCredentialBinding, error) {
	re := regexp.MustCompile(`\r?\n`)
	sanitized := re.ReplaceAllString(raw, "")

	var credentialBinding CSLNCoreCredentialBinding
	err := json.Unmarshal([]byte(sanitized), &credentialBinding)

	if err != nil {
		return nil, err
	}

	return &credentialBinding, nil
}

func parseCSLNCoreParams(ctx context.Context, raw string, cslnDomain string) (*client.OAuthTokenFlowBindingParameters, error) {
	params, err := parseUaaCredentials(raw)

	if err != nil {
		return nil, err
	}

	envMatches := regexp.MustCompile(`([^-]+)$`).FindStringSubmatch(params.Credentials.IdentityZone)

	if len(envMatches) < 1 {
		return nil, fmt.Errorf("Failed to parse environment from identity zone: %s", envMatches, params.Credentials.IdentityZone)
	}

	env := envMatches[0]
	accessTokenUrl, err := url.Parse("https://sequello-core-" + env + regexp.MustCompile(`(\.authentication.*)$`).FindString(params.Credentials.Url) + "/oauth/token")

	if err != nil {
		return nil, err
	}

	serviceUrl, err := url.Parse("https://csln-" + env + "-core." + cslnDomain)

	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "parsed CSLN-Core URLs:", map[string]any{
		"access_token_url": accessTokenUrl.String(),
		"service_url":      serviceUrl.String(),
	})

	return &client.OAuthTokenFlowBindingParameters{
		ClientId:     params.Credentials.ClientId,
		ClientSecret: params.Credentials.ClientSecret,
		TokenUrl:     accessTokenUrl,
		ServiceUrl:   serviceUrl,
	}, nil
}

func parseCSLNParticipantsParams(ctx context.Context, raw string, cslnDomain string) (*client.OAuthTokenFlowStubBindingParameters, error) {
	params, err := parseUaaCredentials(raw)

	if err != nil {
		return nil, err
	}

	env := regexp.MustCompile(`([^-]+)$`).FindString(params.Credentials.IdentityZone)
	accessTokenUrl, err := url.Parse("https://" + regexp.MustCompile(`(authentication.*)$`).FindString(params.Credentials.Url) + "/oauth/token")

	if err != nil {
		return nil, err
	}

	serviceUrl, err := url.Parse("https://csln-" + env + "-participants." + cslnDomain)

	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "parsed CSLN-Participants URLs:", map[string]any{
		"access_token_stub_url": accessTokenUrl.String(),
		"service_url":           serviceUrl.String(),
	})

	return &client.OAuthTokenFlowStubBindingParameters{
		ClientId:     params.Credentials.ClientId,
		ClientSecret: params.Credentials.ClientSecret,
		TokenStubUrl: accessTokenUrl,
		ServiceUrl:   serviceUrl,
	}, nil
}
