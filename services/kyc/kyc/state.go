package kyc

type Status string

const (
	StatusNotStarted  Status = "NOT_STARTED"
	StatusPending     Status = "PENDING"
	StatusUnderReview Status = "UNDER_REVIEW"
	StatusApproved    Status = "APPROVED"
	StatusRejected    Status = "REJECTED"
	StatusExpired     Status = "EXPIRED"
	StatusSuspended   Status = "SUSPENDED"
)

type Profile struct {
	UserID   int64
	Level    int
	Status   Status
	Country  string
	Provider string
}

func (p Profile) CanTrade() bool {
	return p.Status == StatusApproved && p.Level >= 1
}

func (p Profile) CanWithdraw() bool {
	return p.Status == StatusApproved && p.Level >= 2
}
