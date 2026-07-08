package audit

import (
	"testing"
	"time"
)

func TestAuditLogHashChain(t *testing.T) {
	log := NewLog()
	first, err := log.Append(Event{
		ID:           "evt-1",
		ActorID:      "admin-1",
		ActorType:    "ADMIN",
		Action:       "KYC_APPROVE",
		ResourceType: "KYC_CASE",
		ResourceID:   "case-1",
		RequestID:    "req-1",
		Reason:       "verified",
		CreatedAt:    time.Unix(1, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("Append first: %v", err)
	}
	second, err := log.Append(Event{
		ID:           "evt-2",
		ActorID:      "admin-1",
		ActorType:    "ADMIN",
		Action:       "WITHDRAWAL_APPROVE",
		ResourceType: "WITHDRAWAL",
		ResourceID:   "wd-1",
		RequestID:    "req-2",
		Reason:       "policy ok",
		CreatedAt:    time.Unix(2, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("Append second: %v", err)
	}
	if second.PrevHash != first.Hash {
		t.Fatalf("prev hash = %s, want %s", second.PrevHash, first.Hash)
	}
	if err := log.Verify(); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}
