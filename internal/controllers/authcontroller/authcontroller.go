package authcontroller

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Scope string

// We use iota to define the scope levels, where ScopeAdmin has the highest level of access, followed by ScopeWrite and ScopeRead.
// It lets us easily compare the scope level is equal or lower than the required scope level for a given operation.
const (
	ScopeAdmin = iota
	ScopeWrite
	ScopeRead
)

func ParseScope(scope int) Scope {
	switch scope {
	case ScopeAdmin:
		return "admin"
	case ScopeWrite:
		return "write"
	case ScopeRead:
		return "read"
	default:
		return "unknown"
	}
}

var secretKey = []byte("b7351121-4d41-4767-beae-a3dab4c6f275")

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func NewLoginResponse(token string) *LoginResponse {
	return &LoginResponse{
		Token: token,
	}
}

func CreateToken(username string, scope int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
			"scope":    scope,
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func ValidateScope(scope string, allowedScopes ...Scope) bool {

	return false
}

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		err := VerifyToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return secretKey, nil
		})

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AuthorizeScope(scope int, allowedScopes ...int) bool {
	return slices.ContainsFunc(allowedScopes, func(s int) bool {
		if s <= scope {
			return true
		}
		return false
	})
}
