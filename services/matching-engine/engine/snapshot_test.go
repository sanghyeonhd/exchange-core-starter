package engine

import (
	"reflect"
	"testing"
)

func populatedBook(t *testing.T) *OrderBook {
	t.Helper()
	book := NewOrderBook("BTC-USDT")
	orders := []Order{
		{ID: 1, UserID: 1, Symbol: "BTC-USDT", Side: Buy, Type: Limit, TimeInForce: GTC, Price: 49_990_00, Quantity: 100},
		{ID: 2, UserID: 2, Symbol: "BTC-USDT", Side: Buy, Type: Limit, TimeInForce: GTC, Price: 49_990_00, Quantity: 200},
		{ID: 3, UserID: 1, Symbol: "BTC-USDT", Side: Buy, Type: Limit, TimeInForce: GTC, Price: 49_980_00, Quantity: 300},
		{ID: 4, UserID: 2, Symbol: "BTC-USDT", Side: Sell, Type: Limit, TimeInForce: GTC, Price: 50_010_00, Quantity: 400},
		{ID: 5, UserID: 1, Symbol: "BTC-USDT", Side: Sell, Type: Limit, TimeInForce: GTC, Price: 50_020_00, Quantity: 500},
	}
	for _, order := range orders {
		if _, err := book.Submit(order); err != nil {
			t.Fatalf("Submit %d: %v", order.ID, err)
		}
	}
	// Create a partial fill so the snapshot carries executed quantity.
	if _, err := book.Submit(Order{ID: 6, UserID: 3, Symbol: "BTC-USDT", Side: Sell, Type: Limit, TimeInForce: GTC, Price: 49_990_00, Quantity: 50}); err != nil {
		t.Fatalf("Submit taker: %v", err)
	}
	return book
}

func TestSnapshotRestoreRebuildsIdenticalBook(t *testing.T) {
	book := populatedBook(t)
	snapshot := book.Snapshot()

	restored, err := RestoreOrderBook(snapshot)
	if err != nil {
		t.Fatalf("RestoreOrderBook: %v", err)
	}

	if !reflect.DeepEqual(snapshot, restored.Snapshot()) {
		t.Fatalf("restored snapshot differs:\noriginal: %+v\nrestored: %+v", snapshot, restored.Snapshot())
	}

	originalHash, err := snapshot.Hash()
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	restoredHash, err := restored.Snapshot().Hash()
	if err != nil {
		t.Fatalf("Hash restored: %v", err)
	}
	if originalHash != restoredHash {
		t.Fatalf("hash mismatch: %s vs %s", originalHash, restoredHash)
	}
}

func TestRestoredBookMatchesLiveMatching(t *testing.T) {
	book := populatedBook(t)
	restored, err := RestoreOrderBook(book.Snapshot())
	if err != nil {
		t.Fatalf("RestoreOrderBook: %v", err)
	}

	taker := Order{ID: 7, UserID: 3, Symbol: "BTC-USDT", Side: Sell, Type: Limit, TimeInForce: GTC, Price: 49_980_00, Quantity: 400}
	liveResult, err := book.Submit(taker)
	if err != nil {
		t.Fatalf("Submit live: %v", err)
	}
	restoredResult, err := restored.Submit(taker)
	if err != nil {
		t.Fatalf("Submit restored: %v", err)
	}

	if !reflect.DeepEqual(liveResult.Trades, restoredResult.Trades) {
		t.Fatalf("trades differ:\nlive: %+v\nrestored: %+v", liveResult.Trades, restoredResult.Trades)
	}
	if !reflect.DeepEqual(book.Snapshot(), restored.Snapshot()) {
		t.Fatal("books diverged after identical command")
	}
}

func TestRestoreRejectsBadSnapshots(t *testing.T) {
	valid := populatedBook(t).Snapshot()

	wrongVersion := valid
	wrongVersion.SchemaVersion = 99
	if _, err := RestoreOrderBook(wrongVersion); err == nil {
		t.Fatal("expected schema version error")
	}

	duplicated := valid
	duplicated.Asks = append([]Order{}, duplicated.Asks...)
	duplicated.Asks = append(duplicated.Asks, duplicated.Asks[0])
	if _, err := RestoreOrderBook(duplicated); err == nil {
		t.Fatal("expected duplicate order error")
	}

	broken := valid
	broken.Bids = append([]Order{}, broken.Bids...)
	broken.Bids[0].RemainingQuantity = 0
	if _, err := RestoreOrderBook(broken); err == nil {
		t.Fatal("expected invalid resting order error")
	}
}
