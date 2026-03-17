# Changelog

### v0.2.0 — March 17, 2026: Interactive Calendar & Rename

**Renamed from `zdate` to `wen`.**

**Interactive calendar:**
- `wen cal` launches an interactive terminal calendar at the current month
- `wen cal december 2026` starts the calendar at a specific month
- Vim-style navigation: `h`/`l` (day), `j`/`k` (week), `H`/`L` (month)
- `Enter` prints selected date to stdout, `q`/`Esc` cancels (exit 1)
- Today highlighted with bold/underline, cursor with reverse video
- Day/week navigation wraps across month boundaries
- Month jumps clamp day number (e.g., Jan 31 → Feb 28)
- Built with Bubbletea and Lipgloss

**Breaking changes:**
- `--now` flag removed
- Binary renamed from `zdate` to `wen`
- GitHub repo moved to `zachthieme/wen`

### v0.1.0 — March 17, 2026: Initial Release

- Natural language date parsing via `olebedev/when`
- Three input modes: positional args, piped stdin, no-args (prints today)
- Output always `yyyy-mm-dd`
- Error messages to stderr with exit code 1
- Nix flake with binary and source builds
- GoReleaser for cross-platform releases (linux/darwin x amd64/arm64)
- CI pipeline with test and lint

---

For usage details, see [README](README.md).
