package models

// Profile is the authenticated user as resolved from their ID token claims.
// It is the frontend's only source of identity.
type Profile struct {
	Username      string   `json:"username"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Groups        []string `json:"groups"`
	Subject       string   `json:"subject"`
	Issuer        string   `json:"issuer"`
	IssuedAt      int64    `json:"issued_at"`  // "iat" claim in Unix seconds
	ExpiresAt     int64    `json:"expires_at"` // "exp" claim in Unix seconds
}
