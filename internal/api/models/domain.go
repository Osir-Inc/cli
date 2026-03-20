package models

type DomainAvailabilityResponse struct {
	Domain    string  `json:"domain"`
	Available bool    `json:"available"`
	Status    string  `json:"status"`
	Message   string  `json:"message"`
	Price     float64 `json:"price,omitempty"`
	Currency  string  `json:"currency,omitempty"`
}

type DomainRegistrationRequest struct {
	Domain      string   `json:"domain"`
	Period      int      `json:"period"`
	Nameservers []string `json:"nameservers"`
	Privacy     bool     `json:"privacyProtection"`
	AutoRenew   bool     `json:"autoRenew"`
}

type DomainRegistrationResponse struct {
	Success       bool    `json:"success"`
	Message       string  `json:"message"`
	Domain        string  `json:"domain"`
	Status        string  `json:"status"`
	TransactionID string  `json:"transactionId,omitempty"`
	Price         float64 `json:"price,omitempty"`
	Currency      string  `json:"currency,omitempty"`
}

type DomainInfoResponse struct {
	Success   bool            `json:"success"`
	Timestamp string          `json:"timestamp,omitempty"`
	Data      *DomainInfoData `json:"data,omitempty"`
}

type DomainInfoData struct {
	Domain                      string   `json:"domain"`
	Status                      string   `json:"status"`
	Statuses                    []string `json:"statuses,omitempty"`
	ExpiryDate                  string   `json:"expiryDate,omitempty"`
	CreationDate                string   `json:"creationDate,omitempty"`
	Nameservers                 []string `json:"nameservers,omitempty"`
	Locked                      bool     `json:"locked"`
	Expired                     bool     `json:"expired"`
	Premium                     bool     `json:"premium"`
	AutoRenew                   bool     `json:"autoRenew"`
	Privacy                     bool     `json:"privacy"`
	RegistrantEmail             string   `json:"registrantEmail,omitempty"`
	Registrar                   string   `json:"registrar,omitempty"`
	InRedemptionPeriod          bool     `json:"inRedemptionPeriod"`
	RgpStatus                   *string  `json:"rgpStatus,omitempty"`
	RedemptionEndDate           *string  `json:"redemptionEndDate,omitempty"`
	InAutoRenewGracePeriod      bool     `json:"inAutoRenewGracePeriod"`
	AutoRenewGracePeriodEndDate *string  `json:"autoRenewGracePeriodEndDate,omitempty"`
	DnssecEnabled               bool     `json:"dnssecEnabled"`
	DnssecSupported             bool     `json:"dnssecSupported"`
	DnssecRecords               []any    `json:"dnssecRecords,omitempty"`
}

type DomainListResponse struct {
	Success bool           `json:"success"`
	Data    DomainListData `json:"data"`
}

type DomainListData struct {
	Domains       []DomainSummary `json:"domains"`
	Page          int             `json:"page"`
	Size          int             `json:"size"`
	TotalElements int             `json:"totalElements"`
	TotalPages    int             `json:"totalPages"`
	SortBy        string          `json:"sortBy"`
	SortDirection string          `json:"sortDirection"`
}

type DomainSummary struct {
	ID             int      `json:"id"`
	Domain         string   `json:"domain"`
	CreationDate   string   `json:"creationDate,omitempty"`
	ExpirationDate string   `json:"expirationDate,omitempty"`
	Statuses       []string `json:"statuses,omitempty"`
	Nameservers    []string `json:"nameservers,omitempty"`
	AutoRenew      bool     `json:"autoRenew"`
	Privacy        bool     `json:"privacy"`
	Status         string   `json:"status"`
}

type DomainRenewalRequest struct {
	Period int `json:"period"`
}

type DomainRenewalResponse struct {
	Success        bool    `json:"success"`
	Message        string  `json:"message"`
	Domain         string  `json:"domain"`
	ExpirationDate string  `json:"expirationDate,omitempty"`
	Price          float64 `json:"price,omitempty"`
}


type ValidationResult struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

type NameserverUpdateRequest struct {
	Nameservers []string `json:"nameservers"`
}


type DomainSuggestionsResponse struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message"`
	Suggestions []string `json:"suggestions"`
}
