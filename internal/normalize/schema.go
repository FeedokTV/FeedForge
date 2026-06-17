package normalize

import (
	"time"
)

type Type string

const (
	TypeURL    Type = "url"
	TypeDomain Type = "domain"
	TypeIP     Type = "ip"
	TypeIPv6   Type = "ipv6"
	TypeHash   Type = "hash"
	TypeEmail  Type = "email"
)

type Record struct {
	ID         string            `json:"id"`
	Type       Type              `json:"type"`
	Source     string            `json:"source"`
	Value      string            `json:"value"`
	FirstSeen  time.Time         `json:"first_seen"`
	LastSeen   *time.Time        `json:"last_seen,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Confidence *int              `json:"confidence,omitempty"` // 0-100
	Meta       map[string]string `json:"meta,omitempty"`
}
