# `ledger focus` — current-focus primitive (implementation plan)

*Grounded plan produced by a code-exploring agent against the real Go CLI (2026-06-26). No code
written. Adds an explicit current-focus concept so bulwork reads a deliberate focus instead of the
`--last` heuristic. Referenced in `../../ledger/docs/BULWORK_MODE_PLAN.md` decision #11.*

## 1. CLI structure (verified)

- **Cobra**, registered in `cmd/ledger/main.go` (`rootCmd` + `init()` `AddCommand(...)`). Add the new
  command there.
- **Preferred style:** constructor `func NewFocusCmd() *cobra.Command` with closure-captured flag
  vars (like `internal/commands/{note,archive,export}.go`), not the package-level `var XxxCmd` style.
- **`--json`:** reuse `emitResult(jsonOutput bool, payload map[string]any, human string) error` in
  `internal/commands/archive.go` (indented `json.NewEncoder(os.Stdout)`). JSON → stdout; **error path
  prints human text to stderr, never JSON** — preserve this contract.
- **Exit codes:** `internal/exit/exit.go` — `Success=0, AuthError=10, NotFound=20, Network=30,
  Validation=40, Other=50`. Reuse; don't invent codes (ledger-mcp's `runner.ts` mirrors them).
- **Arg convention:** `<id>` is the Firestore **document id** (the `id` from `status --json`), not the
  human name (per `docs/CLI_COMMANDS.md`).

## 2. Firestore (verified)

- **Auth:** `internal/auth/auth.go` `GetFirestoreClient` — project `kaiser-ledger` (override
  `LEDGER_PROJECT_ID`), service account via `GOOGLE_APPLICATION_CREDENTIALS` or gcloud ADC. Wrapped by
  `internal/client/firestore.go` `New(ctx)`.
- **Schema reality:** flat top-level `projects` collection. The **CLI does not filter by `userId`** —
  it reads every project doc → effectively single-user today. **No `/users/<uid>/...`, no `state`
  collection, no settings doc.** No focus/active field exists (only `archived`).

## 3. Design recommendation — single pointer doc

**Recommended: Option A — a single pointer doc at `state/focus` → `{projectId, setAt}`**, over
Option B (an `isFocused` field on projects).

Rationale: the "at most one focus" invariant is structurally free with one doc; Option B needs a
read-all-then-clear transaction on every set and leaks focus state into every project doc + the app's
`Project` type / `normalizeProject`. Focus is a property of the user's session, not a project; `setAt`
gives bulwork a staleness signal; clearing is one `Delete`. Fits the current single-user model.

**Forward-compat:** keep the path in ONE helper `focusDocRef()` so a future `users/<uid>/state/focus`
move (when multi-user scoping lands — decision #11 / Phase-4 consolidation) is a one-line edit.
**Edge case:** focus pointing at an archived/deleted project — `focus` (show) should look the project
up and report `exists`/`archived` so bulwork can fall back.

## 4. Command surface + JSON shapes

- `ledger focus <id>` (set) — validate the project exists first (new `GetProject`); not-found →
  `exit.NotFound`. Returns `{projectId, name, nextAction, setAt, focused:true}`.
- `ledger focus` (show) — no focus → exit **0**, `{focused:false}` (so bulwork distinguishes "no focus"
  from an error); focus set → `{focused:true, projectId, name, nextAction, setAt, exists, archived}`.
- `ledger focus --clear` — delete the pointer (idempotent), `{focused:false, cleared:true}`.
- Cobra: `Args: cobra.MaximumNArgs(1)`; reject `<id>` + `--clear` with `exit.Validation`.

## 5. Smallest implementation

- **New:** `internal/commands/focus.go` — `NewFocusCmd()` (constructor style), captured `jsonOutput`,
  `clear`; `RunE` branches clear / set / show; build `map[string]any` + reuse `emitResult`; map errors
  via `exit.New`.
- **Edit `internal/client/firestore.go`** — add `*Client` methods: `GetProject(ctx,id)` (genuinely
  missing today; needed for validation + show enrichment), `SetFocus(ctx,projectID)`, `GetFocus(ctx)
  → (projectID, setAt, ok, err)`, `ClearFocus(ctx)`, and a private `focusDocRef()` returning
  `c.Collection("state").Doc("focus")` (the single source of the path; comment the single-user
  assumption).
- **Edit `cmd/ledger/main.go`** — `rootCmd.AddCommand(commands.NewFocusCmd())`.
- **Edit `docs/CLI_COMMANDS.md`** — document the command + the three JSON shapes (exit 0 when no
  focus). No changes to `internal/{exit,auth,config}`; no new deps.

## How bulwork consumes it (replaces `--last`)

At session start: run `ledger focus --json`. `focused:false` → prompt the user to pick / `--task`
(no silent default — preserves decision #11's "selecting focus is deliberate"). `focused:true` → use
`nextAction` as the adjudicator's focus task (`name` for display); if `exists:false`/`archived:true`,
treat as stale and re-prompt. Collapses session-start to one command, and gives the Phase-5 exit a
clean `ledger touch` + `ledger focus --clear` pairing.

> Note: bulwork's `--last` lives in `../../bulwork/src/ledger.ts`; this command becomes the preferred
> focus source there once it ships.

## Critical files
`internal/commands/focus.go` (new), `internal/client/firestore.go` (add methods),
`cmd/ledger/main.go` (register), `internal/commands/archive.go` (style/`emitResult` reference),
`docs/CLI_COMMANDS.md` (document).
