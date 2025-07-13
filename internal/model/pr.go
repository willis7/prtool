package model

import "time"

type PR struct {
	Title    string
	Number   int
	MergedAt time.Time
	Author   string
	Repo     string
}
