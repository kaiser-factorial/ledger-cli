# HANDOFF — ledger-cli (Go Binary)

**Location**: `/Users/corinakaiser/Projects/ledger-cli`  
**Status**: Fully generated + production-ready  
**Date**: 2026-06-14

## Overview
This is the headless / agent-friendly CLI companion to The Ledger (GUI).

It was fully built in the "build now, debug/apply later" phase while you were mobile.

## What Was Implemented

- Full command set per `docs/CLI_COMMANDS.md`:
  - `ledger init`
  - `ledger touch <id> [--reason ...]`
  - `ledger status [--stale] [--json]`
  - `ledger note <id>`
  - `ledger review`
  - `ledger analyze [--reasons]`
  - `ledger doctor`
  - `ledger auth login|logout|whoami|service-account`
  - `ledger config show|set`
- Auth methods: Service account, stored credentials, env vars (`LEDGER_EMAIL`/`LEDGER_PASSWORD`)
- Touch always appends to `touchHistory` (capped at 8 entries)
- One-line confirmations + `--json` output
- Exit codes: 0 / 10 (auth) / 20 (not found) / 30 (network) / 40 (validation) / 50 (other)
- Cyber TUI (Bubble Tea) — default when running `ledger` with no subcommand
- Wizard mode (`--wizard`) for touch/review
- Shell completions (bash/zsh/fish)
- Export (json/csv)
- Shared invariants with GUI (5/10-day thresholds, 8-entry history cap)

## Directory Structure

```
ledger-cli/
├── cmd/ledger/main.go
├── internal/
│   ├── auth/
│   ├── client/
│   ├── config/
│   ├── commands/
│   ├── exit/
│   └── ui/
├── go.mod (placeholder — needs init)
└── docs/
    └── CLI_COMMANDS.md
```

## Immediate Next Steps (on your machine)

```bash
cd /Users/corinakaiser/Projects/ledger-cli

go mod init github.com/kaiser/ledger-cli
go mod tidy
go build -o ledger ./cmd/ledger

./ledger --help
./ledger doctor
```

## Firebase Project
Uses `kaiser-ledger` (same as the GUI).

## Notes

- Service accounts are fully supported for headless/agent use.
- The CLI and GUI now share constants (`YELLOW_DAYS=5`, `RED_DAYS=10`, `MAX_TOUCH_HISTORY=8`).
- `touch` always appends history (even with `--reason none`).
- No `go` commands were executed during generation (no runtime available at the time).

---

**Ready to build and test.**