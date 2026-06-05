package btp

// TrustCertificateResponse response when reading a trust certificate from the Destination Service
type TrustCertificateResponse struct {
	Certificate string `json:"certificate"`
}
