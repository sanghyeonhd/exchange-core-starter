# Admin Web App

The admin application is independently deployed from the user frontend.

Current implementation:

- Static operations screen in `index.html`.
- No backend connectivity yet.
- Can be opened directly in a browser for product-shell review.

Expected scope:

- User search and account status.
- KYC review queue.
- AML alert and case queue.
- Withdrawal approval queue.
- Asset/listing whitelist management.
- Market configuration and halt controls.
- Risk settings.
- Audit log search.
- Evidence export for ISMS/ISO and regulator review.

Production gates:

- MFA required.
- RBAC required.
- IP allowlist supported.
- Privileged actions require immutable audit logs.
- High-risk actions require dual approval where policy requires it.
