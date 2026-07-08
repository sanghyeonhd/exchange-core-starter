# billy-coinexchange 분석 보고서

## 1. 저장소 정보

- URL: https://github.com/Billy-Artifice/CoinExchange.git
- Branch: master
- Latest commit: 001c9bd372e42aed678189c085a4cebe944b0767
- License: Apache-2.0
- Language: Java/Vue/Maven
- Stars/Forks: 미확인
- Last active: 미확인
- 위험도: High

## 2. 빌드 결과

- 빌드 명령: 미실행
- 성공 여부: 미확인
- 실패 사유: N/A
- 필요한 의존성: Maven, Node, Java
- 로컬 실행 가능 여부: 추후 검증

## 3. 모듈 구조

- API: `00_framework/*-api`
- Auth/User: `ucenter-api`
- Wallet: `00_framework/wallet`, `01_wallet_rpc`
- Matching: `00_framework/exchange-core`
- Market: `00_framework/market`
- Admin: `04_Web_Admin`
- Futures: 미확인
- Settlement: 미확인
- Risk: 미확인
- Mobile/Web: admin/front/mobile 포함

## 4. 재사용 가능 아이디어

- 그대로 참고 가능한 개념: CoinExchange 계열 기능 범위
- 재작성해야 하는 부분: 보안/지갑/원장/체결 코어
- 폐기해야 하는 부분: 로봇/허위 유동성으로 오해될 수 있는 기능

## 5. 보안 이슈

- 하드코딩 secret: gitleaks 미설치로 미검증
- SQL injection 가능성: 미검증
- 인증/인가 문제: 위험 높음
- 출금 승인 문제: 별도 검토 필요
- RPC wallet 문제: 구조 참고만
- 로그 민감정보 문제: 미검증

## 6. 우리 프로젝트에 반영할 설계

- 반영: 기능 범위와 wallet RPC 분리 개념
- 보류: 세부 API
- 제외: 원본 코드 직접 사용

