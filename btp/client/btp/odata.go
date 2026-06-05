package btp

import "terraform-provider-custom/btp/client/sap"

// BusinessPartnerResponse response from the internal client
type BusinessPartnerResponse struct {
	Context  string                `json:"@odata.context"`
	NextLink string                `json:"@odata.nextLink"`
	Id       string                `json:"@odata.id"`
	ETag     string                `json:"@odata.etag"`
	Value    []sap.BusinessPartner `json:"value"`
}

// BusinessPartnerContactResponse response from the personal data manager client
type BusinessPartnerContactResponse struct {
	Context string                    `json:"@odata.context"`
	Value   []sap.CompanyGroupContact `json:"value"`
}

// ValueResponse response from the internal client
type ValueResponse struct {
	Context string         `json:"@odata.context"`
	Value   OnboardContact `json:"value"`
}

// OnboardContact response from the internal client
type OnboardContact struct {
	ID         string `json:"id"`
	DisplayId  string `json:"displayId"`
	GivenName  string `json:"person_firstName"`
	FamilyName string `json:"person_lastName"`
	Email      string `json:"emailAddress"`
	Phone      string `json:"communicationData_phone"`
	Mobile     string `json:"communicationData_mobile"`
}
