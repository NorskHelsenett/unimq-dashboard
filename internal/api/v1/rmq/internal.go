package rmq

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

var ErrACLForbidden = errors.New("acl does not grant access")

func (rc *RMQHandler) authorizeVhost(ctx context.Context, required models.Scope, vhost string) error {
	if _, err := httpsuite.IsAGroupInClaim(ctx, rc.AdminGroups); err == nil {
		return nil
	}

	groups, err := httpsuite.GetGroupsFromClaim(ctx)
	if err != nil {
		return ErrACLForbidden
	}
	acls, err := rc.DB.GetACLsForGroups(ctx, groups)
	if err != nil {
		return err
	}
	for _, acl := range acls {
		if acl.Allows(required, vhost) {
			return nil
		}
	}
	return ErrACLForbidden
}

func (rc *RMQHandler) VhostACLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vhost, err := url.QueryUnescape(chi.URLParam(r, "vhost-name"))
		if err != nil || vhost == "" {
			httpsuite.WriteJSONError(w, http.StatusBadRequest, httpsuite.WithErrorMessage("invalid vhost name"))
			return
		}

		required := models.ScopeWrite
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			required = models.ScopeRead
		}
		if err := rc.authorizeVhost(r.Context(), required, vhost); err != nil {
			if errors.Is(err, ErrACLForbidden) {
				httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
				return
			}
			httpsuite.WriteJSONError(w, http.StatusInternalServerError, httpsuite.WithError(err), httpsuite.WithErrorMessage("failed to validate ACL"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rc *RMQHandler) FilterAccessibleVhosts(ctx context.Context, vhosts []models.Vhost) ([]models.Vhost, error) {
	if _, err := httpsuite.IsAGroupInClaim(ctx, rc.AdminGroups); err == nil {
		return vhosts, nil
	}

	groups, err := httpsuite.GetGroupsFromClaim(ctx)
	if err != nil {
		return nil, ErrACLForbidden
	}
	acls, err := rc.DB.GetACLsForGroups(ctx, groups)
	if err != nil {
		return nil, err
	}

	accessible := make([]models.Vhost, 0, len(vhosts))
	for i := range vhosts {
		for _, acl := range acls {
			if acl.Allows(models.ScopeRead, vhosts[i].Name) {
				accessible = append(accessible, vhosts[i])
				break
			}
		}
	}
	return accessible, nil
}
