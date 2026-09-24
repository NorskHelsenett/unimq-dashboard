package v1

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-chi/chi/v5"
	api "github.com/sisneve/rabbitmq-dashboard/internal/api/v1"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/v1/rmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/dex"
	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// swaggerOAuthScopes are requested by Swagger UI when the user authorizes.
// The audience scope makes Dex mint a token whose audience the dashboard API
// accepts, so no change to the token verifier is needed.
var swaggerOAuthScopes = []string{"openid", "profile", "email", "groups"}

func swaggerInitOAuth(oidc *config.OIDCConfig) string {
	scopes := append(append([]string{}, swaggerOAuthScopes...),
		fmt.Sprintf("audience:server:client_id:%s", oidc.OIDCClientID))

	clientID, _ := json.Marshal(oidc.OIDCSwaggerClientID)
	scope, _ := json.Marshal(strings.Join(scopes, " "))

	// Inline Javascript to initialize Swagger UI's OAuth2 client with the correct client ID and scopes.
	// A bit hacky, but the httpSwagger package doesn't provide a way to set these values directly.
	return fmt.Sprintf(`ui.initOAuth({
		clientId: %s,
		scopes: %s,
		usePkceWithAuthorizationCodeGrant: true
	})`, clientID, scope)
}

func SetupUtilityRoutes(r chi.Router, apiservice *api.APIService, dex *dex.DexClient, rmqHandler *rmq.RMQHandler, oidc *config.OIDCConfig) {

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.PersistAuthorization(true),
		httpSwagger.AfterScript(swaggerInitOAuth(oidc)),
	))
	r.Get("/healthz", apiservice.HealthzHandler)
	r.Get("/readyz", apiservice.ReadyzHandler(rmqHandler, dex))
}

func SetupInternalRoutes(r chi.Router, apiservice *api.APIService) {

	r.Route("/maintenance", func(r chi.Router) {
		r.Get("/", apiservice.GetMaintenanceHandler)
		r.Get("/{maintenance-id}", apiservice.GetMaintenanceEntryHandler)
		r.Post("/", apiservice.AddMaintenanceHandler)
		r.Patch("/{maintenance-id}", apiservice.PatchMaintenanceHandler)
		r.Put("/{maintenance-id}", apiservice.UpdateMaintenanceStatusHandler)
		r.Delete("/{maintenance-id}", apiservice.DeleteMaintenanceHandler)
		r.Get("/{maintenance-id}/logs", apiservice.GetMaintenanceEditLogsHandler)
	})

	r.Route("/alarms", func(r chi.Router) {
		r.Get("/", apiservice.GetAlarmHistoryAllHandler)
		r.Get("/{rule-id}", apiservice.GetAlarmHistoryHandler)
	})

	r.Get("/status", apiservice.GetCheckerStatusHandler)
	r.Route("/notifications", func(r chi.Router) {
		r.Get("/", apiservice.GetNotificationsHandler)
		r.Route("/{vhost-name}", func(r chi.Router) {
			r.Get("/", apiservice.GetNotificationsVhostHandler)
			r.Delete("/", apiservice.DeleteNotificationsHandler)
			r.Post("/recipients", apiservice.AddNotificationsRecipientHandler)
			r.Get("/recipients/{recipient-id}", apiservice.GetNotificationsRecipientHandler)
			r.Delete("/recipients/{recipient-id}", apiservice.DeleteNotificationsRecipientHandler)

			r.Post("/rules", apiservice.AddNotificationsRuleHandler)
			r.Get("/rules/{rule-id}", apiservice.GetNotificationRuleHandler)
			r.Post("/rules/{rule-id}", apiservice.UpdateNotificationsRuleHandler)
			r.Post("/rules/{rule-id}/toggle", apiservice.ToggleNotificationsRuleHandler)
			r.Post("/rules/{rule-id}/test", apiservice.TestNotificationsRuleHandler)
			r.Delete("/rules/{rule-id}", apiservice.DeleteNotificationsRuleHandler)
		})
	})

}
