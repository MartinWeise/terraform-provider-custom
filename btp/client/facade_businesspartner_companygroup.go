package client

import (
	"context"
	"net/http"
	"net/url"
	"terraform-provider-custom/btp/client/sap"
)

func newBusinessPartnerCompanyGroupFacade(client *v2Client) businessPartnerCompanyGroupFacade {
	return businessPartnerCompanyGroupFacade{client: client}
}

type businessPartnerCompanyGroupFacade struct {
	client *v2Client
}

func (f *businessPartnerCompanyGroupFacade) ReadByDisplayId(ctx context.Context, displayId string) (*http.Response, error) {
	query := "$filter=displayId%20eq%20%27" + url.QueryEscape(displayId) + "%27"
	res, err := f.client.DoServiceCoreRequest(ctx, http.MethodGet, "internalclient/BusinessPartners", &query, nil)

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

func (f *businessPartnerCompanyGroupFacade) UpdateSubdomain(ctx context.Context, id string, subdomain string) (*http.Response, error) {
	res, err := f.client.DoServiceCoreRequest(ctx, http.MethodPost, "internalclient/BusinessPartners("+id+")/InternalClientService.updateSubdomain",
		nil, &sap.UpdateSubdomain{
			Subdomain: subdomain,
		})

	knownErrors := map[int]string{
		http.StatusBadRequest:   "Already exists", // FIXME this circumvents data integrity and should be fixed
		http.StatusUnauthorized: "Authentication failed",
		http.StatusForbidden:    "Operation can not be executed due to insufficient permissions",
		http.StatusNotFound:     "The requested resource could not be found in the application",
	}

	if responseError := handleError(res, err, knownErrors); responseError != nil {
		return nil, responseError
	}

	return res, nil
}

func (f *businessPartnerCompanyGroupFacade) UpdateTenantUuid(ctx context.Context, id string, tenantUuid string) (*http.Response, error) {
	res, err := f.client.DoServiceCoreRequest(ctx, http.MethodPost, "internalclient/BusinessPartners("+id+")/InternalClientService.updateTenantUUID",
		nil, &sap.UpdateTenantUuid{
			TenantUuid: tenantUuid,
		})

	knownErrors := map[int]string{
		http.StatusBadRequest:   "Already exists", // FIXME this circumvents data integrity and should be fixed
		http.StatusUnauthorized: "Authentication failed",
		http.StatusForbidden:    "Operation can not be executed due to insufficient permissions",
		http.StatusNotFound:     "The requested resource could not be found in the application",
	}

	if responseError := handleError(res, err, knownErrors); responseError != nil {
		return nil, responseError
	}

	return res, nil
}
