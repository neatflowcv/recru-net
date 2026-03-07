# Jumpit API cURL 명령어 분석

아래 `curl` 명령은 점핏 API에서 인기 채용 공고 목록을 조회하는 요청입니다.

## 1) 요청 대상 URL

```bash
https://jumpit-api.saramin.co.kr/api/positions?sort=popular&highlight=false
```

- HTTP 메서드: `GET`
- 쿼리 파라미터:
  - `sort=popular`: 인기순 정렬
  - `highlight=false`: 하이라이트 항목 제외(추정)

## 2) 주요 헤더 (`-H`)

- `accept: application/json, text/plain, */*`
  - JSON 응답을 우선 기대
- `accept-language: ko,en-US;q=0.9,en;q=0.8`
  - 한국어 우선 언어 선호
- `origin: https://jumpit.saramin.co.kr`
- `referer: https://jumpit.saramin.co.kr/`
  - 브라우저에서 점핏 사이트를 통해 호출된 요청임을 나타냄
- `sec-ch-ua`, `sec-fetch-*`, `priority`, `user-agent`
  - 브라우저가 자동으로 붙이는 메타 헤더
  - 대부분의 경우 필수는 아님

## 3) 쿠키 (`-b`)

- `_ga`, `_fbp`, `ab180ClientId`, `airbridge_session__jumpit` 등 추적/세션 관련 쿠키 포함
- 개인 식별 가능 정보(PII) 또는 세션 정보가 포함될 수 있어 외부 공유 시 주의 필요

## 4) 정리

- 이 명령은 브라우저 DevTools의 **Copy as cURL** 결과와 유사한 형태
- 재현성은 높지만, 실제 API 호출에는 불필요한 헤더/쿠키가 많이 포함됨
- 보통은 URL + 일부 핵심 헤더(`accept`, 필요 시 `origin/referer`)만으로도 충분히 동작 가능

## 5) 성공/실패 테스트 결과 (2026-03-04)

테스트 목적: 원본 명령에서 옵션을 줄였을 때 실제 호출 성공 여부 확인.

### 실패

1. 로컬 샌드박스 환경에서 최소 명령 실행

```bash
curl -sS -D - 'https://jumpit-api.saramin.co.kr/api/positions?sort=popular&highlight=false'
```

- 결과: 실패
- 에러: `curl: (6) Could not resolve host: jumpit-api.saramin.co.kr`
- 원인: 코드/옵션 문제가 아니라 샌드박스 DNS 제한

### 성공

네트워크 권한이 있는 환경에서 아래 모두 성공 (`HTTP 200`, 응답 메시지: `포지션 리스트가 조회되었습니다.`):

1. 원본 파라미터 유지, 헤더/쿠키 제거

```bash
curl 'https://jumpit-api.saramin.co.kr/api/positions?sort=popular&highlight=false'
```

1. `highlight` 제거

```bash
curl 'https://jumpit-api.saramin.co.kr/api/positions?sort=popular'
```

1. 쿼리 파라미터 모두 제거

```bash
curl 'https://jumpit-api.saramin.co.kr/api/positions'
```

결론: 이 엔드포인트 기준으로 `-H`(헤더)와 `-b`(쿠키)는 성공에 필수 아님.
