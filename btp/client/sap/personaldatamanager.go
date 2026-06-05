package sap

// CompanyGroupContact payload when reading a contact for a Business Partner in the personal data manager client
type CompanyGroupContact struct {
	Id          string `json:"ID"`
	DisplayId   string `json:"displayId"`
	FirstName   string `json:"person_firstName"`
	LastName    string `json:"person_lastName"`
	Email       string `json:"communicationData_email"`
	Phone       string `json:"communicationData_phone"`
	Mobile      string `json:"communicationData_mobile"`
	CompanyName string `json:"assignedCompanyName"`
}
