# Korea Regulated Exchange Requirements

This document is an engineering planning baseline, not legal advice. A licensed Korean legal/compliance advisor must validate every filing, control, and operating procedure before launch.

## Official References Checked

| Topic | Source |
|---|---|
| ISMS/ISMS-P and VASP preliminary ISMS certification | KISA ISMS-P page: https://www.kisa.or.kr/1050602 |
| Virtual Asset User Protection Act | Korean Law Information Center: https://www.law.go.kr/LSW/lsInfoP.do?lsiSeq=261099 |
| VASP filing requirements, ISMS, real-name deposit/withdrawal account, AML readiness | FSC press release: https://fsc.go.kr/no010101/76392 |
| VASP filing manual | KoFIU PDF: https://www.kofiu.go.kr/cmn/file/downloadBoard.do?fileNm=202472010340328g.pdf&fileOrdrNo=2&ordrNo=209&seCd=0007 |
| Travel Rule | FSC press release: https://www.fsc.go.kr/no010101/77579 |
| FATF VASP risk-based guidance | FATF: https://www.fatf-gafi.org/en/publications/Fatfrecommendations/Guidance-rba-virtual-assets-2021.html |
| ISO/IEC 27001:2022 | ISO: https://www.iso.org/standard/27001 |

## Regulatory Baseline

For a Korean centralized exchange, engineering must be planned around the following gates:

- FIU/KOFIU VASP filing readiness.
- ISMS or applicable preliminary ISMS certification path for VASP onboarding.
- Real-name deposit/withdrawal account readiness when fiat-to-virtual-asset exchange is provided.
- AML/CFT program readiness immediately at or before filing acceptance.
- Customer due diligence and enhanced due diligence.
- Suspicious transaction reporting workflow.
- Travel Rule data capture, verification, transmission, retention, and exception handling.
- User asset protection: segregation, custody controls, ledger integrity, proof of holdings, withdrawal controls.
- Incident response, audit trails, log retention, and regulator-ready evidence.

## Product Scope Implications

### Spot Before Derivatives

Regulated launch should start with spot markets only. Perpetual futures and high-leverage products must remain disabled until a separate legal and regulatory review confirms they can be offered in the target jurisdiction.

### Fiat Market Gate

KRW fiat market support is a separate gate. Without real-name deposit/withdrawal account readiness and bank risk assessment readiness, the product should operate only in mock/test environments or non-fiat internal test mode.

### Whitelists

The roadmap treats whitelists as three separate controls:

- Withdrawal address whitelist: user-controlled plus risk/admin approval.
- Counterparty VASP whitelist: Travel Rule and AML counterparty eligibility.
- Asset/listing whitelist: only approved assets/networks/contracts can be listed, deposited, withdrawn, or traded.

## Certification Baseline

### ISMS / ISMS-P

Engineering must produce evidence for:

- Information security governance.
- Risk management.
- Asset inventory.
- Access control.
- Change management.
- Operations security.
- Incident response.
- Business continuity.
- Supplier and outsourcing controls.
- Personal information protection controls where applicable.

### ISO/IEC 27001

ISO/IEC 27001 requires establishing, implementing, maintaining, and continually improving an information security management system. Engineering must support:

- Scope definition.
- Risk assessment and treatment.
- Statement of Applicability.
- Control implementation evidence.
- Internal audit evidence.
- Management review evidence.
- Corrective action tracking.

## Non-Negotiable Engineering Gates

- No production mainnet wallet before signer isolation, HSM/KMS design, and custody runbook.
- No fiat deposit/withdrawal before bank integration and real-name account compliance review.
- No public launch before KYC/AML workflows and audit evidence are operational.
- No asset listing before listing committee, legal review, market surveillance, and wallet risk review.
- No admin privileged action without RBAC, MFA, audit log, and approval workflow.

