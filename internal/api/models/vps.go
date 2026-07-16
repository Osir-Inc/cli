package models

type VpsCatalogResponse struct {
	TotalPackages      int                              `json:"totalPackages"`
	Packages           []VpsPackageDetail               `json:"packages"`
	PackagesByLocation map[string][]VpsPackageDetail    `json:"packagesByLocation,omitempty"`
}

type VpsPackageDetail struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	Description    string             `json:"description,omitempty"`
	MemoryMb       int                `json:"memoryMb"`
	StorageGb      int                `json:"storageGb"`
	CpuCores       int                `json:"cpuCores"`
	TrafficGb      int                `json:"trafficGb"`
	StorageProfile string             `json:"storageProfile"`
	PriceMonthly   int                `json:"priceMonthly"`
	PriceSemiAnnual int               `json:"priceSemiAnnual,omitempty"`
	PriceAnnual    int                `json:"priceAnnual,omitempty"`
	PriceBiennial  int                `json:"priceBiennial,omitempty"`
	PriceTriennial int                `json:"priceTriennial,omitempty"`
	Status         string             `json:"status"`
	Location       *VpsLocationDetail `json:"location,omitempty"`
	AllPrices      map[string]int     `json:"allPrices,omitempty"`
}

type VpsLocationListResponse struct {
	Locations      []VpsLocationDetail `json:"locations"`
	TotalLocations int                 `json:"totalLocations"`
}

type VpsLocationDetail struct {
	ID          string `json:"id"`
	City        string `json:"city"`
	CountryName string `json:"countryName"`
	CountryCode string `json:"countryCode"`
	FlagEmoji   string `json:"flagEmoji,omitempty"`
	DisplayName string `json:"displayName"`
}

type VpsOrderRequest struct {
	PackageID   string `json:"packageId"`
	PaymentTerm string `json:"paymentTerm"`
	Hostname    string `json:"hostname"`
	LocationID  string `json:"locationId,omitempty"`
	// Omit OperatingSystemID and the server is created with NO operating system on it. Template ids are
	// per-install and change when templates are re-imported, so resolve one via GetVpsOsTemplates rather
	// than hardcoding. SSH keys are the access path: they are injected during the install.
	OperatingSystemID int   `json:"operatingSystemId,omitempty"`
	SSHKeyIDs         []int `json:"sshKeyIds,omitempty"`
}

// VpsOsTemplate is an installable operating system, as offered for one specific instance.
type VpsOsTemplate struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Variant     string `json:"variant,omitempty"`
	Arch        int    `json:"arch,omitempty"`
	Description string `json:"description,omitempty"`
	Eol         bool   `json:"eol,omitempty"`
	EolDate     string `json:"eolDate,omitempty"`
	Type        string `json:"type,omitempty"`
}

type VpsOsTemplateListResponse struct {
	Templates []VpsOsTemplate `json:"templates"`
}

// VpsSshKey is a public key stored against your account and injectable at install time.
type VpsSshKey struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	PublicKey   string `json:"publicKey,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
}

type VpsSshKeyListResponse struct {
	Keys []VpsSshKey `json:"keys"`
}

type VpsSshKeyCreateRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"publicKey"`
}

// VpsBuildRequest installs an OS. On a server that already has one this ERASES ALL DATA.
type VpsBuildRequest struct {
	OperatingSystemID int      `json:"operatingSystemId"`
	Hostname          string   `json:"hostname,omitempty"`
	SSHKeyIDs         []int    `json:"sshKeyIds,omitempty"`
	Swap              *float64 `json:"swap,omitempty"`
}

// VpsBuildStatus is returned by the build call and mirrored onto the instance detail endpoint.
type VpsBuildStatus struct {
	BuildState     string `json:"buildState"`
	Built          string `json:"built,omitempty"`
	BuildFailed    bool   `json:"buildFailed,omitempty"`
	OsTemplateID   int    `json:"osTemplateId,omitempty"`
	OsTemplateName string `json:"osTemplateName,omitempty"`
}

type VpsPaymentTermChangeRequest struct {
	NewPaymentTerm string `json:"newPaymentTerm"`
}

// VpsInstance is the full instance object returned by the list/get endpoints.
type VpsInstance struct {
	ID                 string          `json:"id"`
	Hostname           string          `json:"hostname"`
	Status             string          `json:"status"`
	ProvisioningStatus string          `json:"provisioningStatus"`
	IPAddress          string          `json:"ipAddress"`
	IPv6Addresses      string          `json:"ipv6Addresses,omitempty"`
	PaymentTerm        string          `json:"paymentTerm"`
	NextRenewalDate    string          `json:"nextRenewalDate,omitempty"`
	VpsPackage         *VpsPackageBrief `json:"vpsPackage,omitempty"`
	HypervisorGroup    *VpsLocationRef  `json:"hypervisorGroup,omitempty"`
	CreatedAt          string          `json:"createdAt,omitempty"`
	// OS build state. Distinct from ProvisioningStatus, which only covers server *creation* and is
	// terminal at COMPLETED — a COMPLETED instance can still have no operating system on it.
	// UNBUILT | QUEUED | BUILDING | COMPLETE | FAILED
	BuildState     string `json:"buildState,omitempty"`
	Built          string `json:"built,omitempty"`
	BuildFailed    bool   `json:"buildFailed,omitempty"`
	OsTemplateID   int    `json:"osTemplateId,omitempty"`
	OsTemplateName string `json:"osTemplateName,omitempty"`
}

type VpsPackageBrief struct {
	Name       string `json:"name"`
	CpuCores   int    `json:"cpuCores"`
	MemoryMb   int    `json:"memoryMb"`
	StorageGb  int    `json:"storageGb"`
	TrafficGb  int    `json:"trafficGb"`
}

type VpsLocationRef struct {
	DisplayName string `json:"displayName"`
	City        string `json:"city"`
	CountryCode string `json:"countryCode"`
}

type VpsPanelLoginResponse struct {
	InstanceID string `json:"instanceId"`
	Hostname   string `json:"hostname"`
	LoginURL   string `json:"loginUrl"`
	Message    string `json:"message,omitempty"`
}

type VpsOrderResponse struct {
	OrderID       int             `json:"orderId"`
	OrderNumber   string          `json:"orderNumber"`
	OrderStatus   string          `json:"orderStatus"`
	InvoiceID     int             `json:"invoiceId"`
	InvoiceNumber string          `json:"invoiceNumber"`
	TotalAmount   int             `json:"totalAmount"`
	Currency      string          `json:"currency"`
	DueDate       string          `json:"dueDate,omitempty"`
	Instance      *VpsInstanceInfo `json:"instance,omitempty"`
}

type VpsInstanceInfo struct {
	ID                   string `json:"id"`
	Hostname             string `json:"hostname"`
	PackageName          string `json:"packageName"`
	Status               string `json:"status"`
	ProvisioningStatus   string `json:"provisioningStatus"`
	IPAddress            string `json:"ipAddress"`
	IPv6Addresses        string `json:"ipv6Addresses,omitempty"`
	IPv6Subnet           string `json:"ipv6Subnet,omitempty"`
	IPv6Cidr             int    `json:"ipv6Cidr,omitempty"`
	IPv6Gateway          string `json:"ipv6Gateway,omitempty"`
	ControlPanelUrl      string `json:"controlPanelUrl,omitempty"`
	Message              string `json:"message,omitempty"`
	VirtfusionInstanceID string `json:"virtfusionInstanceId,omitempty"`
}
