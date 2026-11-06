# gzh-cli-os-env

## Development workflow — user decision 2026-11-06

- This repository uses ordinary Git and its own Make targets. CE, `ce task`,
  `ce access`, task_manager, and task-manager are not prerequisites and must not
  be invoked for development, worktree lifecycle, task records, verification,
  integration, or cleanup. This product-specific user decision overrides
  inherited instructions requiring those tools or proposing their adoption.
- Use an isolated task branch/worktree; preserve existing work and commits.
  Use Git directly for status, commit, task-branch push, integration and cleanup.
- Run the repository's applicable source checks and keep failures visible.
  Task-board state and card schemas are not source-quality gates. Documentation
  and workflow-only changes require diff checks and relevant Markdown/configuration
  or workflow checks; do not require unrelated Go test suites for prose edits.
- GitLab CI is not required to build, test, or contribute. GitHub Actions and
  server-side branch rules are separate and must still be respected where present.
- Existing `tasks/` documents and historical receipts are retained as records,
  not as an executable task-manager database or a reason to block source work.
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
