# CLAUDE.md - gzh-cli-os-env

OS/Desktop Environment Configuration Library for gzh-cli ecosystem.

## Quick Start

```bash
make build     # Build
make test      # Run tests
make check     # fmt + lint + test
```

## Architecture

```
gzh-cli-os-env/
├── cmd/os-env/          # CLI entry point (NewRootCmd)
│   ├── root.go          # NewRootCmd() — used by gzh-cli wrapper
│   └── detect.go        # detect subcommand
└── pkg/
    └── detector/        # OS / desktop environment detection
        ├── detector.go
        └── detector_test.go
```

## Key API

```go
import osenv "github.com/gizzahub/gzh-cli-os-env/cmd/os-env"

cmd := osenv.NewRootCmd()  // used by gzh-cli wrapper
```

## Platform Support

| Feature | KDE Plasma | GNOME | macOS | Windows |
|---------|------------|-------|-------|---------|
| Detection | ✅ | ✅ | ✅ | ✅ |
| Power | planned | planned | planned | planned |
| Shortcuts | planned | planned | planned | planned |
| Display | planned | planned | planned | planned |
| Input | planned | planned | planned | planned |
| System | planned | planned | planned | planned |

## Module

`github.com/gizzahub/gzh-cli-os-env` — Go 1.26+

## Status

**Phase 1 (scaffolding + detection).** Full feature set — Power, Shortcuts,
Display, Input, System, Backup/Sync — arrives in later phases. See
`tasks/plan/GZH_CLI_OS_ENV.md` in gzh-cli-devbox for the roadmap.
