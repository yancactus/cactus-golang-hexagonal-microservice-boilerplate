package dto

import "time"

// AuditLogResp represents an audit log response
type AuditLogResp struct {
	ID         string                 `json:"id"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Action     string                 `json:"action"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	UserID     *int                   `json:"user_id,omitempty"`
}
