# Launch Readiness Checklist

The `services/compliance/readiness` package encodes these gates so production scope cannot be enabled by configuration alone.

## Spot Launch

Required:

- Legal approval.
- Compliance approval.
- Security approval.
- ISMS readiness.
- VASP filing readiness.
- AML operational readiness.
- KYC operational readiness.
- Closed pilot passed.
- Incident drill passed.
- Recovery drill passed.
- Market surveillance ready.

## Fiat Enablement

Required:

- Legal approval.
- Compliance approval.
- Bank real-name deposit/withdrawal account readiness.
- VASP filing readiness.

## Mainnet Custody

Required:

- Security approval.
- Custody runbook approval.
- Travel Rule operational readiness.
- Incident drill passed.
- Recovery drill passed.

## Derivatives Enablement

Required:

- General legal approval.
- Derivatives-specific legal approval.
- Derivatives risk approval.
- Customer suitability controls.
- Market surveillance readiness.

Default stance: derivatives are blocked unless every derivatives-specific gate is satisfied.

