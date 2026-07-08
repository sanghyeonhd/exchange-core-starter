package surveillance

type AlertType string

const (
	AlertSelfTrade       AlertType = "SELF_TRADE"
	AlertPriceBand       AlertType = "PRICE_BAND"
	AlertLargeOrder      AlertType = "LARGE_ORDER"
	AlertHighCancelRatio AlertType = "HIGH_CANCEL_RATIO"
)

type TradeEvent struct {
	Market      string
	MakerUserID int64
	TakerUserID int64
	Price       int64
	Quantity    int64
	MarkPrice   int64
}

type UserActivity struct {
	UserID        int64
	Orders        int64
	Cancellations int64
}

type Config struct {
	PriceBandBps       int64
	LargeOrderNotional int64
	HighCancelRatioBps int64
}

type Alert struct {
	Type     AlertType
	Market   string
	UserID   int64
	Severity string
	Reason   string
}

func EvaluateTrade(event TradeEvent, cfg Config) []Alert {
	var alerts []Alert
	if event.MakerUserID != 0 && event.MakerUserID == event.TakerUserID {
		alerts = append(alerts, Alert{Type: AlertSelfTrade, Market: event.Market, UserID: event.MakerUserID, Severity: "HIGH", Reason: "maker and taker user match"})
	}
	if cfg.PriceBandBps > 0 && event.MarkPrice > 0 {
		diff := abs(event.Price - event.MarkPrice)
		if diff*10_000 > event.MarkPrice*cfg.PriceBandBps {
			alerts = append(alerts, Alert{Type: AlertPriceBand, Market: event.Market, Severity: "MEDIUM", Reason: "trade price outside mark price band"})
		}
	}
	if cfg.LargeOrderNotional > 0 && event.Price > 0 && event.Quantity > 0 {
		notional := event.Price * event.Quantity
		if notional >= cfg.LargeOrderNotional {
			alerts = append(alerts, Alert{Type: AlertLargeOrder, Market: event.Market, UserID: event.TakerUserID, Severity: "MEDIUM", Reason: "large trade notional"})
		}
	}
	return alerts
}

func EvaluateActivity(activity UserActivity, cfg Config) []Alert {
	if activity.Orders <= 0 || cfg.HighCancelRatioBps <= 0 {
		return nil
	}
	ratioBps := activity.Cancellations * 10_000 / activity.Orders
	if ratioBps >= cfg.HighCancelRatioBps {
		return []Alert{{Type: AlertHighCancelRatio, UserID: activity.UserID, Severity: "LOW", Reason: "high cancellation ratio"}}
	}
	return nil
}

func abs(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
