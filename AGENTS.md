# gzh-cli-os-env

## Development workflow — user decision 2026-11-18

- Use the CE task runtime for worktree lifecycle and integration:
  `ce task run-doctor`, `run-start <card-id> --type <type>`, work and push in
  the task worktree, `run-status`, then `run-finish` from the primary checkout.
  `.ce/task-runtime.yaml` is the tracked declaration; worktree placement comes
  from the machine's CE defaults. This replaces the 2026-11-06 plain-Git rule.
- This repository keeps no task board. Its cards live only in the
  gzh-cli-devbox board; name the run after that card's ID. `tasks/` is ignored
  here; do not create cards or receipts in this repository.
- Run the repository's applicable source checks and keep failures visible.
  Documentation and workflow-only changes require diff checks and relevant
  Markdown/configuration or workflow checks; do not require unrelated Go test
  suites for prose edits.
- GitLab CI is not required to build, test, or contribute. GitHub Actions and
  server-side branch rules are separate and must still be respected where present.
- Preserve in-file locked regions. Do not change machine-global installed tools
  or unrelated repositories as part of this repository's workflow.

OS/Desktop Environment Configuration Library for gzh-cli ecosystem.

## Quick Start

```bash
make build     # Build
make test      # Run tests
make check     # fmt + lint + test
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
| Sync (local remote dir) | ✅ | ✅ | ✅ | ✅ |

## Module

`github.com/gizzahub/gzh-cli-os-env` — Go 1.26+

## Status

**Rebuild complete (Phases 1–6) + restore apply + minimal remote sync.**
macOS/Linux/Windows read paths + backup create/restore/diff; `backup restore
--apply` writes locale/timezone on macOS/Linux (injectable DomainWriter).
`os-env sync status|push|pull --remote <local-dir>` mirrors the backup archive
(secret allowlist: `manifest.json` only). Network git URLs are not dialed
directly — clone first, then pass the local path. See
`tasks/plan/GZH_CLI_OS_ENV.md` and `tasks/done/P2-os-env-rebuild-continue.md`
in gzh-cli-devbox.
