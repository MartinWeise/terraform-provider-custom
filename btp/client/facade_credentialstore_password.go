package client

import (
	"context"
	"net/http"
	"terraform-provider-custom/btp/client/btp"
)

func newCredentialStorePasswordFacade(client *v2Client) credentialStorePasswordFacade {
	return credentialStorePasswordFacade{client: client}
}

type credentialStorePasswordFacade struct {
	client *v2Client
}

func (f *credentialStorePasswordFacade) CreateOrUpdateByNamespace(ctx context.Context, namespace string, args *btp.NewPassword) (*http.Response, error) {
	extraHeaders := map[string]string{
		HeaderContentType:        HeaderJoseValue,
		HeaderCredStoreNamespace: namespace,
	}

	res, err := f.client.DoCredStoreRequest(ctx, http.MethodPost, extraHeaders, "password", nil, args)

	knownErrors := map[int]string{
		http.StatusUnauthorized:          "Authentication failed",
		http.StatusPaymentRequired:       "You have reached the maximum credentials count or size for the service instance",
		http.StatusForbidden:             "Operation can not be executed due to insufficient permissions",
		http.StatusConflict:              "Credential operation can not be executed because the credential is modified concurrently",
		http.StatusRequestEntityTooLarge: "The credential exceed the single credential size limit",
	}

	if responseError := handleError(res, err, knownErrors); responseError != nil {
		return nil, responseError
	}

	return res, nil
}

func (f *credentialStorePasswordFacade) ReadByNamespace(ctx context.Context, namespace string, name string) (*http.Response, error) {
	extraHeaders := map[string]string{
		HeaderCredStoreNamespace: namespace,
	}
	queryParams := map[string]string{
		"name": name,
	}

	res, err := f.client.DoCredStoreRequest(ctx, http.MethodGet, extraHeaders, "password", queryParams, nil)

	knownErrors := map[int]string{
		http.StatusNotModified:  "Credential with the specified name is not changed. I.e. the version identifier provided by the client in the If-None-Match header is the same as the currently stored credential in the credential store. The client can continue to use the value it has.",
		http.StatusUnauthorized: "Authentication failed",
		http.StatusForbidden:    "Operation can not be executed due to insufficient permissions",
		http.StatusNotFound:     "Credential with the specified name is not found",
	}

	if responseError := handleError(res, err, knownErrors); responseError != nil {
		return res, responseError
	}

	return res, nil
}

func (f *credentialStorePasswordFacade) DeleteByNamespace(ctx context.Context, namespace string, name string) (*http.Response, error) {
	extraHeaders := map[string]string{
		HeaderCredStoreNamespace: namespace,
	}
	queryParams := map[string]string{
		"name": name,
	}

	res, err := f.client.DoCredStoreRequest(ctx, http.MethodDelete, extraHeaders, "password", queryParams, nil)

	knownErrors := map[int]string{
		http.StatusUnauthorized: "Authentication failed",
		http.StatusForbidden:    "Operation can not be executed due to insufficient permissions",
		http.StatusNotFound:     "Credential with the specified name is not found",
		http.StatusConflict:     "Credential operation can not be executed because the credential is modified concurrently",
	}

	if responseError := handleError(res, err, knownErrors); responseError != nil {
		return res, responseError
	}

	return res, nil
}
