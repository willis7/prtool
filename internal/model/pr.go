package model

import (
	"time"
)

// PR represents a pull request with all required fields for processing
type PR struct {
	// Basic information
	Repository  string    `json:"repository"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	URL         string    `json:"url"`
	
	// Timestamps
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MergedAt    time.Time `json:"merged_at"`
	
	// Metadata
	Labels      []string  `json:"labels"`
	BaseBranch  string    `json:"base_branch"`
	HeadBranch  string    `json:"head_branch"`
	
	// Summary (to be populated by LLM)
	Summary     string    `json:"summary,omitempty"`
}