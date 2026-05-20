package httpsrv

import (
	"encoding/json"
	"net/http"
)

type healthzResponse struct {
	Status string `json:"status"`
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthzResponse{Status: "ok"})
}
