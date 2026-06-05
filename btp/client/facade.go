package client

func NewClientFacade(client *v2Client) *ClientFacade {
	return &ClientFacade{
		v2Client:         client,
		CredentialStore:  newCredentialStoreFacade(client),
		DestinationTrust: newDestinationTrustFacade(client),
		BusinessPartner:  newBusinessPartnerFacade(client),
	}
}

type ClientFacade struct {
	*v2Client
	CredentialStore  credentialStoreFacade
	DestinationTrust destinationTrustFacade
	BusinessPartner  businessPartnerFacade
}
