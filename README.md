# ledger-cli

Headless / agent-friendly CLI companion to The Ledger (GUI).

## Overview

A Go CLI for tracking project status, touch history, and notes with Firebase Firestore backend. Designed for terminal-based workflows and automation.

## Commands

- `ledger init` - Initialize configuration
- `ledger touch <id> [--reason ...]` - Record a project touch with optional reason
- `ledger status [--stale] [--json]` - Show project status
- `ledger note <id>` - Add notes to a project
- `ledger review` - Interactive review mode
- `ledger analyze [--reasons]` - Analyze project metrics
- `ledger doctor` - Check system connectivity
- `ledger auth login|logout|whoami|service-account` - Auth management
- `ledger config show|set` - Configuration management
- `ledger export` - Export data (json/csv)

## Installation

```bash
go build -o ledger ./cmd/ledger
```

## Configuration

Uses Firebase project `kaiser-ledger`. Supports service account credentials or `LEDGER_EMAIL`/`LEDGER_PASSWORD` environment variables.

## Status Thresholds

- YELLOW_DAYS = 5 (project touched within last 5 days)
- RED_DAYS = 10 (project stale after 10 days)
- MAX_TOUCH_HISTORY = 8 entries

## License

MIT