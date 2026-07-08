package evidence

import (
	"errors"
	"time"
)

var (
	ErrInvalidEvidence = errors.New("invalid evidence")
)

type Control string

const (
	ControlAccessReview         Control = "ACCESS_REVIEW"
	ControlAdminAudit           Control = "ADMIN_AUDIT"
	ControlVulnerabilityScan    Control = "VULNERABILITY_SCAN"
	ControlIncidentDrill        Control = "INCIDENT_DRILL"
	ControlRecoveryDrill        Control = "RECOVERY_DRILL"
	ControlVendorRiskReview     Control = "VENDOR_RISK_REVIEW"
	ControlLedgerReconciliation Control = "LEDGER_RECONCILIATION"
	ControlMatchingReplay       Control = "MATCHING_REPLAY"
)

type Record struct {
	ID          string
	Control     Control
	ArtifactURI string
	Owner       string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

type Register struct {
	records []Record
}

func NewRegister() *Register {
	return &Register{}
}

func (r *Register) Add(record Record) error {
	if record.ID == "" || record.Control == "" || record.ArtifactURI == "" || record.Owner == "" || record.CreatedAt.IsZero() {
		return ErrInvalidEvidence
	}
	if !record.ExpiresAt.IsZero() && !record.ExpiresAt.After(record.CreatedAt) {
		return ErrInvalidEvidence
	}
	r.records = append(r.records, record)
	return nil
}

func (r *Register) Latest(control Control) (Record, bool) {
	var latest Record
	for _, record := range r.records {
		if record.Control != control {
			continue
		}
		if latest.ID == "" || record.CreatedAt.After(latest.CreatedAt) {
			latest = record
		}
	}
	return latest, latest.ID != ""
}

func (r *Register) Missing(required []Control, at time.Time) []Control {
	var missing []Control
	for _, control := range required {
		record, ok := r.Latest(control)
		if !ok || (!record.ExpiresAt.IsZero() && !record.ExpiresAt.After(at)) {
			missing = append(missing, control)
		}
	}
	return missing
}
