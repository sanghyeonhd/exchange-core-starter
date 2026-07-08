# futures-engine 분석 보고서

## 1. 저장소 정보

- URL: https://github.com/Johnny1110/futures_engine.git
- Branch: main
- Latest commit: 2654bceb4ae2b8258d9b6eac74efc0deafbe0488
- License: MIT
- Language: Go
- Stars/Forks: 미확인
- Last active: 미확인
- 위험도: Medium

## 2. 빌드 결과

- 빌드 명령: 미실행
- 성공 여부: 미확인
- 실패 사유: N/A
- 필요한 의존성: Go modules
- 로컬 실행 가능 여부: 추후 검증

## 3. 모듈 구조

- API: `cmd`/`internal` 추가 분석 필요
- Auth/User: 미확인
- Wallet: 미확인
- Matching: 추가 분석 필요
- Market: 추가 분석 필요
- Admin: 미확인
- Futures: `internal`, `pkg`, `docs`
- Settlement: 추가 분석 필요
- Risk: 추가 분석 필요
- Mobile/Web: 없음

## 4. 재사용 가능 아이디어

- 그대로 참고 가능한 개념: 선물 도메인 분리, 포지션/마진/청산 접근
- 재작성해야 하는 부분: 우리 고정소수점/원장 규칙에 맞춘 구현
- 폐기해야 하는 부분: 검토 전 직접 코드 사용

## 5. 보안 이슈

- 하드코딩 secret: gitleaks 기준 no leaks found
- SQL injection 가능성: 미검증
- 인증/인가 문제: 범위 미확인
- 출금 승인 문제: 범위 외 추정
- RPC wallet 문제: 범위 외 추정
- 로그 민감정보 문제: 미검증

## 6. 우리 프로젝트에 반영할 설계

- 반영: futures MVP 설계 비교 자료
- 보류: 개별 알고리즘
- 제외: 코드 복사
