// Package constants defines shared constants used across the T-match backend.
package constants

import (
	"time"
)

// Auth

const (
	// AccessTokenTimeLife is the lifetime of an access JWT token.
	AccessTokenTimeLife = 15 * time.Minute
	// RefreshTokenTimeLife is the lifetime of a refresh JWT token.
	RefreshTokenTimeLife = 7 * 24 * time.Hour
	// VerifyCodeTimeLife is the lifetime of an email verification code.
	VerifyCodeTimeLife = 7 * time.Minute
	// MaxAgeRefreshToken is the max age of the refresh token cookie in seconds.
	MaxAgeRefreshToken = 604800
)

// RateLimit

const (
	// RateLimitAuthStudent is the rate limit for the /auth/students endpoint.
	RateLimitAuthStudent = 20 // /auth/students
	// RateLimitAuthStudentVerify is the rate limit for the /auth/students/verify endpoint.
	RateLimitAuthStudentVerify = 60 // /auth/students/verify
	// RateLimitNewVerifyCode is the rate limit for the /auth/newverify endpoint.
	RateLimitNewVerifyCode = 4 // /auth/newverify
	// RateLimitAuthStudentLogin is the rate limit for the /auth/students/login endpoint.
	RateLimitAuthStudentLogin = 30 // /auth/students/login
	// RateLimitAuthCompany is the rate limit for the /auth/company endpoint.
	RateLimitAuthCompany = 20 // /auth/company
	// RateLimitAuthCompanyVerify is the rate limit for the /auth/company/verify endpoint.
	RateLimitAuthCompanyVerify = 60 // /auth/company/verify
	// RateLimitAuthCompanyLogin is the rate limit for the /auth/company/login endpoint.
	RateLimitAuthCompanyLogin = 30 // /auth/company/login

	// RateLimitUpdateProfile is the rate limit for the /my/profile/put endpoint.
	RateLimitUpdateProfile = 100 // /my/profile/put
	// RateLimitGetProfile is the rate limit for the /my/profile endpoint.
	RateLimitGetProfile = 120 // /my/profile
	// RateLimitAddSkills is the rate limit for the /my/profile/skills/add endpoint.
	RateLimitAddSkills = 5 // /my/profile/skills/add
	// RateLimitDeleteSkills is the rate limit for the /my/profile/skills/delete endpoint.
	RateLimitDeleteSkills = 5 // /my/profile/skills/delete
	// RateLimitUpdateCompany is the rate limit for the /my/company/profile/put endpoint.
	RateLimitUpdateCompany = 100 // /my/company/profile/put
	// RateLimitGetCompany is the rate limit for the /my/company/profile endpoint.
	RateLimitGetCompany = 120 // /my/company/profile
	// RateLimitSetAvatar is the rate limit for the /my/avatar/put endpoint.
	RateLimitSetAvatar = 100 // /my/avatar/put
	// RateLimitCompanyInternship is the rate limit for the company internship endpoints.
	RateLimitCompanyInternship = 100 // company/internship
	// RateLimitMyNotifications is the rate limit for the notifications endpoint.
	RateLimitMyNotifications = 100 // /my/ notifications

	// RateLimitCreateInternship is the rate limit for the /internships (POST) endpoint.
	RateLimitCreateInternship = 12 // /internships (POST)
	// RateLimitUpdateInternship is the rate limit for the /internships/update/:id endpoint.
	RateLimitUpdateInternship = 12 // /internships/update/:id
	// RateLimitArchiveInternship is the rate limit for the /internships/delete/:id endpoint.
	RateLimitArchiveInternship = 5 // /internships/delete/:id
	// RateLimitAddInternshipSkill is the rate limit for the /internship/:id/skill/add endpoint.
	RateLimitAddInternshipSkill = 5 // /internship/:id/skill/add
	// RateLimitDeleteInternshipSkill is the rate limit for the /internship/:id/skill/delete endpoint.
	RateLimitDeleteInternshipSkill = 5 // /internship/:id/skill/delete
	// RateLimitInternshipInvite is the rate limit for the internship invite endpoint.
	RateLimitInternshipInvite = 40

	// RateLimitRespondToInternship is the rate limit for the /internships/:id/respond endpoint.
	RateLimitRespondToInternship = 10 // /internships/:id/respond
	// RateLimitGetMyResponses is the rate limit for the /my/responses endpoint.
	RateLimitGetMyResponses = 20 // /my/responses
	// RateLimitGetInternshipResponses is the rate limit for the /internships/:id/responses endpoint.
	RateLimitGetInternshipResponses = 20 // /internships/:id/responses
	// RateLimitSetResponseStatus is the rate limit for the /responses/:id/status endpoint.
	RateLimitSetResponseStatus = 20 // /responses/:id/status

	// RateLimitSearchInternship is the rate limit for the /internships (GET) endpoint.
	RateLimitSearchInternship = 60 // /internships (GET)
	// RateLimitSearchCompany is the rate limit for the /companies endpoint.
	RateLimitSearchCompany = 60 // /companies
	// RateLimitSearchStudent is the rate limit for the /students endpoint.
	RateLimitSearchStudent = 60 // /students

	// RateLimitRecommendations is the rate limit for the /my/recommendations endpoint.
	RateLimitRecommendations = 60 // /my/recommendations

	// RateLimitAuthAdminLogin is the rate limit for the /admin/login endpoint.
	RateLimitAuthAdminLogin = 10 // /admin/login
	// RateLimitAdminStats is the rate limit for the /admin/stats endpoint.
	RateLimitAdminStats = 30 // /admin/stats
	// RateLimitAdmin is the rate limit for the /admin/ban endpoint.
	RateLimitAdmin = 10 // /admin/ban

)

// Role

const (
	// Intern is the role name for interns.
	Intern = "intern"
	// Company is the role name for companies.
	Company = "company"
	// Admin is the role name for admins.
	Admin = "admin"
)

// MinUserAge is the minimum allowed user age.
const (
	MinUserAge = 16
)

// Status

const (
	// Pending is the initial application status.
	Pending = "pending"
	// Reviewing is the application status while being reviewed.
	Reviewing = "reviewing"
	// Accepted is the application status when accepted.
	Accepted = "accepted"
	// Rejected is the application status when rejected.
	Rejected = "rejected"
)

// S3

const (
	// BucketName is the default S3 bucket for uploaded files.
	BucketName = "t-match-storage"
)

// ServerTimeout

const (
	// ServerReadTimeout is the HTTP server read timeout.
	ServerReadTimeout = time.Second * 15
	// ServerWriteTimeout is the HTTP server write timeout.
	ServerWriteTimeout = time.Second * 15
	// ServerIdleTimeout is the HTTP server idle timeout.
	ServerIdleTimeout = time.Second * 60
)

// Client timeout

const (
	// DadataTimeout is the HTTP client timeout for the dadata service.
	DadataTimeout = time.Second * 10
	// RecsysTimeout is the HTTP client timeout for the recommendation service.
	RecsysTimeout = time.Second * 5
)

// Recsys action types

const (
	// RecsysActionClick is the click action type for the recommendation service.
	RecsysActionClick = "click"
	// RecsysActionApply is the apply action type for the recommendation service.
	RecsysActionApply = "apply"
	// RecsysActionInvate is the invite action type for the recommendation service.
	RecsysActionInvate = "invate"
)

const (
	// MaxSizeImage is the maximum allowed size for uploaded images in bytes.
	MaxSizeImage = 10 * 1024 * 1024
)

// Claims key

type contextKey string

// ClaimsKey is the context key for user claims.
const ClaimsKey contextKey = "claims"

// Postgress const

const (
	// ASC is the ascending sort order suffix.
	ASC = " ASC "
	// DESC is the descending sort order suffix.
	DESC = " DESC "
)

const (
	// AllowSpecialPassword contains special characters allowed in passwords.
	AllowSpecialPassword = "!@#$%^&*()_-+=[]{}|;:',.<>?/~`"
)

const (
	// WSReadBufferSize is the read buffer size for the WebSocket connection.
	WSReadBufferSize = 1024
	// WSWriteBufferSize is the write buffer size for the WebSocket connection.
	WSWriteBufferSize = 1024
)

const (
	// ChangeStatusType is the WebSocket message type for status changes.
	ChangeStatusType = "change_status"
	// InvateType is the WebSocket message type for invites.
	InvateType = "invate"
	// NewApplicationType is the WebSocket message type for new applications.
	NewApplicationType = "new_application"
	// MaxBufferNotificationWS is the max number of buffered WebSocket notifications.
	MaxBufferNotificationWS = 100
)
