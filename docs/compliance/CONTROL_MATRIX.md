# Control Matrix

| Control Area | Product Feature | Evidence Artifact |
|---|---|---|
| Governance | Compliance owner, security owner, risk owner | RACI, policy approval log |
| Asset inventory | Services, databases, wallets, admin systems | Asset register |
| Access control | User roles, admin roles, service accounts | RBAC matrix, access review report |
| Admin security | MFA, privileged action approval | Admin audit log, MFA config evidence |
| KYC | Customer status and review workflow | KYC case record, vendor response archive, `services/kyc/kyc` state tests |
| AML | Risk scoring and case management | AML case, decision log, STR workflow evidence, `services/aml/aml` policy tests |
| Travel Rule | Originator/beneficiary information flow | Transfer record, counterparty whitelist |
| Wallet custody | Signer isolation, hot/cold split | Key ceremony record, signer access log |
| Withdrawal control | Address whitelist, cooldown, approval | Withdrawal audit trail |
| Asset listing | Asset/network/contract whitelist | Listing committee decision record |
| Compliance gates | Trade/deposit/withdraw allow or deny decision | `services/compliance/compliance` decision tests |
| Ledger integrity | Double-entry per asset | Reconciliation report |
| Matching integrity | Deterministic replay | Replay report and book hash |
| Market surveillance | Abnormal trading alerts | Alert record and disposition |
| Vulnerability management | Code and dependency scanning | gitleaks/govulncheck/trivy CI artifact |
| Incident response | Detection, escalation, containment | Incident drill report |
| Business continuity | Backup and recovery | Restore drill report |
| Change management | Production release controls | Release ticket, approval, rollback plan |
| Supplier risk | KYC/AML vendor, custody provider, cloud | Vendor risk review |
| Privacy | Personal data handling | Data map, retention policy, access log |
| ISO 27001 | ISMS operation and improvement | Risk treatment plan, SoA, internal audit |
| Evidence freshness | Expiring control artifacts | `services/compliance/evidence` register tests |
