# Local Development Runbook

## Prerequisites

- Go 1.22 or newer
- Docker or Docker Compose
- Make
- Git
- Optional: `gh`, `gitleaks`, `trivy`, `govulncheck`

## Start Dependencies

```bash
docker compose up -d postgres redis nats
```

## Run Checks

```bash
make test
make vet
```

## Clone Reference Sources

```bash
make clone-sources
make audit-sources
```

Reference repositories are cloned to `~/exchange-lab/sources`. They are analysis inputs only.

## Safety Checklist

- Confirm `.env` is not committed.
- Confirm `MAINNET_ENABLED=false`.
- Confirm `WITHDRAWALS_ENABLED=false`.
- Do not place real private keys or mnemonics in local config.
- Use mock wallet flows until a production security review exists.

