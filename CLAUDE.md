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
│                        # detect/power/system/display/shortcuts/input
└── pkg/
    ├── detector/        # OS / desktop environment detection
    ├── power/           # Battery status (pure parser + macOS dispatch)
    ├── system/          # hosts / locale / timezone
    ├── display/         # Display list (pure parser + macOS dispatch)
    ├── shortcuts/       # Symbolic hotkeys (pure parser + macOS dispatch)
    └── input/           # Keyboard repeat / input sources (pure parser + dispatch)
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
| Power | ✅ (sysfs/upower) | ✅ (sysfs/upower) | ✅ | planned |
| System (hosts) | ✅ | ✅ | ✅ | planned |
| System (locale/timezone) | ✅ | ✅ | ✅ | planned |
| Display | ✅ (xrandr/wlr) | ✅ (xrandr/wlr) | ✅ | planned |
| Shortcuts | ✅ (kglobalaccel) | ✅ (gsettings) | ✅ | planned |
| Input (keyboard) | ✅ (setxkbmap) | ✅ (gsettings) | ✅ | planned |

## Module

`github.com/gizzahub/gzh-cli-os-env` — Go 1.26+

## Status

**Rebuild in progress.** macOS read paths are complete: detection, battery,
system hosts/locale/timezone, display list, shortcuts, input keyboard.
Phase 4 Linux backends landed (power/system/display/shortcuts/input). Remaining — Phase 5 Windows, Phase 6 Backup/Sync. See `tasks/plan/GZH_CLI_OS_ENV.md`
and `tasks/doing/P2-os-env-rebuild-continue.md` in gzh-cli-devbox.
