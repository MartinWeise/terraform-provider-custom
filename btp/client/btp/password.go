package btp

// NewPassword payload for creating a new password in the Credential Store
type NewPassword struct {
	Name         string  `json:"name"`
	Value        string  `json:"value"`
	Unmodifiable bool    `json:"unmodifiable,omitempty"`
	Username     *string `json:"username,omitempty"`
	Metadata     *string `json:"metadata,omitempty"`
}

// Password response when reading a password in the Credential Store
type Password struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Value        string  `json:"value"`
	Type         string  `json:"type"`
	Unmodifiable bool    `json:"unmodifiable"`
	ModifiedAt   *string `json:"modifiedAt,omitempty"`
	Username     *string `json:"username,omitempty"`
	Metadata     *string `json:"metadata,omitempty"`
}
