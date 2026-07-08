package evidence

import (
	"testing"
	"time"
)

func TestRegisterReportsMissingAndExpiredEvidence(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	register := NewRegister()
	if err := register.Add(Record{
		ID:          "scan-1",
		Control:     ControlVulnerabilityScan,
		ArtifactURI: "s3://evidence/scan-1",
		Owner:       "security",
		CreatedAt:   now.Add(-time.Hour),
		ExpiresAt:   now.Add(time.Hour),
	}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := register.Add(Record{
		ID:          "access-1",
		Control:     ControlAccessReview,
		ArtifactURI: "s3://evidence/access-1",
		Owner:       "security",
		CreatedAt:   now.Add(-48 * time.Hour),
		ExpiresAt:   now.Add(-time.Hour),
	}); err != nil {
		t.Fatalf("Add expired: %v", err)
	}

	missing := register.Missing([]Control{ControlVulnerabilityScan, ControlAccessReview, ControlRecoveryDrill}, now)
	want := map[Control]bool{ControlAccessReview: true, ControlRecoveryDrill: true}
	for _, control := range missing {
		delete(want, control)
	}
	if len(want) != 0 || len(missing) != 2 {
		t.Fatalf("missing = %+v, want access review and recovery drill", missing)
	}
}
