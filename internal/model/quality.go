package model

type Quality struct {
	HasSources bool     `json:"has_sources"`
	Score      float64  `json:"score"`
	Warnings   []string `json:"warnings,omitempty"`
}
