# User Web App

The user-facing web application is a separate app from admin.

Current implementation:

- Static screen in `index.html`.
- No backend connectivity yet.
- Can be opened directly in a browser for product-shell review.

Expected scope:

- Public market pages.
- Orderbook and trade stream.
- Authenticated account dashboard.
- Order entry and open orders.
- Wallet deposit/withdrawal UX.
- KYC status and submission flow.
- Security settings: MFA, API keys, withdrawal whitelist.

Production gates:

- No trading if KYC/AML state blocks the user.
- No withdrawal if wallet, KYC, AML, Travel Rule, or whitelist checks fail.
- No private data rendered without authenticated session checks.
