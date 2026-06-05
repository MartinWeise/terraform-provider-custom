package client

import (
	"context"
	"net/http"
)

func newDestinationTrustMetadataFacade(client *v2Client) destinationTrustMetadataFacade {
	return destinationTrustMetadataFacade{client: client}
}

type destinationTrustMetadataFacade struct {
	client *v2Client
}

func (f *destinationTrustMetadataFacade) Read(ctx context.Context) (*http.Response, error) {
	res, err := f.client.DoDestinationTrustRequest(ctx, http.MethodGet, "destination-configuration/v1/saml2Metadata", nil, nil)

	knownErrors := map[int]string{
		http.StatusBadRequest:   "Indicates a problem with the request, e.g., malformed Authorization header.",
		http.StatusUnauthorized: "Authentication failed.",
		http.StatusForbidden:    "Operation can not be executed due to insufficient permissions.",
		http.StatusNotFound:     "The requested resource could not be found or the subaccount has no Key Pair generated.",
	}

	if responseError := handleError(res, err, knownErrors); responseError != nil {
		return res, responseError
	}

	return res, nil
}
