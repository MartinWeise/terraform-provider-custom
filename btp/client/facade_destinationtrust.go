package client

func newDestinationTrustFacade(client *v2Client) destinationTrustFacade {
	return destinationTrustFacade{
		Certificate: newDestinationTrustCertificateFacade(client),
		Metadata:    newDestinationTrustMetadataFacade(client),
	}
}

type destinationTrustFacade struct {
	Certificate destinationTrustCertificateFacade
	Metadata    destinationTrustMetadataFacade
}
