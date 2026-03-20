package models

// SuggestResponse is the response for suggest, spin-word, add-prefix, add-suffix endpoints.
type SuggestResponse struct {
	Results []DomainSuggestion `json:"results"`
}

type DomainSuggestion struct {
	Name         string `json:"name"`
	PunyName     string `json:"punyName,omitempty"`
	Availability string `json:"availability"`
	Token        string `json:"token,omitempty"`
}

// BulkSuggestRequest is the request body for POST /namesuggestions/bulk-suggest.
type BulkSuggestRequest struct {
	Names             []string `json:"names"`
	TLDs              string   `json:"tlds,omitempty"`
	Lang              string   `json:"lang,omitempty"`
	UseNumbers        bool     `json:"useNumbers,omitempty"`
	MaxResultsPerName int      `json:"maxResultsPerName,omitempty"`
}

// BulkSuggestResponse is the response for POST /namesuggestions/bulk-suggest.
type BulkSuggestResponse struct {
	Suggestions []NameSuggestions  `json:"suggestions"`
	Errors      []BulkRequestError `json:"errors,omitempty"`
}

type NameSuggestions struct {
	OriginalName string             `json:"originalName"`
	Suggestions  []DomainSuggestion `json:"suggestions"`
}

type BulkRequestError struct {
	OriginalName string `json:"originalName"`
	ErrorMessage string `json:"errorMessage"`
}

// BulkAvailabilityResponse is the response for keyword-availability endpoints.
type BulkAvailabilityResponse struct {
	Keyword            string              `json:"keyword"`
	Results            []DomainAvailability `json:"results"`
	AvailabilityStats  map[string]int      `json:"availabilityStats,omitempty"`
	TotalDomains       int64               `json:"totalDomains"`
	AvailableDomains   int64               `json:"availableDomains"`
	UnavailableDomains int64               `json:"unavailableDomains"`
	ProcessingTimeMs   int64               `json:"processingTimeMs,omitempty"`
}

type DomainAvailability struct {
	Domain       string  `json:"domain"`
	TLD          string  `json:"tld,omitempty"`
	Registry     string  `json:"registry,omitempty"`
	Availability string  `json:"availability"`
	Reason       string  `json:"reason,omitempty"`
	Premium      bool    `json:"premium,omitempty"`
	FeeClass     string  `json:"feeClass,omitempty"`
	Price        float64 `json:"price,omitempty"`
	IcannFee     int     `json:"icannFee,omitempty"`
	RegistrarFee int     `json:"registrarFee,omitempty"`
	TotalPrice   float64 `json:"totalPrice,omitempty"`
}
