package wal

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/services/matching-engine/engine"
)

func placeCommand(orderID int64, side engine.Side, price, quantity int64) Command {
	return Command{
		Type:   CommandPlace,
		Symbol: "BTC-USDT",
		Order: &engine.Order{
			ID: orderID, UserID: orderID, Symbol: "BTC-USDT",
			Side: side, Type: engine.Limit, TimeInForce: engine.GTC,
			Price: price, Quantity: quantity,
		},
	}
}

func TestLogAssignsContiguousSequences(t *testing.T) {
	log := NewLog()
	for i := int64(1); i <= 3; i++ {
		appended, err := log.Append(placeCommand(i, engine.Buy, 100, 10))
		if err != nil {
			t.Fatalf("Append: %v", err)
		}
		if appended.Sequence != i || appended.SchemaVersion != SchemaVersion {
			t.Fatalf("appended = %+v", appended)
		}
	}
	if got := len(log.After(1)); got != 2 {
		t.Fatalf("After(1) = %d commands, want 2", got)
	}
}

func TestAppendRejectsInvalidCommands(t *testing.T) {
	log := NewLog()
	invalid := []Command{
		{},
		{Type: CommandPlace, Symbol: "BTC-USDT"},
		{Type: CommandCancel, Symbol: "BTC-USDT"},
		{Type: CommandPlace, Symbol: "BTC-USDT", Order: &engine.Order{ID: 1, Symbol: "ETH-USDT"}},
	}
	for i, cmd := range invalid {
		if _, err := log.Append(cmd); !errors.Is(err, ErrInvalidCommand) {
			t.Fatalf("case %d: err = %v, want ErrInvalidCommand", i, err)
		}
	}
}

func TestReplayRebuildsBookAndTrades(t *testing.T) {
	live := engine.NewOrderBook("BTC-USDT")
	log := NewLog()

	var liveTrades []engine.Trade
	commands := []Command{
		placeCommand(1, engine.Buy, 100, 10),
		placeCommand(2, engine.Sell, 100, 4),
		placeCommand(3, engine.Sell, 101, 5),
		{Type: CommandCancel, Symbol: "BTC-USDT", OrderID: 3},
		placeCommand(4, engine.Sell, 100, 6),
	}
	for _, cmd := range commands {
		recorded, err := log.Append(cmd)
		if err != nil {
			t.Fatalf("Append: %v", err)
		}
		switch recorded.Type {
		case CommandPlace:
			result, err := live.Submit(*recorded.Order)
			if err != nil {
				t.Fatalf("Submit: %v", err)
			}
			liveTrades = append(liveTrades, result.Trades...)
		case CommandCancel:
			if _, err := live.Cancel(recorded.OrderID); err != nil {
				t.Fatalf("Cancel: %v", err)
			}
		}
	}

	replayed := engine.NewOrderBook("BTC-USDT")
	replayedTrades, err := Replay(replayed, log.Commands())
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if !reflect.DeepEqual(liveTrades, replayedTrades) {
		t.Fatalf("trades differ:\nlive: %+v\nreplayed: %+v", liveTrades, replayedTrades)
	}
	if !reflect.DeepEqual(live.Snapshot(), replayed.Snapshot()) {
		t.Fatalf("books differ:\nlive: %+v\nreplayed: %+v", live.Snapshot(), replayed.Snapshot())
	}
}

func TestReplayDetectsInconsistentLog(t *testing.T) {
	book := engine.NewOrderBook("BTC-USDT")
	_, err := Replay(book, []Command{
		{Type: CommandCancel, Symbol: "BTC-USDT", OrderID: 99, SchemaVersion: SchemaVersion, Sequence: 1},
	})
	if !errors.Is(err, ErrReplayDiverged) {
		t.Fatalf("err = %v, want ErrReplayDiverged", err)
	}
}

func TestFileLogRoundTripAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "matching.wal")

	log, err := OpenFileLog(path)
	if err != nil {
		t.Fatalf("OpenFileLog: %v", err)
	}
	if _, err := log.Append(placeCommand(1, engine.Buy, 100, 10)); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if _, err := log.Append(placeCommand(2, engine.Sell, 100, 5)); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := log.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Reopen continues the sequence.
	reopened, err := OpenFileLog(path)
	if err != nil {
		t.Fatalf("OpenFileLog reopen: %v", err)
	}
	appended, err := reopened.Append(Command{Type: CommandCancel, Symbol: "BTC-USDT", OrderID: 1})
	if err != nil {
		t.Fatalf("Append after reopen: %v", err)
	}
	if appended.Sequence != 3 {
		t.Fatalf("sequence after reopen = %d, want 3", appended.Sequence)
	}
	if err := reopened.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	commands, err := LoadFileLog(path)
	if err != nil {
		t.Fatalf("LoadFileLog: %v", err)
	}
	if len(commands) != 3 || commands[2].Type != CommandCancel {
		t.Fatalf("commands = %+v", commands)
	}

	book := engine.NewOrderBook("BTC-USDT")
	trades, err := Replay(book, commands)
	if err != nil {
		t.Fatalf("Replay from file: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("trades = %+v, want 1 trade", trades)
	}
}
