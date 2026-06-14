# CLI Commands Reference

## Core Commands

### `ledger init`
Initialize configuration file and set up local credentials directory.

### `ledger touch <slug> [--reason <reason>]`
Record a project touch with optional reason. Always appends to touchHistory (capped at 8 entries).

**Flags:**
- `--reason` - Touch reason (or preset: blocked, research, fix, feat, review, cleanup)
- `--wizard` - Interactive wizard mode
- `--next` - Update next action field
- `--json` - Output JSON format
- `--no-color` - Disable colored status dots

### `ledger status [--stale] [--json]`
Display project status overview.

**Flags:**
- `--stale` - Only show stale projects (RED status)
- `--json` - Output JSON format

### `ledger note <slug>`
Add or edit notes for a project. Opens editor for multiline input.

### `ledger review`
Interactive review mode for touching multiple projects.

**Flags:**
- `--wizard` - Enable wizard mode

### `ledger analyze [--reasons]`
Analyze project metrics and touch patterns.

**Flags:**
- `--reasons` - Include reason breakdown analysis

### `ledger doctor`
Check system connectivity and configuration health.

## Auth Commands

### `ledger auth login`
Interactive login flow.

### `ledger auth logout`
Clear stored credentials.

### `ledger auth whoami`
Display current authenticated user.

### `ledger auth service-account`
Generate or update service account credentials for headless use.

## Config Commands

### `ledger config show`
Display current configuration.

### `ledger config set <key> <value>`
Set configuration value.

## Export Commands

### `ledger export [--format json|csv]`
Export project data.

**Flags:**
- `--format` - Output format (json or csv)

## Exit Codes

- `0` - Success
- `10` - Auth error
- `20` - Not found
- `30` - Network error
- `40` - Validation error
- `50` - Other error