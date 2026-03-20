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
	PackageID    string `json:"packageId"`
	PaymentTerm  string `json:"paymentTerm"`
	Hostname     string `json:"hostname"`
	LocationID   string `json:"locationId,omitempty"`
	RootPassword string `json:"rootPassword,omitempty"`
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
	VirtfusionInstanceID int    `json:"virtfusionInstanceId,omitempty"`
}
