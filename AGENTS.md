# Directory Structure

이 문서는 `/home/biosvos/workspace/recru-net`의 디렉토리 구조만 정리합니다.

완료 사운드는 모든 작업이 실제로 끝난 뒤에만 실행합니다.
구체적으로는 편집, 테스트, 검토, 추가 수정이 모두 끝나고 더 이상의
tool 호출이 필요 없는 상태에서 최종 응답을 보내기 직전에
`paplay /usr/share/sounds/freedesktop/stereo/complete.oga`를 정확히 한 번만
실행합니다.
중간 진행 상황 공유, 부분 완료, 개별 테스트 통과 직후, 병렬 작업 중, 최종 응답 전에 남은 확인 작업이 있는 상태에서는 실행하지 않습니다.

## Root

- `cmd/`: 실행 바이너리 진입점
- `deploy/`: 배포 및 로컬 실행 보조 스크립트
- `docs/`: 문서, 분석 자료, 용어집
- `internal/`: 애플리케이션 내부 구현
- `testdata/`: 테스트용 정적 데이터
- `vendor/`: vendored Go 의존성

## 주요 하위 디렉토리

- `cmd/recru/`: `recru` 실행 파일 엔트리포인트
- `docs/glossary/`: 도메인 용어 정리
- `docs/sources/`: 참고 자료 및 출처 정리
- `internal/app/`: 애플리케이션 서비스 계층
- `internal/app/flow/`: 주요 흐름 로직
- `internal/cli/`: CLI 명령 처리
- `internal/domain/`: 도메인 모델 및 인터페이스
- `internal/providers/`: 외부 채용 소스 연동 구현
- `internal/providers/jumpit/`: Jumpit provider 구현
- `testdata/saramin/`: 사람인 관련 테스트 HTML 샘플
