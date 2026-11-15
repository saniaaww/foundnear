package model

import "time"

// ItemReport Hanya untuk pemilik, petugas, atau admin. 
// Untuk respons publik = PublicItemReport.

type ItemReport struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	CategoryID   int64     `json:"category_id"`
	ReportType   string    `json:"report_type"`
	ItemName     string    `json:"item_name"`
	Description  string    `json:"description"`
	PrivateInfo  *string   `json:"private_info,omitempty"`
	Location     string    `json:"location"`
	IncidentDate time.Time `json:"incident_date"`
	ImageURL     *string   `json:"image_url,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PublicItemReport struct {
	ID           int64     `json:"id"`
	CategoryID   int64     `json:"category_id"`
	CategoryName string    `json:"category_name"`
	ReportType   string    `json:"report_type"`
	ItemName     string    `json:"item_name"`
	Description  string    `json:"description"`
	Location     string    `json:"location"`
	IncidentDate time.Time `json:"incident_date"`
	ImageURL     *string   `json:"image_url,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type ReportFilter struct {
	Keyword    string
	ReportType string
	Status     string
	CategoryID *int64
	Location   string
	Limit      int
	Offset     int
}