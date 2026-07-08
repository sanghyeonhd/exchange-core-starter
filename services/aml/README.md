# AML Service

Responsibilities:

- Customer risk scoring.
- Sanctions/PEP/vendor screening boundary.
- Transaction monitoring rules.
- Suspicious activity case management.
- Travel Rule risk hooks.
- Evidence retention for review.

Initial case states:

```text
OPEN
NEEDS_REVIEW
ESCALATED
CLOSED_FALSE_POSITIVE
CLOSED_REPORTED
CLOSED_NO_ACTION
```

This service blocks trading or withdrawals by policy signal, not by hidden balance mutation.

