package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/services/compliance/readiness"
)

func TestHealthAndMarkets(t *testing.T) {
	server := NewServer([]Market{{Symbol: "BTC-USDT", BaseAsset: "BTC", QuoteAsset: "USDT", Type: "SPOT", Status: "TRADING"}}, readiness.Status{})

	health := request(t, server, "/api/v1/health")
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d, want 200", health.Code)
	}

	markets := request(t, server, "/api/v1/markets")
	if markets.Code != http.StatusOK {
		t.Fatalf("markets status = %d, want 200", markets.Code)
	}
	var body struct {
		Markets []Market `json:"markets"`
	}
	if err := json.Unmarshal(markets.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode markets: %v", err)
	}
	if len(body.Markets) != 1 || body.Markets[0].Symbol != "BTC-USDT" {
		t.Fatalf("markets = %+v", body.Markets)
	}
}

func TestReadinessEndpointsExposeBlockedDerivatives(t *testing.T) {
	server := NewServer(nil, readiness.Status{
		LegalApproved:           true,
		ComplianceApproved:      true,
		SecurityApproved:        true,
		ISMSReady:               true,
		VASPFilingReady:         true,
		AMLOperational:          true,
		KYCOperational:          true,
		ClosedPilotPassed:       true,
		IncidentDrillPassed:     true,
		RecoveryDrillPassed:     true,
		MarketSurveillanceReady: true,
	})

	spot := request(t, server, "/api/v1/readiness/spot-launch")
	var spotDecision readiness.Decision
	if err := json.Unmarshal(spot.Body.Bytes(), &spotDecision); err != nil {
		t.Fatalf("decode spot readiness: %v", err)
	}
	if !spotDecision.Allowed {
		t.Fatalf("spot decision = %+v, want allowed", spotDecision)
	}

	derivatives := request(t, server, "/api/v1/readiness/derivatives")
	var derivativesDecision readiness.Decision
	if err := json.Unmarshal(derivatives.Body.Bytes(), &derivativesDecision); err != nil {
		t.Fatalf("decode derivatives readiness: %v", err)
	}
	if derivativesDecision.Allowed {
		t.Fatalf("derivatives decision = %+v, want blocked", derivativesDecision)
	}
}

func request(t *testing.T, server *Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	return rec
}
