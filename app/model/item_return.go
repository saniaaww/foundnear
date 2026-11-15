package model

import "time"

type ItemReturn struct {
	ID         int64     `json:"id"`
	ClaimID    int64     `json:"claim_id"`
	ReturnedBy int64     `json:"returned_by"`
	ReceivedBy int64     `json:"received_by"`
	ReturnDate time.Time `json:"return_date"`
	ReturnNote string    `json:"return_note"`
}