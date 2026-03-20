package models

type AuditEntry struct {
	ID                    int     `json:"id"`
	Domain                *string `json:"domain"`
	Action                string  `json:"action"`
	Actor                 string  `json:"actor,omitempty"`
	ActorType             string  `json:"actorType,omitempty"`
	TenantID              string  `json:"tenantId,omitempty"`
	Environment           *string `json:"environment"`
	ClientIP              *string `json:"clientIp"`
	UserAgent             *string `json:"userAgent"`
	RegistryTransactionID *string `json:"registryTransactionId"`
	ClientTransactionID   *string `json:"clientTransactionId"`
	Success               bool    `json:"success"`
	ErrorMessage          *string `json:"errorMessage"`
	Details               *string `json:"details"`
	PreviousValues        *string `json:"previousValues"`
	NewValues             *string `json:"newValues"`
	CreatedAt             string  `json:"createdAt"`
}

type AuditPagedResponse struct {
	Data  []AuditEntry `json:"data"`
	Total int          `json:"total"`
}

