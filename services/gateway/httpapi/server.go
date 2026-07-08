package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/exchange-core-starter/exchange-core-starter/services/compliance/readiness"
)

type Market struct {
	Symbol     string `json:"symbol"`
	BaseAsset  string `json:"base_asset"`
	QuoteAsset string `json:"quote_asset"`
	Type       string `json:"type"`
	Status     string `json:"status"`
}

type Server struct {
	markets []Market
	status  readiness.Status
}

func NewServer(markets []Market, status readiness.Status) *Server {
	return &Server{markets: markets, status: status}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", s.health)
	mux.HandleFunc("GET /api/v1/markets", s.marketsHandler)
	mux.HandleFunc("GET /api/v1/readiness/spot-launch", s.spotLaunchReadiness)
	mux.HandleFunc("GET /api/v1/readiness/derivatives", s.derivativesReadiness)
	return withJSON(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) marketsHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string][]Market{"markets": s.markets})
}

func (s *Server) spotLaunchReadiness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, readiness.Evaluate(readiness.ScopeSpotLaunch, s.status))
}

func (s *Server) derivativesReadiness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, readiness.Evaluate(readiness.ScopeDerivatives, s.status))
}

func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
