package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

var ErrACLForbidden = errors.New("acl does not grant access")

func (rc *APIService) authorizeVhost(ctx context.Context, required models.Scope, vhost string) error {
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

func (rc *APIService) VhostACLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vhost, err := url.QueryUnescape(chi.URLParam(r, "vhost"))
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

func (rc *APIService) FilterAccessibleVhosts(ctx context.Context, vhosts []models.Vhost) ([]models.Vhost, error) {
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
	for _, vhost := range vhosts {
		for _, acl := range acls {
			if acl.Allows(models.ScopeRead, vhost.Name) {
				accessible = append(accessible, vhost)
				break
			}
		}
	}
	return accessible, nil
}

func (rc *APIService) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if _, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups); err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return false
	}
	return true
}

func (rc *APIService) GetACLsHandler(w http.ResponseWriter, r *http.Request) {
	if !rc.requireAdmin(w, r) {
		return
	}

	acls, err := rc.DB.GetACLs(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w, http.StatusInternalServerError, httpsuite.WithError(err), httpsuite.WithErrorMessage("failed to fetch ACLs"))
		return
	}
	httpsuite.SendResponse(r.Context(), w, "Fetched ACLs", http.StatusOK, &acls)
}

func (rc *APIService) GetACLHandler(w http.ResponseWriter, r *http.Request) {
	if !rc.requireAdmin(w, r) {
		return
	}

	group, err := url.QueryUnescape(chi.URLParam(r, "group"))
	if err != nil || group == "" {
		httpsuite.WriteJSONError(w, http.StatusBadRequest, httpsuite.WithErrorMessage("invalid group"))
		return
	}
	acl, err := rc.DB.GetACL(r.Context(), group)
	if err != nil {
		if errors.Is(err, database.ErrACLNotFound) {
			httpsuite.WriteJSONError(w, http.StatusNotFound, httpsuite.WithErrorMessage("ACL not found"))
			return
		}
		httpsuite.WriteJSONError(w, http.StatusInternalServerError, httpsuite.WithError(err), httpsuite.WithErrorMessage("failed to fetch ACL"))
		return
	}
	httpsuite.SendResponse(r.Context(), w, "Fetched ACL", http.StatusOK, acl)
}

func (rc *APIService) UpsertACLHandler(w http.ResponseWriter, r *http.Request) {
	if !rc.requireAdmin(w, r) {
		return
	}

	var acl models.ACL
	if err := httpsuite.ReadResponse(r, &acl); err != nil {
		httpsuite.WriteJSONError(w, http.StatusBadRequest, httpsuite.WithError(err), httpsuite.WithErrorMessage("invalid request body"))
		return
	}
	if err := acl.Validate(); err != nil {
		httpsuite.WriteJSONError(w, http.StatusBadRequest, httpsuite.WithError(err), httpsuite.WithErrorMessage("invalid ACL"))
		return
	}
	if err := rc.DB.UpsertACL(r.Context(), &acl); err != nil {
		httpsuite.WriteJSONError(w, http.StatusInternalServerError, httpsuite.WithError(err), httpsuite.WithErrorMessage("failed to save ACL"))
		return
	}
	httpsuite.SendResponse(r.Context(), w, "ACL saved", http.StatusOK, &acl)
}

func (rc *APIService) DeleteACLHandler(w http.ResponseWriter, r *http.Request) {
	if !rc.requireAdmin(w, r) {
		return
	}

	group, err := url.QueryUnescape(chi.URLParam(r, "group"))
	if err != nil || group == "" {
		httpsuite.WriteJSONError(w, http.StatusBadRequest, httpsuite.WithErrorMessage("invalid group"))
		return
	}
	if err := rc.DB.DeleteACL(r.Context(), group); err != nil {
		if errors.Is(err, database.ErrACLNotFound) {
			httpsuite.WriteJSONError(w, http.StatusNotFound, httpsuite.WithErrorMessage("ACL not found"))
			return
		}
		httpsuite.WriteJSONError(w, http.StatusInternalServerError, httpsuite.WithError(err), httpsuite.WithErrorMessage("failed to delete ACL"))
		return
	}
	httpsuite.SendResponse(r.Context(), w, "ACL deleted", http.StatusOK, httpsuite.NewEmptyResponse())
}
