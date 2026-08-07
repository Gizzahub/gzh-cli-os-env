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
│                        # detect/power/system/display/shortcuts/input/backup
└── pkg/
    ├── detector/        # OS / desktop environment detection
    ├── power/           # Battery status (pure parser + platform dispatch)
    ├── system/          # hosts / locale / timezone
    ├── display/         # Display list (pure parser + platform dispatch)
    ├── shortcuts/       # Symbolic hotkeys (pure parser + platform dispatch)
    ├── input/           # Keyboard repeat / input sources (pure parser + dispatch)
    └── backup/          # Config snapshot create/restore/diff (tar.gz)
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
| Power | ✅ (sysfs/upower) | ✅ (sysfs/upower) | ✅ | ✅ (wmic/CIM) |
| System (hosts) | ✅ | ✅ | ✅ | ✅ (read) |
| System (locale/timezone) | ✅ | ✅ | ✅ | ✅ |
| Display | ✅ (xrandr/wlr) | ✅ (xrandr/wlr) | ✅ | ✅ (wmic) |
| Shortcuts | ✅ (kglobalaccel) | ✅ (gsettings) | ✅ | ✅ (accessibility) |
| Input (keyboard) | ✅ (setxkbmap) | ✅ (gsettings) | ✅ | ✅ (WinUserLanguageList) |
| Backup (snapshot) | ✅ | ✅ | ✅ | ✅ |
| Backup apply (locale/tz) | ✅ | ✅ | ✅ | set unsupported |

## Module

`github.com/gizzahub/gzh-cli-os-env` — Go 1.26+

## Status

**Rebuild complete (Phases 1–6) + restore apply.** macOS/Linux/Windows read
paths + backup create/restore/diff; `backup restore --apply` writes
locale/timezone on macOS/Linux (injectable DomainWriter). Remote git sync is a
follow-up. See `tasks/plan/GZH_CLI_OS_ENV.md` and
`tasks/done/P2-os-env-rebuild-continue.md` in gzh-cli-devbox.
