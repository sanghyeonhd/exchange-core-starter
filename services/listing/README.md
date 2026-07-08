# Listing and Whitelist Service

Responsibilities:

- Asset whitelist.
- Network whitelist.
- Token contract whitelist.
- Market listing lifecycle.
- Counterparty VASP whitelist.
- Withdrawal address whitelist policy hooks.

Initial listing states:

```text
PROPOSED
LEGAL_REVIEW
WALLET_REVIEW
RISK_REVIEW
APPROVED
LISTED
HALTED
DELISTED
REJECTED
```

No asset can be deposited, withdrawn, or traded unless the relevant asset, network, and market states allow it.

