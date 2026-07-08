// Package adminapi provides the internal admin HTTP API. It runs on a
// separate port from the public gateway to enforce network-level separation.
// Authentication is a development placeholder: the X-ADMIN-ID header
// identifies the caller. Every state-modifying action is recorded in the
// tamper-evident audit log.
package adminapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/exchange-core-starter/exchange-core-starter/libs/audit"
	"github.com/exchange-core-starter/exchange-core-starter/libs/decimal"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/oms"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/spotexchange"
	"github.com/exchange-core-starter/exchange-core-starter/services/wallet-gateway/wallet"
)

// Deps are the runtime dependencies injected from the gateway main.
type Deps struct {
	Exchange *spotexchange.Exchange
	Wallet   *wallet.Coordinator
	AuditLog *audit.Log
}

// Server exposes admin endpoints.
type Server struct {
	exchange *spotexchange.Exchange
	wallet   *wallet.Coordinator
	auditLog *audit.Log
}

// NewServer creates an admin API server.
func NewServer(deps Deps) *Server {
	return &Server{
		exchange: deps.Exchange,
		wallet:   deps.Wallet,
		auditLog: deps.AuditLog,
	}
}

// Handler returns the HTTP handler for the admin API.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/v1/health", s.health)
	mux.HandleFunc("GET /admin/v1/users", s.adminAuth(s.users))
	mux.HandleFunc("GET /admin/v1/users/{user_id}/balances", s.adminAuth(s.userBalances))
	mux.HandleFunc("GET /admin/v1/markets", s.adminAuth(s.markets))
	mux.HandleFunc("POST /admin/v1/markets/{symbol}/halt", s.adminAuth(s.haltMarket))
	mux.HandleFunc("POST /admin/v1/markets/{symbol}/resume", s.adminAuth(s.resumeMarket))
	mux.HandleFunc("GET /admin/v1/withdrawals/{id}", s.adminAuth(s.getWithdrawal))
	mux.HandleFunc("POST /admin/v1/withdrawals/{id}/approve", s.adminAuth(s.approveWithdrawal))
	mux.HandleFunc("POST /admin/v1/withdrawals/{id}/reject", s.adminAuth(s.rejectWithdrawal))
	mux.HandleFunc("GET /admin/v1/audit-log", s.adminAuth(s.auditLogEntries))
	mux.HandleFunc("GET /admin/v1/ledger/entries", s.adminAuth(s.ledgerEntries))
	mux.HandleFunc("GET /admin/v1/orderbook/{symbol}", s.adminAuth(s.orderbook))
	return withAdminHeaders(mux)
}

// --- handlers ---

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "admin-api"})
}

func (s *Server) users(w http.ResponseWriter, _ *http.Request, _ string) {
	if s.exchange == nil {
		writeError(w, http.StatusServiceUnavailable, "EXCHANGE_UNAVAILABLE", "exchange not configured")
		return
	}
	users := s.exchange.Users()
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

type balanceView struct {
	Asset     string `json:"asset"`
	Available string `json:"available"`
	Locked    string `json:"locked"`
}

func (s *Server) userBalances(w http.ResponseWriter, r *http.Request, _ string) {
	if s.exchange == nil {
		writeError(w, http.StatusServiceUnavailable, "EXCHANGE_UNAVAILABLE", "exchange not configured")
		return
	}
	userID, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "invalid user_id")
		return
	}
	market := s.exchange.Market()
	balances := s.exchange.Balances(userID)
	views := make([]balanceView, 0, len(balances))
	for _, b := range balances {
		scale := assetScale(market, b.Asset)
		views = append(views, balanceView{
			Asset:     b.Asset,
			Available: decimal.New(b.Available, scale).String(),
			Locked:    decimal.New(b.Locked, scale).String(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"user_id": userID, "balances": views})
}

type marketView struct {
	Symbol string `json:"symbol"`
	Status string `json:"status"`
}

func (s *Server) markets(w http.ResponseWriter, _ *http.Request, _ string) {
	if s.exchange == nil {
		writeError(w, http.StatusServiceUnavailable, "EXCHANGE_UNAVAILABLE", "exchange not configured")
		return
	}
	m := s.exchange.Market()
	writeJSON(w, http.StatusOK, map[string]any{"markets": []marketView{{
		Symbol: m.Symbol,
		Status: string(m.Status),
	}}})
}

func (s *Server) haltMarket(w http.ResponseWriter, r *http.Request, adminID string) {
	if s.exchange == nil {
		writeError(w, http.StatusServiceUnavailable, "EXCHANGE_UNAVAILABLE", "exchange not configured")
		return
	}
	symbol := r.PathValue("symbol")
	market := s.exchange.Market()
	if symbol != market.Symbol {
		writeError(w, http.StatusNotFound, "UNKNOWN_SYMBOL", "unknown symbol")
		return
	}
	s.exchange.HaltMarket()
	s.appendAudit(adminID, r, "MARKET_HALT", "MARKET", symbol, "")
	writeJSON(w, http.StatusOK, map[string]any{"symbol": symbol, "status": string(oms.MarketHalted)})
}

func (s *Server) resumeMarket(w http.ResponseWriter, r *http.Request, adminID string) {
	if s.exchange == nil {
		writeError(w, http.StatusServiceUnavailable, "EXCHANGE_UNAVAILABLE", "exchange not configured")
		return
	}
	symbol := r.PathValue("symbol")
	market := s.exchange.Market()
	if symbol != market.Symbol {
		writeError(w, http.StatusNotFound, "UNKNOWN_SYMBOL", "unknown symbol")
		return
	}
	s.exchange.ResumeMarket()
	s.appendAudit(adminID, r, "MARKET_RESUME", "MARKET", symbol, "")
	writeJSON(w, http.StatusOK, map[string]any{"symbol": symbol, "status": string(oms.MarketTrading)})
}

func (s *Server) getWithdrawal(w http.ResponseWriter, r *http.Request, _ string) {
	if s.wallet == nil {
		writeError(w, http.StatusServiceUnavailable, "WALLET_UNAVAILABLE", "wallet not configured")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid withdrawal id")
		return
	}
	withdrawal, ok := s.wallet.Withdrawal(id)
	if !ok {
		writeError(w, http.StatusNotFound, "WITHDRAWAL_NOT_FOUND", "withdrawal not found")
		return
	}
	writeJSON(w, http.StatusOK, withdrawalView(withdrawal))
}

type approveRejectBody struct {
	Reason string `json:"reason"`
}

func (s *Server) approveWithdrawal(w http.ResponseWriter, r *http.Request, adminID string) {
	if s.wallet == nil {
		writeError(w, http.StatusServiceUnavailable, "WALLET_UNAVAILABLE", "wallet not configured")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid withdrawal id")
		return
	}
	var body approveRejectBody
	_ = json.NewDecoder(r.Body).Decode(&body)

	requestID := r.Header.Get("X-REQUEST-ID")
	withdrawal, err := s.wallet.ApproveWithdrawal(id, adminID, requestID, body.Reason)
	if err != nil {
		writeError(w, http.StatusBadRequest, "APPROVAL_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, withdrawalView(withdrawal))
}

func (s *Server) rejectWithdrawal(w http.ResponseWriter, r *http.Request, adminID string) {
	if s.wallet == nil {
		writeError(w, http.StatusServiceUnavailable, "WALLET_UNAVAILABLE", "wallet not configured")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid withdrawal id")
		return
	}
	var body approveRejectBody
	_ = json.NewDecoder(r.Body).Decode(&body)

	requestID := r.Header.Get("X-REQUEST-ID")
	withdrawal, err := s.wallet.RejectWithdrawal(id, adminID, requestID, body.Reason)
	if err != nil {
		writeError(w, http.StatusBadRequest, "REJECTION_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, withdrawalView(withdrawal))
}

type auditEventView struct {
	ID           string `json:"id"`
	ActorID      string `json:"actor_id"`
	ActorType    string `json:"actor_type"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Reason       string `json:"reason,omitempty"`
	Hash         string `json:"hash"`
	PrevHash     string `json:"prev_hash,omitempty"`
	CreatedAt    string `json:"created_at"`
}

func (s *Server) auditLogEntries(w http.ResponseWriter, _ *http.Request, _ string) {
	if s.auditLog == nil {
		writeJSON(w, http.StatusOK, map[string]any{"events": []auditEventView{}})
		return
	}
	events := s.auditLog.Events()
	views := make([]auditEventView, 0, len(events))
	for _, e := range events {
		views = append(views, auditEventView{
			ID:           e.ID,
			ActorID:      e.ActorID,
			ActorType:    e.ActorType,
			Action:       e.Action,
			ResourceType: e.ResourceType,
			ResourceID:   e.ResourceID,
			Reason:       e.Reason,
			Hash:         e.Hash,
			PrevHash:     e.PrevHash,
			CreatedAt:    e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events":      views,
		"chain_valid": s.auditLog.Verify() == nil,
	})
}

type ledgerEntryView struct {
	AccountID      int64  `json:"account_id"`
	Asset          string `json:"asset"`
	Debit          string `json:"debit,omitempty"`
	Credit         string `json:"credit,omitempty"`
	Type           string `json:"type"`
	ReferenceType  string `json:"reference_type,omitempty"`
	ReferenceID    string `json:"reference_id,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

func (s *Server) ledgerEntries(w http.ResponseWriter, _ *http.Request, _ string) {
	if s.exchange == nil {
		writeError(w, http.StatusServiceUnavailable, "EXCHANGE_UNAVAILABLE", "exchange not configured")
		return
	}
	entries := s.exchange.Accounts().Entries()
	views := make([]ledgerEntryView, 0, len(entries))
	for _, e := range entries {
		views = append(views, ledgerEntryView{
			AccountID:      e.AccountID,
			Asset:          e.Asset,
			Debit:          intToDecStr(e.Debit),
			Credit:         intToDecStr(e.Credit),
			Type:           string(e.Type),
			ReferenceType:  e.ReferenceType,
			ReferenceID:    e.ReferenceID,
			IdempotencyKey: e.IdempotencyKey,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": views, "count": len(views)})
}

type levelView struct {
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
}

func (s *Server) orderbook(w http.ResponseWriter, r *http.Request, _ string) {
	if s.exchange == nil {
		writeError(w, http.StatusServiceUnavailable, "EXCHANGE_UNAVAILABLE", "exchange not configured")
		return
	}
	symbol := r.PathValue("symbol")
	market := s.exchange.Market()
	if symbol != market.Symbol {
		writeError(w, http.StatusNotFound, "UNKNOWN_SYMBOL", "unknown symbol")
		return
	}
	bids, asks := s.exchange.Depth(50)
	bidViews := make([]levelView, 0, len(bids))
	for _, l := range bids {
		bidViews = append(bidViews, levelView{
			Price:    decimal.New(l.Price, market.PriceScale).String(),
			Quantity: decimal.New(l.Quantity, market.QuantityScale).String(),
		})
	}
	askViews := make([]levelView, 0, len(asks))
	for _, l := range asks {
		askViews = append(askViews, levelView{
			Price:    decimal.New(l.Price, market.PriceScale).String(),
			Quantity: decimal.New(l.Quantity, market.QuantityScale).String(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"symbol": symbol, "bids": bidViews, "asks": askViews})
}

// --- helpers ---

type adminHandler func(http.ResponseWriter, *http.Request, string)

func (s *Server) adminAuth(next adminHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID := r.Header.Get("X-ADMIN-ID")
		if adminID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "missing X-ADMIN-ID header (development auth placeholder)")
			return
		}
		next(w, r, adminID)
	}
}

func (s *Server) appendAudit(adminID string, r *http.Request, action, resourceType, resourceID, reason string) {
	if s.auditLog == nil {
		return
	}
	_, _ = s.auditLog.Append(audit.Event{
		ID:           action + "-" + resourceID,
		ActorID:      adminID,
		ActorType:    "ADMIN",
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		RequestID:    r.Header.Get("X-REQUEST-ID"),
		Reason:       reason,
	})
}

func withdrawalView(w wallet.Withdrawal) map[string]any {
	return map[string]any{
		"id":      w.Request.ID,
		"user_id": w.Request.UserID,
		"asset":   w.Request.Asset,
		"amount":  w.Request.Amount,
		"fee":     w.Request.Fee,
		"status":  string(w.Status),
		"txid":    w.TxID,
	}
}

func assetScale(market oms.Market, asset string) int32 {
	if asset == market.BaseAsset {
		return market.QuantityScale
	}
	return market.PriceScale
}

func intToDecStr(v int64) string {
	if v == 0 {
		return ""
	}
	return strconv.FormatInt(v, 10)
}

func withAdminHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]any{
		"code":    code,
		"message": message,
	})
}
