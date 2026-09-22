package models

type Profile struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Groups   []string `json:"groups"`
}
