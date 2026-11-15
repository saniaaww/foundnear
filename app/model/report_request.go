package model

type CreateReportRequest struct {
	CategoryID   int64   `json:"category_id"`
	ReportType   string  `json:"report_type"`
	ItemName     string  `json:"item_name"`
	Description  string  `json:"description"`
	PrivateInfo  *string `json:"private_info"`
	Location     string  `json:"location"`
	IncidentDate string  `json:"incident_date"`
	ImageURL     *string `json:"image_url"`
}

type UpdateReportRequest struct {
	CategoryID   int64   `json:"category_id"`
	ItemName     string  `json:"item_name"`
	Description  string  `json:"description"`
	PrivateInfo  *string `json:"private_info"`
	Location     string  `json:"location"`
	IncidentDate string  `json:"incident_date"`
	ImageURL     *string `json:"image_url"`
}
