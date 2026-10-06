package profile

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type ACLStore interface {
	GetACLsForGroups(context.Context, []string) ([]models.ACL, error)
}

type ProfileHandler struct {
	adminGroups []string
	aclStore    ACLStore
}

func NewProfileHandler(groups []string, aclStore ACLStore) *ProfileHandler {
	return &ProfileHandler{
		adminGroups: groups,
		aclStore:    aclStore,
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
	if _, err := httpsuite.IsAGroupInClaim(r.Context(), ps.adminGroups); err != nil {
		groups, err := httpsuite.GetGroupsFromClaim(r.Context())
		if err != nil || len(groups) == 0 {
			httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithErrorMessage("no ACL grants read access"))
			return
		}
		acls, err := ps.aclStore.GetACLsForGroups(r.Context(), groups)
		if err != nil {
			httpsuite.WriteJSONError(w, http.StatusInternalServerError, httpsuite.WithError(err), httpsuite.WithErrorMessage("failed to validate ACL"))
			return
		}
		allowed := false
		for _, acl := range acls {
			for _, vhostID := range acl.VhostIDs {
				if acl.Allows(models.ScopeRead, vhostID) {
					allowed = true
					break
				}
			}
			if allowed {
				break
			}
		}
		if !allowed {
			httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithErrorMessage("no ACL grants read access"))
			return
		}
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
