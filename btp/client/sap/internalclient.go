package sap

// UpdateSubdomain payload when updating the subdomain for a Business Partner in the internal client
type UpdateSubdomain struct {
	Subdomain string `json:"subdomain"`
}

// UpdateTenantUuid payload when updating the subaccount uuid for a Business Partner in the internal client
type UpdateTenantUuid struct {
	TenantUuid string `json:"tenant"`
}

// OnboardContact payload when onboarding contact for a Business Partner in the internal client
type OnboardContact struct {
	DisplayId  string `json:"displayId"`
	GivenName  string `json:"firstName"`
	FamilyName string `json:"lastName"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Mobile     string `json:"mobile"`
}

// BusinessPartner response from the internal client
type BusinessPartner struct {
	Id        string `json:"ID"`
	DisplayId string `json:"displayId"`
}
