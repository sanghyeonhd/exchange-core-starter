# opex-core 분석 보고서

## 1. 저장소 정보

- URL: https://github.com/opexdev/core.git
- Branch: main
- Latest commit: acccd9b462067998ee5c257a5cdcc056233c08cc
- License: MIT
- Language: Kotlin/Java/Maven
- Stars/Forks: 미확인
- Last active: 미확인
- 위험도: Medium

## 2. 빌드 결과

- 빌드 명령: 미실행
- 성공 여부: 미확인
- 실패 사유: N/A
- 필요한 의존성: Maven, JDK, Docker 의존 가능성
- 로컬 실행 가능 여부: 추후 검증

## 3. 모듈 구조

- API: `api`
- Auth/User: `user-management`, Keycloak gateway
- Wallet: `wallet`, `bc-gateway`
- Matching: `matching-engine`, `matching-gateway`
- Market: `market`
- Admin: 별도 명확성 미확인
- Futures: 미확인
- Settlement: `accountant`
- Risk: 미확인
- Mobile/Web: 미확인

## 4. 재사용 가능 아이디어

- 그대로 참고 가능한 개념: CEX 코어 서비스 분리, accountant/matching/wallet/API 경계
- 재작성해야 하는 부분: Go 기반 MVP 코어, 선물/마진/청산 확장
- 폐기해야 하는 부분: 라이선스 검토 전 직접 코드 사용

## 5. 보안 이슈

- 하드코딩 secret: gitleaks 미설치로 미검증
- SQL injection 가능성: 미검증
- 인증/인가 문제: Keycloak 연계 구조 추후 검토
- 출금 승인 문제: wallet/bc-gateway 추후 검토
- RPC wallet 문제: 추후 검토
- 로그 민감정보 문제: 추후 검토

## 6. 우리 프로젝트에 반영할 설계

- 반영: 모듈 경계, accountant 분리, matching gateway 분리 개념
- 보류: Kotlin/Maven 런타임 구조
- 제외: 검토 전 원본 코드 복사

