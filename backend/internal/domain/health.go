package domain

// Health describes process liveness, never business or dependency readiness.
type Health struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}
