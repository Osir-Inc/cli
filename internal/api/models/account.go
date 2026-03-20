package models

type UserProfile struct {
	ID               int             `json:"id"`
	Customer         string          `json:"customer"`
	Details          *ProfileDetails `json:"details"`
	Balance          *SummaryBalance `json:"balance"`
	CreatedAt        string          `json:"createdAt,omitempty"`
	LastLogin        string          `json:"lastLogin,omitempty"`
	HasCompletedDetails bool         `json:"hasCompletedDetails"`
	OrderCount       int             `json:"orderCount"`
	TotalDomains     int             `json:"totalDomains"`
	Active           bool            `json:"active"`
}

type ProfileDetails struct {
	Name           string  `json:"name"`
	Surname        string  `json:"surname"`
	Organization   string  `json:"organization,omitempty"`
	OrganizationID *string `json:"organizationId,omitempty"`
	Street         string  `json:"street,omitempty"`
	Street2        string  `json:"street2,omitempty"`
	City           string  `json:"city,omitempty"`
	State          string  `json:"state,omitempty"`
	PostalCode     string  `json:"postalCode,omitempty"`
	Country        string  `json:"country,omitempty"`
	Phone          string  `json:"phone,omitempty"`
	Fax            *string `json:"fax,omitempty"`
	Email          string  `json:"email,omitempty"`
}

type AccountSummary struct {
	DomainRegistry []SummaryDomain  `json:"domainRegistry"`
	Balances       []SummaryBalance `json:"balances"`
	Orders         []SummaryOrder   `json:"orders"`
	Messages       []SummaryMessage `json:"messages"`
	DomainCount    int              `json:"domainCount"`
	VpsCount       int              `json:"vpsCount"`
	ContactCount   int              `json:"contactCount"`
	TalentCount    int              `json:"talentCount"`
	HasCompleteProfile bool         `json:"hasCompleteProfile"`
}

type SummaryDomain struct {
	Domain         string   `json:"domain"`
	ExpirationDate string   `json:"expirationDate"`
	AutoRenew      bool     `json:"autorenew"`
	Privacy        bool     `json:"privacy"`
	Statuses       []string `json:"statuses"`
}

type SummaryBalance struct {
	ID                  int     `json:"id"`
	Amount              int     `json:"amount"`
	Currency            *string `json:"currency"`
	LastChangedDateTime string  `json:"lastChangedDateTime"`
	Source              string  `json:"source"`
}

type SummaryOrder struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`
	OrderDate string `json:"orderDate"`
	Domain    string `json:"domain"`
	Type      string `json:"type"`
	Amount    int    `json:"amount"`
	Paid      bool   `json:"paid"`
	Message   string `json:"message"`
}

type SummaryMessage struct {
	ID      int    `json:"id"`
	Subject string `json:"subject"`
}
