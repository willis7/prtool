package model

import "time"

// PR represents a simplified Pull Request structure.
type PR struct {
	Title     string
	Body      string
	Author    string
	CreatedAt time.Time
	MergedAt  time.Time
	Labels    []string
	URL       string
}
