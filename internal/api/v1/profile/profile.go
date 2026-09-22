package profile

import (
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
)

// TODO: This is going to become the profile service.
// Currently empty as it's not been implemented.

type ProfileHandler struct {
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
	httpsuite.SendEmptyResponse(r.Context(), w, "profile service not implemented", http.StatusNotImplemented)
}
