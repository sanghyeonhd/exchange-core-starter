package listing

type State string

const (
	StateProposed     State = "PROPOSED"
	StateLegalReview  State = "LEGAL_REVIEW"
	StateWalletReview State = "WALLET_REVIEW"
	StateRiskReview   State = "RISK_REVIEW"
	StateApproved     State = "APPROVED"
	StateListed       State = "LISTED"
	StateHalted       State = "HALTED"
	StateDelisted     State = "DELISTED"
	StateRejected     State = "REJECTED"
)

type Asset struct {
	Symbol string
	State  State
}

type Network struct {
	Asset   string
	Network string
	State   State
}

type Market struct {
	Symbol string
	State  State
}

type CounterpartyVASP struct {
	ID    string
	State State
}

type Policy struct {
	Assets         map[string]Asset
	Networks       map[string]Network
	Markets        map[string]Market
	Counterparties map[string]CounterpartyVASP
}

func (p Policy) CanTrade(market string) bool {
	value, ok := p.Markets[market]
	return ok && value.State == StateListed
}

func (p Policy) CanDeposit(asset string, network string) bool {
	assetValue, assetOK := p.Assets[asset]
	networkValue, networkOK := p.Networks[asset+":"+network]
	return assetOK && networkOK && assetValue.State == StateListed && networkValue.State == StateListed
}

func (p Policy) CanWithdraw(asset string, network string, counterpartyID string) bool {
	if !p.CanDeposit(asset, network) {
		return false
	}
	if counterpartyID == "" {
		return true
	}
	counterparty, ok := p.Counterparties[counterpartyID]
	return ok && counterparty.State == StateApproved
}
