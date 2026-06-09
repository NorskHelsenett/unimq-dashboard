package scraper

import (
	"encoding/json"
	"net/http"
)

func (rc *RestClient) GetClusterHandler(w http.ResponseWriter, _ *http.Request) {
	stats, err := rc.GetClusterStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
