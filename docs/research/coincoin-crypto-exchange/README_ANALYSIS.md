# coincoin-crypto-exchange 분석 보고서

## 1. 저장소 정보

- URL: https://gitee.com/coincoin/crypto-exchange.git
- Branch: master
- Latest commit: 0a6d6dbda24ea143bf76ca9f3948ce88f80639bd
- License: Apache-2.0
- Language: Java/Vue/Android/iOS/Maven
- Stars/Forks: 미확인
- Last active: 미확인
- 위험도: High

## 2. 빌드 결과

- 빌드 명령: 미실행
- 성공 여부: 미확인
- 실패 사유: N/A
- 필요한 의존성: Maven, Node, Android/iOS toolchain 가능성
- 로컬 실행 가능 여부: 추후 검증

## 3. 모듈 구조

- API: `00_framework/*-api`
- Auth/User: `00_framework/ucenter-api`
- Wallet: `00_framework/wallet`, `01_wallet_rpc`
- Matching: `00_framework/exchange-core`
- Market: `00_framework/market`
- Admin: `04_Web_Admin`
- Futures: 미확인
- Settlement: 미확인
- Risk: 미확인
- Mobile/Web: Android, iOS, admin, front 포함

## 4. 재사용 가능 아이디어

- 그대로 참고 가능한 개념: 대형 거래소 기능 인벤토리, wallet RPC 분리
- 재작성해야 하는 부분: 보안/원장/지갑/매칭 핵심
- 폐기해야 하는 부분: 불명확하거나 오래된 운영 방식

## 5. 보안 이슈

- 하드코딩 secret: gitleaks 미설치로 미검증
- SQL injection 가능성: 미검증
- 인증/인가 문제: 위험 높음
- 출금 승인 문제: 별도 검토 필요
- RPC wallet 문제: 구조 참고만
- 로그 민감정보 문제: 미검증

## 6. 우리 프로젝트에 반영할 설계

- 반영: 서비스/프론트/관리 기능 범위 분석
- 보류: 운영 UI 기능 우선순위
- 제외: 원본 코드 직접 사용

