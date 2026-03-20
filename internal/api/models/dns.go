package models

type DnsRecord struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority,omitempty"`
}

type DnsRecordRequest struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

type DnsRecordUpdateRequest struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority,omitempty"`
}

type DnsActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DnsRecordListResponse struct {
	Records []DnsRecord `json:"records"`
}

type ZoneResponse struct {
	Name    string `json:"name,omitempty"`
	Kind    string `json:"kind,omitempty"`
	Account string `json:"account,omitempty"`
	Message string `json:"message,omitempty"`
}

type ZoneExistsResponse struct {
	Exists bool `json:"exists"`
}

type DnssecStatusResponse struct {
	Enabled bool   `json:"enabled"`
	Status  string `json:"status,omitempty"`
}

type DnssecDisableResponse struct {
	Message            string  `json:"message"`
	RegistryPublished  *bool   `json:"registryPublished"`
	RegistryPublishErr *string `json:"registryPublishError"`
}

type DnssecEnableResponse struct {
	Domain             string      `json:"domain"`
	DnssecEnabled      bool        `json:"dnssecEnabled"`
	CryptoKeys         interface{} `json:"cryptoKeys,omitempty"`
	DsRecords          interface{} `json:"dsRecords,omitempty"`
	RegistryPublished  *bool       `json:"registryPublished"`
	RegistryPublishErr *string     `json:"registryPublishError"`
}
