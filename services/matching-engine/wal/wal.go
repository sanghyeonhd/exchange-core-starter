// Package wal is the append-only command log for the matching engine.
// It records every accepted place/cancel command in submission order so
// that snapshot + replay of the tail rebuilds an identical orderbook and
// trade sequence (ADR 0007).
package wal

import (
	"errors"
	"sync"

	"github.com/exchange-core-starter/exchange-core-starter/services/matching-engine/engine"
)

// SchemaVersion identifies the command envelope layout.
const SchemaVersion = 1

var (
	ErrInvalidCommand = errors.New("invalid wal command")
	ErrReplayDiverged = errors.New("wal replay diverged from recorded commands")
)

type CommandType string

const (
	CommandPlace  CommandType = "PLACE"
	CommandCancel CommandType = "CANCEL"
)

// Command is one accepted matching input. For PLACE, Order is the order as
// submitted to the book (before matching). For CANCEL, OrderID names the
// resting order that was successfully canceled.
type Command struct {
	SchemaVersion int           `json:"schema_version"`
	Sequence      int64         `json:"sequence"`
	Type          CommandType   `json:"type"`
	Symbol        string        `json:"symbol"`
	Order         *engine.Order `json:"order,omitempty"`
	OrderID       int64         `json:"order_id,omitempty"`
}

// Appender is the write side of a command log.
type Appender interface {
	// Append assigns the next sequence and durably records the command.
	Append(cmd Command) (Command, error)
}

func validateCommand(cmd Command) error {
	if cmd.Symbol == "" {
		return ErrInvalidCommand
	}
	switch cmd.Type {
	case CommandPlace:
		if cmd.Order == nil || cmd.Order.ID <= 0 || cmd.Order.Symbol != cmd.Symbol {
			return ErrInvalidCommand
		}
	case CommandCancel:
		if cmd.OrderID <= 0 {
			return ErrInvalidCommand
		}
	default:
		return ErrInvalidCommand
	}
	return nil
}

// Log is an in-memory command log for tests and single-process use.
type Log struct {
	mu       sync.Mutex
	commands []Command
	seq      int64
}

func NewLog() *Log {
	return &Log{}
}

func (l *Log) Append(cmd Command) (Command, error) {
	if err := validateCommand(cmd); err != nil {
		return Command{}, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.seq++
	cmd.Sequence = l.seq
	cmd.SchemaVersion = SchemaVersion
	if cmd.Order != nil {
		copied := *cmd.Order
		cmd.Order = &copied
	}
	l.commands = append(l.commands, cmd)
	return cmd, nil
}

// Commands returns a copy of all recorded commands in append order.
func (l *Log) Commands() []Command {
	return l.After(0)
}

// After returns commands with sequence greater than seq, for replaying the
// tail after a snapshot.
func (l *Log) After(seq int64) []Command {
	l.mu.Lock()
	defer l.mu.Unlock()

	var out []Command
	for _, cmd := range l.commands {
		if cmd.Sequence > seq {
			out = append(out, cmd)
		}
	}
	return out
}

// LastSequence returns the sequence of the most recent command.
func (l *Log) LastSequence() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.seq
}

// Replay feeds commands into the book in order and returns the emitted
// trades. The WAL records only accepted commands, so any rejection during
// replay means the log and the starting book state are inconsistent.
func Replay(book *engine.OrderBook, commands []Command) ([]engine.Trade, error) {
	var trades []engine.Trade
	for _, cmd := range commands {
		if err := validateCommand(cmd); err != nil {
			return nil, err
		}
		if cmd.Symbol != book.Symbol {
			return nil, ErrInvalidCommand
		}
		switch cmd.Type {
		case CommandPlace:
			result, err := book.Submit(*cmd.Order)
			if err != nil {
				return nil, errors.Join(ErrReplayDiverged, err)
			}
			trades = append(trades, result.Trades...)
		case CommandCancel:
			if _, err := book.Cancel(cmd.OrderID); err != nil {
				return nil, errors.Join(ErrReplayDiverged, err)
			}
		}
	}
	return trades, nil
}
