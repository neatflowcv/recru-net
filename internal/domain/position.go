package domain

import "time"

type PositionSource string

const (
	PositionSourceJumpit  PositionSource = "jumpit"
	PositionSourceSaramin PositionSource = "saramin"
)

type CareerRange struct {
	MinYears   *int
	MaxYears   *int
	EntryLevel bool
}

type Position struct {
	ID            string
	Source        PositionSource
	ExternalID    string
	Title         string
	CompanyName   string
	JobCategories []string
	TechStacks    []string
	Locations     []string
	Career        CareerRange
	ClosesAt      *time.Time
}
