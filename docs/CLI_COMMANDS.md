# CLI Commands Reference

## Core Commands

> **Note on `<slug>`:** The `<slug>` argument is the Firestore **document id** of the
> project (the `id` field in `ledger status --json`), not the human-readable project
> name. Commands that write by slug (`touch`, `note`) use upsert semantics: if no
> document with that id exists, a **new** project document is created rather than
> erroring — so a mistyped slug, or a project name passed where a document id is
> expected, will silently create a new project instead of updating the intended one.

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

### `ledger note <slug> <note>`
Add a note to a project. The note text is passed **inline as a positional argument** —
this command is not interactive and does not open an editor. The note is recorded as a
touch with the reason `note: <note>`.

**Flags:**
- `--json` - Output JSON format (`{ "slug", "note", "added_at" }`)

### `ledger review`
Show weekly review buckets.

**Flags:**
- `--wizard` - Run interactive wizard for review
- `--json` - Output JSON format
- `--no-color` - Disable colored status dots

### `ledger analyze`
Analyze your touch history.

**Flags:**
- `--spark` - Show ASCII sparkline of touch distribution
- `--json` - Output JSON format

The reason breakdown is always included in the `--json` output (under the `reasons`
key); there is no flag to toggle it. There is no `--reasons` flag — passing it errors
with `unknown flag: --reasons` and exits with code 50.

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
- `--format` - Output format: json or csv (default "json")
- `-o, --output` - Write to a file instead of stdout
- `--stale` - Only export stale projects (≥10 days)

## Exit Codes

- `0` - Success
- `10` - Auth error
- `20` - Not found
- `30` - Network error
- `40` - Validation error
- `50` - Other error