package model

import "time"

type ClaimVerification struct {
	ID               int64     `json:"id"`
	ClaimID          int64     `json:"claim_id"`
	OfficerID        int64     `json:"officer_id"`
	VerificationNote string    `json:"verification_note"`
	Result           string    `json:"result"`
	VerifiedAt       time.Time `json:"verified_at"`
}
