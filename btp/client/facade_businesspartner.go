package client

func newBusinessPartnerFacade(cliClient *v2Client) businessPartnerFacade {
	return businessPartnerFacade{
		CompanyGroup:        newBusinessPartnerCompanyGroupFacade(cliClient),
		CompanyGroupContact: newBusinessPartnerContactFacade(cliClient),
	}
}

type businessPartnerFacade struct {
	CompanyGroup        businessPartnerCompanyGroupFacade
	CompanyGroupContact businessPartnerCompanyGroupContactFacade
}
