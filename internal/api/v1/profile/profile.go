package profile

import (
	"log/slog"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type ProfileHandler struct {
	adminGroups []string
}

func NewProfileHandler(groups []string) *ProfileHandler {
	return &ProfileHandler{
		adminGroups: groups,
	}
}

// @Summary		Get user profile
// @Description	Get the profile information of the authenticated user
// @Tags			Profile
// @Produce		json
// @Success		200	{object}	models.Profile
// @Failure		401	{object}	httpsuite.ErrorResponse
// @Failure		403	{object}	httpsuite.ErrorResponse
// @Router			/v1/profile [get]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (ps *ProfileHandler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), ps.adminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithError(err))
		return
	}

	username, err := httpsuite.GetUsernameFromContext(r.Context())
	if err != nil {
		slog.WarnContext(r.Context(), "failed to get username from context", "error", err)
	}

	email, err := httpsuite.GetEmailFromContext(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to get email from context"),
		)
		return
	}

	groups, err := httpsuite.GetGroupsFromContext(r.Context())
	if err != nil {
		slog.WarnContext(r.Context(), "failed to get groups from context", "error", err)
	}

	profile := models.Profile{
		Username: username,
		Email:    email,
		Groups:   groups,
	}

	// The remaining claims are optional functionalities for the profile page,
	// so a provider that omits them yields zero values rather than an error.
	if claims, err := httpsuite.GetClaimsFromContext(r.Context()); err == nil {
		profile.EmailVerified, _ = claims["email_verified"].(bool)
		profile.Subject, _ = claims["sub"].(string)
		profile.Issuer, _ = claims["iss"].(string)
		// Numeric claims arrive as float64 from encoding/json.
		if iat, ok := claims["iat"].(float64); ok {
			profile.IssuedAt = int64(iat)
		}
		if exp, ok := claims["exp"].(float64); ok {
			profile.ExpiresAt = int64(exp)
		}
	}

	httpsuite.SendResponse(r.Context(), w, "User profile fetched successfully", http.StatusOK, &profile)
}
