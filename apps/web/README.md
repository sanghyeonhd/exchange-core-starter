# User Web App

The user-facing web application is a separate app from admin.

Current implementation:

- Product shell in `index.html` + `app.js`.
- Connects to the local gateway REST API (`http://localhost:8080/api/v1`) for
  markets, orderbook, recent trades, balances, order entry, and withdrawal
  requests. Uses the development auth placeholder (`X-USER-ID: 1`).
- Falls back to static mock data when the gateway is offline, so the shell
  stays reviewable on its own.

Run locally:

```bash
# terminal 1: backend
go run ./services/gateway/cmd/gateway

# terminal 2: static server for the web shell
python3 -m http.server 8081 --directory apps/web
# open http://localhost:8081
```

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
