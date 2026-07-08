package surveillance

import "testing"

func TestEvaluateTradeDetectsSelfTradeAndPriceBand(t *testing.T) {
	alerts := EvaluateTrade(TradeEvent{
		Market:      "BTC-USDT",
		MakerUserID: 1,
		TakerUserID: 1,
		Price:       52_000_00,
		Quantity:    10_000_000,
		MarkPrice:   50_000_00,
	}, Config{PriceBandBps: 300})

	if !hasAlert(alerts, AlertSelfTrade) || !hasAlert(alerts, AlertPriceBand) {
		t.Fatalf("alerts = %+v, want self trade and price band", alerts)
	}
}

func TestEvaluateActivityDetectsHighCancelRatio(t *testing.T) {
	alerts := EvaluateActivity(UserActivity{UserID: 10, Orders: 100, Cancellations: 80}, Config{HighCancelRatioBps: 7000})
	if !hasAlert(alerts, AlertHighCancelRatio) {
		t.Fatalf("alerts = %+v, want high cancel ratio", alerts)
	}
}

func hasAlert(alerts []Alert, typ AlertType) bool {
	for _, alert := range alerts {
		if alert.Type == typ {
			return true
		}
	}
	return false
}
