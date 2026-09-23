package profile

import (
	"log/slog"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

// TODO: This is going to become the profile service.
// Currently empty as it's not been implemented.

type ProfileHandler struct {
	adminGroups []string
}

func NewProfileHandler(groups []string) *ProfileHandler {
	return &ProfileHandler{
		adminGroups: groups,
	}
}

// TODO: Implement the profile service to return user profile information.

// @Summary		Get user profile
// @Description	Get the profile information of the authenticated user
// @Tags			Profile
// @Produce		json
// @Success		200	{object}	models.Profile
// @Failure		401	{object}	httpsuite.ErrorResponse
// @Failure		403	{object}	httpsuite.ErrorResponse
// @Router			/v1/profile [get]
// @security		bearer
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

	httpsuite.SendResponse(r.Context(), w, "User profile fetched successfully", http.StatusOK, &profile)
}
