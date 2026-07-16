# Product Goals (No-PRD)

**Project**: gzh-cli-os-env (library — `NewRootCmd()`, 바이너리 없음)
**Doc Type**: Goals + Constraints + Quality Gates
**Status**: Active — **Rebuild in progress (Phase 3/6)**
**Last Updated**: 2026-07-16

______________________________________________________________________

## Product Intent

gzh-cli-os-env reads and reports **OS / desktop environment configuration** across
KDE · GNOME · macOS · Windows. It:

- detects the running OS and desktop environment,
- exposes per-domain state (power, system, display) through pure parsers with
  platform dispatch,
- and is **being rebuilt from zero** — the repo was Code Lost, and its scope is
  governed by the design doc, not by ad-hoc growth.

This is a feature-library project — a single PRODUCT.md is sufficient. It
replaces a PRD.

| 제공하는 것 (Is)                              | 되지 않을 것 (Is Not)                       |
| --------------------------------------------- | ------------------------------------------- |
| OS·데스크톱 환경 감지 (KDE/GNOME/macOS/Win)   | OS·데스크톱 환경 자체 대체                  |
| 도메인별 설정 조회 (power·system·display)     | 설계문서 6개 영역 밖으로의 확장             |
| 순수 파서 + 플랫폼 dispatch (net-env 패턴)    | 독립 실행 바이너리                          |
| gzh-cli wrapper가 마운트하는 라이브러리       | 권한 상승·GUI                               |

______________________________________________________________________

## Goals (Measurable Targets)

로드맵의 SSoT는 devbox `tasks/plan/GZH_CLI_OS_ENV.md`이며, 진행 상태는
`tasks/doing/P2-os-env-rebuild-continue.md`가 추적한다. 아래 목표는 그 Phase
구분과 1:1로 대응한다.

G1. **Environment detection** (Phase 1 — 완료)

- Target: KDE·GNOME·macOS·Windows 4종 감지 (`runtime.GOOS` + `XDG_CURRENT_DESKTOP`)
  — 현재 4/4

G2. **Domain coverage** (설계문서 6개 영역)

- Target: detect · power · system · display · shortcuts · input = 6/6
- 현재 **4/6** — Shortcuts·Input 미착수 (Phase 3 잔여)

G3. **Platform backends**

- Target: 각 도메인의 macOS·Linux(KDE/GNOME)·Windows 백엔드 완비
- 현재 **macOS 우선 구현**. Power·Display·locale/timezone은 macOS만 동작하고
  그 외는 `ErrUnsupported`; Linux는 Phase 4, Windows는 Phase 5

G4. **Pure-function separation**

- Target: 모든 파싱 로직은 외부 명령과 분리된 순수 함수 + 테이블 테스트로 검증
  (net-env 패턴) — 현재 충족 (`ParseBatteryOutput` 등)

G5. **Test reliability**

- Target: 커버리지 >= 80%
- 현재 pkg/display 80.0% · pkg/detector 78.6% · pkg/power 57.9% ·
  pkg/system 51.5% · cmd/os-env 32.5%

______________________________________________________________________

## Non-Goals (Explicitly Out of Scope)

- No 독립 실행 바이너리 — 라이브러리로 존재한다 (SOUL 게이트 2)
- No OS·데스크톱 환경 자체 대체 — 기존 도구(`pmset`, `gsettings`, `powercfg` 등)를
  감쌀 뿐이다 (SOUL 신념 1)
- No 설계문서 6개 영역 밖 확장 — 범위는 `tasks/plan/GZH_CLI_OS_ENV.md`가 고정한다
- No 권한 상승(sudo) — 시스템 전역 설정 자동 변경 금지
- No GUI·데몬

______________________________________________________________________

## Guardrails and Technical Constraints

**Architecture**

- `pkg/<domain>/`: 순수 파싱 함수 + 플랫폼 dispatch; 미지원 플랫폼은 `ErrUnsupported`
- `cmd/os-env`: `NewRootCmd()` export — gzh-cli wrapper가 마운트하는 단일 진입점
- 참조 구현은 `gzh-cli-net-env` (cmd/netenv + pkg/netenv 패턴)

**Dependency Boundaries**

- `gzh-cli-core`만 의존 가능; 다른 feature 라이브러리 의존 금지 (GUIDELINES §2)
- 현재 직접 의존은 cobra 뿐 — core 유틸이 필요해지면 그때 도입한다

**Compatibility**

- Go 1.26 (`go.mod`; devbox 툴체인 1.26과 동일 — gitforge와 함께 정렬 완료)

**Safety**

- 현재 전 명령이 **읽기 전용**이다. Phase 6(Backup/Sync)에서 쓰기를 도입할 때
  백업·`--dry-run`을 함께 제공하지 않으면 머지할 수 없다

**Baseline**

- GUIDELINES §3 베이스라인 충족 — `Makefile`·`.golangci.yml`(v2)·CI·`LICENSE`(MIT)·
  문서·본 PRODUCT.md 보유
- devbox `go.work` use 목록에 등록됨

______________________________________________________________________

## Quality Gates (Release Readiness)

**Build and Lint**

- `make check` (fmt + lint + test) pass with no warnings

**Testing**

- `go test ./... -cover` pass; 커버리지 >= 80%

**Roadmap Fidelity**

- Phase 완료 시 `CLAUDE.md`의 Platform Support 매트릭스를 실제 구현 상태로 갱신한다
  (planned → ✅). 매트릭스와 코드의 불일치는 릴리스 차단 사유다

**Docs**

- 명령 레퍼런스가 실제 명령·플래그와 일치

______________________________________________________________________

## Decision Rules

- **Phase 순서(3 → 4 → 5 → 6)를 건너뛰지 않는다.** macOS 완결 전 Linux/Windows
  백엔드를 먼저 늘리면 도메인별 파서 계약이 굳기 전에 분기가 퍼진다
- 새 도메인은 설계문서 6개 영역 안에서만 추가한다 — 영역 확장은 설계문서 개정을
  선행한다
- 플랫폼 명령 재구현은 SOUL 게이트 1(재발명 금지)에서 거절된다
- 새 기능은 SOUL.md 4-게이트(틈 · 라이브러리 · 대량/전환 · 날카로움)를 통과해야 한다
- Quality Gates 미충족 시 릴리스는 차단된다

______________________________________________________________________

**End of Document**
