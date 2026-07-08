# coinglobalvip 분석 보고서

## 1. 저장소 정보

- URL: https://gitee.com/source/coinglobalvip.git
- Branch: master
- Latest commit: bf4042f62cc5bfe3753bc84f60c84da1b8825a72
- License: Apache-2.0
- Language: Java/Vue/Maven
- Stars/Forks: 미확인
- Last active: 미확인
- 위험도: High

## 2. 빌드 결과

- 빌드 명령: 미실행
- 성공 여부: 미확인
- 실패 사유: N/A
- 필요한 의존성: Maven, Node, SpringCloud 계열 런타임 가능성
- 로컬 실행 가능 여부: 추후 검증

## 3. 모듈 구조

- API: `00_framework/*-api`
- Auth/User: `00_framework/ucenter-api`
- Wallet: `00_framework/wallet`, `01_wallet_rpc`
- Matching: `00_framework/exchange`, `exchange-core`
- Market: `00_framework/market`
- Admin: `04_Web_Admin`
- Futures: 문서/코드 추가 확인 필요
- Settlement: `exchange-core` 추정
- Risk: 미확인
- Mobile/Web: `02_App_Android`, `03_APP_IOS`, `05_Web_Front`

## 4. 재사용 가능 아이디어

- 그대로 참고 가능한 개념: wallet RPC 모듈 분리, admin/front/wallet/market 기능 범위
- 재작성해야 하는 부분: 인증, 권한, 지갑, 원장, 매칭 핵심 로직
- 폐기해야 하는 부분: 보안성 미확인 운영 구조, 로봇/시세 조작으로 오해될 수 있는 구성

## 5. 보안 이슈

- 하드코딩 secret: gitleaks leak 후보 발견, 구조 참고만
- SQL injection 가능성: 미검증
- 인증/인가 문제: 위험 높음, 구조 참고만
- 출금 승인 문제: wallet RPC 보안 모델 별도 검토 필요
- RPC wallet 문제: 구조 참고만
- 로그 민감정보 문제: 미검증

## 6. 우리 프로젝트에 반영할 설계

- 반영: wallet adapter와 admin 승인 플로우의 기능 범위
- 보류: 일부 운영 화면 기능 목록
- 제외: 원본 운영 코드 직접 사용
