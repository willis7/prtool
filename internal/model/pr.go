package model

import "time"

// PR represents a GitHub pull request with the essential fields we need
type PR struct {
	Title      string
	Body       string
	Author     string
	CreatedAt  time.Time
	MergedAt   *time.Time
	Labels     []string
	FilePaths  []string
	HTMLURL    string
	Number     int
	Repository string
	State      string
}
