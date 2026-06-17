package api

import (
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

func (rc *APIService) AddNotificationsRecipientHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		httpsuite.WriteJSONError(w, "missing required name parameter", http.StatusBadRequest)
		return
	}

	urlP := r.FormValue("url")
	if urlP == "" {
		httpsuite.WriteJSONError(w, "missing required url parameter", http.StatusBadRequest)
		return
	}

	typ := r.FormValue("type")
	if typ == "" {
		httpsuite.WriteJSONError(w, "missing required type parameter", http.StatusBadRequest)
		return
	}

	if slices.Contains(models.GetReceipientTypes(), models.RecipientType(typ)) == false {
		httpsuite.WriteJSONError(w, "invalid type parameter, expected one of "+models.GetRecipientTypesString(), http.StatusBadRequest)
		return
	}

	vhostObject, err := rc.DB.GetVhost(r.Context(), vhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhost: "+err.Error(), http.StatusInternalServerError)
		return
	}
	id := strconv.FormatInt(int64(len(vhostObject.Recipients)+1), 10)

	recipient := models.Recipient{
		ID:   id,
		Name: name,
		URL:  urlP,
		Type: models.ParseRecipientType(typ),
	}

	err = rc.DB.AddNotificationRecipient(r.Context(), vhost, recipient)
	if err != nil {
		httpsuite.WriteJSONError(w, "error adding recipient: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

func (rc *APIService) DeleteNotificationsRecipientHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}
	id := chi.URLParam(r, "recipient")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required id parameter", http.StatusBadRequest)
		return
	}

	err := rc.DB.DeleteNotificationRecipient(r.Context(), vhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error deleting recipient: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}
