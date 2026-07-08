package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrInvalidEvent = errors.New("invalid audit event")
)

type Event struct {
	ID           string
	ActorID      string
	ActorType    string
	Action       string
	ResourceType string
	ResourceID   string
	RequestID    string
	Reason       string
	CreatedAt    time.Time
	PrevHash     string
	Hash         string
}

type Log struct {
	mu     sync.Mutex
	events []Event
}

func NewLog() *Log {
	return &Log{}
}

func (l *Log) Append(event Event) (Event, error) {
	if event.ID == "" || event.ActorID == "" || event.ActorType == "" || event.Action == "" || event.ResourceType == "" || event.ResourceID == "" {
		return Event{}, ErrInvalidEvent
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.events) > 0 {
		event.PrevHash = l.events[len(l.events)-1].Hash
	}
	event.Hash = hashEvent(event)
	l.events = append(l.events, event)
	return event, nil
}

func (l *Log) Verify() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	prev := ""
	for i, event := range l.events {
		if event.PrevHash != prev {
			return fmt.Errorf("%w: event %d previous hash mismatch", ErrInvalidEvent, i)
		}
		if event.Hash != hashEvent(event) {
			return fmt.Errorf("%w: event %d hash mismatch", ErrInvalidEvent, i)
		}
		prev = event.Hash
	}
	return nil
}

func (l *Log) Events() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()

	copied := make([]Event, len(l.events))
	copy(copied, l.events)
	return copied
}

func hashEvent(event Event) string {
	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%d",
		event.ID,
		event.ActorID,
		event.ActorType,
		event.Action,
		event.ResourceType,
		event.ResourceID,
		event.RequestID,
		event.Reason,
		event.PrevHash,
		event.CreatedAt.UnixNano(),
	)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
