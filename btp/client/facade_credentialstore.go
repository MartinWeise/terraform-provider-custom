package client

func newCredentialStoreFacade(client *v2Client) credentialStoreFacade {
	return credentialStoreFacade{
		Password: newCredentialStorePasswordFacade(client),
	}
}

type credentialStoreFacade struct {
	Password credentialStorePasswordFacade
}
