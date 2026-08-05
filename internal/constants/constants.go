package constants

import (
	"time"
)

// Auth

const (
	AccessTokenTimeLife  = 15 * time.Minute
	RefreshTokenTimeLife = 7 * 24 * time.Hour
	VerifyCodeTimeLife   = 7 * time.Minute
	MaxAgeRefreshToken   = 604800
)

// RateLimit

const (
	// Auth endpoints
	RateLimitAuthStudent       = 20 // /auth/students
	RateLimitAuthStudentVerify = 60 // /auth/students/verify
	RateLimitNewVerifyCode     = 4  // /auth/newverify
	RateLimitAuthStudentLogin  = 30 // /auth/students/login
	RateLimitAuthCompany       = 20 // /auth/company
	RateLimitAuthCompanyVerify = 60 // /auth/company/verify
	RateLimitAuthCompanyLogin  = 30 // /auth/company/login

	// Profile endpoints
	RateLimitUpdateProfile     = 100 // /my/profile/put
	RateLimitGetProfile        = 120 // /my/profile
	RateLimitAddSkills         = 5   // /my/profile/skills/add
	RateLimitDeleteSkills      = 5   // /my/profile/skills/delete
	RateLimitUpdateCompany     = 100 // /my/company/profile/put
	RateLimitGetCompany        = 120 // /my/company/profile
	RateLimitSetAvatar         = 100 // /my/avatar/put
	RateLimitCompanyInternship = 100 // company/internship
	RateLimitMyNotifications   = 100 // /my/ notifications

	// Internship endpoints
	RateLimitCreateInternship      = 12 // /internships (POST)
	RateLimitUpdateInternship      = 12 // /internships/update/:id
	RateLimitArchiveInternship     = 5  // /internships/delete/:id
	RateLimitAddInternshipSkill    = 5  // /internship/:id/skill/add
	RateLimitDeleteInternshipSkill = 5  // /internship/:id/skill/delete
	RateLimitInternshipInvite      = 40

	// Response endpoints
	RateLimitRespondToInternship    = 10 // /internships/:id/respond
	RateLimitGetMyResponses         = 20 // /my/responses
	RateLimitGetInternshipResponses = 20 // /internships/:id/responses
	RateLimitSetResponseStatus      = 20 // /responses/:id/status

	// Search endpoints (typically higher limits)
	RateLimitSearchInternship = 60 // /internships (GET)
	RateLimitSearchCompany    = 60 // /companies
	RateLimitSearchStudent    = 60 // /students
)

// Role

const (
	Intern  = "intern"
	Company = "company"
	Admin   = "admin"
)

// MinUserAge

const (
	MinUserAge = 16
)

// Status
const (
	Pending   = "pending"
	Reviewing = "reviewing"
	Accepted  = "accepted"
	Rejected  = "rejected"
)

// S3
const (
	BucketName = "t-match-storage"
)

// ServerTimeout
const (
	ServerReadTimeout  = time.Second * 15
	ServerWriteTimeout = time.Second * 15
	ServerIdleTimeout  = time.Second * 60
)

// Client timeout
const (
	DadataTimeout = time.Second * 10
)

const (
	MaxSizeImage = 10 * 1024 * 1024
)

// Claims key
type contextKey string

const ClaimsKey contextKey = "claims"

// Postgress const
const (
	ASC  = " ASC "
	DESC = " DESC "
)

const (
	AllowSpecialPassword = "!@#$%^&*()_-+=[]{}|;:',.<>?/~`"
)

const (
	WSReadBufferSize  = 1024
	WSWriteBufferSize = 1024
)

const (
	ChangeStatusType        = "change_status"
	InvateType              = "invate"
	NewApplicationType      = "new_application"
	MaxBufferNotificationWS = 100
)
