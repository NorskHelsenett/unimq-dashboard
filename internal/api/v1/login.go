package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/controllers/authcontroller"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Login handles user login requests and returns a JWT token upon successful authentication.
// @Description	This endpoint accepts a JSON payload containing the username and password. If the credentials are valid, it returns a JWT token that can be used for subsequent requests to protected endpoints.
// @Description	To use the token, include it in the Authorization header of your requests as follows: `Authorization: Bearer <token>`.
// @Tags			Authentication
// @Accept			json
// @Produce		json
// @Param			login	body		authcontroller.User				true	"Login credentials"
// @Success		200		{object}	authcontroller.LoginResponse	"Login successful"
// @Failure		401		{object}	httpsuite.APIError				"Invalid credentials"
// @Router			/login [post]
func (rc *APIService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var u authcontroller.User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	slog.DebugContext(r.Context(), "login request", "username", u.Username)

	// TODO: Implement proper authentication and authorization logic here. For now, we will just check if the username and password match a hardcoded value.
	if u.Username == "admin" && u.Password == "123456" {
		// TODO: Implement proper scope handling. For now, we will just use a hardcoded value.
		scope := authcontroller.ScopeAdmin
		tokenString, err := authcontroller.CreateToken(u.Username, scope)
		if err != nil {
			httpsuite.WriteJSONError(w, "Invalid credentials", http.StatusUnauthorized)
		}
		httpsuite.SendResponse(r.Context(), w, "Login successful", http.StatusOK, authcontroller.NewLoginResponse(tokenString))
		return
	}

	httpsuite.WriteJSONError(w, "Invalid credentials", http.StatusUnauthorized)
}
