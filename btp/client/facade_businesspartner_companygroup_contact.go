package client

import (
	"context"
	"net/http"
	"net/url"
	"terraform-provider-custom/btp/client/sap"
)

func newBusinessPartnerContactFacade(client *v2Client) businessPartnerCompanyGroupContactFacade {
	return businessPartnerCompanyGroupContactFacade{client: client}
}

type businessPartnerCompanyGroupContactFacade struct {
	client *v2Client
}

func (f *businessPartnerCompanyGroupContactFacade) ReadByDisplayId(ctx context.Context, displayId string,
	tenantSubdomain string) (*http.Response, error) {
	query := "$filter=displayId%20eq%20%27" + url.QueryEscape(displayId) + "%27"
	res, err := f.client.DoServiceTenantRequest(ctx, http.MethodGet, tenantSubdomain, "personal-data-manager/BusinessPartnerContactInfo", &query, nil)

	knownErrors := map[int]string{
		http.StatusUnauthorized: "Authentication failed",
		http.StatusForbidden:    "Operation can not be executed due to insufficient permissions",
		http.StatusNotFound:     "The requested resource could not be found in the application",
	}

	if responseError := handleError(res, err, knownErrors); responseError != nil {
		return nil, responseError
	}

	return res, nil
}

func (f *businessPartnerCompanyGroupContactFacade) OnboardContact(ctx context.Context, tenantSubdomain string,
	contact *sap.OnboardContact) (*http.Response, error) {
	res, err := f.client.DoServiceTenantRequest(ctx, http.MethodPost, tenantSubdomain, "internalclient/onboardContact", nil, contact)

	knownErrors := map[int]string{
		http.StatusUnauthorized: "Authentication failed",
		http.StatusForbidden:    "Operation can not be executed due to insufficient permissions",
		http.StatusNotFound:     "The requested resource could not be found in the application",
	}

	if responseError := handleError(res, err, knownErrors); responseError != nil {
		return nil, responseError
	}

	return res, nil
}
