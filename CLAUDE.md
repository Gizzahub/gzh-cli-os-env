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
├── cmd/os-env/          # CLI entry point (NewRootCmd) + detect/power/system/display
└── pkg/
    ├── detector/        # OS / desktop environment detection
    ├── power/           # Battery status (pure parser + macOS dispatch)
    ├── system/          # hosts / locale / timezone
    └── display/         # Display list (pure parser + macOS dispatch)
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
| Power | planned | planned | ✅ | planned |
| System (hosts) | ✅ | ✅ | ✅ | planned |
| System (locale/timezone) | planned | planned | ✅ | planned |
| Display | planned | planned | ✅ | planned |
| Shortcuts | planned | planned | planned | planned |
| Input | planned | planned | planned | planned |

## Module

`github.com/gizzahub/gzh-cli-os-env` — Go 1.26+

## Status

**Rebuild in progress.** Implemented: detection, battery (macOS),
system hosts/locale/timezone, display list (macOS). Remaining — Shortcuts,
Input, Backup/Sync, non-macOS backends. See `tasks/plan/GZH_CLI_OS_ENV.md`
and `tasks/doing/P2-os-env-rebuild-continue.md` in gzh-cli-devbox.
