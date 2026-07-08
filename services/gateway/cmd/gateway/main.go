package main

import (
	"log"
	"net/http"

	"github.com/exchange-core-starter/exchange-core-starter/services/compliance/readiness"
	"github.com/exchange-core-starter/exchange-core-starter/services/gateway/httpapi"
)

func main() {
	server := httpapi.NewServer([]httpapi.Market{
		{Symbol: "BTC-USDT", BaseAsset: "BTC", QuoteAsset: "USDT", Type: "SPOT", Status: "TRADING"},
	}, readiness.Status{})

	log.Println("gateway listening on :8080")
	if err := http.ListenAndServe(":8080", server.Handler()); err != nil {
		log.Fatal(err)
	}
}
