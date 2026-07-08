# omx-engine 분석 보고서

## 1. 저장소 정보

- URL: https://github.com/0xae/omx-engine.git
- Branch: master
- Latest commit: bfc0139092957d89191083f7dccab74850bdbfc2
- License: 루트 라이선스 없음
- Language: C/C++/Python
- Stars/Forks: 미확인
- Last active: 미확인
- 위험도: High

## 2. 빌드 결과

- 빌드 명령: 미실행
- 성공 여부: 미확인
- 실패 사유: N/A
- 필요한 의존성: C/C++ toolchain 추정
- 로컬 실행 가능 여부: 추후 검증

## 3. 모듈 구조

- API: 미확인
- Auth/User: 범위 외
- Wallet: 범위 외
- Matching: `src` 하위 추정
- Market: 미확인
- Admin: 범위 외
- Futures: 옵션/선물 개념 참고
- Settlement: 추가 분석 필요
- Risk: 추가 분석 필요
- Mobile/Web: 없음

## 4. 재사용 가능 아이디어

- 그대로 참고 가능한 개념: 파생상품 엔진 분리 철학
- 재작성해야 하는 부분: Go 기반 선물 MVP
- 폐기해야 하는 부분: 라이선스 없는 코드 직접 사용

## 5. 보안 이슈

- 하드코딩 secret: gitleaks 기준 no leaks found
- SQL injection 가능성: 범위 미확인
- 인증/인가 문제: 범위 외 추정
- 출금 승인 문제: 범위 외 추정
- RPC wallet 문제: 범위 외 추정
- 로그 민감정보 문제: 미검증

## 6. 우리 프로젝트에 반영할 설계

- 반영: derivatives matching/settlement 개념 검토
- 보류: 자료구조 세부
- 제외: 코드 복사
