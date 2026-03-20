package models

type ContactDetail struct {
	ID           string `json:"id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `json:"email"`
	Phone        string `json:"phone,omitempty"`
	Organization string `json:"organization,omitempty"`
	Street1      string `json:"street1,omitempty"`
	Street2      string `json:"street2,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	PostalCode   string `json:"postalCode,omitempty"`
	Country      string `json:"country,omitempty"`
}

type ContactCreateRequest struct {
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `json:"email"`
	Phone        string `json:"phone,omitempty"`
	Organization string `json:"organization,omitempty"`
	Street1      string `json:"street1,omitempty"`
	Street2      string `json:"street2,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	PostalCode   string `json:"postalCode,omitempty"`
	Country      string `json:"country,omitempty"`
}

type ContactActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DomainContactsResponse struct {
	Success  bool            `json:"success"`
	Message  string          `json:"message"`
	Contacts []ContactDetail `json:"contacts"`
}
