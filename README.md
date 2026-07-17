# gzh-cli-os-env

> OS / 데스크톱 환경(KDE Plasma, GNOME, macOS, Windows) 설정을 조회·관리하는
> gzh-cli 계열 라이브러리.

**모듈**: `github.com/gizzahub/gzh-cli-os-env` · **Root cmd**: `os-env` · **Go**: 1.26+

## 상태

재구현 Phases 1–6 완료(2026-07-17). macOS + Linux(KDE/GNOME) + Windows 읽기
경로와 설정 스냅샷 백업(`backup create|restore|diff`)이 있다. 호스트 적용
(apply-restore)과 git remote sync는 후속 작업.

## 주요 기능

- **데스크톱 환경 감지** — macOS, Windows, KDE Plasma, GNOME (Linux는
  `XDG_CURRENT_DESKTOP` / `DESKTOP_SESSION` 기반)
- **배터리 상태 조회** — macOS (`pmset`) · Linux (sysfs/upower)
- **hosts 파일 조회** — `/etc/hosts` 파싱, macOS/Linux
- **locale / timezone 조회** — macOS · Linux (LANG/localectl · localtime/timedatectl)
- **디스플레이 목록 조회** — macOS · Linux (xrandr / wlr-randr)

## 명령 구조

독립 바이너리는 아직 없다(main 패키지 없음). 아래 서브커맨드 트리는
gzh-cli wrapper(`gz os-env ...`)를 통해 노출될 구조다:

```bash
os-env detect                # 현재 OS/데스크톱 환경 감지
os-env power battery          # 배터리 상태 (macOS)
os-env system hosts           # /etc/hosts 항목 목록
os-env system locale          # 현재 locale (macOS)
os-env system timezone        # 현재 timezone (macOS)
os-env display list           # 연결된 디스플레이 목록
os-env backup create -o snap.tar.gz
os-env backup restore -i snap.tar.gz
os-env backup diff --from snap.tar.gz
```

## Library Usage

```go
import osenv "github.com/gizzahub/gzh-cli-os-env/cmd/os-env"

cmd := osenv.NewRootCmd()  // gzh-cli wrapper에서 사용
```

## Architecture

```
cmd/os-env/          CLI entry point (NewRootCmd) + detect/power/system/display/shortcuts/input/backup
pkg/detector/        OS / 데스크톱 환경 감지
pkg/power/           배터리 상태 (순수 파싱 + 플랫폼 dispatch)
pkg/system/          hosts / locale / timezone
pkg/display/         디스플레이 목록
pkg/shortcuts/       단축키
pkg/input/           키보드
pkg/backup/          설정 스냅샷 create/restore/diff
```

## Platform Support

| Feature                  | KDE Plasma | GNOME   | macOS | Windows |
|---------------------------|------------|---------|-------|---------|
| Detection                 | ✅         | ✅      | ✅    | ✅      |
| Power                      | ✅ (sysfs)  | ✅ (sysfs) | ✅    | ✅      |
| System (hosts)             | ✅         | ✅      | ✅    | planned |
| System (locale/timezone)   | ✅         | ✅      | ✅    | ✅      |
| Display                    | ✅ (xrandr) | ✅ (xrandr) | ✅    | ✅      |
| Shortcuts                  | ✅         | ✅      | ✅    | ✅      |
| Input                      | ✅         | ✅      | ✅    | ✅      |
| Backup                     | ✅         | ✅      | ✅    | ✅      |

## 개발

```bash
make build     # go build ./...
make test      # go test ./...
make lint      # golangci-lint
make check     # fmt + lint + test (pre-commit)
```

## 관련 프로젝트

- [gzh-cli](https://github.com/gizzahub/gzh-cli)
- [gzh-cli-net-env](https://github.com/Gizzahub/gzh-cli-net-env) — 같은
  순수 함수 분리 + 플랫폼 dispatch 패턴을 따르는 참조 구현
- [gzh-cli-devbox](https://github.com/gizzahub/gzh-cli-devbox)
