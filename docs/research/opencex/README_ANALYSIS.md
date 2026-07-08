# opencex 분석 보고서

## 1. 저장소 정보

- URL: https://github.com/Polygant/OpenCEX.git
- Branch: master
- Latest commit: 12d721419cc4cf1378f43be3514cfe982a3d3515
- License: Apache-2.0
- Language: config/templates 중심
- Stars/Forks: 미확인
- Last active: 미확인
- 위험도: High

## 2. 빌드 결과

- 빌드 명령: 미실행
- 성공 여부: 미확인
- 실패 사유: N/A
- 필요한 의존성: 추후 확인
- 로컬 실행 가능 여부: not maintained 여부 검토 필요

## 3. 모듈 구조

- API: `backend`
- Auth/User: 미확인
- Wallet: 미확인
- Matching: 미확인
- Market: 미확인
- Admin: `admin`
- Futures: 미확인
- Settlement: 미확인
- Risk: 미확인
- Mobile/Web: `frontend`, `nuxt`

## 4. 재사용 가능 아이디어

- 그대로 참고 가능한 개념: 기능 체크리스트, 운영 화면 구성
- 재작성해야 하는 부분: 전체 코어
- 폐기해야 하는 부분: 운영 기반 사용

## 5. 보안 이슈

- 하드코딩 secret: gitleaks 미설치로 미검증
- SQL injection 가능성: 미검증
- 인증/인가 문제: 미검증
- 출금 승인 문제: 미검증
- RPC wallet 문제: 미검증
- 로그 민감정보 문제: 미검증

## 6. 우리 프로젝트에 반영할 설계

- 반영: 기능 목록 점검
- 보류: UI 우선순위
- 제외: 실서비스 베이스 사용

