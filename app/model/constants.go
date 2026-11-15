package model

const (
	RoleAdmin   = "admin"
	RoleOfficer = "officer"
	RoleUser    = "user"
)

const (
	ReportTypeLost  = "lost"
	ReportTypeFound = "found"
)

const (
	ReportStatusActive   = "active"
	ReportStatusMatched  = "matched"
	ReportStatusReturned = "returned"
	ReportStatusClosed   = "closed"
)

const (
	ClaimStatusPending   = "pending"
	ClaimStatusApproved  = "approved"
	ClaimStatusRejected  = "rejected"
	ClaimStatusCancelled = "cancelled"
)

const (
	VerificationApproved = "approved"
	VerificationRejected = "rejected"
)