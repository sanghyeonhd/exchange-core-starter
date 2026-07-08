# exchange-core 분석 보고서

## 1. 저장소 정보

- URL: https://github.com/exchange-core/exchange-core.git
- Branch: master
- Latest commit: 2f8548749839e9095c8dc597e4b61521d259fa5d
- License: Apache-2.0
- Language: Java/Maven
- Stars/Forks: 미확인
- Last active: 미확인
- 위험도: Medium

## 2. 빌드 결과

- 빌드 명령: 미실행
- 성공 여부: 미확인
- 실패 사유: N/A
- 필요한 의존성: Maven, JDK
- 로컬 실행 가능 여부: 추후 검증

## 3. 모듈 구조

- API: 추가 분석 필요
- Auth/User: 범위 외
- Wallet: 범위 외
- Matching: `src` 하위
- Market: matching domain 내 추정
- Admin: 범위 외
- Futures: 추가 분석 필요
- Settlement: 추가 분석 필요
- Risk: 추가 분석 필요
- Mobile/Web: 없음

## 4. 재사용 가능 아이디어

- 그대로 참고 가능한 개념: 고성능 오더북 자료구조, 벤치마크 접근
- 재작성해야 하는 부분: Go 기반 matching shard와 WAL/snapshot
- 폐기해야 하는 부분: 언어/런타임 직접 의존

## 5. 보안 이슈

- 하드코딩 secret: gitleaks 미설치로 미검증
- SQL injection 가능성: 범위 외 추정
- 인증/인가 문제: 범위 외 추정
- 출금 승인 문제: 범위 외 추정
- RPC wallet 문제: 범위 외 추정
- 로그 민감정보 문제: 미검증

## 6. 우리 프로젝트에 반영할 설계

- 반영: 벤치마크 기준과 자료구조 검토
- 보류: 성능 최적화 세부 구현
- 제외: Java 코드 복사

