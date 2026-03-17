# Changelog

### v1.1.0 — March 17, 2026

Internal quality improvements. No user-facing behavior changes.

**Bubble Tea modernization:**
- Key bindings defined with `bubbles/key` — single source of truth for keys and help text
- Help bar rendered with `bubbles/help` instead of hardcoded string
- `Update()` uses `key.Matches()` instead of nested switch statements
- Model fields encapsulated behind accessors

**Go idiom fixes:**
- Config validation returns warnings instead of writing directly to stderr
- Error messages follow Go conventions (no `"error: "` prefix)
- Added `.golangci.yml` with stricter linters (gocritic, revive, misspell)
- Fixed `TestMain` defer-after-exit pattern

---

### v1.0.0 — March 17, 2026

First stable release. Natural language date parsing + interactive calendar picker.

**Date parsing:**
- Natural language dates via `olebedev/when`: "tomorrow", "2 weeks ago", "march 25 2026"
- Custom "this/next/last" weekday handling: "this thursday" vs "next thursday" give correct, different results
- Three input modes: positional args, piped stdin, no-args (prints today)
- Output always `yyyy-mm-dd`
- `--help` (`-h`) and `--version` (`-v`) flags

**Interactive calendar (`wen cal`):**
- `wen cal` opens at current month, `wen cal december 2026` at a specific month
- Vim-style navigation: `h`/`l` (day), `j`/`k` (week), `H`/`L` (month), `J`/`K` (year)
- Arrow keys mirror vim keys
- `t` to jump to today
- `Enter` prints selected date to stdout, `q`/`Esc` cancels
- `w` toggles week numbers (US and ISO 8601)
- `y` yanks cursor date to clipboard (pbcopy, wl-copy, xclip, xsel)
- `?` toggles help bar
- Day/week navigation wraps across boundaries, month/year jumps clamp day

**Themes and config:**
- Four built-in color themes: default, catppuccin-mocha, dracula, nord
- Custom color overrides per element
- `~/.config/wen/config.yaml` (auto-created on first run, respects `$XDG_CONFIG_HOME`)
- Configurable week start day (Sunday/Monday) and week numbering system

**Infrastructure:**
- CI pipeline with test and lint
- GoReleaser for cross-platform releases (linux/darwin x amd64/arm64)
- Nix flake with binary and source builds
- MIT license

---

For usage details, see [README](README.md).
