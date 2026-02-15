# Cali Logger UX Enhancements Plan

## Goal
Improve terminal experience so logging is faster, clearer, and safer while keeping workflows simple and script-friendly.

## UX Principles
- Keep logging under 20 seconds for common sessions.
- Never block workout capture due to optional features.
- Prefer explicit defaults and clear errors.
- Keep all enhancements optional and backward compatible.

## Phase 1: Quick Wins (High Value, Low Complexity)

### 1. Startup Status Line
- Show active backend (`sheets` or `local`) and target destination before prompts.
- Example: `Backend: sheets | Sheet: Log`.
- Value: reduces configuration confusion.

### 2. Smart Input Defaults
- Auto-suggest previous `Day`, `Exercise`, and `Level`.
- Allow Enter to accept suggested value.
- Value: speeds repeated training patterns.

### 3. Better Prompt Validation
- Re-prompt for invalid day/number instead of silent fallback.
- Show valid range and expected format.
- Value: prevents accidental wrong entries.

### 4. Success Summary
- After log save, print compact summary row.
- Example: `Saved: 2026-02-14 | A | Pushups | Incline | 40x3`.
- Value: immediate confidence and auditability.

### 5. Consistent Command Aliases
- Add short aliases for common actions (`-h`, `-p`, `-s`, `-r`, `-yt`).
- Ensure help page lists all aliases clearly.
- Value: lower memory load.

## Phase 2: Flow Improvements (Medium Complexity)

### 6. Non-Interactive Quick Add
- Add command:
  - `cali add -d A -e Pushups -l Incline -r 40x3 -c "solid"`
- Value: fast logging from scripts and shell history.

### 7. `today` and `week` Views
- `cali today`: show today’s entries.
- `cali week`: grouped summary with counts.
- Value: quick progress check without spreadsheets.

### 8. `status` Command
- `cali status` prints backend, auth file path, sheet/tab, and last sync/read status.
- Value: troubleshooting without guessing.

### 9. Better Table Output
- Align columns in history/search output.
- Truncate long comments safely with optional full view flag.
- Value: better readability in terminal.

### 10. Remove-Flow Safety
- Show exact row preview before deletion and require confirmation.
- Value: prevents accidental data loss.

## Phase 3: Reliability + Resilience

### 11. Offline Queue for Sheets Mode
- If Google API fails, enqueue unsent entries locally.
- Add `cali sync` to retry queued writes.
- Value: no missed logs during network issues.

### 12. Retry with Helpful Errors
- Add structured error messages for auth, permissions, missing env vars, and API failures.
- Include direct “Fix this” hint in each case.
- Value: faster recovery.

### 13. Config File Support
- Add `~/.config/cali/config.yaml` for non-secret defaults.
- Keep env vars as overrides.
- Value: cleaner shell setup.

## Phase 4: Advanced UX (Optional)

### 14. Interactive Picker Mode
- Add optional fuzzy exercise/level selection.
- Keep plain mode as fallback.
- Value: faster navigation when command list grows.

### 15. Shell Autocomplete
- Generate completion scripts for bash/zsh/fish.
- Include commands, flags, and known exercises.
- Value: fewer typing errors.

### 16. Session Templates
- Predefined routines (`A`, `B`, `C`) with one command kickoff.
- Value: speed + consistency.

### 17. Goal Progress Indicator
- Show achieved vs target ratio after each log.
- Value: motivational feedback without spreadsheet open.

## Tutorial UX Enhancements

### 18. Tutorial Policy Modes
- Add setting for tutorial behavior:
  - `off`
  - `ask` (current behavior)
  - `auto-open`
  - `open-and-exit` (current implementation)
- Value: user control by preference.

### 19. Tutorial Source Versioning
- Load links from `toturial links.yaml` with validation at startup.
- Warn if required exercise/level link is missing.
- Value: data quality and maintainability.

## Documentation UX

### 20. Command-Centric Quickstart
- Add `QUICKSTART.md` with 90-second setup flow.
- Add copy-paste blocks for Linux/macOS/Windows.
- Value: faster onboarding.

### 21. Troubleshooting Matrix
- Error -> Cause -> Fix table in README.
- Value: self-service support.

## Prioritized Delivery Roadmap

### Sprint 1
- Startup status line
- Smart defaults
- Better validation
- Success summary

### Sprint 2
- `add` command
- `today` / `week`
- `status`
- Better output formatting

### Sprint 3
- Offline queue + `sync`
- Config file
- Retry/error taxonomy

### Sprint 4
- Interactive picker
- Autocomplete
- Templates
- Goal progress indicator

## Acceptance Criteria
- New user can log first entry in under 3 minutes.
- Returning user can log entry in under 20 seconds.
- No silent fallback on invalid input.
- Sheets mode failures never lose user input.
- Help and README stay aligned with implemented commands.

## Next Recommended Implementation Order
1. Smart defaults + validation improvements.
2. `status` and `today` command.
3. `add` non-interactive command.
4. Offline queue + `sync`.
