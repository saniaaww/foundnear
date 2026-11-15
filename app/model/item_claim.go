package model

import "time"

type ItemClaim struct {
	ID               int64     `json:"id"`
	ReportID         int64     `json:"report_id"`
	ClaimantID       int64     `json:"claimant_id"`
	ClaimDescription string    `json:"claim_description"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
