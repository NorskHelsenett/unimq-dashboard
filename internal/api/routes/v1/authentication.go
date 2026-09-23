package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/dex"
)

func SetupAuthenticationRoutes(r chi.Router, dex *dex.DexClient) {

	r.Route("/login", func(r chi.Router) {
		r.Post("/", dex.LoginHandler)
		r.Get("/redirect", dex.RedirectHandler)
		r.Get("/callback", dex.OauthCallbackHandler)
	})
}
