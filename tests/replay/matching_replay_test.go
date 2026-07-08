// Package replay proves the recovery gates from ADR 0007 and the recovery
// runbook: snapshot + WAL replay of the tail rebuilds the same orderbook
// and the same trade sequence as live matching.
package replay

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/services/matching-engine/engine"
	"github.com/exchange-core-starter/exchange-core-starter/services/matching-engine/wal"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/oms"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/spotexchange"
)

const symbol = "BTC-USDT"

func newJournaledExchange(t *testing.T) (*spotexchange.Exchange, *wal.Log) {
	t.Helper()
	ex, err := spotexchange.New(oms.Market{
		Symbol:        symbol,
		BaseAsset:     "BTC",
		QuoteAsset:    "USDT",
		Status:        oms.MarketTrading,
		PriceScale:    2,
		QuantityScale: 8,
		MinOrderQty:   100_000,
		MinNotional:   10_00,
		TickSize:      1_00,
		LotSize:       100_000,
		MakerFeePPM:   100,
		TakerFeePPM:   200,
	})
	if err != nil {
		t.Fatalf("spotexchange.New: %v", err)
	}
	journal := wal.NewLog()
	ex.SetJournal(journal)
	return ex, journal
}

func seedBalances(t *testing.T, ex *spotexchange.Exchange) {
	t.Helper()
	for user := int64(1); user <= 4; user++ {
		if err := ex.Deposit(depositID("usdt", user), user, "USDT", 1_000_000_000_00); err != nil {
			t.Fatalf("seed USDT user %d: %v", user, err)
		}
		if err := ex.Deposit(depositID("btc", user), user, "BTC", 1_000_00_000_000); err != nil {
			t.Fatalf("seed BTC user %d: %v", user, err)
		}
	}
}

func depositID(asset string, user int64) string {
	return "seed-" + asset + "-" + string(rune('0'+user))
}

// runWorkload drives a deterministic pseudo-random mix of placements and
// cancels through the exchange. Returns the checkpoint taken mid-way.
func runWorkload(t *testing.T, ex *spotexchange.Exchange, rng *rand.Rand, operations int) spotexchange.Checkpoint {
	t.Helper()
	var checkpoint spotexchange.Checkpoint
	var openOrders []int64
	ownerOf := make(map[int64]int64)

	for i := 0; i < operations; i++ {
		if i == operations/2 {
			checkpoint = ex.Checkpoint()
		}

		if len(openOrders) > 0 && rng.Intn(5) == 0 {
			index := rng.Intn(len(openOrders))
			orderID := openOrders[index]
			if _, err := ex.CancelOrder(ownerOf[orderID], orderID); err == nil {
				openOrders = append(openOrders[:index], openOrders[index+1:]...)
			}
			continue
		}

		userID := int64(1 + rng.Intn(4))
		side := engine.Buy
		if rng.Intn(2) == 0 {
			side = engine.Sell
		}
		price := int64(49_990_00 + int64(rng.Intn(21))*1_00) // 49990.00 .. 50010.00
		quantity := int64(1+rng.Intn(20)) * 100_000          // 0.001 .. 0.02 BTC

		result, err := ex.PlaceOrder(spotexchange.PlaceOrderRequest{
			UserID:   userID,
			Symbol:   symbol,
			Side:     side,
			Type:     engine.Limit,
			Price:    price,
			Quantity: quantity,
		})
		if err != nil {
			t.Fatalf("op %d: PlaceOrder: %v", i, err)
		}
		if result.Order.Status == engine.OrderAccepted || result.Order.Status == engine.OrderPartiallyFilled {
			openOrders = append(openOrders, result.Order.ID)
			ownerOf[result.Order.ID] = userID
		}
	}
	return checkpoint
}

func TestFullReplayRebuildsBookAndTrades(t *testing.T) {
	ex, journal := newJournaledExchange(t)
	seedBalances(t, ex)
	rng := rand.New(rand.NewSource(42))
	runWorkload(t, ex, rng, 400)

	final := ex.Checkpoint()
	liveTrades := ex.Trades(0)

	replayBook := engine.NewOrderBook(symbol)
	replayTrades, err := wal.Replay(replayBook, journal.Commands())
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}

	if !reflect.DeepEqual(final.Book, replayBook.Snapshot()) {
		t.Fatal("full replay produced a different book")
	}
	if len(liveTrades) == 0 {
		t.Fatal("workload produced no trades; test is not exercising matching")
	}
	if !reflect.DeepEqual(liveTrades, replayTrades) {
		t.Fatalf("full replay produced different trades: live=%d replayed=%d", len(liveTrades), len(replayTrades))
	}
}

func TestSnapshotPlusTailReplayRebuildsBook(t *testing.T) {
	ex, journal := newJournaledExchange(t)
	seedBalances(t, ex)
	rng := rand.New(rand.NewSource(7))
	checkpoint := runWorkload(t, ex, rng, 400)

	final := ex.Checkpoint()
	liveTrades := ex.Trades(0)

	restored, err := engine.RestoreOrderBook(checkpoint.Book)
	if err != nil {
		t.Fatalf("RestoreOrderBook: %v", err)
	}
	tailTrades, err := wal.Replay(restored, journal.After(checkpoint.WALSequence))
	if err != nil {
		t.Fatalf("Replay tail: %v", err)
	}

	if !reflect.DeepEqual(final.Book, restored.Snapshot()) {
		t.Fatal("snapshot + tail replay produced a different book")
	}

	wantHash, err := final.Book.Hash()
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	gotHash, err := restored.Snapshot().Hash()
	if err != nil {
		t.Fatalf("Hash restored: %v", err)
	}
	if wantHash != gotHash {
		t.Fatalf("book hash mismatch: %s vs %s", wantHash, gotHash)
	}

	var liveTail []engine.Trade
	for _, trade := range liveTrades {
		if trade.Sequence > checkpoint.Book.Sequence {
			liveTail = append(liveTail, trade)
		}
	}
	if len(liveTail) == 0 {
		t.Fatal("no trades after checkpoint; test is not exercising tail replay")
	}
	if !reflect.DeepEqual(liveTail, tailTrades) {
		t.Fatalf("tail replay produced different trades: live=%d replayed=%d", len(liveTail), len(tailTrades))
	}
}

func TestRepeatedRecoveryIsIdentical(t *testing.T) {
	ex, journal := newJournaledExchange(t)
	seedBalances(t, ex)
	rng := rand.New(rand.NewSource(99))
	checkpoint := runWorkload(t, ex, rng, 200)
	final := ex.Checkpoint()

	var hashes []string
	for i := 0; i < 3; i++ {
		restored, err := engine.RestoreOrderBook(checkpoint.Book)
		if err != nil {
			t.Fatalf("RestoreOrderBook: %v", err)
		}
		if _, err := wal.Replay(restored, journal.After(checkpoint.WALSequence)); err != nil {
			t.Fatalf("Replay: %v", err)
		}
		hash, err := restored.Snapshot().Hash()
		if err != nil {
			t.Fatalf("Hash: %v", err)
		}
		hashes = append(hashes, hash)
	}

	finalHash, err := final.Book.Hash()
	if err != nil {
		t.Fatalf("Hash final: %v", err)
	}
	for i, hash := range hashes {
		if hash != finalHash {
			t.Fatalf("recovery %d hash %s != live %s", i, hash, finalHash)
		}
	}
}
