package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/v1/profile"
)

func SetupProfileRoutes(r chi.Router, profileHandler *profile.ProfileHandler) {
	r.Route("/profile", func(r chi.Router) {
		r.Get("/", profileHandler.GetProfileHandler)
	})
}
